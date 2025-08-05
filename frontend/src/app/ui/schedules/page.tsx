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

  Card,

  SimpleGrid,
  Tooltip,
  Code,
} from "@mantine/core";
import {
  IconClock,
  IconPlus,
  IconEdit,
  IconTrash,
  IconRefresh,

  IconX,
  IconInfoCircle,
  IconDatabase,
  IconCloud,
  IconCalendarStats,
  IconPlayerPlay,
  IconPlayerPause,
  IconActivity,

  IconCheckbox,
  IconClockHour3,
  IconHistory,
} from "@tabler/icons-react";

import { get, post, put, del } from "@/lib/backendRequests";
import BottomRightNotification from "@/components/Notifications";
import { NotificationData } from "@/components/Notifications";
import {
  BackupDestination,
  DatabaseConnection,
  ApiResponse,
  BackupSchedule,
} from "@/lib/types";

interface ScheduleFormData {
  connection_id: string;
  destination_id: string;
  schedule: string;
  enabled: boolean;
}

interface ScheduleStats {
  total: number;
  enabled: number;
  disabled: number;
  recentRuns: number;
  upcomingRuns: number;
}

const COMMON_SCHEDULES = [
  { value: "0 2 * * *", label: "Daily at 2:00 AM" },
  { value: "0 2 * * 0", label: "Weekly on Sunday at 2:00 AM" },
  { value: "0 2 1 * *", label: "Monthly on 1st at 2:00 AM" },
  { value: "0 */6 * * *", label: "Every 6 hours" },
  { value: "0 */12 * * *", label: "Every 12 hours" },
  { value: "0 3 * * 1-5", label: "Weekdays at 3:00 AM" },
  { value: "custom", label: "Custom Cron Expression" },
];

export default function BackupScheduleDashboard() {
  const [schedules, setSchedules] = useState<BackupSchedule[]>([]);
  const [connections, setConnections] = useState<DatabaseConnection[]>([]);
  const [destinations, setDestinations] = useState<BackupDestination[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [drawerOpened, setDrawerOpened] = useState<boolean>(false);
  const [confirmModalOpened, setConfirmModalOpened] = useState<boolean>(false);
  const [scheduleToDelete, setScheduleToDelete] = useState<{
    id: number;
    name: string;
  } | null>(null);
  const [editingSchedule, setEditingSchedule] = useState<BackupSchedule | null>(
    null,
  );
  const [formData, setFormData] = useState<ScheduleFormData>({
    connection_id: "",
    destination_id: "",
    schedule: "",
    enabled: true,
  });
  const [customSchedule, setCustomSchedule] = useState<string>("");
  const [notification, setNotification] = useState<NotificationData | null>(
    null,
  );
  const [formLoading, setFormLoading] = useState<boolean>(false);
  const [stats, setStats] = useState<ScheduleStats>({
    total: 0,
    enabled: 0,
    disabled: 0,
    recentRuns: 0,
    upcomingRuns: 0,
  });

  useEffect(() => {
    loadData();
  }, []);

  useEffect(() => {
    calculateStats();
  }, [schedules]);

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
      const [connectionsRes, destinationsRes, schedulesRes] = await Promise.all(
        [
          get("connections/list") as Promise<ApiResponse<DatabaseConnection[]>>,
          get("backup-destinations/s3/list") as Promise<
            ApiResponse<BackupDestination[]>
          >,
          get("schedules/list") as Promise<ApiResponse<BackupSchedule[]>>,
        ],
      );

      setConnections(connectionsRes.data || []);
      setDestinations(destinationsRes.data || []);
      setSchedules(schedulesRes.data || []);
    } catch (err) {
      showNotification("error", "Error", "Failed to load data");
      console.error("Error loading data:", err);
    } finally {
      setLoading(false);
    }
  };

  const calculateStats = (): void => {
    const now = new Date();
    const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);
    const oneDayFromNow = new Date(now.getTime() + 24 * 60 * 60 * 1000);

    const newStats: ScheduleStats = {
      total: schedules.length,
      enabled: schedules.filter((s) => s.enabled).length,
      disabled: schedules.filter((s) => !s.enabled).length,
      recentRuns: schedules.filter(
        (s) => s.last_run && new Date(s.last_run) >= oneDayAgo,
      ).length,
      upcomingRuns: schedules.filter(
        (s) => s.next_run && new Date(s.next_run) <= oneDayFromNow && s.enabled,
      ).length,
    };

    setStats(newStats);
  };

  const handleSubmit = async (): Promise<void> => {
    try {
      setFormLoading(true);
      let response: ApiResponse;

      const scheduleValue =
        formData.schedule === "custom" ? customSchedule : formData.schedule;
      const payload = {
        ...formData,
        schedule: scheduleValue,
      };

      if (editingSchedule) {
        response = await put(
          `schedules/update?schedule_id=${editingSchedule.id}`,
          payload,
        );
      } else {
        response = await post("schedules/create", payload);
      }

      if (
        response.status == 200 ) {
        showNotification(
          "success",
          "Success",
          `Schedule ${editingSchedule ? "updated" : "created"} successfully`,
        );
        setDrawerOpened(false);
        resetForm();
        loadData();
      }
    } catch (err) {
      showNotification(
        "error",
        "Error",
        `Failed to ${editingSchedule ? "update" : "create"} schedule`,
      );
      console.error("Error submitting form:", err);
    } finally {
      setFormLoading(false);
    }
  };

  const handleDelete = async (scheduleId: number): Promise<void> => {
    const schedule = schedules.find((s) => s.id === scheduleId);
    const name = schedule
      ? `${getConnectionName(schedule.connection_id)} → ${getDestinationName(schedule.destination_id)}`
      : "Unknown Schedule";
    setScheduleToDelete({ id: scheduleId, name });
    setConfirmModalOpened(true);
  };

  const confirmDelete = async (): Promise<void> => {
    if (!scheduleToDelete) return;

    try {
      const response: ApiResponse = await del(
        `schedules/delete?schedule_id=${scheduleToDelete.id}`,
      );
      showNotification("success", "Success", "Schedule deleted successfully");
      loadData();
    } catch (err) {
      showNotification("error", "Error", "Failed to delete schedule");
      console.error("Error deleting schedule:", err);
    } finally {
      setConfirmModalOpened(false);
      setScheduleToDelete(null);
    }
  };

  const handleToggleEnabled = async (
    scheduleId: number,
    enabled: boolean,
  ): Promise<void> => {
    try {
      const endpoint = enabled
        ? `schedules/enable?schedule_id=${scheduleId}`
        : `schedules/disable?schedule_id=${scheduleId}`;
      await post(endpoint, {});
      showNotification(
        "success",
        "Success",
        `Schedule ${enabled ? "enabled" : "disabled"} successfully`,
      );
      loadData();
    } catch (err) {
      showNotification(
        "error",
        "Error",
        `Failed to ${enabled ? "enable" : "disable"} schedule`,
      );
      console.error("Error toggling schedule:", err);
    }
  };

  const openCreateDrawer = (): void => {
    setEditingSchedule(null);
    resetForm();
    setDrawerOpened(true);
  };

  const openEditDrawer = (schedule: BackupSchedule): void => {
    setEditingSchedule(schedule);

    // Check if schedule matches a common pattern
    const commonSchedule = COMMON_SCHEDULES.find(
      (cs) => cs.value === schedule.schedule,
    );

    setFormData({
      connection_id: schedule.connection_id.toString(),
      destination_id: schedule.destination_id.toString(),
      schedule: commonSchedule ? schedule.schedule : "custom",
      enabled: schedule.enabled,
    });

    if (!commonSchedule) {
      setCustomSchedule(schedule.schedule);
    }

    setDrawerOpened(true);
  };

  const resetForm = (): void => {
    setFormData({
      connection_id: "",
      destination_id: "",
      schedule: "",
      enabled: true,
    });
    setCustomSchedule("");
  };

  const closeDrawer = (): void => {
    setDrawerOpened(false);
    setEditingSchedule(null);
    resetForm();
  };

  const formatDate = (dateString?: string): string => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "Never";
    return new Date(dateString).toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatRelativeTime = (dateString?: string): string => {
    if (!dateString || dateString === "0001-01-01T00:00:00Z") return "Never";

    const date = new Date(dateString);
    const now = new Date();
    const diffMs = date.getTime() - now.getTime();
    const diffMins = Math.round(diffMs / (1000 * 60));
    const diffHours = Math.round(diffMs / (1000 * 60 * 60));
    const diffDays = Math.round(diffMs / (1000 * 60 * 60 * 24));

    if (Math.abs(diffMins) < 60) {
      return diffMins > 0 ? `in ${diffMins}m` : `${Math.abs(diffMins)}m ago`;
    } else if (Math.abs(diffHours) < 24) {
      return diffHours > 0 ? `in ${diffHours}h` : `${Math.abs(diffHours)}h ago`;
    } else {
      return diffDays > 0 ? `in ${diffDays}d` : `${Math.abs(diffDays)}d ago`;
    }
  };

  const handleFormDataChange = <K extends keyof ScheduleFormData>(
    field: K,
    value: ScheduleFormData[K],
  ): void => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const isFormValid = (): boolean => {
    const scheduleValue =
      formData.schedule === "custom" ? customSchedule : formData.schedule;
    return !!(
      formData.connection_id &&
      formData.destination_id &&
      scheduleValue
    );
  };

  const getConnectionName = (connectionId: number): string => {
    const connection = connections.find((c) => c.id === connectionId);
    return connection
      ? `${connection.postgres_db_name} (${connection.postgres_host})`
      : "Unknown";
  };

  const getDestinationName = (destinationId: number): string => {
    const destination = destinations.find((d) => d.id === destinationId);
    return destination ? destination.name : "Unknown";
  };

  const getAvailableDestinations = () => {
    if (!formData.connection_id) return [];
    const connectionId = parseInt(formData.connection_id);
    return destinations.filter((d) => d.connection_id === connectionId);
  };

  const connectionOptions = connections.map((conn) => ({
    value: conn.id.toString(),
    label: `${conn.postgres_db_name} (${conn.postgres_host})`,
  }));

  const destinationOptions = getAvailableDestinations().map((dest) => ({
    value: dest.id.toString(),
    label: dest.name,
  }));

  const rows = schedules.map((schedule: BackupSchedule) => (
    <Table.Tr key={schedule.id}>
      <Table.Td>
        <Flex align="center" gap="sm">
          <IconClock
            size={18}
            color={schedule.enabled ? "#495057" : "#adb5bd"}
          />
          <Box>
            <Text
              fw={500}
              size="md"
              c={schedule.enabled ? undefined : "dimmed"}
            >
              {getConnectionName(schedule.connection_id)}
            </Text>
            <Text size="xs" c="dimmed">
              to {getDestinationName(schedule.destination_id)}
            </Text>
          </Box>
        </Flex>
      </Table.Td>
      <Table.Td>
        <Tooltip label={schedule.schedule}>
          <Code style={{ cursor: "help" }}>{schedule.schedule}</Code>
        </Tooltip>
      </Table.Td>
      <Table.Td>
        <Badge
          variant={schedule.enabled ? "outline" : "outline"}
          color={schedule.enabled ? "success" : "slate"}
          size="sm"
          radius="sm"
        >
          {schedule.enabled ? "Enabled" : "Disabled"}
        </Badge>
      </Table.Td>
      <Table.Td>
        <Flex direction="column" gap={2}>
          <Text size="sm">{formatDate(schedule.last_run)}</Text>
          {schedule.last_run && (
            <Text size="xs" c="dimmed">
              {formatRelativeTime(schedule.last_run)}
            </Text>
          )}
        </Flex>
      </Table.Td>
      <Table.Td>
        <Flex direction="column" gap={2}>
          <Text size="sm">{formatDate(schedule.next_run)}</Text>
          {schedule.next_run && schedule.enabled && (
            <Text size="xs" c="dimmed">
              {formatRelativeTime(schedule.next_run)}
            </Text>
          )}
        </Flex>
      </Table.Td>
      <Table.Td>
        <Group gap="xs">
          <Tooltip label={schedule.enabled ? "Disable" : "Enable"}>
            <ActionIcon
              variant="outline"
              color={schedule.enabled ? "warning" : "success"}
              onClick={() =>
                handleToggleEnabled(schedule.id, !schedule.enabled)
              }
              aria-label={`${schedule.enabled ? "Disable" : "Enable"} schedule`}
            >
              {schedule.enabled ? (
                <IconPlayerPause size={16} />
              ) : (
                <IconPlayerPlay size={16} />
              )}
            </ActionIcon>
          </Tooltip>
          <ActionIcon
            variant="outline"
            onClick={() => openEditDrawer(schedule)}
            aria-label="Edit schedule"
          >
            <IconEdit size={16} />
          </ActionIcon>
          <ActionIcon
            variant="outline"
            color="error"
            onClick={() => handleDelete(schedule.id)}
            aria-label="Delete schedule"
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
          <IconCalendarStats size={32} />
          <Box>
            <Title order={1} size="2rem" fw={600} mb="xs">
              Backup Schedules
            </Title>
            <Text size="lg" c="dimmed">
              Manage automated backup schedules for your databases
            </Text>
          </Box>
        </Flex>
      </Box>

      {/* Statistics Cards */}
      <SimpleGrid cols={{ base: 2, sm: 3, lg: 5 }} spacing="lg" mb="xl">
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Flex align="center" gap="md">
            <IconActivity size={24} color="grey" />
            <Box>
              <Text size="lg" fw={600}>
                {stats.total}
              </Text>
              <Text size="sm" c="dimmed">
                Total Schedules
              </Text>
            </Box>
          </Flex>
        </Card>

        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Flex align="center" gap="md">
            <IconCheckbox size={24} color="green" />
            <Box>
              <Text size="lg" fw={600}>
                {stats.enabled}
              </Text>
              <Text size="sm" c="dimmed">
                Enabled
              </Text>
            </Box>
          </Flex>
        </Card>

        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Flex align="center" gap="md">
            <IconX size={24} color="grey" />
            <Box>
              <Text size="lg" fw={600}>
                {stats.disabled}
              </Text>
              <Text size="sm" c="dimmed">
                Disabled
              </Text>
            </Box>
          </Flex>
        </Card>

        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Flex align="center" gap="md">
            <IconHistory size={24} color="grey" />
            <Box>
              <Text size="lg" fw={600}>
                {stats.recentRuns}
              </Text>
              <Text size="sm" c="dimmed">
                Ran Today
              </Text>
            </Box>
          </Flex>
        </Card>

        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Flex align="center" gap="md">
            <IconClockHour3 size={24} color="grey" />
            <Box>
              <Text size="lg" fw={600}>
                {stats.upcomingRuns}
              </Text>
              <Text size="sm" c="dimmed">
                Due Soon
              </Text>
            </Box>
          </Flex>
        </Card>
      </SimpleGrid>

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
              {schedules.length} schedule{schedules.length !== 1 ? "s" : ""}
            </Badge>
          </Group>
          <Button
            leftSection={<IconPlus size={16} />}
            onClick={openCreateDrawer}
          >
            Add Schedule
          </Button>
        </Flex>
      </Box>

      {/* Schedules Table */}
      <Paper shadow="sm" p="lg" pos="relative" style={{ width: "100%" }}>
        <LoadingOverlay visible={loading} />

        {schedules.length === 0 && !loading ? (
          <Alert
            icon={<IconInfoCircle size={20} />}
            title="No backup schedules found"
            color="blue"
          >
            <Text>
              Get started by creating your first automated backup schedule.
            </Text>
          </Alert>
        ) : (
          <Table striped highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Database → Destination</Table.Th>
                <Table.Th>Schedule</Table.Th>
                <Table.Th>Status</Table.Th>
                <Table.Th>Last Run</Table.Th>
                <Table.Th>Next Run</Table.Th>
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
            {editingSchedule
              ? "Edit Backup Schedule"
              : "Create New Backup Schedule"}
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
            onChange={(value) => {
              handleFormDataChange("connection_id", value || "");
              handleFormDataChange("destination_id", ""); // Reset destination when connection changes
            }}
            leftSection={<IconDatabase size={16} />}
            disabled={!!editingSchedule}
          />

          <Select
            label="Backup Destination"
            placeholder="Select a backup destination"
            required
            data={destinationOptions}
            value={formData.destination_id}
            onChange={(value) =>
              handleFormDataChange("destination_id", value || "")
            }
            leftSection={<IconCloud size={16} />}
            disabled={!formData.connection_id || !!editingSchedule}
          />

          <Select
            label="Schedule Frequency"
            placeholder="Select schedule frequency"
            required
            data={COMMON_SCHEDULES}
            value={formData.schedule}
            onChange={(value) => handleFormDataChange("schedule", value || "")}
            leftSection={<IconClock size={16} />}
          />

          {formData.schedule === "custom" && (
            <TextInput
              label="Custom Cron Expression"
              placeholder="0 2 * * * (every day at 2 AM)"
              required
              value={customSchedule}
              onChange={(event) => setCustomSchedule(event.currentTarget.value)}
              description="Use standard cron format: minute hour day month weekday"
            />
          )}

          <Switch
            label="Enable Schedule"
            description="Schedule will run automatically when enabled"
            checked={formData.enabled}
            onChange={(event) =>
              handleFormDataChange("enabled", event.currentTarget.checked)
            }
          />

          <Group justify="flex-end" mt="lg">
            <Button
              onClick={handleSubmit}
              loading={formLoading}
              disabled={!isFormValid()}
            >
              {editingSchedule ? "Update" : "Create"} Schedule
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
        size="xl"
        centered
      >
        <Stack gap="lg">
          <Text size="md">
            Are you sure you want to delete the backup schedule for{" "}
            <Text span fw={600}>
              "{scheduleToDelete?.name}"
            </Text>
            ?
          </Text>
          <Text size="sm" c="dimmed">
            This action cannot be undone. Future automated backups will not run.
          </Text>

          <Group justify="flex-end" mt="lg">
            <Button
              variant="subtle"
              onClick={() => setConfirmModalOpened(false)}
            >
              Cancel
            </Button>
            <Button color="error" onClick={confirmDelete}>
              Delete Schedule
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}
