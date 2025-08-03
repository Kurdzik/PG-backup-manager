export const setAuthCookie = (token: string) => {
  const payload = JSON.parse(atob(token.split('.')[1]));
  const expDate = new Date(payload.exp * 1000);
  
  document.cookie = `auth_token=${token}; expires=${expDate.toUTCString()}; path=/; secure; samesite=strict`;
};