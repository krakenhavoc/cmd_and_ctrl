// @vitest-environment jsdom
//
// phaseDisplay.extraTurn.render.test.ts — ADR 0059 Decision 11 (#753).
// PhaseDisplay is the action dock's header since ADR 0111 PR 2; it is
// mounted on its own here because the turn line is all this tests.
// "T{n}" stays the round on an extra turn (owner decision 1), so the
// extra turn is MARKED instead, and a queued extra turn names who takes
// it next.

import { describe, it, expect, afterEach } from "vitest";

import PhaseDisplay from "./components/board/PhaseDisplay.svelte";
import type { PlayerView, TurnView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seats = [
  { id: "a", name: "Alice", seat: 0 },
  { id: "b", name: "Bob", seat: 1 },
] as unknown as PlayerView[];

function mount(turn: TurnView) {
  return render(
    PhaseDisplay as never,
    {
      turn,
      seats,
      mulligansOpen: false,
    } as never,
  ).container;
}

const base: TurnView = {
  seq: 5,
  number: 3,
  active_seat: 0,
  priority_holder: 0,
  phase: "precombat_main",
  step: "precombat_main",
};

describe("PhaseDisplay extra turns", () => {
  it("shows no mark on a normal turn", () => {
    const c = mount(base);
    expect(c.querySelector(".turn-no")?.textContent).toBe("T3");
    expect(c.querySelector(".extra-turn")).toBeNull();
    expect(c.querySelector(".next-extra")).toBeNull();
  });

  it("marks an extra turn and keeps the round number", () => {
    const c = mount({ ...base, seq: 6, extra: true });
    expect(c.querySelector(".turn-no")?.textContent).toBe("T3");
    expect(c.querySelector(".extra-turn")?.textContent).toBe("Extra turn");
  });

  it("names the player who takes the next queued extra turn", () => {
    const c = mount({ ...base, extra_turns: [1, 0] });
    expect(c.querySelector(".next-extra")?.textContent?.trim()).toBe("Next: Bob (extra)");
  });
});
