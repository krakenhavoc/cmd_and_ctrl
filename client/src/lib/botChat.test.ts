import { describe, it, expect } from "vitest";
import { visibleBotLines, isBotLine, isImprovisation, BOT_FEED_LIMIT } from "./botChat";
import type { ChatMessage } from "./ws";
import type { ChatKind } from "./protocol";

function msg(kind: ChatKind, text: string, reason = "", id = text): ChatMessage {
  return {
    id,
    authorID: "bot-seat",
    authorName: "Kess",
    text,
    at: new Date("2026-09-11T12:00:00Z"),
    kind,
    reason,
  };
}

const human = msg("say", "nice board", "");
const improv = msg("bot_improvisation", "Grim Tutor: lose 3 life (improvised…)", "no catalog spec");
const reasoning = msg("bot_reasoning", "Cast Lightning Bolt", "kills the biggest threat");

describe("isBotLine / isImprovisation", () => {
  it("separates bot lines from people", () => {
    expect(isBotLine(human)).toBe(false);
    expect(isBotLine(improv)).toBe(true);
    expect(isBotLine(reasoning)).toBe(true);
    expect(isImprovisation(improv)).toBe(true);
    expect(isImprovisation(reasoning)).toBe(false);
  });
});

describe("visibleBotLines", () => {
  it("always shows improvisations, whatever the setting says", () => {
    for (const show of [false, true]) {
      const lines = visibleBotLines([human, improv], show);
      expect(lines).toHaveLength(1);
      expect(lines[0].improvised).toBe(true);
      expect(lines[0].text).toContain("Grim Tutor");
    }
  });

  it("hides reasoning unless the viewer asked for it", () => {
    expect(visibleBotLines([reasoning], false)).toHaveLength(0);
    expect(visibleBotLines([reasoning], true)).toHaveLength(1);
  });

  it("attaches an improvisation's reason only behind the setting", () => {
    expect(visibleBotLines([improv], false)[0].reason).toBe("");
    expect(visibleBotLines([improv], true)[0].reason).toBe("no catalog spec");
  });

  it("never surfaces a human's message", () => {
    expect(visibleBotLines([human], true)).toHaveLength(0);
  });

  it("keeps the newest lines when over the cap", () => {
    const many = Array.from({ length: BOT_FEED_LIMIT + 3 }, (_, i) =>
      msg("bot_improvisation", `line ${i}`, "", `id-${i}`),
    );
    const lines = visibleBotLines(many, false);
    expect(lines).toHaveLength(BOT_FEED_LIMIT);
    expect(lines[lines.length - 1].text).toBe(`line ${many.length - 1}`);
  });

  it("gives every line a distinct key", () => {
    const lines = visibleBotLines([improv, reasoning], true);
    expect(new Set(lines.map((l) => l.id)).size).toBe(lines.length);
  });
});
