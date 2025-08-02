"use client";
import { createTheme, rem } from "@mantine/core";

export const theme = createTheme({
  defaultRadius: "md",
  cursorType: "pointer",

  primaryColor: "brand",
  primaryShade: { light: 6, dark: 5 },

  // Microsoft Fluent Design inspired typography
  fontFamily: `"Segoe UI", -apple-system, BlinkMacSystemFont, "Roboto", "Helvetica Neue", Arial, sans-serif`,

  fontSizes: {
    xs: rem(11),
    sm: rem(13),
    md: rem(15),
    lg: rem(17),
    xl: rem(20),
  },

  colors: {
    brand: [
      "#f8fafc",
      "#e6eff7",
      "#cadcf0",
      "#9fc0e2",
      "#7ba4d4",
      "#5c8acc", // ⬅️ primary
      "#4e78b5",
      "#40679e",
      "#355883",
      "#2b4869",
    ],
    neutral: [
      "#f9f9fa",
      "#f1f2f4",
      "#dddddf",
      "#c5c6cb",
      "#a1a2a9",
      "#7e7f88", // previous: #71717a
      "#63646d", // slightly darker
      "#4c4d55", // good for borders
      "#37383e",
      "#24252b",
    ],
    success: [
      "#f4f8f5",
      "#def1e3",
      "#c5e4ce",
      "#a3d2b3",
      "#7fbe97",
      "#5ca07e", // ⬅️ slight pop, still muted
      "#488669",
      "#386d55",
      "#2b5442",
      "#203d31",
    ],
    warning: [
      "#fffaf4",
      "#fef0da",
      "#fde0b3",
      "#fbc87f",
      "#f8ae4f",
      "#e49130", // ⬅️ earthy amber
      "#c47321",
      "#9d5814",
      "#7a410b",
      "#572d06",
    ],
    error: [
      "#fdf7f7",
      "#f6e3e3",
      "#edcdcd",
      "#dda6a6",
      "#ca7d7d",
      "#b25e5e", // ⬅️ balanced red tone
      "#934848",
      "#763636",
      "#5a2727",
      "#421b1b",
    ],
    info: [
      "#f6fafd",
      "#eaf1f9",
      "#d6e3f3",
      "#b7cde7",
      "#94b3da",
      "#7499cb", // ⬅️ slate-blue for feedback/info
      "#5d7eac",
      "#4c6991",
      "#3b5477",
      "#2b405d",
    ],
  },

  headings: {
    fontFamily: `"Segoe UI", -apple-system, BlinkMacSystemFont, "Roboto", "Helvetica Neue", Arial, sans-serif`,
    fontWeight: "600",
    sizes: {
      h1: {
        fontSize: rem(32),
        lineHeight: "1.25",
        fontWeight: "700",
      },
      h2: {
        fontSize: rem(24),
        lineHeight: "1.3",
        fontWeight: "600",
      },
      h3: {
        fontSize: rem(20),
        lineHeight: "1.35",
        fontWeight: "600",
      },
      h4: {
        fontSize: rem(18),
        lineHeight: "1.4",
        fontWeight: "600",
      },
      h5: {
        fontSize: rem(16),
        lineHeight: "1.45",
        fontWeight: "600",
      },
      h6: {
        fontSize: rem(14),
        lineHeight: "1.5",
        fontWeight: "600",
      },
    },
  },

  spacing: {
    xs: rem(4),
    sm: rem(8),
    md: rem(16),
    lg: rem(24),
    xl: rem(40),
  },

  // Modern shadow system
  shadows: {
    xs: "0 1px 2px 0 rgb(0 0 0 / 0.05)",
    sm: "0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)",
    md: "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)",
    lg: "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)",
    xl: "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)",
  },

  // Component-specific customizations
  components: {
    Button: {
      defaultProps: {
        size: "sm",
      },
      styles: {
        root: {
          borderRadius: rem(6),
          fontWeight: 500,
          fontSize: rem(14),
        },
      },
    },

    Card: {
      defaultProps: {
        shadow: "sm",
        radius: "md",
      },
      styles: {
        root: {
          border: "1px solid var(--mantine-color-neutral-2)",
        },
      },
    },

    Input: {
      styles: {
        input: {
          borderRadius: rem(6),
          border: "1px solid var(--mantine-color-neutral-3)",
          fontSize: rem(14),
          "&:focus": {
            borderColor: "var(--mantine-color-brand-6)",
            boxShadow: "0 0 0 2px var(--mantine-color-brand-1)",
          },
        },
      },
    },

    Paper: {
      defaultProps: {
        shadow: "xs",
        radius: "md",
      },
    },

    Modal: {
      styles: {
        content: {
          borderRadius: rem(12),
        },
        header: {
          borderBottom: "1px solid var(--mantine-color-neutral-2)",
          paddingBottom: rem(16),
          marginBottom: rem(16),
        },
      },
    },
  },

  other: {
    borderRadius: rem(8),
    transitionSpeed: "150ms",
    borderColor: "var(--mantine-color-neutral-4)", // NEW
    elevation: {
      subtle: "0 1px 2px rgba(0,0,0,0.04)",
      card: "0 2px 4px rgba(0,0,0,0.08)",
      elevated: "0 4px 8px rgba(0,0,0,0.12)",
      floating: "0 8px 16px rgba(0,0,0,0.16)",
    }, 
  },
});

// Very muted professional color palette for charts and data visualization
export const colors = [
  // "#5c8acc", // brand blue
  "#5ca07e", // balanced green
  "#e49130", // amber
  "#b25e5e", // mellow red
  "#8581b9", // desaturated violet
  "#679aa6", // dusty cyan
  "#85a354", // olive green
  "#b9864a", // copper orange
  "#a77895", // mauve pink
  "#6e7cb3", // indigo
  "#5b998d", // teal
  "#b29a3d", // muted gold
  "#80858e", // neutral slate
  "#9d6363", // rust red
  "#917cb3", // soft purple
  "#52876d", // hunter green
  "#a68a5f", // tan amber
  "#5d94a1", // dusty cyan (alt)
  "#78905e", // sage green
  "#ba7742", // orange-brown
];
