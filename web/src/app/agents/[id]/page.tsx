"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Save } from "lucide-react";
import { AppShell } from "@/components/layout/app-shell";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { api, type Agent } from "@/services/api";

const PROVIDERS = [
  { value: "openai", label: "OpenAI", models: ["gpt-4o", "gpt-4o-mini", "gpt-4.1", "o4-mini"] },
  { value: "anthropic", label: "Anthropic", models: ["claude-sonnet-4-20250514", "claude-haiku-4-5-20251001", "claude-opus-4-20250514"] },
];

export default function AgentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const [agent, setAgent] = useState<Agent | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (id) {
      api.agents.get(id).then(setAgent);
    }
  }, [id]);

  if (!agent) {
    return (
      <AppShell>
        <div className="flex h-full items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        </div>
      </AppShell>
    );
  }

  const provider = PROVIDERS.find((p) => p.value === agent.provider);

  const handleSave = async () => {
    setSaving(true);
    try {
      const updated = await api.agents.update(agent.id, {
        name: agent.name,
        description: agent.description,
        system_prompt: agent.system_prompt,
        model: agent.model,
        provider: agent.provider,
        temperature: agent.temperature,
        max_tokens: agent.max_tokens,
        persona: agent.persona,
        memory: agent.memory,
        is_active: agent.is_active,
      });
      setAgent(updated);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } finally {
      setSaving(false);
    }
  };

  return (
    <AppShell>
      <div className="p-6 space-y-6 max-w-4xl">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push("/agents")}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div className="flex-1">
            <h1 className="text-2xl font-bold">{agent.name}</h1>
            <p className="text-sm text-muted-foreground">Agent Configuration</p>
          </div>
          <div className="flex items-center gap-2">
            <Label htmlFor="active" className="text-sm">Active</Label>
            <Switch
              id="active"
              checked={agent.is_active}
              onCheckedChange={(v) => setAgent({ ...agent, is_active: v })}
            />
          </div>
          <Button onClick={handleSave} disabled={saving}>
            <Save className="mr-2 h-4 w-4" />
            {saved ? "Saved!" : saving ? "Saving..." : "Save"}
          </Button>
        </div>

        <Tabs defaultValue="general">
          <TabsList>
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="prompt">Prompt</TabsTrigger>
            <TabsTrigger value="persona">Persona & Memory</TabsTrigger>
          </TabsList>

          <TabsContent value="general" className="mt-4 space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Basic Info</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label>Name</Label>
                  <Input
                    value={agent.name}
                    onChange={(e) => setAgent({ ...agent, name: e.target.value })}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Description</Label>
                  <Input
                    value={agent.description || ""}
                    onChange={(e) => setAgent({ ...agent, description: e.target.value })}
                    placeholder="What does this agent do?"
                  />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">Model Settings</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Provider</Label>
                    <Select
                      value={agent.provider}
                      onValueChange={(v) => {
                        if (!v) return;
                        const prov = PROVIDERS.find((p) => p.value === v);
                        setAgent({ ...agent, provider: v, model: prov ? prov.models[0] : "" });
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {PROVIDERS.map((p) => (
                          <SelectItem key={p.value} value={p.value}>
                            {p.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label>Model</Label>
                    <Select
                      value={agent.model}
                      onValueChange={(v) => v && setAgent({ ...agent, model: v })}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {provider?.models.map((m) => (
                          <SelectItem key={m} value={m}>{m}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Temperature: {agent.temperature}</Label>
                    <input
                      type="range"
                      min="0"
                      max="2"
                      step="0.1"
                      value={agent.temperature}
                      onChange={(e) =>
                        setAgent({ ...agent, temperature: parseFloat(e.target.value) })
                      }
                      className="w-full accent-primary"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Max Tokens</Label>
                    <Input
                      type="number"
                      value={agent.max_tokens}
                      onChange={(e) =>
                        setAgent({ ...agent, max_tokens: parseInt(e.target.value) || 4096 })
                      }
                    />
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="prompt" className="mt-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">System Prompt</CardTitle>
              </CardHeader>
              <CardContent>
                <Textarea
                  value={agent.system_prompt}
                  onChange={(e) => setAgent({ ...agent, system_prompt: e.target.value })}
                  rows={12}
                  className="font-mono text-sm"
                  placeholder="You are a helpful assistant..."
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="persona" className="mt-4 space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Persona (soul.md)</CardTitle>
              </CardHeader>
              <CardContent>
                <Textarea
                  value={agent.persona || ""}
                  onChange={(e) => setAgent({ ...agent, persona: e.target.value })}
                  rows={6}
                  className="font-mono text-sm"
                  placeholder="Define the agent's personality, tone, and character..."
                />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">Memory (memory.md)</CardTitle>
              </CardHeader>
              <CardContent>
                <Textarea
                  value={agent.memory || ""}
                  onChange={(e) => setAgent({ ...agent, memory: e.target.value })}
                  rows={6}
                  className="font-mono text-sm"
                  placeholder="Long-term memory and context the agent should always remember..."
                />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </AppShell>
  );
}
