// tutorial.test.ts — the tutorial's step machine (ADR 0076 §2.1–§2.4,
// #1079). Pure: no DOM, fake timers for the hint and the watch timeout.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import {
  HINT_AFTER_MS,
  TUTORIAL_STEP_COUNT,
  anchorsOf,
  copyText,
  createTutorialRun,
  spotlit,
  type CoachState,
  type TutorialStep,
} from "./tutorial";
import { HANDOFF, TUTORIAL_STEPS, WELCOME } from "./tutorialSteps";
import type { GameView } from "./protocol";

beforeEach(() => {
  vi.useFakeTimers();
});
afterEach(() => {
  vi.useRealTimers();
});

const v = (n: number) => ({ id: "g", turn: { number: n } }) as unknown as GameView;
const turnOf = (g: GameView | null) =>
  (g as unknown as { turn: { number: number } } | null)?.turn.number ?? 0;

const action = (over: Partial<TutorialStep> = {}): TutorialStep => ({
  id: "act",
  n: 5,
  kind: "action",
  title: "Tap a land for mana",
  body: "Click one of the Forests.",
  hint: "A land already on the table.",
  anchor: { label: "lands", within: "your board" },
  ...over,
});

const script = (...middle: TutorialStep[]) => [WELCOME, ...middle, HANDOFF];

describe("the skeleton script", () => {
  it("is steps 1 and 11, opening and done, with the canvas's copy", () => {
    expect(TUTORIAL_STEPS.map((s) => [s.n, s.kind])).toEqual([
      [1, "opening"],
      [11, "done"],
    ]);
    expect(TUTORIAL_STEP_COUNT).toBe(11);
    expect(copyText(WELCOME.title, { helpKey: "?", settingsKey: "," })).toBe(
      "A five-minute practice game",
    );
    // Neither button step points at anything: step 1 leaves the board lit.
    expect(anchorsOf(WELCOME)).toEqual([]);
    expect(anchorsOf(HANDOFF)).toEqual([]);
  });

  it("names the player's own keys in the handoff, and copes with none", () => {
    expect(copyText(HANDOFF.body, { helpKey: "?", settingsKey: "," })).toMatch(
      /^Press \? for the keymap and , for settings\. Your zone piles are on the rail/,
    );
    expect(copyText(HANDOFF.body, { helpKey: "", settingsKey: "" })).toMatch(
      /^The keymap is in Settings\. /,
    );
  });
});

describe("createTutorialRun", () => {
  it("opens on step 1 and Start goes to the next step", () => {
    const run = createTutorialRun(TUTORIAL_STEPS, { viewerID: "me" });
    expect(run.current()).toMatchObject({ index: 0, coach: "opening", visible: true });
    run.start();
    expect(run.current()).toMatchObject({ index: 1, coach: "done", visible: true });
    expect(run.current().step.id).toBe("handoff");
  });

  it("closes on Skip tutorial or Finish, and stays closed", () => {
    const run = createTutorialRun(script(action()), { viewerID: "me" });
    run.close();
    expect(run.current().visible).toBe(false);
    run.start();
    run.skipStep();
    run.observe(v(2));
    expect(run.current()).toMatchObject({ index: 0, visible: false });
  });

  it("skips a step silently, and never past the last one", () => {
    const run = createTutorialRun(script(action()), { viewerID: "me" });
    run.start();
    expect(run.current().step.id).toBe("act");
    run.skipStep();
    expect(run.current().step.id).toBe("handoff");
    run.skipStep();
    expect(run.current()).toMatchObject({ step: HANDOFF, visible: true });
  });

  it("advances an action step when its predicate holds, measured from the step's start", () => {
    const step = action({ done: (c) => turnOf(c.view) === turnOf(c.start) + 1 });
    const run = createTutorialRun(script(step), { viewerID: "me", view: v(3) });
    run.start();
    run.observe(v(3));
    expect(run.current().step.id).toBe("act");
    run.observe(v(4));
    expect(run.current().step.id).toBe("handoff");
  });

  it("hands bus events to the predicate", () => {
    const step = action({ done: (c) => c.event === "hand-hovered" });
    const run = createTutorialRun(script(step), { viewerID: "me" });
    run.start();
    run.observe(null, "pile-hovered");
    expect(run.current().step.id).toBe("act");
    run.observe(null, "hand-hovered");
    expect(run.current().step.id).toBe("handoff");
  });

  it("shows the hint only after HINT_AFTER_MS idle", () => {
    const run = createTutorialRun(script(action()), { viewerID: "me" });
    run.start();
    expect(run.current().coach).toBe("action");
    vi.advanceTimersByTime(HINT_AFTER_MS - 1);
    expect(run.current().coach).toBe("action");
    vi.advanceTimersByTime(1);
    expect(run.current().coach).toBe("hint");
  });

  it("has no hint state for a step with no hint", () => {
    const run = createTutorialRun(script(action({ hint: undefined })), { viewerID: "me" });
    run.start();
    vi.advanceTimersByTime(HINT_AFTER_MS * 2);
    expect(run.current().coach).toBe("action");
  });

  it("recovers from the common wrong action without blocking the step", () => {
    const step = action({
      done: (c) => turnOf(c.view) === 9,
      recover: { when: (c) => turnOf(c.view) === 5, title: "Close", body: "That played a land." },
    });
    const run = createTutorialRun(script(step), { viewerID: "me" });
    run.start();
    run.observe(v(5));
    expect(run.current()).toMatchObject({ coach: "recovered", step: { id: "act" } });
    // The hint timer no longer fires over the recovered card.
    vi.advanceTimersByTime(HINT_AFTER_MS);
    expect(run.current().coach).toBe("recovered");
    // The step stays live.
    run.observe(v(9));
    expect(run.current().step.id).toBe("handoff");
  });

  it("advances a step whose anchor is missing, and logs it", () => {
    const log = vi.fn();
    const run = createTutorialRun(script(action()), { viewerID: "me", log });
    run.start();
    run.anchorMissing("act");
    expect(run.current().step.id).toBe("handoff");
    expect(log).toHaveBeenCalledWith(expect.stringMatching(/step act has no anchor.*advancing/));
  });

  it("ignores a missing-anchor report for a step it has already left", () => {
    const log = vi.fn();
    const run = createTutorialRun(script(action(), action({ id: "act2" })), {
      viewerID: "me",
      log,
    });
    run.start();
    run.skipStep();
    run.anchorMissing("act");
    expect(run.current().step.id).toBe("act2");
    expect(log).not.toHaveBeenCalled();
  });

  it("lets a watch step time out on its own, so a stalled bot never wedges it", () => {
    const log = vi.fn();
    const watch: TutorialStep = {
      id: "watch",
      n: 9,
      kind: "watch",
      title: "The bot is taking its turn",
      body: "Anything it casts appears in the strip.",
      timeoutMs: 30_000,
      done: () => false,
    };
    const run = createTutorialRun(script(watch), { viewerID: "me", log });
    run.start();
    expect(run.current().coach).toBe("watch");
    vi.advanceTimersByTime(30_000);
    expect(run.current().step.id).toBe("handoff");
    expect(log).toHaveBeenCalledWith(expect.stringMatching(/watch timed out/));
  });

  it("publishes every change to subscribers", () => {
    const run = createTutorialRun(TUTORIAL_STEPS, { viewerID: "me" });
    const seen: CoachState[] = [];
    const off = run.subscribe((s) => seen.push(s.coach));
    run.start();
    off();
    expect(seen).toEqual(["opening", "done"]);
  });
});

describe("spotlit", () => {
  it("lights only the states that ask the player for something", () => {
    const lit = (["opening", "action", "hint", "watch", "recovered", "done"] as const).filter(
      spotlit,
    );
    expect(lit).toEqual(["action", "hint", "recovered"]);
  });
});
