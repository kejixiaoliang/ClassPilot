export class ApiClientError extends Error {
  constructor(
    readonly code: string,
    message: string,
    readonly status: number
  ) {
    super(message);
    this.name = "ApiClientError";
  }
}

type Envelope<T> = {
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: unknown;
  };
};

export function bootstrapToken() {
  const fragment = new URLSearchParams(window.location.hash.slice(1));
  const token = fragment.get("token");
  if (!token) return;

  sessionStorage.setItem("classpilot-token", token);
  window.history.replaceState(
    window.history.state,
    "",
    `${window.location.pathname}${window.location.search}`
  );
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const token = sessionStorage.getItem("classpilot-token");
  const response = await fetch(`/api/v1${path}`, {
    method,
    headers: {
      Accept: "application/json",
      ...(body === undefined ? {} : { "Content-Type": "application/json" }),
      ...(token ? { "X-ClassPilot-Token": token } : {})
    },
    body: body === undefined ? undefined : JSON.stringify(body)
  });

  const envelope = (await response.json()) as Envelope<T>;
  if (!response.ok || envelope.error) {
    const error = envelope.error ?? {
      code: "HTTP_ERROR",
      message: `请求失败（${response.status}）`
    };
    throw new ApiClientError(error.code, error.message, response.status);
  }
  return envelope.data as T;
}

export const api = {
  get: <T>(path: string) => request<T>("GET", path),
  post: <T>(path: string, body?: unknown) => request<T>("POST", path, body),
  patch: <T>(path: string, body: unknown) => request<T>("PATCH", path, body),
  delete: <T>(path: string) => request<T>("DELETE", path)
};
