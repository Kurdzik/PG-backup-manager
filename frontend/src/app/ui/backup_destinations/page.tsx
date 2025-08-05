"use client";

import { useState, useEffect } from "react";
import {
  Title,
  Text,
  Button,
  Table,
  Modal,
  TextInput,
  Select,
  Switch,
  Group,
  Stack,
  Paper,
  Badge,
  ActionIcon,
  LoadingOverlay,
  Flex,
  Alert,
  Box,
  Drawer,
} from "@mantine/core";
import {
  IconCloud,
  IconPlus,
  IconEdit,
  IconTrash,
  IconRefresh,
  IconInfoCircle,
  IconDatabase,
  IconPlugConnected,
} from "@tabler/icons-react";

import { get, post, put, del } from "@/lib/backendRequests";
import BottomRightNotification from "@/components/Notifications";
import { NotificationData } from "@/components/Notifications";
import {
  BackupDestination,
  DatabaseConnection,
  ApiResponse,
} from "@/lib/types";

interface DestinationFormData {
  connection_id: string;
  name: string;
  endpoint_url: string;
  region: string;
  bucket_name: string;
  access_key_id: string;
  secret_access_key: string;
  path_prefix: string;
  use_ssl: boolean;
  verify_ssl: boolean;
}

export default function S3BackupDestinations() {
  const [destinations, setDestinations] = useState<BackupDestination[]>([]);
  const [connections, setConnections] = useState<DatabaseConnection[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [drawerOpened, setDrawerOpened] = useState<boolean>(false);
  const [confirmModalOpened, setConfirmModalOpened] = useState<boolean>(false);
  const [destinationToDelete, setDestinationToDelete] = useState<{
    id: number;
    name: string;
  } | null>(null);
  const [editingDestination, setEditingDestination] =
    useState<BackupDestination | null>(null);
  const [formData, setFormData] = useState<DestinationFormData>({
    connection_id: "",
    name: "",
    endpoint_url: "",
    region: "",
    bucket_name: "",
    access_key_id: "",
    secret_access_key: "",
    path_prefix: "",
    use_ssl: true,
    verify_ssl: true,
  });
  const [notification, setNotification] = useState<NotificationData | null>(
    null,
  );
  const [formLoading, setFormLoading] = useState<boolean>(false);
  const [testLoading, setTestLoading] = useState<boolean>(false);

  useEffect(() => {
    loadData();
  }, []);

  const showNotification = (
    type: "success" | "error",
    title: string,
    message: string,
  ): void => {
    setNotification({ type, title, message });
    setTimeout(() => setNotification(null), 5000);
  };

  const loadData = async (): Promise<void> => {
    try {
      setLoading(true);
      const [connectionsRes, destinationsRes] = await Promise.all([
        get("connections/list") as Promise<ApiResponse<DatabaseConnection[]>>,
        get("backup-destinations/s3/list") as Promise<
          ApiResponse<BackupDestination[]>
        >,
      ]);

      setConnections(connectionsRes.data || []);
      setDestinations(destinationsRes.data || []);
    } catch (err) {
      showNotification("error", "Error", "Failed to load data");
      console.error("Error loading data:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async (): Promise<void> => {
    try {
      setTestLoading(true);

      // Convert connection_id to number for API and ensure path_prefix is properly handled
      const payload = {
        ...formData,
        connection_id: parseInt(formData.connection_id, 10),
        path_prefix: formData.path_prefix.trim(), // Trim whitespace but keep empty string
      };

      const response: ApiResponse = await post(
        "backup-destinations/s3/create?test_connection=true",
        payload,
      );

      // Check for successful response
      if (
        response.status==200
      ) {
        showNotification(
          "success",
          "Connection Test Successful",
          "S3 destination configuration is valid and connection successful",
        );
      } else {
        showNotification(
          "error",
          "Connection Test Failed",
          "Unable to connect to S3 destination with provided configuration",
        );
      }
    } catch (err: any) {
      // Handle different types of errors
      let errorMessage =
        "Connection test failed. Please check your configuration.";

      if (err.response?.data?.message) {
        errorMessage = err.response.data.message;
      } else if (err.message) {
        errorMessage = err.message;
      }

      showNotification("error", "Connection Test Failed", errorMessage);
      console.error("Error testing connection:", err);
    } finally {
      setTestLoading(false);
    }
  };

  const handleSubmit = async (): Promise<void> => {
    try {
      setFormLoading(true);
      let response: ApiResponse;

      // Convert connection_id to number for API and ensure path_prefix is properly handled
      const payload = {
        ...formData,
        connection_id: parseInt(formData.connection_id, 10),
        path_prefix: formData.path_prefix.trim(), // Trim whitespace but keep empty string
      };

      if (editingDestination) {
        response = await put(
          `backup-destinations/s3/update?destination_id=${editingDestination.id}`,
          payload,
        );
      } else {
        response = await post("backup-destinations/s3/create", payload);
      }

      if (response.status==200) {
        showNotification(
          "success",
          "Success",
          response?.message || "Operation completed successfully",
        );
        setDrawerOpened(false);
        resetForm();
        loadData();
      }
    } catch (err) {
      showNotification(
        "error",
        "Error",
        `Failed to ${editingDestination ? "update" : "create"} backup destination`,
      );
      console.error("Error submitting form:", err);
    } finally {
      setFormLoading(false);
    }
  };

  const handleDelete = async (
    destinationId: number,
    destinationName: string,
  ): Promise<void> => {
    setDestinationToDelete({ id: destinationId, name: destinationName });
    setConfirmModalOpened(true);
  };

  const confirmDelete = async (): Promise<void> => {
    if (!destinationToDelete) return;

    try {
      const response: ApiResponse = await del(
        `backup-destinations/s3/delete?destination_id=${destinationToDelete.id}`,
      );
      showNotification(
        "success",
        "Success",
        "Backup destination deleted successfully",
      );
      loadData();
    } catch (err) {
      showNotification("error", "Error", "Failed to delete backup destination");
      console.error("Error deleting destination:", err);
    } finally {
      setConfirmModalOpened(false);
      setDestinationToDelete(null);
    }
  };

  const openCreateDrawer = (): void => {
    setEditingDestination(null);
    resetForm();
    setDrawerOpened(true);
  };

  const openEditDrawer = (destination: BackupDestination): void => {
    setEditingDestination(destination);
    setFormData({
      connection_id: destination.connection_id.toString(),
      name: destination.name,
      endpoint_url: destination.endpoint_url,
      region: destination.region,
      bucket_name: destination.bucket_name,
      access_key_id: destination.access_key_id,
      secret_access_key: destination.secret_access_key, // Show current value for editing
      path_prefix: destination.path_prefix || "", // Ensure empty string if null/undefined
      use_ssl: destination.use_ssl,
      verify_ssl: destination.verify_ssl,
    });
    setDrawerOpened(true);
  };

  const resetForm = (): void => {
    setFormData({
      connection_id: "",
      name: "",
      endpoint_url: "",
      region: "",
      bucket_name: "",
      access_key_id: "",
      secret_access_key: "",
      path_prefix: "", // Explicitly set to empty string
      use_ssl: true,
      verify_ssl: true,
    });
  };

  const closeDrawer = (): void => {
    setDrawerOpened(false);
    setEditingDestination(null);
    resetForm();
  };

  const formatDate = (dateString: string): string => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "N/A";
    return new Date(dateString).toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const handleFormDataChange = <K extends keyof DestinationFormData>(
    field: K,
    value: DestinationFormData[K],
  ): void => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const isFormValid = (): boolean => {
    return !!(
      formData.connection_id &&
      formData.name &&
      formData.endpoint_url &&
      formData.region &&
      formData.bucket_name &&
      formData.access_key_id &&
      formData.secret_access_key // Always require secret key
    );
  };

  const getConnectionName = (connectionId: number): string => {
    const connection = connections.find((c) => c.id === connectionId);
    return connection
      ? `${connection.postgres_db_name} (${connection.postgres_host})`
      : "Unknown";
  };

  const connectionOptions = connections.map((conn) => ({
    value: conn.id.toString(),
    label: `${conn.postgres_db_name} (${conn.postgres_host})`,
  }));

  const rows = destinations.map((destination: BackupDestination) => (
    <Table.Tr key={destination.id}>
      <Table.Td>
        <Flex align="center" gap="sm">
          <IconCloud size={18} color="#495057" />
          <Box>
            <Text fw={500} size="md">
              {destination.name}
            </Text>
            <Text size="xs" c="dimmed">
              {destination.bucket_name}
            </Text>
          </Box>
        </Flex>
      </Table.Td>
      <Table.Td>
        <Flex align="center" gap="xs">
          <IconDatabase size={14} />
          <Text size="sm">{getConnectionName(destination.connection_id)}</Text>
        </Flex>
      </Table.Td>
      <Table.Td>
        <Text size="sm">{destination.region}</Text>
      </Table.Td>
      <Table.Td>
        <Text size="sm" style={{ maxWidth: 200 }} truncate>
          {destination.endpoint_url}
        </Text>
      </Table.Td>
      <Table.Td>
        <Group gap="xs">
          <Badge variant="outline" size="sm" radius={"sm"}>
            SSL: {destination.use_ssl ? "On" : "Off"}
          </Badge>
          <Badge variant="outline" size="sm" radius={"sm"}>
            Verify: {destination.verify_ssl ? "On" : "Off"}
          </Badge>
        </Group>
      </Table.Td>
      <Table.Td>
        <Text size="sm" c="dimmed">
          {formatDate(destination.created_at)}
        </Text>
      </Table.Td>
      <Table.Td>
        <Group gap="xs">
          <ActionIcon
            variant="outline"
            onClick={() => openEditDrawer(destination)}
            aria-label={`Edit destination ${destination.name}`}
          >
            <IconEdit size={16} />
          </ActionIcon>
          <ActionIcon
            variant="outline"
            color="error"
            onClick={() => handleDelete(destination.id, destination.name)}
            aria-label={`Delete destination ${destination.name}`}
          >
            <IconTrash size={16} />
          </ActionIcon>
        </Group>
      </Table.Td>
    </Table.Tr>
  ));

  return (
    <Box style={{ width: "90%", margin: "0 auto", padding: "2rem 0" }}>
      {/* Bottom-right notification */}
      <BottomRightNotification
        notification={notification}
        onClose={() => setNotification(null)}
      />

      {/* Header */}
      <Box mb="xl">
        <Flex align="center" gap="md" mb="lg">
          <IconCloud size={32} />
          <Box>
            <Title order={1} size="2rem" fw={600} mb="xs">
              S3 Backup Destinations
            </Title>
            <Text size="lg" c="dimmed">
              Manage S3-compatible storage destinations for database backups
            </Text>
          </Box>
        </Flex>
      </Box>

      {/* Action Bar */}
      <Box mb="xl">
        <Flex justify="space-between" align="center" style={{ width: "100%" }}>
          <Group>
            <Button
              leftSection={<IconRefresh size={16} />}
              variant="light"
              onClick={loadData}
              loading={loading}
            >
              Refresh
            </Button>
            <Badge variant="outline" radius={"sm"} size="md">
              {destinations.length} destination
              {destinations.length !== 1 ? "s" : ""}
            </Badge>
          </Group>
          <Button
            leftSection={<IconPlus size={16} />}
            onClick={openCreateDrawer}
          >
            Add Destination
          </Button>
        </Flex>
      </Box>

      {/* Destinations Table */}
      <Paper shadow="sm" p="lg" pos="relative" style={{ width: "100%" }}>
        <LoadingOverlay visible={loading} />

        {destinations.length === 0 && !loading ? (
          <Alert
            icon={<IconInfoCircle size={20} />}
            title="No backup destinations found"
            color="blue"
          >
            <Text>
              Get started by creating your first S3 backup destination.
            </Text>
          </Alert>
        ) : (
          <Table striped highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Destination</Table.Th>
                <Table.Th>Database Connection</Table.Th>
                <Table.Th>Region</Table.Th>
                <Table.Th>Endpoint</Table.Th>
                <Table.Th>Security</Table.Th>
                <Table.Th>Created</Table.Th>
                <Table.Th>Actions</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>{rows}</Table.Tbody>
          </Table>
        )}
      </Paper>

      {/* Create/Edit Drawer */}
      <Drawer
        opened={drawerOpened}
        onClose={closeDrawer}
        title={
          <Text fw={600} size="lg">
            {editingDestination
              ? "Edit Backup Destination"
              : "Create New Backup Destination"}
          </Text>
        }
        position="left"
        size="md"
        padding="xl"
      >
        <Stack gap="lg">
          <Select
            label="Database Connection"
            placeholder="Select a database connection"
            required
            data={connectionOptions}
            value={formData.connection_id}
            onChange={(value) =>
              handleFormDataChange("connection_id", value || "")
            }
            disabled={!!editingDestination}
          />

          <TextInput
            label="Destination Name"
            placeholder="e.g., AWS Production, Synology C2"
            required
            value={formData.name}
            onChange={(event) =>
              handleFormDataChange("name", event.currentTarget.value)
            }
          />

          <TextInput
            label="Endpoint URL"
            placeholder="https://s3.amazonaws.com"
            required
            value={formData.endpoint_url}
            onChange={(event) =>
              handleFormDataChange("endpoint_url", event.currentTarget.value)
            }
          />

          <TextInput
            label="Region"
            placeholder="us-east-1"
            required
            value={formData.region}
            onChange={(event) =>
              handleFormDataChange("region", event.currentTarget.value)
            }
          />

          <TextInput
            label="Bucket Name"
            placeholder="my-backup-bucket"
            required
            value={formData.bucket_name}
            onChange={(event) =>
              handleFormDataChange("bucket_name", event.currentTarget.value)
            }
          />

          <TextInput
            label="Access Key ID"
            placeholder="AKIA..."
            required
            value={formData.access_key_id}
            onChange={(event) =>
              handleFormDataChange("access_key_id", event.currentTarget.value)
            }
          />

          <TextInput
            label="Secret Access Key"
            placeholder="Enter secret key"
            type="password"
            required
            value={formData.secret_access_key}
            onChange={(event) =>
              handleFormDataChange(
                "secret_access_key",
                event.currentTarget.value,
              )
            }
          />

          <TextInput
            label="Path Prefix"
            placeholder="postgres/ (optional)"
            description="Leave blank for root path"
            value={formData.path_prefix}
            onChange={(event) =>
              handleFormDataChange("path_prefix", event.currentTarget.value)
            }
          />

          <Switch
            label="Use SSL"
            checked={formData.use_ssl}
            onChange={(event) =>
              handleFormDataChange("use_ssl", event.currentTarget.checked)
            }
          />

          <Switch
            label="Verify SSL Certificate"
            checked={formData.verify_ssl}
            onChange={(event) =>
              handleFormDataChange("verify_ssl", event.currentTarget.checked)
            }
          />

          <Group justify="flex-end" mt="lg">
            {!editingDestination && (
              <Button
                leftSection={<IconPlugConnected size={16} />}
                variant="light"
                onClick={handleTestConnection}
                loading={testLoading}
                disabled={!isFormValid()}
              >
                Test Connection
              </Button>
            )}
            <Button
              onClick={handleSubmit}
              loading={formLoading}
              disabled={!isFormValid()}
            >
              {editingDestination ? "Update" : "Create"} Destination
            </Button>
          </Group>
        </Stack>
      </Drawer>

      {/* Delete Confirmation Modal */}
      <Modal
        opened={confirmModalOpened}
        onClose={() => setConfirmModalOpened(false)}
        title={
          <Text fw={600} size="lg">
            Confirm Deletion
          </Text>
        }
        size="lg"
        centered
      >
        <Stack gap="lg">
          <Text size="md">
            Are you sure you want to delete the backup destination{" "}
            <Text span fw={600}>
              "{destinationToDelete?.name}"
            </Text>
            ?
          </Text>
          <Text size="sm" c="dimmed">
            This action cannot be undone.
          </Text>

          <Group justify="flex-end" mt="lg">
            <Button
              variant="subtle"
              onClick={() => setConfirmModalOpened(false)}
            >
              Cancel
            </Button>
            <Button color="error" onClick={confirmDelete}>
              Delete Destination
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}