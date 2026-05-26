"use client";

import { useEffect, useRef, useState } from "react";
import { Bot } from "lucide-react";
import { AppShell } from "@/components/layout/app-shell";
import { SessionList } from "@/components/chat/session-list";
import { MessageBubble, StreamingBubble } from "@/components/chat/message-bubble";
import { MessageInput } from "@/components/chat/message-input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useChat } from "@/stores/chat";
import { useAuth } from "@/stores/auth";
import { api, type Agent } from "@/services/api";

export default function ChatPage() {
  const { token } = useAuth();
  const {
    sessions,
    currentSession,
    messages,
    isStreaming,
    streamContent,
    loadSessions,
    selectSession,
    createSession,
    deleteSession,
    sendMessage,
    connectWS,
  } = useChat();

  const [agents, setAgents] = useState<Agent[]>([]);
  const [showPicker, setShowPicker] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadSessions();
    api.agents.list().then((r) => setAgents(r.data || []));
  }, [loadSessions]);

  useEffect(() => {
    if (token) connectWS(token);
  }, [token, connectWS]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, streamContent]);

  const handleNewChat = () => {
    if (agents.length === 1) {
      createSession(agents[0].id, `Chat with ${agents[0].name}`);
    } else {
      setShowPicker(true);
    }
  };

  const handlePickAgent = (agent: Agent) => {
    setShowPicker(false);
    createSession(agent.id, `Chat with ${agent.name}`);
  };

  return (
    <AppShell>
      <div className="flex h-full">
        <SessionList
          sessions={sessions}
          currentId={currentSession?.id}
          onSelect={selectSession}
          onDelete={deleteSession}
          onNew={handleNewChat}
        />

        <div className="flex flex-1 flex-col">
          {currentSession ? (
            <>
              <div className="flex h-14 items-center border-b px-4">
                <Bot className="mr-2 h-5 w-5 text-muted-foreground" />
                <h2 className="font-semibold">{currentSession.title}</h2>
              </div>

              <ScrollArea className="flex-1">
                <div className="mx-auto max-w-3xl py-4">
                  {messages.map((msg) => (
                    <MessageBubble key={msg.id} message={msg} />
                  ))}
                  {isStreaming && <StreamingBubble content={streamContent} />}
                  <div ref={messagesEndRef} />
                </div>
              </ScrollArea>

              <MessageInput onSend={sendMessage} disabled={isStreaming} />
            </>
          ) : (
            <div className="flex flex-1 flex-col items-center justify-center text-center">
              <Bot className="h-12 w-12 text-muted-foreground/50" />
              <h2 className="mt-4 text-lg font-semibold">Start a conversation</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Select a conversation or create a new one
              </p>
              <Button className="mt-4" onClick={handleNewChat}>
                New Chat
              </Button>
            </div>
          )}
        </div>
      </div>

      <Dialog open={showPicker} onOpenChange={setShowPicker}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Choose an Agent</DialogTitle>
          </DialogHeader>
          <div className="space-y-2">
            {agents.map((agent) => (
              <div
                key={agent.id}
                className="flex cursor-pointer items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-muted"
                onClick={() => handlePickAgent(agent)}
              >
                <div className="flex h-9 w-9 items-center justify-center rounded-full bg-primary/10">
                  <Bot className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <p className="font-medium text-sm">{agent.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {agent.description || `${agent.provider} / ${agent.model}`}
                  </p>
                </div>
              </div>
            ))}
            {agents.length === 0 && (
              <p className="py-4 text-center text-sm text-muted-foreground">
                No agents available. Create one first.
              </p>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </AppShell>
  );
}
