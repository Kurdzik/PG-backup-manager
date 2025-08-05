"use client";
import { Text, Paper, ActionIcon, Flex, Box, rem } from "@mantine/core";
import { IconCheck, IconX } from "@tabler/icons-react";

export interface NotificationData {
  type: "success" | "error";
  title: string;
  message: string;
}

const BottomRightNotification = ({
  notification,
  onClose,
}: {
  notification: NotificationData | null;
  onClose: () => void;
}) => {
  if (!notification) return null;

  return (
    <Box
      style={{
        position: "fixed",
        bottom: rem(24),
        right: rem(24),
        zIndex: 1000,
        minWidth: rem(320),
        maxWidth: rem(400),
      }}
    >
      <Paper
        shadow="sm"
        p="md"
        radius="sm"
        style={{
          backgroundColor:
            notification.type === "success"
              ? "var(--mantine-color-neutral-0)"
              : "var(--mantine-color-neutral-0)",
          border: `1px solid ${notification.type === "success" ? "var(--mantine-color-success-6)" : "var(--mantine-color-error-6)"}`,
          borderLeft: `3px solid ${notification.type === "success" ? "var(--mantine-color-success-6)" : "var(--mantine-color-error-6)"}`,
        }}
      >
        <Flex align="flex-start" gap="sm">
          <Box
            style={{
              color:
                notification.type === "success"
                  ? "var(--mantine-color-success-6)"
                  : "var(--mantine-color-error-6)",
              marginTop: rem(2),
            }}
          >
            {notification.type === "success" ? (
              <IconCheck size={18} stroke={1.5} />
            ) : (
              <IconX size={18} stroke={1.5} />
            )}
          </Box>
          <Box style={{ flex: 1 }}>
            <Text fw={500} size="sm" mb={4}>
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
            radius="sm"
          >
            <IconX size={14} stroke={1.5} />
          </ActionIcon>
        </Flex>
      </Paper>
    </Box>
  );
};

export default BottomRightNotification;
