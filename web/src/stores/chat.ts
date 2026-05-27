import { create } from "zustand";
import { api, createWebSocket, type ChatSession, type Message } from "@/services/api";

interface ToolCallEvent {
  name: string;
  status: "calling" | "done";
}

interface ChatState {
  sessions: ChatSession[];
  currentSession: ChatSession | null;
  messages: Message[];
  isStreaming: boolean;
  streamContent: string;
  activeToolCall: ToolCallEvent | null;
  ws: WebSocket | null;

  loadSessions: () => Promise<void>;
  selectSession: (session: ChatSession) => Promise<void>;
  createSession: (agentId: string, title?: string) => Promise<ChatSession>;
  deleteSession: (id: string) => Promise<void>;
  sendMessage: (content: string) => void;
  connectWS: (token: string) => void;
  disconnectWS: () => void;
}

export const useChat = create<ChatState>((set, get) => ({
  sessions: [],
  currentSession: null,
  messages: [],
  isStreaming: false,
  streamContent: "",
  activeToolCall: null,
  ws: null,

  loadSessions: async () => {
    const res = await api.chat.listSessions();
    set({ sessions: res.data || [] });
  },

  selectSession: async (session) => {
    set({ currentSession: session, messages: [], streamContent: "", activeToolCall: null });
    const res = await api.chat.getMessages(session.id);
    set({ messages: res.data || [] });
  },

  createSession: async (agentId, title) => {
    const session = await api.chat.createSession({
      agent_id: agentId,
      title,
    });
    set((s) => ({ sessions: [session, ...s.sessions], currentSession: session, messages: [] }));
    return session;
  },

  deleteSession: async (id) => {
    await api.chat.deleteSession(id);
    set((s) => ({
      sessions: s.sessions.filter((sess) => sess.id !== id),
      currentSession: s.currentSession?.id === id ? null : s.currentSession,
      messages: s.currentSession?.id === id ? [] : s.messages,
    }));
  },

  sendMessage: (content) => {
    const { ws, currentSession } = get();
    if (!ws || !currentSession) return;

    const userMsg: Message = {
      id: crypto.randomUUID(),
      session_id: currentSession.id,
      role: "user",
      content,
      created_at: new Date().toISOString(),
    };

    set((s) => ({
      messages: [...s.messages, userMsg],
      isStreaming: true,
      streamContent: "",
      activeToolCall: null,
    }));

    ws.send(
      JSON.stringify({
        type: "chat",
        session_id: currentSession.id,
        content,
      })
    );
  },

  connectWS: (token) => {
    const existing = get().ws;
    if (existing && existing.readyState === WebSocket.OPEN) return;

    const ws = createWebSocket(token);

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);

      switch (data.type) {
        case "chunk":
          set((s) => ({ streamContent: s.streamContent + data.payload, activeToolCall: null }));
          break;

        case "tool_call":
          set({ activeToolCall: { name: data.payload.name, status: "calling" } });
          break;

        case "tool_result": {
          const state = get();
          const toolMsg: Message = {
            id: crypto.randomUUID(),
            session_id: state.currentSession?.id || "",
            role: "tool",
            content: data.payload.content,
            tool_call_id: data.payload.tool_call_id,
            created_at: new Date().toISOString(),
          };
          set((s) => ({
            messages: [...s.messages, toolMsg],
            activeToolCall: { name: state.activeToolCall?.name || "tool", status: "done" },
          }));
          break;
        }

        case "done": {
          const state = get();
          const assistantMsg: Message = {
            id: crypto.randomUUID(),
            session_id: state.currentSession?.id || "",
            role: "assistant",
            content: state.streamContent,
            created_at: new Date().toISOString(),
          };
          set((s) => ({
            messages: [...s.messages, assistantMsg],
            isStreaming: false,
            streamContent: "",
            activeToolCall: null,
          }));
          break;
        }

        case "error":
          set({ isStreaming: false, streamContent: "", activeToolCall: null });
          break;
      }
    };

    ws.onclose = () => set({ ws: null });
    set({ ws });
  },

  disconnectWS: () => {
    const { ws } = get();
    if (ws) ws.close();
    set({ ws: null });
  },
}));
