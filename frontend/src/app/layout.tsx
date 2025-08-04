"use client";

import "@mantine/core/styles.css";
import "@mantine/notifications/styles.css";
import { ColorSchemeScript, MantineProvider } from "@mantine/core";
import { Notifications } from "@mantine/notifications";

import React from "react";
import { theme } from "../../theme";

// biome-ignore lint/suspicious/noExplicitAny: <explanation>
export default function RootLayout({ children }: { children: any }) {
  return (
    <html lang="en">
      <title>PG Backup Manager</title>
      <head>
        <ColorSchemeScript />
        <link rel="shortcut icon" href="/favicon.ico" />
        <meta
          name="viewport"
          content="minimum-scale=1, initial-scale=1, width=device-width, user-scalable=no"
        />
      </head>
      <body
        style={{
          background:
            "linear-gradient(135deg, var(--mantine-color-slate-0) 0%, var(--mantine-color-slate-0) 25%, var(--mantine-color-slate-2) 100%)",
        }}
      >
        <MantineProvider theme={theme}>
          <Notifications position="bottom-center" />
          {children}
        </MantineProvider>
      </body>
    </html>
  );
}
