// tutorial.test.ts — the tutorial's step machine (ADR 0076 §2.1–§2.4,
// #1079). Pure: no DOM, fake timers for the hint and the watch timeout.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import {
  HINT_AFTER_MS,
  TUTORIAL_STEP_COUNT,
  anchorsOf,
  copyText,
  createTutorialRun,
  spotAnchors,
  spotlit,
  statusText,
  type CoachState,
  type TutorialStep,
} from "./tutorial";
import { HANDOFF, TUTORIAL_STEPS, WELCOME } from "./tutorialSteps";
import { L } from "./labels";
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
  anchor: { label: L.lands, within: L.yourBoard },
  ...over,
});

const script = (...middle: TutorialStep[]) => [WELCOME, ...middle, HANDOFF];

describe("the script's two button steps", () => {
  it("opens with step 1 and closes with step 14, with the canvas's copy", () => {
    expect([TUTORIAL_STEPS[0], TUTORIAL_STEPS.at(-1)]).toEqual([WELCOME, HANDOFF]);
    expect([WELCOME.n, WELCOME.kind, HANDOFF.n, HANDOFF.kind]).toEqual([1, "opening", 14, "done"]);
    expect(TUTORIAL_STEP_COUNT).toBe(14);
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

  it("says where Help is, which replays any of it (ADR 0125 §5.1)", () => {
    expect(copyText(HANDOFF.body, { helpKey: "?", settingsKey: "," })).toMatch(
      /Help, in the ⋯ menu here and in the header elsewhere, replays any of this\.$/,
    );
  });
});

describe("createTutorialRun", () => {
  it("opens on step 1 and Start goes to the next step", () => {
    const run = createTutorialRun(script(), { viewerID: "me" });
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
    const run = createTutorialRun(script(), { viewerID: "me" });
    const seen: CoachState[] = [];
    const off = run.subscribe((s) => seen.push(s.coach));
    run.start();
    off();
    expect(seen).toEqual(["opening", "done"]);
  });
});

// Sub-PR 4 (#1081): what the nine middle steps added to the machine.
describe("detours, cannot, hover and advance", () => {
  const detour = { id: "to-main", title: "First, your main phase", body: "Press next." };

  it("shows a detour while the board needs one, and drops it once it does not", () => {
    const step = action({
      done: (c) => turnOf(c.view) === 9,
      first: (c) => (turnOf(c.view) < 3 ? detour : null),
    });
    const run = createTutorialRun(script(step), { viewerID: "me", view: v(1) });
    const seen: Array<string | null> = [];
    run.subscribe((s) => seen.push(s.detour?.id ?? null));
    run.start();
    expect(run.current()).toMatchObject({ coach: "action", detour: { id: "to-main" } });
    // The same detour again is not republished.
    const before = seen.length;
    run.observe(v(2));
    expect(seen.length).toBe(before);
    run.observe(v(3));
    expect(run.current().detour).toBeNull();
    expect(run.current().step.id).toBe("act");
    run.observe(v(9));
    expect(run.current().step.id).toBe("handoff");
  });

  it("starts each step with no detour", () => {
    const run = createTutorialRun(script(action({ first: () => detour }), action({ id: "b" })), {
      viewerID: "me",
    });
    run.start();
    expect(run.current().detour?.id).toBe("to-main");
    run.skipStep();
    expect(run.current()).toMatchObject({ step: { id: "b" }, detour: null });
  });

  it("advances a step the board cannot produce, and logs why", () => {
    const log = vi.fn();
    const step = action({ cannot: (c) => (turnOf(c.view) === 4 ? "no land in hand" : null) });
    const run = createTutorialRun(script(step), { viewerID: "me", log });
    run.start();
    expect(run.current().step.id).toBe("act");
    run.observe(v(4));
    expect(run.current().step.id).toBe("handoff");
    expect(log).toHaveBeenCalledWith(
      "tutorial: step act cannot happen (no land in hand); advancing",
    );
  });

  it("checks done before cannot: a step that just completed is not 'cannot'", () => {
    const log = vi.fn();
    const step = action({ done: () => true, cannot: () => "nothing left" });
    const run = createTutorialRun(script(step), { viewerID: "me", log });
    run.start();
    expect(run.current().step.id).toBe("handoff");
    expect(log).not.toHaveBeenCalled();
  });

  it("completes a hover step on hovered(), and nothing else on it", () => {
    const hover = action({ id: "h", hover: { ms: 600, event: "hand-hovered" } });
    const run = createTutorialRun(script(action(), hover), { viewerID: "me" });
    run.start();
    // Not a hover step: hovered() does nothing.
    run.hovered("act");
    expect(run.current().step.id).toBe("act");
    run.skipStep();
    run.hovered("act");
    expect(run.current().step.id).toBe("h");
    run.hovered("h");
    expect(run.current().step.id).toBe("handoff");
  });

  it("advance() moves on from the current step only, logs, and never past the end", () => {
    const log = vi.fn();
    const run = createTutorialRun(script(action()), { viewerID: "me", log });
    run.start();
    run.advance("welcome", "stale");
    expect(run.current().step.id).toBe("act");
    run.advance("act", "cannot be hovered on this device");
    expect(run.current().step.id).toBe("handoff");
    expect(log).toHaveBeenCalledWith(
      "tutorial: step act cannot be hovered on this device; advancing",
    );
    run.advance("handoff", "done");
    expect(run.current()).toMatchObject({ step: HANDOFF, visible: true });
  });

  it("lets an action step carry a timeout too", () => {
    const log = vi.fn();
    const run = createTutorialRun(script(action({ timeoutMs: 5_000 })), { viewerID: "me", log });
    run.start();
    vi.advanceTimersByTime(4_999);
    expect(run.current().step.id).toBe("act");
    vi.advanceTimersByTime(1);
    expect(run.current().step.id).toBe("handoff");
    expect(log).toHaveBeenCalledWith(expect.stringMatching(/act timed out/));
  });

  it("hands anchors, statuses and predicates the board as it is now", () => {
    const step = action({
      anchor: (c) => (turnOf(c.view) > 1 ? { cardID: `c${turnOf(c.view)}` } : null),
      status: (c) => `turn ${turnOf(c.view)}`,
    });
    const run = createTutorialRun(script(step), { viewerID: "me", view: v(1) });
    run.start();
    expect(anchorsOf(step, run.context())).toEqual([]);
    expect(statusText(step, run.context())).toBe("turn 1");
    run.observe(v(2));
    expect(anchorsOf(step, run.context())).toEqual([{ cardID: "c2" }]);
    expect(run.context()).toMatchObject({ viewerID: "me", event: null });
    expect(turnOf(run.context().start)).toBe(1);
    // A computed anchor with no board names nothing.
    expect(anchorsOf(step)).toEqual([]);
  });

  it("spotlights a detour's own anchor, else the step's", () => {
    const step = action();
    expect(spotAnchors(step, null)).toEqual([{ label: L.lands, within: L.yourBoard }]);
    expect(spotAnchors(step, { ...detour, anchor: { label: L.actions } })).toEqual([
      { label: L.actions },
    ]);
    expect(spotAnchors(step, detour)).toEqual([{ label: L.lands, within: L.yourBoard }]);
    expect(statusText(action(), { view: null, start: null, viewerID: null, event: null })).toBe(
      undefined,
    );
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

// ADR 0125 §5.3: a step that completes reports itself, so the coach can
// mark the hints it teaches as seen. One that is skipped, gives up,
// times out or loses its anchor reports nothing.
describe("onComplete", () => {
  const teaching = (over: Partial<TutorialStep> = {}) =>
    action({ teaches: ["table.stack"], ...over });

  function runWith(steps: TutorialStep[]) {
    const done: string[] = [];
    const run = createTutorialRun(steps, {
      viewerID: "me",
      log: () => {},
      onComplete: (s) => done.push(s.id),
    });
    return { run, done };
  }

  it("reports Start, a predicate, a hover, and Finish on the last step", () => {
    const hover = action({ id: "h", hover: { ms: 600 } });
    const pred = teaching({ id: "p", done: (c) => turnOf(c.view) === 2 });
    const { run, done } = runWith(script(pred, hover));
    run.start();
    run.observe(v(2));
    run.hovered("h");
    expect(run.current().step).toBe(HANDOFF);
    run.close();
    expect(done).toEqual(["welcome", "p", "h", "handoff"]);
    // A second close reports nothing more.
    run.close();
    expect(done).toHaveLength(4);
  });

  it("reports nothing for a skip, a give-up, a timeout or a missing anchor", () => {
    const steps = script(
      teaching({ id: "skipped" }),
      teaching({ id: "gave-up", cannot: (c) => (turnOf(c.view) === 3 ? "gone" : null) }),
      teaching({ id: "timed-out", timeoutMs: 1_000 }),
      teaching({ id: "no-anchor" }),
      teaching({ id: "advanced" }),
    );
    const { run, done } = runWith(steps);
    run.start();
    run.skipStep();
    run.observe(v(3));
    expect(run.current().step.id).toBe("timed-out");
    vi.advanceTimersByTime(1_000);
    run.anchorMissing("no-anchor");
    run.advance("advanced", "cannot be hovered on this device");
    expect(run.current().step).toBe(HANDOFF);
    expect(done).toEqual(["welcome"]);
  });

  it("reports nothing when Skip tutorial closes the opening card", () => {
    const { run, done } = runWith(script(teaching()));
    run.close();
    expect(done).toEqual([]);
  });

  it("moves on, and logs, when the report throws", () => {
    const log = vi.fn();
    const run = createTutorialRun(script(teaching({ done: () => true })), {
      viewerID: "me",
      log,
      onComplete: () => {
        throw new Error("boom");
      },
    });
    run.start();
    expect(run.current().step).toBe(HANDOFF);
    expect(log).toHaveBeenCalledWith(expect.stringMatching(/step act completion failed/));
  });
});

// ADR 0125 §5.2: while the opening roll is open every step but the roll's
// own is held, and a held step's hint and timeout wait for the deal.
describe("the opening roll's hold and a step's timers", () => {
  const rolling = () =>
    ({
      id: "g",
      turn: { number: 1 },
      opening_roll: { rounds: [{ seats: [0, 1], rolls: [] }] },
    }) as unknown as GameView;

  it("starts a held step's timeout when the hold lifts, for its full length", () => {
    const log = vi.fn();
    const run = createTutorialRun(script(action({ timeoutMs: 5_000 })), {
      viewerID: "me",
      view: rolling(),
      log,
    });
    run.start();
    expect(run.current()).toMatchObject({ step: { id: "act" }, held: true });
    // The roll outlasts the step's whole timeout: nothing happens yet.
    vi.advanceTimersByTime(8_000);
    expect(run.current().step.id).toBe("act");
    run.observe(v(1));
    expect(run.current().held).toBe(false);
    vi.advanceTimersByTime(4_999);
    expect(run.current().step.id).toBe("act");
    vi.advanceTimersByTime(1);
    expect(run.current().step).toBe(HANDOFF);
    expect(log).toHaveBeenCalledWith("tutorial: step act timed out waiting; advancing");
  });

  it("starts a held step's hint when the hold lifts", () => {
    const run = createTutorialRun(script(action()), { viewerID: "me", view: rolling() });
    run.start();
    vi.advanceTimersByTime(HINT_AFTER_MS * 2);
    expect(run.current().coach).toBe("action");
    run.observe(v(1));
    vi.advanceTimersByTime(HINT_AFTER_MS - 1);
    expect(run.current().coach).toBe("action");
    vi.advanceTimersByTime(1);
    expect(run.current().coach).toBe("hint");
  });

  it("stops a step's timers while a hold lasts and starts them afresh after it", () => {
    const run = createTutorialRun(script(action({ timeoutMs: 5_000 })), {
      viewerID: "me",
      view: v(1),
      log: () => {},
    });
    run.start();
    vi.advanceTimersByTime(4_000);
    run.observe(rolling());
    vi.advanceTimersByTime(5_000);
    expect(run.current().step.id).toBe("act");
    run.observe(v(1));
    vi.advanceTimersByTime(4_999);
    expect(run.current().step.id).toBe("act");
    vi.advanceTimersByTime(1);
    expect(run.current().step).toBe(HANDOFF);
  });

  it("does not hold the step that teaches the roll", () => {
    const log = vi.fn();
    const roll = action({
      id: "roll",
      duringOpeningRoll: true,
      done: (c) => !!c.view && !(c.view as { opening_roll?: unknown }).opening_roll,
      first: () => ({ id: "choose", title: "You won", body: "Choose." }),
    });
    const run = createTutorialRun(script(roll), { viewerID: "me", view: rolling(), log });
    run.start();
    // Its detour and its hint work during the roll; the roll's end completes it.
    expect(run.current()).toMatchObject({ step: { id: "roll" }, held: false });
    expect(run.current().detour?.id).toBe("choose");
    vi.advanceTimersByTime(HINT_AFTER_MS);
    expect(run.current().coach).toBe("hint");
    run.observe(v(1));
    expect(run.current().step).toBe(HANDOFF);
    expect(log).not.toHaveBeenCalled();
  });
});
