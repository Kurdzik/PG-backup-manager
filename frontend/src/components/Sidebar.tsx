import { 
  Box, 
  Stack, 
  Text, 
  ScrollArea,
  Group,
  ThemeIcon,
  UnstyledButton,
  Divider
} from "@mantine/core";
import { 
  IconHome, 
  IconUser, 
  IconSettings, 
  IconChartBar, 
  IconFiles,
  IconLogout,
  IconDashboard
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
}

interface MainItemType {
  icon: React.ReactNode;
  label: string;
  route: string;
}

const mainItems: MainItemType[] = [
    { icon: <IconHome size={18} />, label: "Database Connections", route: "/ui/db_connections" },
    { icon: <IconChartBar size={18} />, label: "Backups", route: "/ui/backups" },
    { icon: <IconFiles size={18} />, label: "Schedules", route: "/ui/schedules" },
];

const bottomItems = [
    { icon: <IconLogout size={18} />, label: "Logout" },
];

const SidebarItem = ({ icon, label, active = false, onClick, href }: SidebarItemProps) => {
  const content = (
    <UnstyledButton
      onClick={onClick}
      data-active={active || undefined}
      className={classes.sidebarItem}
      style={{ width: '100%' }}
    >
      <Group gap="md" align="center">
        <ThemeIcon
          variant={active ? "filled" : "light"}
          size={25}
          radius="sm"
          className={classes.itemIcon}
        >
          {icon}
        </ThemeIcon>
        <Text 
          size="sm" 
          fw={active ? 600 : 400}
          className={classes.itemText}
        >
          {label}
        </Text>
      </Group>
    </UnstyledButton>
  );

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

  return (
    <Box className={classes.sidebarContainer}>
      {/* Header */}

      <Divider className={classes.headerDivider} />

      {/* Main Navigation */}
      <ScrollArea flex={1} p="md" className={classes.scrollArea}>
        <Stack gap="xs">
          <Text 
            size="xs" 
            fw={600} 
            c="dimmed" 
            tt="uppercase" 
            mb="sm"
            className={classes.sectionLabel}
          >
            Navigation
          </Text>
          
          <Stack gap={2}>
            {mainItems.map((item) => (
              <SidebarItem
                key={item.label}
                icon={item.icon}
                label={item.label}
                active={activeItem === item.label}
                href={item.route}
                onClick={() => setActiveItem(item.label)}
              />
            ))}
          </Stack>
        </Stack>
      </ScrollArea>

      {/* Bottom Section */}
      <Box className={classes.sidebarBottom}>
        <Divider className={classes.bottomDivider} />
        
        <Box p="md">
          <Text 
            size="xs" 
            fw={600} 
            c="dimmed" 
            tt="uppercase" 
            mb="sm"
            className={classes.sectionLabel}
          >
            Account
          </Text>
          
          <Stack gap={2}>
            {bottomItems.map((item) => (
              <SidebarItem
                key={item.label}
                icon={item.icon}
                label={item.label}
                active={activeItem === item.label}
                onClick={() => setActiveItem(item.label)}
              />
            ))}
          </Stack>
        </Box>
      </Box>
    </Box>
  );
};