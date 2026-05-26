"use client";

import { useEffect, useState } from "react";
import { Plus, Trash2, Zap, Clock, Globe } from "lucide-react";
import { AppShell } from "@/components/layout/app-shell";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { api, type Agent, type Trigger, type CreateTriggerInput } from "@/services/api";

export default function TriggersPage() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [triggers, setTriggers] = useState<Trigger[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState<CreateTriggerInput>({
    agent_id: "",
    name: "",
    type: "cron",
    config: '{"expression": "0 0 9 * * *"}',
    action: "",
    is_active: true,
  });

  const loadData = async () => {
    setLoading(true);
    try {
      const agentRes = await api.agents.list();
      const agentList = agentRes.data || [];
      setAgents(agentList);

      const allTriggers: Trigger[] = [];
      for (const agent of agentList) {
        const res = await api.triggers.listByAgent(agent.id);
        if (res.data) allTriggers.push(...res.data);
      }
      setTriggers(allTriggers);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    await api.triggers.create(form);
    setOpen(false);
    setForm({
      agent_id: "",
      name: "",
      type: "cron",
      config: '{"expression": "0 0 9 * * *"}',
      action: "",
      is_active: true,
    });
    loadData();
  };

  const handleDelete = async (id: string) => {
    await api.triggers.delete(id);
    loadData();
  };

  const getAgentName = (id: string) => agents.find((a) => a.id === id)?.name || "Unknown";

  return (
    <AppShell>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Triggers</h1>
            <p className="text-muted-foreground">Automate agent actions with cron and webhooks</p>
          </div>

          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button />}>
              <Plus className="mr-2 h-4 w-4" />
              New Trigger
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Trigger</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleCreate} className="space-y-4">
                <div className="space-y-2">
                  <Label>Agent</Label>
                  <Select
                    value={form.agent_id}
                    onValueChange={(v) => v && setForm({ ...form, agent_id: v })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select an agent" />
                    </SelectTrigger>
                    <SelectContent>
                      {agents.map((a) => (
                        <SelectItem key={a.id} value={a.id}>
                          {a.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label>Name</Label>
                    <Input
                      value={form.name}
                      onChange={(e) => setForm({ ...form, name: e.target.value })}
                      placeholder="Daily Summary"
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Type</Label>
                    <Select
                      value={form.type}
                      onValueChange={(v) => v && setForm({ ...form, type: v })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="cron">Cron</SelectItem>
                        <SelectItem value="webhook">Webhook</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                {form.type === "cron" && (
                  <div className="space-y-2">
                    <Label>Cron Expression (with seconds)</Label>
                    <Input
                      value={JSON.parse(form.config).expression || ""}
                      onChange={(e) =>
                        setForm({
                          ...form,
                          config: JSON.stringify({ expression: e.target.value }),
                        })
                      }
                      placeholder="0 0 9 * * *"
                      className="font-mono"
                    />
                    <p className="text-xs text-muted-foreground">
                      Format: sec min hour day month weekday (e.g. &quot;0 0 9 * * *&quot; = every day at 9am)
                    </p>
                  </div>
                )}

                <div className="space-y-2">
                  <Label>Action Prompt</Label>
                  <Textarea
                    value={form.action}
                    onChange={(e) => setForm({ ...form, action: e.target.value })}
                    rows={4}
                    placeholder="Summarize the latest news about AI..."
                    required
                  />
                </div>

                <div className="flex items-center gap-2">
                  <Switch
                    checked={form.is_active}
                    onCheckedChange={(v) => setForm({ ...form, is_active: v })}
                  />
                  <Label>Active</Label>
                </div>

                <div className="flex justify-end gap-2">
                  <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={!form.agent_id || !form.name || !form.action}>
                    Create
                  </Button>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {loading ? (
          <div className="flex justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
          </div>
        ) : triggers.length === 0 ? (
          <div className="flex flex-col items-center py-16 text-center">
            <Zap className="h-10 w-10 text-muted-foreground/50" />
            <p className="mt-4 text-lg font-medium">No triggers yet</p>
            <p className="mt-1 text-sm text-muted-foreground">
              Create triggers to automate agent actions
            </p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {triggers.map((trigger) => (
              <Card key={trigger.id} className="group">
                <CardHeader className="flex flex-row items-start gap-3 pb-2">
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
                    {trigger.type === "cron" ? (
                      <Clock className="h-4 w-4" />
                    ) : (
                      <Globe className="h-4 w-4" />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <CardTitle className="text-sm">{trigger.name}</CardTitle>
                    <div className="mt-1 flex items-center gap-2">
                      <Badge variant="outline" className="text-[10px]">
                        {trigger.type}
                      </Badge>
                      <Badge
                        variant={trigger.is_active ? "default" : "secondary"}
                        className="text-[10px]"
                      >
                        {trigger.is_active ? "Active" : "Inactive"}
                      </Badge>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 opacity-0 group-hover:opacity-100"
                    onClick={() => handleDelete(trigger.id)}
                  >
                    <Trash2 className="h-3.5 w-3.5 text-destructive" />
                  </Button>
                </CardHeader>
                <CardContent className="pt-0 space-y-1">
                  <p className="text-xs text-muted-foreground">
                    Agent: {getAgentName(trigger.agent_id)}
                  </p>
                  <p className="text-xs text-muted-foreground line-clamp-2">
                    {trigger.action}
                  </p>
                  {trigger.last_run_at && (
                    <p className="text-[10px] text-muted-foreground">
                      Last run: {new Date(trigger.last_run_at).toLocaleString()}
                    </p>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </AppShell>
  );
}
