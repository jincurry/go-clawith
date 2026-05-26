"use client";

import Link from "next/link";
import { Bot, Settings, Trash2 } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import type { Agent } from "@/services/api";

interface Props {
  agent: Agent;
  onDelete: (id: string) => void;
  onChat: (agent: Agent) => void;
}

export function AgentCard({ agent, onDelete, onChat }: Props) {
  return (
    <Card className="group relative transition-shadow hover:shadow-md">
      <CardHeader className="flex flex-row items-start gap-3 pb-2">
        <Avatar className="h-10 w-10">
          <AvatarFallback className="bg-primary/10 text-primary">
            <Bot className="h-5 w-5" />
          </AvatarFallback>
        </Avatar>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold truncate">{agent.name}</h3>
            <Badge variant={agent.is_active ? "default" : "secondary"} className="text-[10px]">
              {agent.is_active ? "Active" : "Inactive"}
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground truncate mt-0.5">
            {agent.provider} / {agent.model}
          </p>
        </div>
      </CardHeader>

      <CardContent className="pt-0">
        <p className="text-sm text-muted-foreground line-clamp-2 min-h-[2.5rem]">
          {agent.description || "No description"}
        </p>

        <div className="mt-3 flex items-center gap-2">
          <Button size="sm" className="flex-1" onClick={() => onChat(agent)}>
            Chat
          </Button>
          <Link href={`/agents/${agent.id}`}>
            <Button variant="outline" size="icon" className="h-8 w-8">
              <Settings className="h-3.5 w-3.5" />
            </Button>
          </Link>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8 text-destructive hover:bg-destructive hover:text-destructive-foreground"
            onClick={() => onDelete(agent.id)}
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
