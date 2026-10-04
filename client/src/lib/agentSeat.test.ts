import { describe, expect, it } from "vitest";

import {
  agentChipLabel,
  agentChipText,
  agentChipTitle,
  agentLines,
  isAgentSeat,
} from "./agentSeat";
import type { ChatMessage } from "./ws";

const cc = { is_agent: true, agent_client: "claude-code" };

describe("the agent chip's wording", () => {
  it("reads 'AI · <client>' with the ADR's tooltip", () => {
    expect(agentChipText(cc)).toBe("AI · claude-code");
    expect(agentChipTitle(cc)).toBe(
      "Played by an AI agent (claude-code). It sees only what this seat sees.",
    );
    expect(agentChipLabel(cc)).toBe("AI agent, claude-code");
  });
  it("drops an unknown or empty client rather than printing it", () => {
    for (const agent_client of ["unknown", "", undefined]) {
      const s = { is_agent: true, agent_client };
      expect(agentChipText(s)).toBe("AI");
      expect(agentChipLabel(s)).toBe("AI agent");
      expect(agentChipTitle(s)).toBe("Played by an AI agent. It sees only what this seat sees.");
    }
  });
  it("reads thinking… while the seat holds priority", () => {
    expect(agentChipText(cc, true)).toBe("AI · thinking…");
    expect(agentChipLabel(cc, true)).toBe("AI agent, claude-code, thinking");
  });
  it("is only an agent when the wire says so", () => {
    expect(isAgentSeat(cc)).toBe(true);
    expect(isAgentSeat({ agent_client: "x" })).toBe(false);
    expect(isAgentSeat(null)).toBe(false);
  });
});

function say(id: string, authorID: string, kind: ChatMessage["kind"] = "say"): ChatMessage {
  return { id, authorID, authorName: "n", text: id, at: new Date(0), kind, reason: "" };
}

describe("agentLines", () => {
  const seats = [{ id: "a", ...cc }, { id: "h" }, { id: "b" }];
  it("keeps only plain lines typed by an agent seat", () => {
    const chat = [
      say("1", "a"),
      say("2", "h"),
      say("3", "b"),
      say("4", "a", "bot_reasoning"),
      say("5", ""),
    ];
    expect(agentLines(chat, seats).map((l) => l.id)).toEqual(["1"]);
  });
  it("caps at the newest few", () => {
    const chat = ["1", "2", "3", "4", "5"].map((i) => say(i, "a"));
    expect(agentLines(chat, seats).map((l) => l.id)).toEqual(["3", "4", "5"]);
  });
});
