const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function request<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("token") : null;

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }

  if (res.status === 204) return {} as T;
  return res.json();
}

export const api = {
  auth: {
    register: (data: { email: string; username: string; password: string }) =>
      request<{ token: string; user: User }>("/api/auth/register", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    login: (data: { email: string; password: string }) =>
      request<{ token: string; user: User }>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    me: () => request<User>("/api/auth/me"),
  },

  agents: {
    list: (page = 1) =>
      request<{ data: Agent[]; total: number }>(`/api/agents?page=${page}`),
    get: (id: string) => request<Agent>(`/api/agents/${id}`),
    create: (data: CreateAgentInput) =>
      request<Agent>("/api/agents", { method: "POST", body: JSON.stringify(data) }),
    update: (id: string, data: UpdateAgentInput) =>
      request<Agent>(`/api/agents/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id: string) =>
      request<void>(`/api/agents/${id}`, { method: "DELETE" }),
  },

  chat: {
    createSession: (data: { agent_id: string; title?: string }) =>
      request<ChatSession>("/api/chat/sessions", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    listSessions: (page = 1) =>
      request<{ data: ChatSession[]; total: number }>(
        `/api/chat/sessions?page=${page}`
      ),
    getMessages: (sessionId: string, page = 1) =>
      request<{ data: Message[] }>(
        `/api/chat/sessions/${sessionId}/messages?page=${page}&page_size=50`
      ),
    sendMessage: (sessionId: string, content: string) =>
      request<Message>(`/api/chat/sessions/${sessionId}/messages`, {
        method: "POST",
        body: JSON.stringify({ content }),
      }),
    deleteSession: (sessionId: string) =>
      request<void>(`/api/chat/sessions/${sessionId}`, { method: "DELETE" }),
  },

  tools: {
    list: (page = 1) =>
      request<{ data: Tool[]; total: number }>(`/api/tools?page=${page}`),
    get: (id: string) => request<Tool>(`/api/tools/${id}`),
    create: (data: CreateToolInput) =>
      request<Tool>("/api/tools", { method: "POST", body: JSON.stringify(data) }),
    delete: (id: string) =>
      request<void>(`/api/tools/${id}`, { method: "DELETE" }),
  },

  triggers: {
    listByAgent: (agentId: string) =>
      request<{ data: Trigger[] }>(`/api/triggers/agent/${agentId}`),
    create: (data: CreateTriggerInput) =>
      request<Trigger>("/api/triggers", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    delete: (id: string) =>
      request<void>(`/api/triggers/${id}`, { method: "DELETE" }),
  },
};

export function createWebSocket(token: string): WebSocket {
  const wsBase = API_BASE.replace(/^http/, "ws");
  return new WebSocket(`${wsBase}/ws?token=${token}`);
}

// Types
export interface User {
  id: string;
  email: string;
  username: string;
  avatar?: string;
  role: string;
  is_active: boolean;
  created_at: string;
}

export interface Agent {
  id: string;
  user_id: string;
  name: string;
  avatar?: string;
  description?: string;
  system_prompt: string;
  model: string;
  provider: string;
  temperature: number;
  max_tokens: number;
  memory?: string;
  persona?: string;
  is_active: boolean;
  tools?: { id: string; tool_id: string }[];
  created_at: string;
  updated_at: string;
}

export interface CreateAgentInput {
  name: string;
  description?: string;
  system_prompt: string;
  model: string;
  provider: string;
  temperature?: number;
  max_tokens?: number;
  persona?: string;
  tool_ids?: string[];
}

export interface UpdateAgentInput {
  name?: string;
  description?: string;
  system_prompt?: string;
  model?: string;
  provider?: string;
  temperature?: number;
  max_tokens?: number;
  persona?: string;
  memory?: string;
  is_active?: boolean;
  tool_ids?: string[];
}

export interface ChatSession {
  id: string;
  user_id: string;
  agent_id: string;
  title: string;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  session_id: string;
  role: "user" | "assistant" | "system" | "tool";
  content: string;
  tool_call_id?: string;
  tool_name?: string;
  token_count?: number;
  created_at: string;
}

export interface Tool {
  id: string;
  name: string;
  display_name: string;
  description: string;
  type: string;
  schema: string;
  config?: string;
  endpoint?: string;
  is_active: boolean;
  created_at: string;
}

export interface CreateToolInput {
  name: string;
  display_name?: string;
  description?: string;
  type: string;
  schema?: string;
  config?: string;
  endpoint?: string;
}

export interface Trigger {
  id: string;
  agent_id: string;
  name: string;
  type: string;
  config: string;
  action: string;
  is_active: boolean;
  last_run_at?: string;
  created_at: string;
}

export interface CreateTriggerInput {
  agent_id: string;
  name: string;
  type: string;
  config: string;
  action: string;
  is_active: boolean;
}
