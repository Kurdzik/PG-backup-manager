"use client"

import { useState, useEffect } from "react";
import {
  Container,
  Title,
  Text,
  Button,
  Table,
  Modal,
  TextInput,
  NumberInput,
  PasswordInput,
  Group,
  Stack,
  Paper,
  Badge,
  ActionIcon,
  LoadingOverlay,
  Flex,
  Alert,
  Box,
  rem,
  Drawer,
} from '@mantine/core';
import {
  IconDatabase,
  IconPlus,
  IconEdit,
  IconTrash,
  IconRefresh,
  IconCheck,
  IconX,
  IconInfoCircle,
  IconServer,
  IconPlugConnected,
} from '@tabler/icons-react';

import { get, post, put, del } from "../../../lib/backendRequests";

interface DatabaseConnection {
  id: number;
  postgres_host: string;
  postgres_port: string;
  postgres_db_name: string;
  postgres_user: string;
  postgres_password: string;
  created_at: string;
  updated_at: string;
}

interface ConnectionFormData {
  postgres_host: string;
  postgres_port: string;
  postgres_db_name: string;
  postgres_user: string;
  postgres_password: string;
}

interface NotificationData {
  type: 'success' | 'error';
  title: string;
  message: string;
}

interface ApiResponse<T = any> {
  data?: T;
  status?: string;
  count?: number;
  message?: string;
}

// Custom notification component for bottom-right positioning
const BottomRightNotification = ({ 
  notification, 
  onClose 
}: { 
  notification: NotificationData | null; 
  onClose: () => void; 
}) => {
  if (!notification) return null;

  return (
    <Box
      style={{
        position: 'fixed',
        bottom: rem(24),
        right: rem(24),
        zIndex: 1000,
        minWidth: rem(320),
        maxWidth: rem(400),
      }}
    >
      <Paper
        shadow="lg"
        p="md"
        style={{
          backgroundColor: notification.type === 'success' ? '#f8f9fa' : '#fff5f5',
          borderLeft: `4px solid ${notification.type === 'success' ? '#51cf66' : '#ff6b6b'}`,
        }}
      >
        <Flex align="flex-start" gap="sm">
          <Box
            style={{
              color: notification.type === 'success' ? '#51cf66' : '#ff6b6b',
              marginTop: rem(2),
            }}
          >
            {notification.type === 'success' ? 
              <IconCheck size={20} /> : 
              <IconX size={20} />
            }
          </Box>
          <Box style={{ flex: 1 }}>
            <Text fw={600} size="sm" mb={4}>
              {notification.title}
            </Text>
            <Text size="sm" c="dimmed">
              {notification.message}
            </Text>
          </Box>
          <ActionIcon
            variant="subtle"
            color="gray"
            size="sm"
            onClick={onClose}
          >
            <IconX size={16} />
          </ActionIcon>
        </Flex>
      </Paper>
    </Box>
  );
};

export default function DatabaseConnectionsDashboard() {
  const [connections, setConnections] = useState<DatabaseConnection[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [drawerOpened, setDrawerOpened] = useState<boolean>(false);
  const [confirmModalOpened, setConfirmModalOpened] = useState<boolean>(false);
  const [connectionToDelete, setConnectionToDelete] = useState<{id: number, name: string} | null>(null);
  const [editingConnection, setEditingConnection] = useState<DatabaseConnection | null>(null);
  const [formData, setFormData] = useState<ConnectionFormData>({
    postgres_host: "",
    postgres_port: "5432",
    postgres_db_name: "",
    postgres_user: "",
    postgres_password: "",
  });
  const [notification, setNotification] = useState<NotificationData | null>(null);
  const [formLoading, setFormLoading] = useState<boolean>(false);
  const [testLoading, setTestLoading] = useState<boolean>(false);

  useEffect(() => {
    loadConnections();
  }, []);

  const showNotification = (type: 'success' | 'error', title: string, message: string): void => {
    setNotification({ type, title, message });
    setTimeout(() => setNotification(null), 5000);
  };

  const loadConnections = async (): Promise<void> => {
    try {
      setLoading(true);
      const response: ApiResponse<DatabaseConnection[]> = await get("connections/list");
      setConnections(response.data || []);
    } catch (err) {
      showNotification('error', 'Error', 'Failed to load connections');
      console.error("Error loading connections:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleTestConnection = async (): Promise<void> => {
    try {
      setTestLoading(true);
      
      const response: ApiResponse = await post("connections/create?test_connection=true", formData);
      
      // Check for successful response
      if (response.message || response.status?.includes("successfully") || response.status?.includes("success")) {
        showNotification('success', 'Connection Test Successful', 
          response.message || 'Database connection is valid and working correctly');
      } else {
        showNotification('error', 'Connection Test Failed', 
          response.message || 'Unable to connect to database with provided credentials');
      }
    } catch (err: any) {
      // Handle different types of errors
      let errorMessage = 'Connection test failed. Please check your database credentials and network connectivity.';
      
      if (err.response?.data?.message) {
        errorMessage = err.response.data.message;
      } else if (err.message) {
        errorMessage = err.message;
      }
      
      showNotification('error', 'Connection Test Failed', errorMessage);
      console.error("Error testing connection:", err);
    } finally {
      setTestLoading(false);
    }
  };

  const handleSubmit = async (): Promise<void> => {
    try {
      setFormLoading(true);
      let response: ApiResponse;
      
      if (editingConnection) {
        response = await put(`connections/update?connection_id=${editingConnection.id}`, formData);
      } else {
        response = await post("connections/create", formData);
      }
      
      if (response.status?.includes("successfully") || response.message) {
        showNotification('success', 'Success', response.message || response.status || 'Operation completed successfully');
        setDrawerOpened(false);
        resetForm();
        loadConnections();
      }
    } catch (err) {
      showNotification('error', 'Error', `Failed to ${editingConnection ? 'update' : 'create'} connection`);
      console.error("Error submitting form:", err);
    } finally {
      setFormLoading(false);
    }
  };

  const handleDelete = async (connectionId: number, connectionName: string): Promise<void> => {
    setConnectionToDelete({ id: connectionId, name: connectionName });
    setConfirmModalOpened(true);
  };

  const confirmDelete = async (): Promise<void> => {
    if (!connectionToDelete) return;
    
    try {
      const response: ApiResponse = await del(`connections/delete?connection_id=${connectionToDelete.id}`);
      if (response.status?.includes("successfully")) {
        showNotification('success', 'Success', 'Connection deleted successfully');
        loadConnections();
      }
    } catch (err) {
      showNotification('error', 'Error', 'Failed to delete connection');
      console.error("Error deleting connection:", err);
    } finally {
      setConfirmModalOpened(false);
      setConnectionToDelete(null);
    }
  };

  const openCreateDrawer = (): void => {
    setEditingConnection(null);
    resetForm();
    setDrawerOpened(true);
  };

  const openEditDrawer = (connection: DatabaseConnection): void => {
    setEditingConnection(connection);
    setFormData({
      postgres_host: connection.postgres_host,
      postgres_port: connection.postgres_port,
      postgres_db_name: connection.postgres_db_name,
      postgres_user: connection.postgres_user,
      postgres_password: connection.postgres_password,
    });
    setDrawerOpened(true);
  };

  const resetForm = (): void => {
    setFormData({
      postgres_host: "",
      postgres_port: "5432",
      postgres_db_name: "",
      postgres_user: "",
      postgres_password: "",
    });
  };

  const closeDrawer = (): void => {
    setDrawerOpened(false);
    setEditingConnection(null);
    resetForm();
  };

  const formatDate = (dateString: string): string => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "N/A";
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const handleFormDataChange = <K extends keyof ConnectionFormData>(
    field: K,
    value: ConnectionFormData[K]
  ): void => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const isFormValid = (): boolean => {
    return !!(
      formData.postgres_host &&
      formData.postgres_port &&
      formData.postgres_db_name &&
      formData.postgres_user &&
      formData.postgres_password
    );
  };

  const rows = connections.map((connection: DatabaseConnection) => (
    <Table.Tr key={connection.id}>
      <Table.Td>
        <Flex align="center" gap="sm">
          <IconServer size={18} color="#495057" />
          <Box>
            <Text fw={500} size="sm">{connection.postgres_db_name}</Text>
            <Text size="xs" c="dimmed">{connection.postgres_host}:{connection.postgres_port}</Text>
          </Box>
        </Flex>
      </Table.Td>
      <Table.Td>
        <Badge variant="light" color="blue" size="sm">
          {connection.postgres_user}
        </Badge>
      </Table.Td>
      <Table.Td>
        <Text size="sm" c="dimmed">{formatDate(connection.created_at)}</Text>
      </Table.Td>
      <Table.Td>
        <Text size="sm" c="dimmed">{formatDate(connection.updated_at)}</Text>
      </Table.Td>
      <Table.Td>
        <Group gap="xs">
          <ActionIcon
            variant="subtle"
            color="blue"
            onClick={() => openEditDrawer(connection)}
            aria-label={`Edit connection ${connection.postgres_db_name}`}
          >
            <IconEdit size={16} />
          </ActionIcon>
          <ActionIcon
            variant="subtle"
            color="red"
            onClick={() => handleDelete(connection.id, connection.postgres_db_name)}
            aria-label={`Delete connection ${connection.postgres_db_name}`}
          >
            <IconTrash size={16} />
          </ActionIcon>
        </Group>
      </Table.Td>
    </Table.Tr>
  ));

  return (
    <Box style={{ width: '90%', margin: '0 auto', padding: '2rem 0' }}>
      {/* Bottom-right notification */}
      <BottomRightNotification 
        notification={notification} 
        onClose={() => setNotification(null)} 
      />

      {/* Header */}
      <Box mb="xl">
        <Flex align="center" gap="md" mb="lg">
          <IconDatabase size={32} />
          <Box>
            <Title order={1} size="2rem" fw={600} mb="xs">
              Database Connections
            </Title>
            <Text size="lg" c="dimmed">
              Manage PostgreSQL connections
            </Text>
          </Box>
        </Flex>
      </Box>

      {/* Action Bar */}
      <Box mb="xl">
        <Flex justify="space-between" align="center" style={{ width: '100%' }}>
          <Group>
            <Button
              leftSection={<IconRefresh size={16} />}
              variant="light"
              onClick={loadConnections}
              loading={loading}
            >
              Refresh
            </Button>
            <Badge variant="light" size="lg">
              {connections.length} connection{connections.length !== 1 ? 's' : ''}
            </Badge>
          </Group>
          <Button
            leftSection={<IconPlus size={16} />}
            onClick={openCreateDrawer}
          >
            Add Connection
          </Button>
        </Flex>
      </Box>

      {/* Connections Table */}
      <Paper shadow="sm" p="lg" pos="relative" style={{ width: '100%' }}>
        <LoadingOverlay visible={loading} />
        
        {connections.length === 0 && !loading ? (
          <Alert icon={<IconInfoCircle size={20} />} title="No connections found" color="blue">
            <Text>Get started by creating your first database connection.</Text>
          </Alert>
        ) : (
          <Table striped highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Database</Table.Th>
                <Table.Th>User</Table.Th>
                <Table.Th>Created</Table.Th>
                <Table.Th>Updated</Table.Th>
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
            {editingConnection ? "Edit Connection" : "Create New Connection"}
          </Text>
        }
        position="right"
        size="md"
        padding="xl"
      >
        <Stack gap="lg">
          <TextInput
            label="Host"
            placeholder="localhost or IP address"
            required
            value={formData.postgres_host}
            onChange={(event) =>
              handleFormDataChange('postgres_host', event.currentTarget.value)
            }
          />

          <NumberInput
            label="Port"
            placeholder="5432"
            required
            min={1}
            max={65535}
            value={formData.postgres_port}
            onChange={(value) =>
              handleFormDataChange('postgres_port', String(value) || "5432")
            }
          />

          <TextInput
            label="Database Name"
            placeholder="mydb"
            required
            value={formData.postgres_db_name}
            onChange={(event) =>
              handleFormDataChange('postgres_db_name', event.currentTarget.value)
            }
          />

          <TextInput
            label="Username"
            placeholder="postgres"
            required
            value={formData.postgres_user}
            onChange={(event) =>
              handleFormDataChange('postgres_user', event.currentTarget.value)
            }
          />

          <PasswordInput
            label="Password"
            placeholder="Enter password"
            required
            value={formData.postgres_password}
            onChange={(event) =>
              handleFormDataChange('postgres_password', event.currentTarget.value)
            }
          />

          <Group justify="flex-end" mt="lg">
            {!editingConnection && (
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
              {editingConnection ? "Update" : "Create"} Connection
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
        size="md"
        centered
      >
        <Stack gap="lg">
          <Text size="md">
            Are you sure you want to delete the connection{' '}
            <Text span fw={600}>
              "{connectionToDelete?.name}"
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
            <Button 
              color="red" 
              onClick={confirmDelete}
            >
              Delete Connection
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}