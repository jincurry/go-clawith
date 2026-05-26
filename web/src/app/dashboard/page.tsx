"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Bot, MessageSquare, Wrench, Zap, Plus } from "lucide-react";
import { AppShell } from "@/components/layout/app-shell";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { api, type Agent, type ChatSession } from "@/services/api";

export default function DashboardPage() {
  const router = useRouter();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [toolCount, setToolCount] = useState(0);

  useEffect(() => {
    api.agents.list().then((r) => setAgents(r.data || []));
    api.chat.listSessions().then((r) => setSessions(r.data || []));
    api.tools.list().then((r) => setToolCount(r.total || 0));
  }, []);

  const stats = [
    { label: "Agents", value: agents.length, icon: Bot, color: "text-blue-500" },
    { label: "Conversations", value: sessions.length, icon: MessageSquare, color: "text-green-500" },
    { label: "Tools", value: toolCount, icon: Wrench, color: "text-orange-500" },
    { label: "Active Agents", value: agents.filter((a) => a.is_active).length, icon: Zap, color: "text-purple-500" },
  ];

  return (
    <AppShell>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Dashboard</h1>
            <p className="text-muted-foreground">Overview of your AI agents</p>
          </div>
          <Button onClick={() => router.push("/agents")}>
            <Plus className="mr-2 h-4 w-4" />
            New Agent
          </Button>
        </div>

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {stats.map((stat) => (
            <Card key={stat.label}>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {stat.label}
                </CardTitle>
                <stat.icon className={`h-4 w-4 ${stat.color}`} />
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold">{stat.value}</div>
              </CardContent>
            </Card>
          ))}
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Recent Agents</CardTitle>
            </CardHeader>
            <CardContent>
              {agents.length === 0 ? (
                <p className="text-sm text-muted-foreground py-4 text-center">
                  No agents yet. Create one to get started.
                </p>
              ) : (
                <div className="space-y-3">
                  {agents.slice(0, 5).map((agent) => (
                    <div
                      key={agent.id}
                      className="flex items-center gap-3 rounded-lg border p-3 cursor-pointer hover:bg-muted/50 transition-colors"
                      onClick={() => router.push(`/agents/${agent.id}`)}
                    >
                      <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/10">
                        <Bot className="h-4 w-4 text-primary" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-sm truncate">{agent.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {agent.provider} / {agent.model}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Recent Conversations</CardTitle>
            </CardHeader>
            <CardContent>
              {sessions.length === 0 ? (
                <p className="text-sm text-muted-foreground py-4 text-center">
                  No conversations yet. Start chatting with an agent.
                </p>
              ) : (
                <div className="space-y-3">
                  {sessions.slice(0, 5).map((session) => (
                    <div
                      key={session.id}
                      className="flex items-center gap-3 rounded-lg border p-3 cursor-pointer hover:bg-muted/50 transition-colors"
                      onClick={() => router.push("/chat")}
                    >
                      <MessageSquare className="h-4 w-4 text-muted-foreground" />
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-sm truncate">{session.title}</p>
                        <p className="text-xs text-muted-foreground">
                          {new Date(session.updated_at).toLocaleDateString()}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </AppShell>
  );
}
