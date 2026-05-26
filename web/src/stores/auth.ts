import { create } from "zustand";
import { api, type User } from "@/services/api";

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, username: string, password: string) => Promise<void>;
  logout: () => void;
  loadUser: () => Promise<void>;
}

export const useAuth = create<AuthState>((set) => ({
  user: null,
  token: typeof window !== "undefined" ? localStorage.getItem("token") : null,
  isLoading: true,

  login: async (email, password) => {
    const res = await api.auth.login({ email, password });
    localStorage.setItem("token", res.token);
    set({ user: res.user, token: res.token });
  },

  register: async (email, username, password) => {
    const res = await api.auth.register({ email, username, password });
    localStorage.setItem("token", res.token);
    set({ user: res.user, token: res.token });
  },

  logout: () => {
    localStorage.removeItem("token");
    set({ user: null, token: null });
  },

  loadUser: async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        set({ isLoading: false });
        return;
      }
      const user = await api.auth.me();
      set({ user, token, isLoading: false });
    } catch {
      localStorage.removeItem("token");
      set({ user: null, token: null, isLoading: false });
    }
  },
}));
