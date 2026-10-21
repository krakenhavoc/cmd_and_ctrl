// @vitest-environment jsdom
//
// phaseDisplay.attackDirection.render.test.ts — ADR 0206 (#4211). The
// dock's header says which way attacks go while a Pramikon or a Teyo
// −2 is in force: one status line per rule, nothing otherwise.

import { describe, it, expect, afterEach } from "vitest";

import PhaseDisplay from "./components/board/PhaseDisplay.svelte";
import type { AttackDirectionView, PlayerView, TurnView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seats = [
  { id: "a", name: "Alice", seat: 0 },
  { id: "b", name: "Bob", seat: 1 },
  { id: "c", name: "Carol", seat: 2 },
] as unknown as PlayerView[];

const turn: TurnView = {
  seq: 5,
  number: 3,
  active_seat: 0,
  priority_holder: 0,
  phase: "combat",
  step: "declare_attackers",
};

const pramikon: AttackDirectionView = {
  direction: "left",
  source: "Pramikon, Sky Rampart",
  nearest: { a: "b", b: "c", c: "a" },
};

function mount(attackDirections?: AttackDirectionView[], viewerId: string | null = "a") {
  return render(
    PhaseDisplay as never,
    {
      turn,
      seats,
      mulligansOpen: false,
      attackDirections,
      viewerId,
    } as never,
  ).container;
}

describe("PhaseDisplay attack direction", () => {
  it("shows no line while no rule is in force", () => {
    expect(mount().querySelector("[data-attack-direction]")).toBeNull();
    expect(mount([]).querySelector("[data-attack-direction]")).toBeNull();
  });

  it("names the direction, the opponent and the source", () => {
    const row = mount([pramikon]).querySelector("[data-attack-direction]");
    expect(row?.getAttribute("data-attack-direction")).toBe("left");
    expect(row?.getAttribute("role")).toBe("status");
    expect(row?.textContent?.replace(/\s+/g, " ").trim()).toBe(
      "⚔ Attacks go left: you may attack only Bob (Pramikon, Sky Rampart)",
    );
    expect(row?.getAttribute("title")).toMatch(/nearest opponent to the left/);
    // The sword is decoration: the sentence carries the meaning.
    expect(row?.querySelector(".dir-glyph")?.getAttribute("aria-hidden")).toBe("true");
  });

  it("tells a spectator the direction without a target", () => {
    const row = mount([pramikon], null).querySelector("[data-attack-direction]");
    expect(row?.textContent?.replace(/\s+/g, " ").trim()).toBe(
      "⚔ Attacks go left (Pramikon, Sky Rampart)",
    );
  });

  it("flags rules that disagree", () => {
    const rows = mount([
      pramikon,
      { direction: "right", source: "Teyo", nearest: { a: "c", b: "a", c: "b" } },
    ]).querySelectorAll("[data-attack-direction]");
    expect([...rows].map((r) => r.getAttribute("data-attack-direction"))).toEqual([
      "left",
      "right",
      "conflict",
    ]);
  });
});
