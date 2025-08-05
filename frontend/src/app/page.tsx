"use client";

import { useState, useEffect } from "react";
import {
  Container,
  Title,
  Text,
  Button,
  TextInput,
  PasswordInput,
  Stack,
  Paper,
  Box,
  rem,
  Alert,
  Flex,
  Loader,
  Center,
} from "@mantine/core";
import {
  IconUser,
  IconLock,
  IconLogin,
  IconInfoCircle,
  IconDatabase,
  IconUserPlus,
} from "@tabler/icons-react";

import { post, get } from "@/lib/backendRequests";
import { setAuthCookie } from "@/lib/cookies";

interface FormData {
  username: string;
  password: string;
}

interface LoginResponse {
  payload?: string;
  message: string;
  status: number;
}

interface UserListResponse {
  data: Array<{
    id: number;
    username: string;
    password: string;
    created_at: string;
    updated_at: string;
  }>;
  status: number;
}

interface CreateUserResponse {
  status: number;
  message: string;
}

export default function AuthPage() {
  const [formData, setFormData] = useState<FormData>({
    username: "",
    password: "",
  });
  const [loading, setLoading] = useState<boolean>(false);
  const [checkingUsers, setCheckingUsers] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [hasUsers, setHasUsers] = useState<boolean>(false);
  const [isRegistration, setIsRegistration] = useState<boolean>(false);

  // Check if users exist on component mount
  useEffect(() => {
    checkForExistingUsers();
  }, []);

  const checkForExistingUsers = async (): Promise<void> => {
    try {
      setCheckingUsers(true);
      const response: UserListResponse = await get("user/list", false);

      if (response.status == 200 && response.data && response.data.length > 0) {
        setHasUsers(true);
        setIsRegistration(false);
      } else {
        setHasUsers(false);
        setIsRegistration(true);
      }
    } catch (err: any) {
      console.error("Error checking for users:", err);
      // If we can't check, assume we need registration
      setHasUsers(false);
      setIsRegistration(true);
    } finally {
      setCheckingUsers(false);
    }
  };

  const handleFormDataChange = <K extends keyof FormData>(
    field: K,
    value: FormData[K],
  ): void => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const isFormValid = (): boolean => {
    return !!(formData.username && formData.password);
  };

  const handleLogin = async (): Promise<void> => {
    if (!isFormValid()) return;

    try {
      setLoading(true);
      setError(null);

      const response: LoginResponse = await post(
        "user/login",
        {
          username: formData.username,
          password: formData.password,
        },
        false,
      );

      if (response.payload && response.status == 200) {
        setAuthCookie(response.payload);
        window.location.href = "/ui/db_connections";
      } else {
        setError(response.message);
      }
    } catch (err: any) {
      setError(err.message || "An error occurred during login");
      console.error("Login error:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleRegistration = async (): Promise<void> => {
    if (!isFormValid()) return;

    try {
      setLoading(true);
      setError(null);

      const response: CreateUserResponse = await post(
        "user/create",
        {
          username: formData.username,
          password: formData.password,
        },
        false,
      );

      if (response.status == 200) {
        // After successful registration, automatically log in
        await handleLogin();
      } else {
        setError(response.message || "Registration failed");
      }
    } catch (err: any) {
      setError(err.message || "An error occurred during registration");
      console.error("Registration error:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (): Promise<void> => {
    if (isRegistration) {
      await handleRegistration();
    } else {
      await handleLogin();
    }
  };

  const handleKeyPress = (event: React.KeyboardEvent): void => {
    if (event.key === "Enter" && isFormValid()) {
      handleSubmit();
    }
  };

  // Show loading spinner while checking for users
  if (checkingUsers) {
    return (
      <Box
        style={{
          minHeight: "100vh",
          background:
            "linear-gradient(135deg, var(--mantine-color-slate-0) 0%, var(--mantine-color-slate-3) 50%, var(--mantine-color-slate-5) 100%)",
        }}
      >
        <Center style={{ minHeight: "100vh" }}>
          <Stack align="center" gap="md">
            <Loader size="lg" color="slate" />
            <Text c="slate.6" size="sm">
              Checking system status...
            </Text>
          </Stack>
        </Center>
      </Box>
    );
  }

  return (
    <Box
      style={{
        minHeight: "100vh",
        position: "relative",
        overflow: "hidden",
        background:
          "linear-gradient(135deg, var(--mantine-color-slate-0) 0%, var(--mantine-color-slate-3) 50%, var(--mantine-color-slate-5) 100%)",
      }}
    >
      {/* Blurred Background Elements */}
      <Box
        style={{
          position: "absolute",
          top: "-15%",
          left: "-8%",
          width: "35%",
          height: "50%",
          borderRadius: "50%",
          background: "var(--mantine-color-slate-4)",
          filter: "blur(120px)",
          opacity: 0.15,
          zIndex: 0,
        }}
      />
      <Box
        style={{
          position: "absolute",
          bottom: "-25%",
          right: "-12%",
          width: "45%",
          height: "60%",
          borderRadius: "50%",
          background: "var(--mantine-color-slate-3)",
          filter: "blur(140px)",
          opacity: 0.12,
          zIndex: 0,
        }}
      />
      <Box
        style={{
          position: "absolute",
          top: "25%",
          right: "15%",
          width: "25%",
          height: "35%",
          borderRadius: "50%",
          background: "var(--mantine-color-slate-5)",
          filter: "blur(100px)",
          opacity: 0.1,
          zIndex: 0,
        }}
      />

      {/* Main Content */}
      <Container
        size="xs"
        style={{
          position: "relative",
          zIndex: 1,
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Paper
          shadow="md"
          p="xl"
          radius="sm"
          withBorder
          style={{
            width: "100%",
            maxWidth: rem(400),
            backgroundColor: "var(--mantine-color-neutral-0)",
            backdropFilter: "blur(12px)",
            border: "1px solid var(--mantine-color-neutral-2)",
            transition: "all 150ms ease",
          }}
        >
          <Stack gap="lg">
            {/* Header */}
            <Box ta="center" mb="md">
              <Flex justify="center" mb="md">
                <Box
                  style={{
                    padding: rem(14),
                    borderRadius: rem(8),
                    backgroundColor: "var(--mantine-color-slate-1)",
                    border: "1px solid var(--mantine-color-slate-2)",
                    transition: "all 150ms ease",
                  }}
                >
                  <IconDatabase
                    size={28}
                    color="var(--mantine-color-slate-6)"
                  />
                </Box>
              </Flex>
              <Title
                order={2}
                size="1.5rem"
                fw={600}
                mb="xs"
                c="slate.8"
                style={{
                  fontFamily: "var(--mantine-font-family)",
                }}
              >
                {isRegistration ? "Get Started" : "Welcome Back"}
              </Title>
              <Text
                size="sm"
                c="slate.5"
                style={{
                  fontFamily: "var(--mantine-font-family)",
                }}
              >
                {isRegistration
                  ? "Create your admin account to get started"
                  : "Sign in to manage your database connections"}
              </Text>
            </Box>

            {/* Error Alert */}
            {error && (
              <Alert
                icon={<IconInfoCircle size={18} />}
                color="error"
                radius="sm"
                style={{
                  backgroundColor: "var(--mantine-color-error-0)",
                  border: "1px solid var(--mantine-color-error-2)",
                  color: "var(--mantine-color-error-7)",
                }}
              >
                {error}
              </Alert>
            )}

            {/* Form */}
            <Stack gap="md">
              <TextInput
                label="Username"
                placeholder="Enter your username"
                required
                size="sm"
                leftSection={
                  <IconUser size={16} color="var(--mantine-color-slate-5)" />
                }
                value={formData.username}
                onChange={(event) =>
                  handleFormDataChange("username", event.currentTarget.value)
                }
                onKeyPress={handleKeyPress}
                styles={{
                  label: {
                    color: "var(--mantine-color-slate-7)",
                    fontSize: rem(14),
                    fontWeight: 500,
                    marginBottom: rem(6),
                  },
                  input: {
                    fontSize: rem(14),
                    backgroundColor: "var(--mantine-color-neutral-0)",
                    border: "1px solid var(--mantine-color-neutral-4)",
                    borderRadius: rem(6),
                    "&:focus": {
                      borderColor: "var(--mantine-color-slate-6)",
                      boxShadow: "0 0 0 2px var(--mantine-color-slate-1)",
                    },
                  },
                }}
              />

              <PasswordInput
                label="Password"
                placeholder="Enter your password"
                required
                size="sm"
                leftSection={
                  <IconLock size={16} color="var(--mantine-color-slate-5)" />
                }
                value={formData.password}
                onChange={(event) =>
                  handleFormDataChange("password", event.currentTarget.value)
                }
                onKeyPress={handleKeyPress}
                styles={{
                  label: {
                    color: "var(--mantine-color-slate-7)",
                    fontSize: rem(14),
                    fontWeight: 500,
                    marginBottom: rem(6),
                  },
                  input: {
                    fontSize: rem(14),
                    backgroundColor: "var(--mantine-color-neutral-0)",
                    border: "1px solid var(--mantine-color-neutral-4)",
                    borderRadius: rem(6),
                    "&:focus": {
                      borderColor: "var(--mantine-color-slate-6)",
                      boxShadow: "0 0 0 2px var(--mantine-color-slate-1)",
                    },
                  },
                }}
              />

              <Button
                fullWidth
                size="sm"
                leftSection={
                  isRegistration ? (
                    <IconUserPlus size={16} />
                  ) : (
                    <IconLogin size={16} />
                  )
                }
                onClick={handleSubmit}
                loading={loading}
                disabled={!isFormValid()}
                mt="md"
                color="slate"
                styles={{
                  root: {
                    height: rem(40),
                    fontSize: rem(14),
                    fontWeight: 500,
                    borderRadius: rem(6),
                    backgroundColor: "var(--mantine-color-slate-6)",
                    transition: "all 150ms ease",
                    "&:hover:not([data-disabled])": {
                      backgroundColor: "var(--mantine-color-slate-7)",
                      transform: "translateY(-1px)",
                      boxShadow: "var(--mantine-shadow-sm)",
                    },
                    "&[data-disabled]": {
                      backgroundColor: "var(--mantine-color-slate-3)",
                      color: "var(--mantine-color-slate-5)",
                    },
                  },
                }}
              >
                {isRegistration ? "Create Account" : "Sign In"}
              </Button>
            </Stack>

            {/* Footer */}
            <Box ta="center" mt="md"></Box>
          </Stack>
        </Paper>
      </Container>
    </Box>
  );
}
