"use client"

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
  rem,
  Card,
  Grid,
  Progress,
  Divider,
} from '@mantine/core';
import {
  IconDatabase,
  IconDownload,
  IconRestore,
  IconTrash,
  IconRefresh,
  IconCheck,
  IconX,
  IconInfoCircle,
  IconChartLine,
  IconCloudUpload,
  IconServer,
  IconCalendar,
  IconClock,
} from '@tabler/icons-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

import { get, post } from "../../../lib/backendRequests";

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

interface NotificationData {
  type: 'success' | 'error';
  title: string;
  message: string;
}

interface ApiResponse<T = any> {
  data?: T;
  msg?: string[];
  status?: string;
  count?: number;
}

interface ChartDataPoint {
  date: string;
  count: number;
  fullDate: Date;
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

export default function BackupManagerDashboard() {
  const [connections, setConnections] = useState<DatabaseConnection[]>([]);
  const [selectedDatabase, setSelectedDatabase] = useState<string | null>(null);
  const [backupDestination, setBackupDestination] = useState<string>('local');
  const [backups, setBackups] = useState<BackupFile[]>([]);
  const [stats, setStats] = useState<BackupStats>({
    totalBackups: 0,
    lastBackup: null,
    backupsThisMonth: 0,
    backupsToday: 0,
  });
  const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [backupsLoading, setBackupsLoading] = useState<boolean>(false);
  const [createBackupLoading, setCreateBackupLoading] = useState<boolean>(false);
  const [restoreModalOpened, setRestoreModalOpened] = useState<boolean>(false);
  const [confirmModalOpened, setConfirmModalOpened] = useState<boolean>(false);
  const [selectedBackupFile, setSelectedBackupFile] = useState<string | null>(null);
  const [actionType, setActionType] = useState<'restore' | 'delete'>('restore');
  const [notification, setNotification] = useState<NotificationData | null>(null);

  useEffect(() => {
    loadConnections();
  }, []);

  useEffect(() => {
    if (selectedDatabase) {
      loadBackups();
    }
  }, [selectedDatabase, backupDestination]);

  const showNotification = (type: 'success' | 'error', title: string, message: string): void => {
    setNotification({ type, title, message });
    setTimeout(() => setNotification(null), 5000);
  };

  const loadConnections = async (): Promise<void> => {
    try {
      setLoading(true);
      const response: ApiResponse<DatabaseConnection[]> = await get("connections/list");
      const connectionsList = response.data || [];
      setConnections(connectionsList);
      
      // Auto-select first database if available
      if (connectionsList.length > 0 && !selectedDatabase) {
        setSelectedDatabase(connectionsList[0].id.toString());
      }
    } catch (err) {
      showNotification('error', 'Error', 'Failed to load database connections');
      console.error("Error loading connections:", err);
    } finally {
      setLoading(false);
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
    
    const backupsToday = backupFiles.filter(backup => 
      backup.timestamp >= today
    ).length;
    
    const backupsThisMonth = backupFiles.filter(backup => 
      backup.timestamp >= thisMonth
    ).length;
    
    const lastBackup = backupFiles.length > 0 
      ? new Date(Math.max(...backupFiles.map(b => b.timestamp.getTime())))
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
    
    backupFiles.forEach(backup => {
      const dateStr = backup.timestamp.toISOString().split('T')[0]; // YYYY-MM-DD
      dateGroups.set(dateStr, (dateGroups.get(dateStr) || 0) + 1);
    });

    // Get the date range (last 30 days or from first backup, whichever is shorter)
    const now = new Date();
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
    const firstBackupDate = backupFiles.length > 0 
      ? new Date(Math.min(...backupFiles.map(b => b.timestamp.getTime())))
      : now;
    
    const startDate = firstBackupDate > thirtyDaysAgo ? firstBackupDate : thirtyDaysAgo;
    
    // Create data points for each day
    const chartData: ChartDataPoint[] = [];
    const currentDate = new Date(startDate);
    
    while (currentDate <= now) {
      const dateStr = currentDate.toISOString().split('T')[0];
      const count = dateGroups.get(dateStr) || 0;
      
      chartData.push({
        date: currentDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        count,
        fullDate: new Date(currentDate)
      });
      
      currentDate.setDate(currentDate.getDate() + 1);
    }

    return chartData;
  };

  const loadBackups = async (): Promise<void> => {
    if (!selectedDatabase) return;
    
    try {
      setBackupsLoading(true);
      const response: ApiResponse = await get(`backup/list?database_id=${selectedDatabase}&backup_destination=${backupDestination}`);
      
      const backupFiles: BackupFile[] = (response.msg || []).map(filename => ({
        filename,
        timestamp: parseBackupTimestamp(filename),
      }));
      
      // Sort by timestamp descending (newest first)
      backupFiles.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
      
      setBackups(backupFiles);
      setStats(calculateStats(backupFiles));
      setChartData(generateChartData(backupFiles));
    } catch (err) {
      showNotification('error', 'Error', 'Failed to load backups');
      console.error("Error loading backups:", err);
      setBackups([]);
      setStats({ totalBackups: 0, lastBackup: null, backupsThisMonth: 0, backupsToday: 0 });
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
        database_id: parseInt(selectedDatabase),
        backup_destination: backupDestination,
      });
      
      if (response.status === 'OK') {
        showNotification('success', 'Success', 'Backup created successfully');
        loadBackups(); // Refresh the backup list
      }
    } catch (err) {
      showNotification('error', 'Error', 'Failed to create backup');
      console.error("Error creating backup:", err);
    } finally {
      setCreateBackupLoading(false);
    }
  };

  const handleRestore = async (): Promise<void> => {
    if (!selectedDatabase || !selectedBackupFile) return;
    
    try {
      const response: ApiResponse = await post("backup/restore", {
        database_id: parseInt(selectedDatabase),
        backup_destination: backupDestination,
        backup_filename: selectedBackupFile,
      });
      
      if (response.status === 'OK') {
        showNotification('success', 'Success', 'Database restored successfully');
      }
    } catch (err) {
      showNotification('error', 'Error', 'Failed to restore database');
      console.error("Error restoring database:", err);
    } finally {
      setRestoreModalOpened(false);
      setSelectedBackupFile(null);
    }
  };

  const openRestoreModal = (filename: string): void => {
    setSelectedBackupFile(filename);
    setActionType('restore');
    setRestoreModalOpened(true);
  };

  const getSelectedConnectionName = (): string => {
    if (!selectedDatabase) return '';
    const connection = connections.find(c => c.id.toString() === selectedDatabase);
    return connection ? connection.postgres_db_name : '';
  };

  const formatDate = (date: Date): string => {
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getTimeAgo = (date: Date): string => {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffHours / 24);
    
    if (diffDays > 0) return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
    if (diffHours > 0) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  const connectionOptions = connections.map(conn => ({
    value: conn.id.toString(),
    label: `${conn.postgres_db_name} (${conn.postgres_host}:${conn.postgres_port})`,
  }));

  const backupRows = backups.map((backup) => (
    <Table.Tr key={backup.filename}>
      <Table.Td>
        <Flex align="center" gap="sm">
          <IconDatabase size={16} color="#495057" />
          <Box>
            <Text fw={500} size="sm">{backup.filename}</Text>
            <Text size="xs" c="dimmed">{getTimeAgo(backup.timestamp)}</Text>
          </Box>
        </Flex>
      </Table.Td>
      <Table.Td>
        <Text size="sm">{formatDate(backup.timestamp)}</Text>
      </Table.Td>
      <Table.Td>
        <Badge variant="light" color="blue" size="sm">
          {backupDestination}
        </Badge>
      </Table.Td>
      <Table.Td>
        <Group gap="xs">
          <ActionIcon
            variant="subtle"
            color="green"
            onClick={() => openRestoreModal(backup.filename)}
            aria-label={`Restore backup ${backup.filename}`}
          >
            <IconRestore size={16} />
          </ActionIcon>
        </Group>
      </Table.Td>
    </Table.Tr>
  ));

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <Paper shadow="md" p="xs" style={{ backgroundColor: 'white', border: '1px solid #e9ecef' }}>
          <Text size="sm" fw={500}>{label}</Text>
          <Text size="sm" c="blue">
            {payload[0].value} backup{payload[0].value !== 1 ? 's' : ''}
          </Text>
        </Paper>
      );
    }
    return null;
  };

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
          <IconCloudUpload size={32} />
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

      {/* Database Selection */}
      <Paper shadow="sm" p="lg" mb="xl">
        <Grid>
          <Grid.Col span={6}>
            <Select
              label="Select Database"
              placeholder="Choose a database connection"
              data={connectionOptions}
              value={selectedDatabase}
              onChange={setSelectedDatabase}
              leftSection={<IconServer size={16} />}
              required
            />
          </Grid.Col>
          <Grid.Col span={6}>
            <Select
              label="Backup Location"
              placeholder="Select backup location"
              data={[
                { value: 'local', label: 'Local Storage' },
                // Future: S3 buckets will be added here
              ]}
              value={backupDestination}
              onChange={(value) => setBackupDestination(value || 'local')}
              leftSection={<IconCloudUpload size={16} />}
            />
          </Grid.Col>
        </Grid>
      </Paper>

      {selectedDatabase && (
        <>
          {/* Statistics Cards */}
          <Grid mb="xl">
            <Grid.Col span={3}>
              <Card shadow="sm" p="md">
                <Flex align="center" gap="sm">
                  <IconDatabase color="#228be6" size={24} />
                  <Box>
                    <Text size="xs" tt="uppercase" fw={700} c="dimmed">
                      Total Backups
                    </Text>
                    <Text fw={700} size="xl">
                      {stats.totalBackups}
                    </Text>
                  </Box>
                </Flex>
              </Card>
            </Grid.Col>
            <Grid.Col span={3}>
              <Card shadow="sm" p="md">
                <Flex align="center" gap="sm">
                  <IconCalendar color="#40c057" size={24} />
                  <Box>
                    <Text size="xs" tt="uppercase" fw={700} c="dimmed">
                      This Month
                    </Text>
                    <Text fw={700} size="xl">
                      {stats.backupsThisMonth}
                    </Text>
                  </Box>
                </Flex>
              </Card>
            </Grid.Col>
            <Grid.Col span={3}>
              <Card shadow="sm" p="md">
                <Flex align="center" gap="sm">
                  <IconClock color="#fd7e14" size={24} />
                  <Box>
                    <Text size="xs" tt="uppercase" fw={700} c="dimmed">
                      Today
                    </Text>
                    <Text fw={700} size="xl">
                      {stats.backupsToday}
                    </Text>
                  </Box>
                </Flex>
              </Card>
            </Grid.Col>
            <Grid.Col span={3}>
              <Card shadow="sm" p="md">
                <Flex align="center" gap="sm">
                  <IconChartLine color="#9775fa" size={24} />
                  <Box>
                    <Text size="xs" tt="uppercase" fw={700} c="dimmed">
                      Last Backup
                    </Text>
                    <Text fw={700} size="sm">
                      {stats.lastBackup ? getTimeAgo(stats.lastBackup) : 'Never'}
                    </Text>
                  </Box>
                </Flex>
              </Card>
            </Grid.Col>
          </Grid>

          {/* Backup Frequency Chart */}
          <Paper shadow="sm" p="lg" mb="xl">
            <Flex align="center" gap="sm" mb="md">
              <IconChartLine size={20} color="#495057" />
              <Text fw={600} size="lg">Backup Frequency</Text>
            </Flex>
            <Box style={{ height: 200 }}>
              {chartData.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={chartData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#f1f3f4" />
                    <XAxis 
                      dataKey="date" 
                      stroke="#868e96"
                      fontSize={12}
                      tickLine={false}
                    />
                    <YAxis 
                      stroke="#868e96"
                      fontSize={12}
                      tickLine={false}
                      allowDecimals={false}
                    />
                    <Tooltip content={<CustomTooltip />} />
                    <Line 
                      type="monotone" 
                      dataKey="count" 
                      stroke="#228be6" 
                      strokeWidth={2}
                      dot={{ fill: '#228be6', strokeWidth: 2, r: 4 }}
                      activeDot={{ r: 6, stroke: '#228be6', strokeWidth: 2 }}
                    />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <Flex align="center" justify="center" style={{ height: '100%' }}>
                  <Text c="dimmed" size="sm">No backup data to display</Text>
                </Flex>
              )}
            </Box>
          </Paper>

          {/* Action Bar */}
          <Box mb="xl">
            <Flex justify="space-between" align="center">
              <Group>
                <Button
                  leftSection={<IconRefresh size={16} />}
                  variant="light"
                  onClick={loadBackups}
                  loading={backupsLoading}
                >
                  Refresh
                </Button>
                <Badge variant="light" size="lg">
                  {getSelectedConnectionName()}
                </Badge>
              </Group>
              <Button
                leftSection={<IconDownload size={16} />}
                onClick={createBackup}
                loading={createBackupLoading}
              >
                Create Backup
              </Button>
            </Flex>
          </Box>

          {/* Backups Table */}
          <Paper shadow="sm" p="lg" pos="relative">
            <LoadingOverlay visible={backupsLoading} />
            
            {backups.length === 0 && !backupsLoading ? (
              <Alert icon={<IconInfoCircle size={20} />} title="No backups found" color="blue">
                <Text>Create your first backup for this database to get started.</Text>
              </Alert>
            ) : (
              <Table striped highlightOnHover>
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>Backup File</Table.Th>
                    <Table.Th>Created</Table.Th>
                    <Table.Th>Location</Table.Th>
                    <Table.Th>Actions</Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>{backupRows}</Table.Tbody>
              </Table>
            )}
          </Paper>
        </>
      )}

      {/* Restore Confirmation Modal */}
      <Modal
        opened={restoreModalOpened}
        onClose={() => setRestoreModalOpened(false)}
        title={
          <Text fw={600} size="lg">
            Confirm Database Restore
          </Text>
        }
        size="xl"
        centered
      >
        <Stack gap="lg">
          <Text size="md">
            Are you sure you want to restore the database{' '}
            <Text span fw={600}>
              "{getSelectedConnectionName()}"
            </Text>
            {' '}from backup{' '}
            <Text span fw={600} >
              "{selectedBackupFile}"
            </Text>
            ?
          </Text>
          <Alert color="orange" icon={<IconInfoCircle size={16} />}>
            <Text size="sm">
              This will overwrite all current data in the database. This action cannot be undone.
            </Text>
          </Alert>
          
          <Group justify="flex-end" mt="lg">
            <Button 
              variant="subtle" 
              onClick={() => setRestoreModalOpened(false)}
            >
              Cancel
            </Button>
            <Button 
              onClick={handleRestore}
            >
              Restore Database
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Box>
  );
}