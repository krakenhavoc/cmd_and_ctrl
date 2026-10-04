// @vitest-environment jsdom
//
// phaseDisplay.track.render.test.ts — #2214 (S60). The dock's phase
// track is CR 500.1's five phases, each a labelled group with a tiny
// caption under its steps. What must not move: every step still has a
// button whose name is its title ("Upkeep — click to pin a one-time
// stop"), clicking one still toggles a manual stop, the current step
// is still `.current` with aria-current="step", a pinned one is still
// `.pinned`, and `turn and phase indicator` / `#dock-phase-track` are
// where they were (AGENTS.md §5: labels are a contract).

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import PhaseDisplay from "./components/board/PhaseDisplay.svelte";
import { _resetForTests as resetStops, hasManualStop } from "./priorityStops";
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

const turn = (step: StepID): TurnView => ({
  seq: 5,
  number: 3,
  active_seat: 0,
  priority_holder: 0,
  phase: step,
  step,
});

function mount(step: StepID = "precombat_main") {
  return render(PhaseDisplay as never, { turn: turn(step), seats, mulligansOpen: false } as never)
    .container;
}

const button = (c: Element, id: StepID) =>
  c.querySelector<HTMLButtonElement>(`button.step-icon[data-step="${id}"]`)!;

// The five phases, their accessible names, captions and steps.
const PHASES: Array<[string, string, StepID[]]> = [
  ["beginning phase", "Begin", ["untap", "upkeep", "draw"]],
  ["precombat main phase", "Main 1", ["precombat_main"]],
  [
    "combat phase",
    "Combat",
    [
      "begin_combat",
      "declare_attackers",
      "declare_blockers",
      "first_strike_damage",
      "combat_damage",
      "end_combat",
    ],
  ],
  ["postcombat main phase", "Main 2", ["postcombat_main"]],
  ["ending phase", "End", ["end", "cleanup"]],
];

describe("the phase track's groups", () => {
  it("keeps the names the e2e suite and the dock's toggle read", () => {
    const c = mount();
    expect(c.querySelector('[aria-label="turn and phase indicator"]')).not.toBeNull();
    const track = c.querySelector("#dock-phase-track")!;
    expect(track.getAttribute("aria-label")).toBe("phase track");
  });

  it("is five labelled groups, in turn order, each with its caption", () => {
    const groups = [...mount().querySelectorAll('#dock-phase-track [role="group"]')];
    expect(groups.map((g) => g.getAttribute("aria-label"))).toEqual(PHASES.map(([n]) => n));
    groups.forEach((g, i) => {
      const caption = g.querySelector(".phase-caption")!;
      expect(caption.textContent?.trim()).toBe(PHASES[i][1]);
      // The group's name speaks for it; the caption is the sighted half.
      expect(caption.getAttribute("aria-hidden")).toBe("true");
      const steps = [...g.querySelectorAll("button.step-icon")].map((b) =>
        b.getAttribute("data-step"),
      );
      expect(steps).toEqual(PHASES[i][2]);
    });
  });

  it("puts every step of turn.ts in exactly one group, in order", () => {
    const c = mount();
    const steps = [...c.querySelectorAll("#dock-phase-track button.step-icon")].map((b) =>
      b.getAttribute("data-step"),
    );
    expect(steps).toEqual([...STEP_IDS]);
  });

  it("draws a separator between groups and none at the ends", () => {
    const track = mount().querySelector("#dock-phase-track")!;
    const kids = [...track.children];
    expect(kids.filter((k) => k.classList.contains("track-gap"))).toHaveLength(4);
    expect(kids[0].classList.contains("phase-group")).toBe(true);
    expect(kids[kids.length - 1].classList.contains("phase-group")).toBe(true);
    for (const gap of track.querySelectorAll(".track-gap")) {
      expect(gap.getAttribute("aria-hidden")).toBe("true");
    }
  });

  it("marks the current step's phase", () => {
    const c = mount("first_strike_damage");
    const current = [...c.querySelectorAll(".phase-group.current-phase")];
    expect(current.map((g) => g.getAttribute("aria-label"))).toEqual(["combat phase"]);
  });
});

describe("the phase track's steps", () => {
  for (const id of STEP_IDS) {
    it(`${id} renders its icon and names the step in its title`, () => {
      const b = button(mount(), id);
      expect(b).not.toBeNull();
      const svg = b.querySelector("svg.phase-icon")!;
      expect(svg.getAttribute("data-step")).toBe(id);
      expect(svg.getAttribute("aria-hidden")).toBe("true");
      expect(svg.querySelectorAll("path, circle, rect, polygon").length).toBeGreaterThan(0);
      // The accessible name is the title: unchanged by #2214.
      expect(b.getAttribute("aria-label")).toBeNull();
      expect(b.textContent?.trim()).toBe("");
      expect(b.getAttribute("title")).toBe(
        id === "untap" || id === "cleanup"
          ? `${STEP_LABELS[id]} — no priority`
          : `${STEP_LABELS[id]} — click to pin a one-time stop`,
      );
    });
  }

  it("highlights the current step only", () => {
    const c = mount("declare_blockers");
    const current = [...c.querySelectorAll("button.step-icon.current")];
    expect(current).toEqual([button(c, "declare_blockers")]);
    expect(current[0].getAttribute("aria-current")).toBe("step");
    expect(c.querySelectorAll('[aria-current="step"]')).toHaveLength(1);
  });

  it("pins and unpins a stop on click, with the pinned mark and title", () => {
    const c = mount();
    const b = button(c, "postcombat_main");
    expect(b.classList.contains("pinned")).toBe(false);
    expect(b.getAttribute("aria-pressed")).toBe("false");

    click(b);
    flushSync();
    expect(hasManualStop("postcombat_main")).toBe(true);
    expect(b.classList.contains("pinned")).toBe(true);
    expect(b.getAttribute("aria-pressed")).toBe("true");
    expect(b.getAttribute("title")).toBe("Main 2 — click to unpin");

    click(b);
    flushSync();
    expect(hasManualStop("postcombat_main")).toBe(false);
    expect(b.classList.contains("pinned")).toBe(false);
    expect(b.getAttribute("title")).toBe("Main 2 — click to pin a one-time stop");
  });

  it("does not pin a step with no priority", () => {
    const c = mount();
    for (const id of ["untap", "cleanup"] as const) {
      const b = button(c, id);
      expect(b.disabled).toBe(true);
      expect(b.getAttribute("aria-pressed")).toBeNull();
      click(b);
      flushSync();
      expect(hasManualStop(id)).toBe(false);
      expect(b.classList.contains("pinned")).toBe(false);
    }
  });
});
