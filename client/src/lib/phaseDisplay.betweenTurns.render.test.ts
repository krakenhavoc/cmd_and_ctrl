// @vitest-environment jsdom
//
// phaseDisplay.betweenTurns.render.test.ts — ADR 0188 Decision 12
// (#3562). While the turn change waits on the which-skip question the
// old turn is over and the next has not begun, so the step label reads
// "Choosing the next turn" instead of the cleanup step's.

import { describe, it, expect, afterEach } from "vitest";

import PhaseDisplay from "./components/board/PhaseDisplay.svelte";
import { L } from "./labels";
import type { PlayerView, TurnView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const seats = [
  { id: "a", name: "Alice", seat: 0 },
  { id: "b", name: "Bob", seat: 1 },
] as unknown as PlayerView[];

function mount(turn: TurnView) {
  return render(PhaseDisplay as never, { turn, seats, mulligansOpen: false } as never).container;
}

const base: TurnView = {
  seq: 5,
  number: 3,
  active_seat: 0,
  priority_holder: -1,
  phase: "ending",
  step: "cleanup",
} as unknown as TurnView;

describe("the step label between turns", () => {
  it("reads Choosing the next turn while between_turns is set", () => {
    const c = mount({ ...base, between_turns: true });
    const label = c.querySelector(".step-label")!;
    expect(label.textContent).toContain(L.choosingNextTurn);
    expect(L.choosingNextTurn).toBe("Choosing the next turn");
    expect(label.textContent).not.toContain("Cleanup");
    // Nobody holds priority, but the cleanup step's explanation does not
    // apply to a turn change that has not begun.
    expect(label.querySelector(".no-priority")).toBeNull();
  });

  it("is the step's own name otherwise", () => {
    for (const between_turns of [undefined, false]) {
      const c = mount({ ...base, between_turns });
      const label = c.querySelector(".step-label")!;
      expect(label.textContent).toContain("Cleanup");
      expect(label.textContent).not.toContain(L.choosingNextTurn);
      expect(label.querySelector(".no-priority")).not.toBeNull();
    }
  });
});
