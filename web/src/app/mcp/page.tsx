"use client";

import { useEffect, useState } from "react";
import {
  Plus,
  Trash2,
  RefreshCw,
  Server,
  Globe,
  Terminal,
  Wrench,
  CheckCircle,
} from "lucide-react";
import { AppShell } from "@/components/layout/app-shell";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import { api, type MCPServerInfo, type ConnectMCPInput } from "@/services/api";

export default function MCPPage() {
  const [servers, setServers] = useState<MCPServerInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [lastResult, setLastResult] = useState<string | null>(null);
  const [form, setForm] = useState<ConnectMCPInput>({
    name: "",
    transport: "http",
    endpoint: "",
    command: "",
    args: [],
    env: [],
  });

  const loadServers = async () => {
    setLoading(true);
    try {
      const res = await api.mcp.listServers();
      setServers(res.data || []);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadServers();
  }, []);

  const handleConnect = async (e: React.FormEvent) => {
    e.preventDefault();
    setConnecting(true);
    setLastResult(null);
    try {
      const result = await api.mcp.connect(form);
      setLastResult(
        `Connected to "${result.name}" — ${result.tools_synced} tools synced`
      );
      setOpen(false);
      setForm({
        name: "",
        transport: "http",
        endpoint: "",
        command: "",
        args: [],
        env: [],
      });
      loadServers();
    } catch (err: unknown) {
      setLastResult(
        `Failed: ${err instanceof Error ? err.message : "unknown error"}`
      );
    } finally {
      setConnecting(false);
    }
  };

  const handleDisconnect = async (id: string) => {
    await api.mcp.disconnect(id);
    loadServers();
  };

  const handleRefresh = async (id: string) => {
    await api.mcp.refresh(id);
    loadServers();
  };

  return (
    <AppShell>
      <div className="p-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">MCP Servers</h1>
            <p className="text-muted-foreground">
              Connect to Model Context Protocol servers to extend agent capabilities
            </p>
          </div>

          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button />}>
              <Plus className="mr-2 h-4 w-4" />
              Connect Server
            </DialogTrigger>
            <DialogContent className="max-w-lg">
              <DialogHeader>
                <DialogTitle>Connect MCP Server</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleConnect} className="space-y-4">
                <div className="space-y-2">
                  <Label>Server Name</Label>
                  <Input
                    value={form.name}
                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                    placeholder="My MCP Server"
                    required
                  />
                </div>

                <div className="space-y-2">
                  <Label>Transport</Label>
                  <Select
                    value={form.transport}
                    onValueChange={(v) => v && setForm({ ...form, transport: v })}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="http">HTTP (Streamable HTTP)</SelectItem>
                      <SelectItem value="stdio">Stdio (Local Process)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {form.transport === "http" && (
                  <div className="space-y-2">
                    <Label>Endpoint URL</Label>
                    <Input
                      value={form.endpoint || ""}
                      onChange={(e) =>
                        setForm({ ...form, endpoint: e.target.value })
                      }
                      placeholder="http://localhost:3001/mcp"
                      required
                    />
                  </div>
                )}

                {form.transport === "stdio" && (
                  <>
                    <div className="space-y-2">
                      <Label>Command</Label>
                      <Input
                        value={form.command || ""}
                        onChange={(e) =>
                          setForm({ ...form, command: e.target.value })
                        }
                        placeholder="npx"
                        required
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>Arguments (one per line)</Label>
                      <Textarea
                        value={(form.args || []).join("\n")}
                        onChange={(e) =>
                          setForm({
                            ...form,
                            args: e.target.value
                              .split("\n")
                              .filter((a) => a.trim()),
                          })
                        }
                        rows={3}
                        className="font-mono text-sm"
                        placeholder={"-y\n@modelcontextprotocol/server-filesystem\n/tmp"}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>Environment Variables (KEY=VALUE, one per line)</Label>
                      <Textarea
                        value={(form.env || []).join("\n")}
                        onChange={(e) =>
                          setForm({
                            ...form,
                            env: e.target.value
                              .split("\n")
                              .filter((a) => a.trim()),
                          })
                        }
                        rows={2}
                        className="font-mono text-sm"
                        placeholder="API_KEY=xxx"
                      />
                    </div>
                  </>
                )}

                {lastResult && (
                  <div
                    className={`rounded-md px-3 py-2 text-sm ${
                      lastResult.startsWith("Failed")
                        ? "bg-destructive/10 text-destructive"
                        : "bg-green-500/10 text-green-500"
                    }`}
                  >
                    {lastResult}
                  </div>
                )}

                <div className="flex justify-end gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" disabled={connecting || !form.name}>
                    {connecting ? "Connecting..." : "Connect"}
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
        ) : servers.length === 0 ? (
          <div className="flex flex-col items-center py-16 text-center">
            <Server className="h-10 w-10 text-muted-foreground/50" />
            <p className="mt-4 text-lg font-medium">No MCP servers connected</p>
            <p className="mt-1 text-sm text-muted-foreground">
              Connect an MCP server to give your agents access to external tools
            </p>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {servers.map((server) => (
              <Card key={server.id} className="group">
                <CardHeader className="flex flex-row items-start gap-3 pb-2">
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
                    {server.transport === "http" ? (
                      <Globe className="h-4 w-4" />
                    ) : (
                      <Terminal className="h-4 w-4" />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <CardTitle className="text-sm flex items-center gap-2">
                      {server.name}
                      <CheckCircle className="h-3 w-3 text-green-500" />
                    </CardTitle>
                    <div className="mt-1 flex items-center gap-2">
                      <Badge variant="outline" className="text-[10px]">
                        {server.transport}
                      </Badge>
                      <Badge variant="secondary" className="text-[10px]">
                        {server.tools.length} tools
                      </Badge>
                    </div>
                  </div>
                </CardHeader>

                <CardContent className="pt-0 space-y-2">
                  {server.endpoint && (
                    <p className="text-[11px] text-muted-foreground font-mono truncate">
                      {server.endpoint}
                    </p>
                  )}
                  {server.command && (
                    <p className="text-[11px] text-muted-foreground font-mono truncate">
                      {server.command}
                    </p>
                  )}

                  {server.tools.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {server.tools.slice(0, 6).map((tool) => (
                        <Badge
                          key={tool}
                          variant="outline"
                          className="text-[10px] gap-1"
                        >
                          <Wrench className="h-2.5 w-2.5" />
                          {tool}
                        </Badge>
                      ))}
                      {server.tools.length > 6 && (
                        <Badge variant="outline" className="text-[10px]">
                          +{server.tools.length - 6} more
                        </Badge>
                      )}
                    </div>
                  )}

                  <div className="flex gap-2 pt-1">
                    <Button
                      variant="outline"
                      size="sm"
                      className="flex-1"
                      onClick={() => handleRefresh(server.id)}
                    >
                      <RefreshCw className="mr-1 h-3 w-3" />
                      Refresh
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-destructive hover:bg-destructive hover:text-destructive-foreground"
                      onClick={() => handleDisconnect(server.id)}
                    >
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </AppShell>
  );
}
