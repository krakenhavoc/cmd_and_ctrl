// @vitest-environment jsdom
//
// phaseDisplay.readOnly.render.test.ts — #3378. A spectator's copy of
// the dock header, in the top bar: the same turn line and track, but
// no step can be pinned (a spectator holds no priority to stop at), a
// pin the viewer's own store holds is not drawn, and every step's name
// is just its label.

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import PhaseDisplay from "./components/board/PhaseDisplay.svelte";
import { _resetForTests as resetStops, hasManualStop, toggleManualStop } from "./priorityStops";
import type { PlayerView, TurnView } from "./protocol";
import { STEP_IDS, STEP_LABELS, type StepID } from "./turn";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

beforeEach(resetStops);
afterEach(() => {
  cleanup();
  resetStops();
});

const seats = [
  { id: "a", name: "Alice", seat: 0 },
  { id: "b", name: "Bob", seat: 1 },
] as unknown as PlayerView[];

const turn: TurnView = {
  seq: 5,
  number: 3,
  active_seat: 1,
  priority_holder: 1,
  phase: "declare_attackers",
  step: "declare_attackers",
};

function mount() {
  return render(
    PhaseDisplay as never,
    {
      turn,
      seats,
      mulligansOpen: false,
      readOnly: true,
      inline: true,
    } as never,
  ).container;
}

const button = (c: Element, id: StepID) =>
  c.querySelector<HTMLButtonElement>(`button.step-icon[data-step="${id}"]`)!;

describe("a read-only phase display", () => {
  it("says the round, whose turn and which step", () => {
    const c = mount();
    expect(c.querySelector(".turn-no")?.textContent).toBe("R3");
    expect(c.querySelector(".active-name")?.textContent).toBe("Bob");
    expect(c.querySelector(".step-label")?.textContent).toContain(STEP_LABELS.declare_attackers);
    expect(button(c, "declare_attackers").getAttribute("aria-current")).toBe("step");
    expect(c.querySelector(".phase-display")?.classList.contains("inline")).toBe(true);
  });

  it("pins nothing and takes no clicks", () => {
    const c = mount();
    for (const id of STEP_IDS) {
      const b = button(c, id);
      expect(b.disabled).toBe(true);
      expect(b.classList.contains("clickable")).toBe(false);
      expect(b.getAttribute("aria-pressed")).toBeNull();
      expect(b.title).toBe(STEP_LABELS[id]);
    }
    click(button(c, "upkeep"));
    flushSync();
    expect(hasManualStop("upkeep")).toBe(false);
  });

  it("does not draw a pin from the viewer's store", () => {
    toggleManualStop("end");
    const c = mount();
    expect(button(c, "end").classList.contains("pinned")).toBe(false);
  });
});
