export const setAuthCookie = (token: string) => {
  const payload = JSON.parse(atob(token.split('.')[1]));
  const expDate = new Date(payload.exp * 1000);
  
  document.cookie = `auth_token=${token}; expires=${expDate.toUTCString()}; path=/; secure; samesite=strict`;
};

export const getAuthCookie = (): string => {
  const cookies = document.cookie.split(';');
  
  for (let cookie of cookies) {
    const [name, value] = cookie.trim().split('=');
    if (name === 'auth_token') {
      return value || '';
    }
  }
  
  return '';
};

export const removeAuthCookie = (): void => {
  document.cookie = 'auth_token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; secure; samesite=strict';
};

export const isAuthenticated = (): boolean => {
  const token = getAuthCookie();
  
  if (!token) return false;
  
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const currentTime = Math.floor(Date.now() / 1000);
    
    return payload.exp > currentTime;
  } catch (error) {
    return false;
  }
};