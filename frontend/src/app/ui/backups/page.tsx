"use client";

import { useState, useEffect } from "react";
import {
  Title,
  Text,
  Button,
  Table,
  Modal,
  Select,
  Group,
  Stack,
  Paper,
  Badge,
  ActionIcon,
  LoadingOverlay,
  Flex,
  Alert,
  Box,
  Card,
  Grid,
  SimpleGrid,
} from "@mantine/core";
import {
  IconDatabase,
  IconHistory,
  IconFileDatabase,
  IconTrash,
  IconRefresh,
  IconInfoCircle,
  IconChartLine,
  IconCloudUpload,
  IconCloud,
  IconCalendar,
  IconClock,
  IconAlertTriangle,
  IconArchive,
} from "@tabler/icons-react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import BottomRightNotification from "@/components/Notifications";
import { NotificationData } from "@/components/Notifications";
import { get, post, del } from "@/lib/backendRequests";
import {
  BackupDestination,
  DatabaseConnection,
  ApiResponse,
} from "@/lib/types";

interface BackupFile {
  filename: string;
  timestamp: Date;
  size?: string;
}

interface BackupStats {
  totalBackups: number;
  lastBackup: Date | null;
  backupsThisMonth: number;
  backupsToday: number;
}

interface ChartDataPoint {
  date: string;
  count: number;
  fullDate: Date;
}

export default function BackupManagerDashboard() {
  const [connections, setConnections] = useState<DatabaseConnection[]>([]);
  const [selectedDatabase, setSelectedDatabase] = useState<string | null>(null);
  const [backupDestinations, setBackupDestinations] = useState<
    BackupDestination[]
  >([]);
  const [backupDestination, setBackupDestination] = useState<string>("");
  const [backups, setBackups] = useState<BackupFile[]>([]);
  const [stats, setStats] = useState<BackupStats>({
    totalBackups: 0,
    lastBackup: null,
    backupsThisMonth: 0,
    backupsToday: 0,
  });
  const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [destinationsLoading, setDestinationsLoading] =
    useState<boolean>(false);
  const [backupsLoading, setBackupsLoading] = useState<boolean>(false);
  const [createBackupLoading, setCreateBackupLoading] =
    useState<boolean>(false);
  const [restoreModalOpened, setRestoreModalOpened] = useState<boolean>(false);
  const [deleteModalOpened, setDeleteModalOpened] = useState<boolean>(false);
  const [selectedBackupFile, setSelectedBackupFile] = useState<string | null>(
    null,
  );
  const [actionType, setActionType] = useState<"restore" | "delete">("restore");
  const [notification, setNotification] = useState<NotificationData | null>(
    null,
  );
  const [deleteLoading, setDeleteLoading] = useState<boolean>(false);
  const [restoreLoading, setRestoreLoading] = useState<boolean>(false);

  useEffect(() => {
    loadConnections();
  }, []);

  useEffect(() => {
    if (selectedDatabase) {
      loadBackupDestinations();
    }
  }, [selectedDatabase]);

  useEffect(() => {
    if (selectedDatabase && backupDestination) {
      loadBackups();
    }
  }, [selectedDatabase, backupDestination]);

  const showNotification = (
    type: "success" | "error",
    title: string,
    message: string,
  ): void => {
    setNotification({ type, title, message });
    setTimeout(() => setNotification(null), 5000);
  };

  const loadConnections = async (): Promise<void> => {
    try {
      setLoading(true);
      const response: ApiResponse<DatabaseConnection[]> =
        await get("connections/list");
      const connectionsList = response.data || [];
      setConnections(connectionsList);

      // Auto-select first database if available
      if (connectionsList.length > 0 && !selectedDatabase) {
        setSelectedDatabase(connectionsList[0].id.toString());
      }
    } catch (err) {
      showNotification("error", "Error", "Failed to load database connections");
      console.error("Error loading connections:", err);
    } finally {
      setLoading(false);
    }
  };

  const loadBackupDestinations = async (): Promise<void> => {
    if (!selectedDatabase) return;

    try {
      setDestinationsLoading(true);
      const response: ApiResponse<BackupDestination[]> = await get(
        `backup-destinations/s3/list?connection_id=${selectedDatabase}`,
      );
      const destinations = response.data || [];
      setBackupDestinations(destinations);
    } catch (err) {
      showNotification("error", "Error", "Failed to load backup destinations");
      console.error("Error loading backup destinations:", err);
      setBackupDestinations([]);
    } finally {
      setDestinationsLoading(false);
    }
  };

  const parseBackupTimestamp = (filename: string): Date => {
    // Parse timestamp from filename format: backup_YYYYMMDD_HHMMSS.dump
    const match = filename.match(/backup_(\d{8})_(\d{6})\.dump/);
    if (match) {
      const dateStr = match[1]; // YYYYMMDD
      const timeStr = match[2]; // HHMMSS

      const year = parseInt(dateStr.substring(0, 4));
      const month = parseInt(dateStr.substring(4, 6)) - 1; // JS months are 0-indexed
      const day = parseInt(dateStr.substring(6, 8));
      const hour = parseInt(timeStr.substring(0, 2));
      const minute = parseInt(timeStr.substring(2, 4));
      const second = parseInt(timeStr.substring(4, 6));

      return new Date(year, month, day, hour, minute, second);
    }
    return new Date(); // fallback
  };

  const calculateStats = (backupFiles: BackupFile[]): BackupStats => {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const thisMonth = new Date(now.getFullYear(), now.getMonth(), 1);

    const backupsToday = backupFiles.filter(
      (backup) => backup.timestamp >= today,
    ).length;

    const backupsThisMonth = backupFiles.filter(
      (backup) => backup.timestamp >= thisMonth,
    ).length;

    const lastBackup =
      backupFiles.length > 0
        ? new Date(Math.max(...backupFiles.map((b) => b.timestamp.getTime())))
        : null;

    return {
      totalBackups: backupFiles.length,
      lastBackup,
      backupsThisMonth,
      backupsToday,
    };
  };

  const generateChartData = (backupFiles: BackupFile[]): ChartDataPoint[] => {
    if (backupFiles.length === 0) return [];

    // Group backups by date
    const dateGroups = new Map<string, number>();

    backupFiles.forEach((backup) => {
      const dateStr = backup.timestamp.toISOString().split("T")[0]; // YYYY-MM-DD
      dateGroups.set(dateStr, (dateGroups.get(dateStr) || 0) + 1);
    });

    // Get the date range (last 30 days or from first backup, whichever is shorter)
    const now = new Date();
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
    const firstBackupDate =
      backupFiles.length > 0
        ? new Date(Math.min(...backupFiles.map((b) => b.timestamp.getTime())))
        : now;

    const startDate =
      firstBackupDate > thirtyDaysAgo ? firstBackupDate : thirtyDaysAgo;

    // Create data points for each day
    const chartData: ChartDataPoint[] = [];
    const currentDate = new Date(startDate);

    while (currentDate <= now) {
      const dateStr = currentDate.toISOString().split("T")[0];
      const count = dateGroups.get(dateStr) || 0;

      chartData.push({
        date: currentDate.toLocaleDateString("en-US", {
          month: "short",
          day: "numeric",
        }),
        count,
        fullDate: new Date(currentDate),
      });

      currentDate.setDate(currentDate.getDate() + 1);
    }

    return chartData;
  };

  const loadBackups = async (): Promise<void> => {
    if (!selectedDatabase || !backupDestination) return;

    try {
      setBackupsLoading(true);
      const response: ApiResponse = await get(
        `backup/list?database_id=${selectedDatabase}&backup_destination=${backupDestination}`,
      );

      const backupFiles: BackupFile[] = (response.payload || []).map(
        (filename) => ({
          filename,
          timestamp: parseBackupTimestamp(filename),
        }),
      );

      // Sort by timestamp descending (newest first)
      backupFiles.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());

      setBackups(backupFiles);
      setStats(calculateStats(backupFiles));
      setChartData(generateChartData(backupFiles));
    } catch (err) {
      showNotification("error", "Error", "Failed to load backups");
      console.error("Error loading backups:", err);
      setBackups([]);
      setStats({
        totalBackups: 0,
        lastBackup: null,
        backupsThisMonth: 0,
        backupsToday: 0,
      });
      setChartData([]);
    } finally {
      setBackupsLoading(false);
    }
  };

  const createBackup = async (): Promise<void> => {
    if (!selectedDatabase) return;

    try {
      setCreateBackupLoading(true);
      const response: ApiResponse = await post("backup/create", {
        database_id: selectedDatabase,
        backup_destination: backupDestination,
      });

      if (response.status === "OK") {
        showNotification("success", "Success", "Backup created successfully");
        loadBackups(); // Refresh the backup list
      }
    } catch (err) {
      showNotification("error", "Error", "Failed to create backup");
      console.error("Error creating backup:", err);
    } finally {
      setCreateBackupLoading(false);
    }
  };

  const handleRestore = async (): Promise<void> => {
    if (!selectedDatabase || !selectedBackupFile) return;

    try {
      setRestoreLoading(true);
      const response: ApiResponse = await post("backup/restore", {
        database_id: selectedDatabase,
        backup_destination: backupDestination,
        backup_filename: selectedBackupFile,
      });

      if (response.status === "OK") {
        showNotification(
          "success",
          "Success",
          "Database restored successfully",
        );
      }
    } catch (err) {
      showNotification("error", "Error", "Failed to restore database");
      console.error("Error restoring database:", err);
    } finally {
      setRestoreLoading(false);
      setRestoreModalOpened(false);
      setSelectedBackupFile(null);
    }
  };

  const handleDelete = async (): Promise<void> => {
    if (!selectedDatabase || !selectedBackupFile) return;

    try {
      setDeleteLoading(true);
      const response: ApiResponse = await del(
        `backup/delete?database_id=${selectedDatabase}&destination=${backupDestination}&filename=${selectedBackupFile}`,
      );

      if (response.status === "OK") {
        showNotification("success", "Success", "Backup deleted successfully");
        loadBackups(); // Refresh the backup list
      }
    } catch (err) {
      showNotification("error", "Error", "Failed to delete backup");
      console.error("Error deleting backup:", err);
    } finally {
      setDeleteLoading(false);
      setDeleteModalOpened(false);
      setSelectedBackupFile(null);
    }
  };

  const openRestoreModal = (filename: string): void => {
    setSelectedBackupFile(filename);
    setActionType("restore");
    setRestoreModalOpened(true);
  };

  const openDeleteModal = (filename: string): void => {
    setSelectedBackupFile(filename);
    setActionType("delete");
    setDeleteModalOpened(true);
  };

  const getSelectedConnectionName = (): string => {
    if (!selectedDatabase) return "";
    const connection = connections.find(
      (c) => c.id.toString() === selectedDatabase,
    );
    return connection ? connection.postgres_db_name : "";
  };

  const getDestinationDisplayName = (destinationValue: string): string => {
    if (destinationValue === "local") return "Local Storage";
    const destination = backupDestinations.find(
      (d) => d.id.toString() === destinationValue,
    );
    return destination ? `${destination.name}` : destinationValue;
  };

  const formatDate = (date: Date): string => {
    return date.toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getTimeAgo = (date: Date): string => {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays} day${diffDays !== 1 ? "s" : ""} ago`;
    if (diffHours > 0)
      return `${diffHours} hour${diffHours !== 1 ? "s" : ""} ago`;
    return "Just now";
  };

  const connectionOptions = connections.map((conn) => ({
    value: conn.id.toString(),
    label: `${conn.postgres_db_name} (${conn.postgres_host}:${conn.postgres_port})`,
  }));

  // Create backup destination options including local and S3 destinations
  const destinationOptions = [
    { value: "local", label: "Local Storage" },
    ...backupDestinations.map((dest) => ({
      value: dest.id.toString(),
      label: `${dest.name} (S3)`,
    })),
  ];

  const backupRows = backups.map((backup) => (
    <Table.Tr key={backup.filename}>
      <Table.Td>
        <Flex align="center" gap="sm">
          <IconArchive size={18} color="#495057" />
          <Box>
            <Text fw={500} size="md">
              {backup.filename}
            </Text>
            <Text size="xs" c="dimmed">
              {getTimeAgo(backup.timestamp)}
            </Text>
          </Box>
        </Flex>
      </Table.Td>
      <Table.Td>
        <Text size="sm" c="dimmed">
          {formatDate(backup.timestamp)}
        </Text>
      </Table.Td>
      <Table.Td>
        <Badge variant="outline" radius={"sm"} size="sm">
          {getDestinationDisplayName(backupDestination)}
        </Badge>
      </Table.Td>
      <Table.Td>
        <Group gap="xs">
          <ActionIcon
            variant="outline"
            onClick={() => openRestoreModal(backup.filename)}
            aria-label={`Restore backup ${backup.filename}`}
          >
            <IconHistory size={16} />
          </ActionIcon>
          <ActionIcon
            variant="outline"
            color="error"
            onClick={() => openDeleteModal(backup.filename)}
            aria-label={`Delete backup ${backup.filename}`}
          >
            <IconTrash size={16} />
          </ActionIcon>
        </Group>
      </Table.Td>
    </Table.Tr>
  ));

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <Paper shadow="sm" p="xs" radius="sm">
          <Text size="sm" fw={500}>
            {label}
          </Text>
          <Text size="sm" c="dimmed">
            {payload[0].value} backup{payload[0].value !== 1 ? "s" : ""}
          </Text>
        </Paper>
      );
    }
    return null;
  };

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
          <IconFileDatabase size={32} />
          <Box>
            <Title order={1} size="2rem" fw={600} mb="xs">
              Backup Manager
            </Title>
            <Text size="lg" c="dimmed">
              Create, manage, and restore PostgreSQL database backups
            </Text>
          </Box>
        </Flex>
      </Box>

      {/* Controls */}
      <Paper shadow="sm" p="lg" mb="xl" style={{ width: "100%" }}>
        <Grid>
          <Grid.Col span={{ base: 12, md: 4 }}>
            <Select
              label="Database Connection"
              placeholder="Select a database"
              data={connectionOptions}
              value={selectedDatabase}
              onChange={setSelectedDatabase}
              leftSection={<IconDatabase size={16} />}
              disabled={loading}
            />
          </Grid.Col>
          <Grid.Col span={{ base: 12, md: 4 }}>
            <Select
              label="Backup Destination"
              placeholder="Select destination"
              data={destinationOptions}
              value={backupDestination}
              onChange={(value) => setBackupDestination(value || "local")}
              leftSection={<IconCloud size={16} />}
              disabled={!selectedDatabase || destinationsLoading}
            />
          </Grid.Col>
          <Grid.Col span={{ base: 12, md: 4 }}>
            <Text size="sm" fw={500} mb="xs">
              Actions
            </Text>
            <Group>
              <Button
                leftSection={<IconCloudUpload size={16} />}
                onClick={createBackup}
                loading={createBackupLoading}
                disabled={!selectedDatabase || !backupDestination}
              >
                Create Backup
              </Button>
              <Button
                leftSection={<IconRefresh size={16} />}
                variant="light"
                onClick={loadBackups}
                loading={backupsLoading}
                disabled={!selectedDatabase || !backupDestination}
              >
                Refresh
              </Button>
            </Group>
          </Grid.Col>
        </Grid>
      </Paper>

      {/* Action Bar */}
      <Box mb="xl">
        <Flex justify="space-between" align="center" style={{ width: "100%" }}>
          <Group>
            <Badge variant="outline" radius={"sm"} size="md">
              {backups.length} backup{backups.length !== 1 ? "s" : ""}
            </Badge>
          </Group>
        </Flex>
      </Box>

      {/* Stats Cards */}
      {selectedDatabase && (
        <SimpleGrid cols={{ base: 2, sm: 4 }} mb="xl" spacing="md">
          <Card shadow="sm" p="md">
            <Flex align="center" gap="sm">
              <IconFileDatabase size={20} color="#495057" />
              <Box>
                <Text size="xl" fw={600}>
                  {stats.totalBackups}
                </Text>
                <Text size="sm" c="dimmed">
                  Total Backups
                </Text>
              </Box>
            </Flex>
          </Card>

          <Card shadow="sm" p="md">
            <Flex align="center" gap="sm">
              <IconClock size={20} color="#495057" />
              <Box>
                <Text size="xl" fw={600}>
                  {stats.backupsToday}
                </Text>
                <Text size="sm" c="dimmed">
                  Today
                </Text>
              </Box>
            </Flex>
          </Card>

          <Card shadow="sm" p="md">
            <Flex align="center" gap="sm">
              <IconCalendar size={20} color="#495057" />
              <Box>
                <Text size="xl" fw={600}>
                  {stats.backupsThisMonth}
                </Text>
                <Text size="sm" c="dimmed">
                  This Month
                </Text>
              </Box>
            </Flex>
          </Card>

          <Card shadow="sm" p="md">
            <Flex align="center" gap="sm">
              <IconChartLine size={20} color="#495057" />
              <Box>
                <Text size="sm" fw={500}>
                  {stats.lastBackup ? getTimeAgo(stats.lastBackup) : "Never"}
                </Text>
                <Text size="sm" c="dimmed">
                  Last Backup
                </Text>
              </Box>
            </Flex>
          </Card>
        </SimpleGrid>
      )}

      {/* Chart */}
      {selectedDatabase && chartData.length > 0 && (
        <Paper shadow="sm" p="lg" mb="xl" style={{ width: "100%" }}>
          <Title order={3} mb="md" fw={500}>
            Backup Activity (Last 30 Days)
          </Title>
          <Box style={{ height: 200 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" fontSize={12} />
                <YAxis fontSize={12} />
                <Tooltip content={<CustomTooltip />} />
                <Line
                  type="monotone"
                  dataKey="count"
                  stroke="#495057"
                  strokeWidth={2}
                  dot={{ fill: "#495057", strokeWidth: 2, r: 3 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </Box>
        </Paper>
      )}

      {/* Backups Table */}
      <Paper shadow="sm" p="lg" pos="relative" style={{ width: "100%" }}>
        <LoadingOverlay visible={backupsLoading} />

        <Flex justify="space-between" align="center" mb="md">
          <Title order={3} fw={500}>
            Backup Files
            {selectedDatabase && (
              <Text span c="dimmed" fw={400}>
                {" "}
                - {getSelectedConnectionName()}
              </Text>
            )}
          </Title>
        </Flex>

        {!selectedDatabase ? (
          <Alert
            icon={<IconInfoCircle size={20} />}
            title="No database selected"
            color="blue"
          >
            <Text>Please select a database connection to view backups.</Text>
          </Alert>
        ) : backups.length === 0 && !backupsLoading ? (
          <Alert
            icon={<IconInfoCircle size={20} />}
            title="No backups found"
            color="blue"
          >
            <Text>
              No backup files found for the selected database and destination.
            </Text>
          </Alert>
        ) : (
          <Table striped highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Backup File</Table.Th>
                <Table.Th>Created</Table.Th>
                <Table.Th>Destination</Table.Th>
                <Table.Th>Actions</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>{backupRows}</Table.Tbody>
          </Table>
        )}
      </Paper>

      {/* Restore Confirmation Modal */}
      <Modal
        opened={restoreModalOpened}
        onClose={() => !restoreLoading && setRestoreModalOpened(false)}
        title={
          <Text fw={600} size="lg">
            Restore Database
          </Text>
        }
        size="md"
        centered
        closeOnClickOutside={!restoreLoading}
        closeOnEscape={!restoreLoading}
      >
        <LoadingOverlay visible={restoreLoading} />
        <Stack gap="lg">
          <Alert
            icon={<IconAlertTriangle size={20} />}
            title="Warning"
            color="orange"
          >
            <Text>
              This will restore the database "{getSelectedConnectionName()}"
              from the backup file{" "}
              <Text span fw={500}>
                "{selectedBackupFile}"
              </Text>
              .
            </Text>
            <Text mt="xs">
              All current data in the database will be replaced. This action
              cannot be undone.
            </Text>
          </Alert>

          <Group justify="flex-end" mt="lg">
            <Button
              variant="subtle"
              onClick={() => setRestoreModalOpened(false)}
              disabled={restoreLoading}
            >
              Cancel
            </Button>
            <Button
              color="orange"
              onClick={handleRestore}
              loading={restoreLoading}
            >
              Restore Database
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        opened={deleteModalOpened}
        onClose={() => setDeleteModalOpened(false)}
        title={
          <Text fw={600} size="lg">
            Delete Backup
          </Text>
        }
        size="md"
        centered
      >
        <Stack gap="lg">
          <Text size="md">
            Are you sure you want to delete the backup file{" "}
            <Text span fw={600}>
              "{selectedBackupFile}"
            </Text>
            ?
          </Text>
          <Text size="sm" c="dimmed">
            This action cannot be undone.
          </Text>

          <Group justify="flex-end" mt="lg">
            <Button
              variant="subtle"
              onClick={() => setDeleteModalOpened(false)}
            >
              Cancel
            </Button>
            <Button color="error" onClick={handleDelete} loading={deleteLoading}>
              Delete Backup
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}