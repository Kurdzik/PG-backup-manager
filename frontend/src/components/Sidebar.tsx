"use client"

import { 
  Box, 
  Stack, 
  Text, 
  ScrollArea,
  Group,
  ThemeIcon,
  UnstyledButton,
  Divider,
  ActionIcon,
  Tooltip
} from "@mantine/core";
import { 
  IconDatabase, 
  IconServer2, 
  IconCloud, 
  IconLogout,
  IconShield,
  IconActivity,
  IconServer,
  IconFolder,
  IconChevronLeft,
  IconChevronRight
} from "@tabler/icons-react";
import React from "react";
import Link from "next/link";
import classes from "./Sidebar.module.css";

interface SidebarItemProps {
  icon: React.ReactNode;
  label: string;
  active?: boolean;
  onClick?: () => void;
  href?: string;
  badge?: string | number;
  status?: 'online' | 'warning' | 'error' | 'offline';
  collapsed?: boolean;
}

interface MainItemType {
  icon: React.ReactNode;
  label: string;
  route: string;
  badge?: string | number;
  status?: 'online' | 'warning' | 'error' | 'offline';
}

const mainItems: MainItemType[] = [
    { 
      icon: <IconDatabase size={16} stroke={1.5} />, 
      label: "Database Connections", 
      route: "/ui/db_connections",
    },
    { 
      icon: <IconCloud size={16} stroke={1.5} />, 
      label: "Backup Destinations", 
      route: "/ui/backup_destinations",
    },
    { 
      icon: <IconServer2  size={16} stroke={1.5} />, 
      label: "Backups", 
      route: "/ui/backups",
    },
];

const bottomItems = [
    { 
      icon: <IconLogout size={16} stroke={1.5} />, 
      label: "Logout" 
    },
];

const StatusDot = ({ status }: { status?: 'online' | 'warning' | 'error' | 'offline' }) => {
  if (!status) return null;
  
  return (
    <Box 
      className={`${classes.statusDot} ${classes[status]}`}
    />
  );
};

const ItemBadge = ({ badge }: { badge?: string | number }) => {
  if (!badge) return null;
  
  return (
    <Text 
      size="xs" 
      className={classes.itemBadge}
    >
      {badge}
    </Text>
  );
};

const SidebarItem = ({ 
  icon, 
  label, 
  active = false, 
  onClick, 
  href, 
  badge, 
  status,
  collapsed = false
}: SidebarItemProps) => {
  const buttonContent = (
    <UnstyledButton
      onClick={onClick}
      data-active={active || undefined}
      className={classes.sidebarItem}
      style={{ width: '100%' }}
    >
      {collapsed ? (
        <Group justify="center" align="center">
          <ThemeIcon
            variant="light"
            size={28}
            radius={4}
            className={classes.itemIcon}
            color={active ? "slate" : "gray"}
          >
            {icon}
          </ThemeIcon>
        </Group>
      ) : (
        <Group gap="sm" align="center" wrap="nowrap">
          <ThemeIcon
            variant="light"
            size={28}
            radius={4}
            className={classes.itemIcon}
            color={active ? "slate" : "gray"}
          >
            {icon}
          </ThemeIcon>
          
          <Text 
            size="sm" 
            fw={active ? 500 : 400}
            className={classes.itemText}
            style={{ flex: 1 }}
          >
            {label}
          </Text>
          
          <Group gap={6} align="center">
            <ItemBadge badge={badge} />
            <StatusDot status={status} />
          </Group>
        </Group>
      )}
    </UnstyledButton>
  );

  const content = collapsed ? (
    <Tooltip label={label} position="right" withArrow>
      {buttonContent}
    </Tooltip>
  ) : buttonContent;

  if (href) {
    return (
      <Link href={href} style={{ textDecoration: 'none', color: 'inherit' }}>
        {content}
      </Link>
    );
  }

  return content;
};

export const SidebarComponent = () => {
  const [activeItem, setActiveItem] = React.useState("Database Connections");
  const [collapsed, setCollapsed] = React.useState(false);

  // Load state from localStorage on mount
  React.useEffect(() => {
    if (typeof window !== 'undefined') {
      const savedActiveItem = localStorage.getItem('sidebar-active-item');
      const savedCollapsed = localStorage.getItem('sidebar-collapsed');
      
      if (savedActiveItem) {
        setActiveItem(savedActiveItem);
      }
      
      if (savedCollapsed) {
        setCollapsed(JSON.parse(savedCollapsed));
      }
    }
  }, []);

  // Save active item to localStorage
  const handleItemClick = (label: string) => {
    setActiveItem(label);
    if (typeof window !== 'undefined') {
      localStorage.setItem('sidebar-active-item', label);
    }
  };

  // Toggle collapsed state
  const toggleCollapsed = () => {
    const newCollapsed = !collapsed;
    setCollapsed(newCollapsed);
    if (typeof window !== 'undefined') {
      localStorage.setItem('sidebar-collapsed', JSON.stringify(newCollapsed));
    }
  };

  return (
    <Box 
      className={classes.sidebarContainer}
      style={{ 
        width: collapsed ? '80px' : '280px',
        transition: 'width 0.2s ease'
      }}
    >
      {/* Header */}
      <Box className={classes.sidebarHeader} p={collapsed ? "sm" : "lg"}>
        {collapsed ? (
          <Group justify="center">
            <ThemeIcon
              size={32}
              radius={6}
              variant="light"
              color="slate"
            >
              <IconShield size={18} stroke={1.5} />
            </ThemeIcon>
          </Group>
        ) : (
          <Group gap="sm" align="center" justify="space-between">
            <Group gap="sm" align="center">
              <ThemeIcon
                size={32}
                radius={6}
                variant="light"
                color="slate"
              >
                <IconShield size={18} stroke={1.5} />
              </ThemeIcon>
              <Box>
                <Text size="md" fw={600} c="slate.8">
                  PG Backup Manager
                </Text>
              </Box>
            </Group>
          </Group>
        )}
      </Box>

      <Divider className={classes.headerDivider} />

      {/* Collapse Toggle */}
      <Box p="xs">
        <Group justify={collapsed ? "center" : "flex-end"}>
          <ActionIcon
            variant="subtle"
            size="sm"
            onClick={toggleCollapsed}
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          >
            {collapsed ? (
              <IconChevronRight size={16} />
            ) : (
              <IconChevronLeft size={16} />
            )}
          </ActionIcon>
        </Group>
      </Box>

      {/* Main Navigation */}
      <ScrollArea flex={1} p={collapsed ? "xs" : "md"} className={classes.scrollArea}>
        <Stack gap="lg">
          <Box className={classes.menuSection}>
            {!collapsed && (
              <Text 
                size="xs" 
                fw={500} 
                c="dimmed" 
                tt="uppercase" 
                mb="sm"
                className={classes.sectionLabel}
              >
                Management
              </Text>
            )}
            
            <Stack gap={2}>
              {mainItems.map((item) => (
                <SidebarItem
                  key={item.label}
                  icon={item.icon}
                  label={item.label}
                  active={activeItem === item.label}
                  href={item.route}
                  badge={item.badge}
                  status={item.status}
                  collapsed={collapsed}
                  onClick={() => handleItemClick(item.label)}
                />
              ))}
            </Stack>
          </Box>
        </Stack>
      </ScrollArea>

      {/* Bottom Section */}
      <Box className={classes.sidebarBottom}>
        <Divider className={classes.bottomDivider} />
        
        <Box p={collapsed ? "xs" : "md"}>
          {!collapsed && (
            <Text 
              size="xs" 
              fw={500} 
              c="dimmed" 
              tt="uppercase" 
              mb="sm"
              className={classes.sectionLabel}
            >
              Account
            </Text>
          )}
          
          <Stack gap={2}>
            {bottomItems.map((item) => (
              <SidebarItem
                key={item.label}
                icon={item.icon}
                label={item.label}
                active={activeItem === item.label}
                collapsed={collapsed}
                onClick={() => handleItemClick(item.label)}
              />
            ))}
          </Stack>
        </Box>
      </Box>
    </Box>
  );
};