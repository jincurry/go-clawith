"use client";

import { MessageSquare, Plus, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import type { ChatSession } from "@/services/api";

interface Props {
  sessions: ChatSession[];
  currentId?: string;
  onSelect: (session: ChatSession) => void;
  onDelete: (id: string) => void;
  onNew: () => void;
}

export function SessionList({ sessions, currentId, onSelect, onDelete, onNew }: Props) {
  return (
    <div className="flex h-full w-64 flex-col border-r bg-muted/20">
      <div className="flex items-center justify-between p-3">
        <h2 className="text-sm font-semibold">Conversations</h2>
        <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onNew}>
          <Plus className="h-4 w-4" />
        </Button>
      </div>

      <ScrollArea className="flex-1">
        <div className="space-y-0.5 p-2">
          {sessions.length === 0 && (
            <p className="px-3 py-8 text-center text-xs text-muted-foreground">
              No conversations yet
            </p>
          )}
          {sessions.map((session) => (
            <div
              key={session.id}
              className={cn(
                "group flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2 text-sm transition-colors hover:bg-muted",
                currentId === session.id && "bg-muted font-medium"
              )}
              onClick={() => onSelect(session)}
            >
              <MessageSquare className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="flex-1 truncate">{session.title}</span>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 opacity-0 group-hover:opacity-100"
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete(session.id);
                }}
              >
                <Trash2 className="h-3 w-3" />
              </Button>
            </div>
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}
