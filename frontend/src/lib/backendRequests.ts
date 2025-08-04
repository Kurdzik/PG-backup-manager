import { removeAuthCookie, getAuthCookie } from "@/lib/cookies";

const conf = {
  backendUrl: process.env.NEXT_PUBLIC_BACKEND_URL || "NO BACKEND URL PROVIDED",
  apiKey: process.env.NEXT_PUBLIC_API_KEY || "NO API KEY PROVIDED",
  apiVersion: "v1",
};

const handleUnauthorized = (): void => {
  removeAuthCookie();
  window.location.href = "/";
};

const createHeaders = (secure: boolean = true): Record<string, string> => {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "x-api-key": conf.apiKey,
  };

  if (secure) {
    const token = getAuthCookie();
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
  }

  return headers;
};

const handleResponse = async (response: Response) => {
  if (response.status === 401) {
    handleUnauthorized();
    throw new Error('Unauthorized - redirecting to login');
  }
  
  return response.json();
};

export async function get(endpoint: string, secure: boolean = true) {
  const url = `${conf.backendUrl}/api/${conf.apiVersion}/${endpoint}`;
  const options = {
    method: "GET",
    headers: createHeaders(secure),
  };

  return fetch(url, options)
    .then(handleResponse)
    .catch(err => {
      throw err;
    });
}

export async function post(endpoint: string, requestBody?: any, secure: boolean = true) {
  const url = `${conf.backendUrl}/api/${conf.apiVersion}/${endpoint}`;
  const options = {
    method: "POST",
    headers: createHeaders(secure),
    body: JSON.stringify(requestBody),
  };

  return fetch(url, options)
    .then(handleResponse)
    .catch(err => {
      throw err;
    });
}

export async function put(endpoint: string, requestBody?: any, secure: boolean = true) {
  const url = `${conf.backendUrl}/api/${conf.apiVersion}/${endpoint}`;
  const options = {
    method: "PUT",
    headers: createHeaders(secure),
    body: JSON.stringify(requestBody),
  };

  return fetch(url, options)
    .then(handleResponse)
    .catch(err => {
      throw err;
    });
}

export async function del(endpoint: string, secure: boolean = true) {
  const url = `${conf.backendUrl}/api/${conf.apiVersion}/${endpoint}`;
  const options = {
    method: "DELETE",
    headers: createHeaders(secure),
  };

  return fetch(url, options)
    .then(handleResponse)
    .catch(err => {
      throw err;
    });
}