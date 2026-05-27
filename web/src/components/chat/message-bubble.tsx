"use client";

import { Bot, User, Wrench } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { cn } from "@/lib/utils";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import type { Message } from "@/services/api";

function MarkdownContent({ content, className }: { content: string; className?: string }) {
  return (
    <div className={cn("prose prose-sm dark:prose-invert max-w-none break-words", className)}>
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        pre: ({ children }) => (
          <pre className="overflow-x-auto rounded-lg bg-background/50 p-3 text-xs">
            {children}
          </pre>
        ),
        code: ({ children, className: codeClassName }) => {
          const isInline = !codeClassName;
          if (isInline) {
            return (
              <code className="rounded bg-background/50 px-1.5 py-0.5 text-xs font-mono">
                {children}
              </code>
            );
          }
          return <code className={cn("text-xs font-mono", codeClassName)}>{children}</code>;
        },
        table: ({ children }) => (
          <div className="overflow-x-auto">
            <table className="text-xs">{children}</table>
          </div>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
    </div>
  );
}

export function MessageBubble({ message }: { message: Message }) {
  const isUser = message.role === "user";
  const isTool = message.role === "tool";

  if (isTool) {
    return (
      <div className="flex gap-3 px-4 py-2">
        <div className="ml-11 flex items-start gap-2 rounded-lg border border-dashed px-3 py-2 text-xs text-muted-foreground">
          <Wrench className="mt-0.5 h-3 w-3 shrink-0" />
          <div className="min-w-0">
            {message.tool_name && (
              <Badge variant="outline" className="mb-1 text-[10px]">
                {message.tool_name}
              </Badge>
            )}
            <div className="line-clamp-3 whitespace-pre-wrap">{message.content}</div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={cn("flex gap-3 px-4 py-3", isUser && "flex-row-reverse")}>
      <Avatar className="h-8 w-8 shrink-0">
        <AvatarFallback className={cn(isUser ? "bg-primary text-primary-foreground" : "bg-muted")}>
          {isUser ? <User className="h-4 w-4" /> : <Bot className="h-4 w-4" />}
        </AvatarFallback>
      </Avatar>

      <div
        className={cn(
          "max-w-[70%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed",
          isUser
            ? "bg-primary text-primary-foreground rounded-br-md"
            : "bg-muted rounded-bl-md"
        )}
      >
        {isUser ? (
          <div className="whitespace-pre-wrap break-words">{message.content}</div>
        ) : (
          <MarkdownContent content={message.content} />
        )}
        <div
          className={cn(
            "mt-1 text-[10px] opacity-50",
            isUser ? "text-right" : "text-left"
          )}
        >
          {new Date(message.created_at).toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          })}
        </div>
      </div>
    </div>
  );
}

export function StreamingBubble({ content }: { content: string }) {
  if (!content) return null;

  return (
    <div className="flex gap-3 px-4 py-3">
      <Avatar className="h-8 w-8 shrink-0">
        <AvatarFallback className="bg-muted">
          <Bot className="h-4 w-4" />
        </AvatarFallback>
      </Avatar>
      <div className="max-w-[70%] rounded-2xl rounded-bl-md bg-muted px-4 py-2.5 text-sm leading-relaxed">
        <MarkdownContent content={content} />
        <span className="inline-block h-4 w-1 animate-pulse bg-foreground/70 ml-0.5" />
      </div>
    </div>
  );
}

export function ToolCallBubble({ name, status }: { name: string; status?: string }) {
  return (
    <div className="flex gap-3 px-4 py-1.5">
      <div className="ml-11 flex items-center gap-2 text-xs text-muted-foreground">
        <Wrench className="h-3 w-3 animate-spin" />
        <span>Calling <strong>{name}</strong>...</span>
        {status && <span className="text-[10px]">({status})</span>}
      </div>
    </div>
  );
}
