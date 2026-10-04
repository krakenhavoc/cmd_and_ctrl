// @vitest-environment jsdom
//
// gameLogPanel.render.test.ts — ADR 0119 §5 (S60, #2204). The log's
// `trigger` and `activate` entries carry a decorative mark beside the
// server's sentence; every other entry renders as it did.

import { afterEach, describe, expect, it } from "vitest";

import GameLogPanel from "./components/board/GameLogPanel.svelte";
import type { GameView, LogEvent, PlayerView } from "./protocol";
import { cleanup, render } from "./test/render.svelte";

afterEach(() => cleanup());

function line(seq: number, kind: LogEvent["kind"], text: string): LogEvent {
  return { seq, kind, seat: 0, text } as LogEvent;
}

function view(log: LogEvent[]): GameView {
  const seat = { id: "p1", name: "Alice" } as PlayerView;
  return { seats: [seat], log } as unknown as GameView;
}

describe("GameLogPanel", () => {
  it("marks a trigger and an activation, and no other line", () => {
    const { container } = render(GameLogPanel, {
      view: view([
        line(1, "step", "Turn 1 — Alice · precombat main"),
        line(2, "cast", "Alice cast Mulldrifter"),
        line(3, "trigger", "Alice's trigger: Mulldrifter — draw two cards"),
        line(4, "activate", "Alice activated Prodigal Sorcerer — {T}: deal 1 damage to any target"),
        line(5, "resolve", "Mulldrifter — draw two cards resolved"),
      ]),
      viewerID: "p1",
      onClose: () => {},
    });
    const rows = [...container.querySelectorAll("li.log-entry")];
    const marked = rows.filter((r) => r.querySelector(".log-icon"));
    expect(marked.map((r) => r.querySelector(".log-text")?.textContent)).toEqual([
      "Alice's trigger: Mulldrifter — draw two cards",
      "Alice activated Prodigal Sorcerer — {T}: deal 1 damage to any target",
    ]);
    for (const r of marked) {
      expect(r.querySelector(".log-icon")?.getAttribute("aria-hidden")).toBe("true");
      expect(r.classList.contains("tone-cast")).toBe(true);
    }
  });
});
