const API_BASE = (import.meta.env.VITE_API_URL?.trim() || (import.meta.env.DEV ? "http://localhost:3000" : "")).replace(/\/+$/, "");
const TOKEN_KEY = "multicatalogo_token";

export const readToken = () => localStorage.getItem(TOKEN_KEY);
export const saveToken = (token: string) => localStorage.setItem(TOKEN_KEY, token);
export const clearToken = () => localStorage.removeItem(TOKEN_KEY);

export async function api<T>(path: string, init: RequestInit = {}, options: { auth?: boolean } = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  const token = readToken();
  if (options.auth !== false && token) headers.set("Authorization", `Bearer ${token}`);

  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, { ...init, headers });
  } catch {
    throw new Error(navigator.onLine
      ? "No se pudo conectar con el servidor. Comprueba que la API esté disponible."
      : "No hay conexión a internet. Comprueba tu conexión e inténtalo de nuevo.");
  }
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
