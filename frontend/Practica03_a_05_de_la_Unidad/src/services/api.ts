const API_BASE = import.meta.env.VITE_API_URL ?? "http://localhost:3000";
const TOKEN_KEY = "multicatalogo_token";

export const readToken = () => localStorage.getItem(TOKEN_KEY);
export const saveToken = (token: string) => localStorage.setItem(TOKEN_KEY, token);
export const clearToken = () => localStorage.removeItem(TOKEN_KEY);

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const token = readToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const response = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (!response.ok) {
    let message = `La solicitud falló (${response.status})`;
    try {
      const body = await response.json();
      message = body.message ?? body.error ?? message;
    } catch { /* La respuesta no tenía JSON */ }
    if (response.status === 401) {
      clearToken();
      window.dispatchEvent(new Event("session-expired"));
    }
    throw new Error(message);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export const apiUrl = API_BASE;
