const conf = {
  backendUrl: process.env.NEXT_PUBLIC_BACKEND_URL || "NO BACKEND URL PROVIDED",
  apiKey: process.env.NEXT_PUBLIC_API_KEY || "NO API KEY PROVIDED",
  apiVersion: "v1",
};

export async function get(endpoint: string) {
  const url = `${conf.backendUrl}/api/${conf.apiVersion}/${endpoint}`;
  const options = {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": conf.apiKey,
    },
  };

  return fetch(url, options)
    .then(res => res.json())
    .catch(err => {
      throw err;
    });
}

// biome-ignore lint/suspicious/noExplicitAny: <explanation>
export async function post(endpoint: string, requestBody?: any) {
  const url = `${conf.backendUrl}/api/${conf.apiVersion}/${endpoint}`;
  const options = {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": conf.apiKey,
    },
    body: JSON.stringify(requestBody),
  };

  return fetch(url, options)
    .then(res => res.json())
    .catch(err => {
      throw err;
    });
}

// biome-ignore lint/suspicious/noExplicitAny: <explanation>
export async function put(endpoint: string, requestBody?: any) {
  const url = `${conf.backendUrl}/api/${conf.apiVersion}/${endpoint}`;
  const options = {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": conf.apiKey,
    },
    body: JSON.stringify(requestBody),
  };

  return fetch(url, options)
    .then(res => res.json())
    .catch(err => {
      throw err;
    });
}


export async function del(endpoint: string) {
  const url = `${conf.backendUrl}/api/${conf.apiVersion}/${endpoint}`;
  const options = {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": conf.apiKey,
    },
  };

  return fetch(url, options)
    .then(res => res.json())
    .catch(err => {
      throw err;
    });
}