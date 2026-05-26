"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { AppShell } from "@/components/layout/app-shell";
import { AgentCard } from "@/components/agent/agent-card";
import { CreateAgentDialog } from "@/components/agent/create-agent-dialog";
import { api, type Agent } from "@/services/api";
import { useChat } from "@/stores/chat";

export default function AgentsPage() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const { createSession } = useChat();
  const router = useRouter();

  const loadAgents = async () => {
    setLoading(true);
    try {
      const res = await api.agents.list();
      setAgents(res.data || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAgents();
  }, []);

  const handleDelete = async (id: string) => {
    await api.agents.delete(id);
    loadAgents();
  };

  const handleChat = async (agent: Agent) => {
    await createSession(agent.id, `Chat with ${agent.name}`);
    router.push("/chat");
  };

  return (
    <AppShell>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Agents</h1>
            <p className="text-muted-foreground">Manage your AI agents</p>
          </div>
          <CreateAgentDialog onCreated={loadAgents} />
        </div>

        {loading ? (
          <div className="flex justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          </div>
        ) : agents.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <p className="text-lg font-medium">No agents yet</p>
            <p className="mt-1 text-sm text-muted-foreground">
              Create your first agent to get started
            </p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {agents.map((agent) => (
              <AgentCard
                key={agent.id}
                agent={agent}
                onDelete={handleDelete}
                onChat={handleChat}
              />
            ))}
          </div>
        )}
      </div>
    </AppShell>
  );
}
