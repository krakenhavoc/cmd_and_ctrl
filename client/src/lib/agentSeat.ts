// agentSeat: what the table says about a seat played by an outside AI
// agent (ADR 0122 §7). Plain functions over plain data, so the wording
// is testable without mounting anything.
//
// An agent is NOT a bot. A bot is the server's own seat (a policy
// runner in this process); an agent is a separate program, the MCP
// seat binary, joined as a guest. The two never share a chip.

import type { ChatMessage } from "./ws";

/** The two wire fields every viewer gets on PlayerView and SeatInfo. */
export interface AgentFields {
  is_agent?: boolean;
  agent_client?: string;
}

export function isAgentSeat(s: AgentFields | null | undefined): boolean {
  return s?.is_agent === true;
}

// The server already cuts the client to [a-z0-9._-], 32 characters,
// or "unknown". "unknown" and empty say nothing, so they are dropped.
function clientOf(s: AgentFields): string {
  const c = (s.agent_client ?? "").trim();
  return c === "" || c === "unknown" ? "" : c;
}

/** Visible chip text: "AI · claude-code", or just "AI". */
export function agentChipText(s: AgentFields, thinking = false): string {
  if (thinking) return "AI · thinking…";
  const c = clientOf(s);
  return c ? `AI · ${c}` : "AI";
}

/** The chip's accessible name. */
export function agentChipLabel(s: AgentFields, thinking = false): string {
  const c = clientOf(s);
  const who = c ? `AI agent, ${c}` : "AI agent";
  return thinking ? `${who}, thinking` : who;
}

/** The ADR's tooltip. */
export function agentChipTitle(s: AgentFields): string {
  const c = clientOf(s);
  return `Played by an AI agent${c ? ` (${c})` : ""}. It sees only what this seat sees.`;
}

export interface AgentLine {
  id: string;
  authorName: string;
  text: string;
  at: Date;
  seat: AgentFields;
}

export const AGENT_FEED_LIMIT = 3;

/**
 * Plain chat lines typed by an agent seat, newest last. No other chat
 * UI renders a "say" line, and an agent's recourse for an
 * unimplemented card is to say so in chat (ADR 0122 §3), so the table
 * has to be able to read it, labelled.
 */
export function agentLines(
  chat: readonly ChatMessage[],
  seats: readonly (AgentFields & { id: string })[],
  limit: number = AGENT_FEED_LIMIT,
): AgentLine[] {
  const byID = new Map<string, AgentFields>();
  for (const s of seats) if (isAgentSeat(s)) byID.set(s.id, s);
  const out: AgentLine[] = [];
  for (const m of chat) {
    if (m.kind !== "say" || !m.authorID) continue;
    const seat = byID.get(m.authorID);
    if (!seat) continue;
    out.push({ id: m.id, authorName: m.authorName, text: m.text, at: m.at, seat });
  }
  return out.slice(-limit);
}
