"use client";

import { useEffect, useState } from "react";
import { Plus, Trash2, Wrench, Code, Globe } from "lucide-react";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { api, type Tool, type CreateToolInput } from "@/services/api";

const typeIcons: Record<string, React.ReactNode> = {
  builtin: <Wrench className="h-4 w-4" />,
  mcp: <Globe className="h-4 w-4" />,
  custom: <Code className="h-4 w-4" />,
};

export default function ToolsPage() {
  const [tools, setTools] = useState<Tool[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [form, setForm] = useState<CreateToolInput>({
    name: "",
    display_name: "",
    description: "",
    type: "custom",
    schema: "",
    endpoint: "",
  });

  const loadTools = async () => {
    setLoading(true);
    try {
      const res = await api.tools.list();
      setTools(res.data || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTools();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    await api.tools.create(form);
    setOpen(false);
    setForm({ name: "", display_name: "", description: "", type: "custom", schema: "", endpoint: "" });
    loadTools();
  };

  const handleDelete = async (id: string) => {
    await api.tools.delete(id);
    loadTools();
  };

  return (
    <AppShell>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Tools</h1>
            <p className="text-muted-foreground">Manage tools available to your agents</p>
          </div>

          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button />}>
              <Plus className="mr-2 h-4 w-4" />
              Add Tool
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add Tool</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleCreate} className="space-y-4">
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label>Name</Label>
                    <Input
                      value={form.name}
                      onChange={(e) => setForm({ ...form, name: e.target.value })}
                      placeholder="web_search"
                      required
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Display Name</Label>
                    <Input
                      value={form.display_name || ""}
                      onChange={(e) => setForm({ ...form, display_name: e.target.value })}
                      placeholder="Web Search"
                    />
                  </div>
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
                      <SelectItem value="builtin">Built-in</SelectItem>
                      <SelectItem value="mcp">MCP</SelectItem>
                      <SelectItem value="custom">Custom</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>Description</Label>
                  <Input
                    value={form.description || ""}
                    onChange={(e) => setForm({ ...form, description: e.target.value })}
                    placeholder="Search the web for information"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Endpoint</Label>
                  <Input
                    value={form.endpoint || ""}
                    onChange={(e) => setForm({ ...form, endpoint: e.target.value })}
                    placeholder="https://api.example.com/tool"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Schema (JSON)</Label>
                  <Textarea
                    value={form.schema || ""}
                    onChange={(e) => setForm({ ...form, schema: e.target.value })}
                    rows={4}
                    className="font-mono text-sm"
                    placeholder='{"type": "object", "properties": {...}}'
                  />
                </div>
                <div className="flex justify-end gap-2">
                  <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={!form.name}>
                    Add
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
        ) : tools.length === 0 ? (
          <div className="flex flex-col items-center py-16 text-center">
            <Wrench className="h-10 w-10 text-muted-foreground/50" />
            <p className="mt-4 text-lg font-medium">No tools yet</p>
            <p className="mt-1 text-sm text-muted-foreground">
              Add tools to give your agents extra capabilities
            </p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {tools.map((tool) => (
              <Card key={tool.id} className="group">
                <CardHeader className="flex flex-row items-start gap-3 pb-2">
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
                    {typeIcons[tool.type] || <Wrench className="h-4 w-4" />}
                  </div>
                  <div className="flex-1 min-w-0">
                    <CardTitle className="text-sm">{tool.display_name || tool.name}</CardTitle>
                    <Badge variant="outline" className="mt-1 text-[10px]">
                      {tool.type}
                    </Badge>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 opacity-0 group-hover:opacity-100"
                    onClick={() => handleDelete(tool.id)}
                  >
                    <Trash2 className="h-3.5 w-3.5 text-destructive" />
                  </Button>
                </CardHeader>
                <CardContent className="pt-0">
                  <p className="text-xs text-muted-foreground line-clamp-2">
                    {tool.description || "No description"}
                  </p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </AppShell>
  );
}
