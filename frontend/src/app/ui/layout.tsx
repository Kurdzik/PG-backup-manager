"use client";

import "@mantine/core/styles.css";
import { Box, Grid, Paper } from "@mantine/core";
import React from "react";
import { SidebarComponent } from "@/components/Sidebar";

export default function UILayout({ children }: { children: any }) {
  return (
    <Box
      style={{
        margin: 0,
        height: "100vh",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <Box style={{ flex: 1, display: "flex", overflow: "hidden" }}>
        <Grid columns={27} style={{ flex: 1, display: "flex" }}>
          <Grid.Col
            span={"content"}
            style={{ height: "100%", display: "flex", flexDirection: "column" }}
          >
            <SidebarComponent />
          </Grid.Col>

          <Grid.Col
            span={"auto"}
            h={"100%"}
            style={{ overflow: "auto", scrollbarWidth: "none" }}
          >
            {/* <NavbarComponent /> */}
            <Paper
              shadow="xl"
              pt={"1%"}
              pl={"0.5%"}
              pr={"1%"}
              style={{
                display: "flex",
                flexDirection: "column",
                justifyItems: "center",
              }}
            >
              {children}
            </Paper>
          </Grid.Col>
        </Grid>
      </Box>
    </Box>
  );
}
