// #1508: drag a card out of the hand onto the table to cast it. The
// gesture's decisions — when a press becomes a drag, where the line is,
// what a release does, what the ghost says while it moves — are pinned
// here; Hand.svelte only wires pointer events to them.

import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import {
  CAST_ZONE_MARGIN_PX,
  DRAG_ACTIVATION_PX,
  IDLE,
  LAND_DRAG_REASON,
  autoTapHighlight,
  cantPayReason,
  clearAutoTapHighlight,
  dragVerdict,
  inCastZone,
  pastActivation,
  previewDecidesMana,
  previewSourceIDs,
  release,
  step,
  wantsPreview,
  type DragInput,
  type DragOutcome,
  type DragState,
  type Verdict,
} from "./dragCast";
import type { CardView } from "./protocol";
import type { AutoTapPreview } from "./api";

const HAND_TOP = 600;
const OK: Verdict = { castable: true };

function card(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "bolt",
    name: "Lightning Bolt",
    owner: "me",
    controller: "me",
    type_line: "Instant",
    mana_cost: "{R}",
    ...extras,
  };
}

// run feeds a gesture through the machine and returns the final state
// and the last non-"none" outcome.
function run(inputs: DragInput[]): { state: DragState; outcome: DragOutcome } {
  let state = IDLE;
  let outcome: DragOutcome = { kind: "none" };
  for (const input of inputs) {
    const res = step(state, input);
    state = res.state;
    if (res.outcome.kind !== "none") outcome = res.outcome;
  }
  return { state, outcome };
}

const down = (x = 100, y = 650): DragInput => ({
  type: "down",
  cardID: "bolt",
  x,
  y,
  handTop: HAND_TOP,
});
const move = (x: number, y: number): DragInput => ({ type: "move", x, y });
const up = (x: number, y: number, verdict: Verdict = OK): DragInput => ({
  type: "up",
  x,
  y,
  verdict,
});

// Far enough above the hand to be on the table.
const TABLE_Y = HAND_TOP - CAST_ZONE_MARGIN_PX - 1;

describe("the activation distance: click vs drag", () => {
  it("is about 8px", () => {
    expect(DRAG_ACTIVATION_PX).toBe(8);
  });

  it("a press that never moves is a click", () => {
    const { state, outcome } = run([down(), up(100, 650)]);
    expect(outcome).toEqual({ kind: "click" });
    expect(state).toEqual(IDLE);
  });

  it("a press that jitters under the threshold is still a click", () => {
    const { state, outcome } = run([down(), move(104, 646), move(97, 653), up(97, 653)]);
    expect(outcome).toEqual({ kind: "click" });
    expect(state.phase).toBe("idle");
  });

  it("a press that travels the threshold becomes a drag", () => {
    const res = run([down(), move(100, 650 - DRAG_ACTIVATION_PX)]);
    expect(res.state.phase).toBe("dragging");
  });

  it("measures the distance, not one axis", () => {
    expect(pastActivation(0, 0, 5, 5)).toBe(false); // ≈7.07px
    expect(pastActivation(0, 0, 6, 6)).toBe(true); // ≈8.49px
  });

  it("ignores a second pointerdown while a gesture is live", () => {
    const first = run([down(), move(100, 600)]);
    const res = step(first.state, { ...down(300, 650), cardID: "other" } as DragInput);
    expect(res.state).toBe(first.state);
    expect(res.outcome).toEqual({ kind: "none" });
  });

  it("ignores moves and releases with nothing pressed", () => {
    expect(step(IDLE, move(10, 10))).toEqual({ state: IDLE, outcome: { kind: "none" } });
    expect(step(IDLE, up(10, 10))).toEqual({ state: IDLE, outcome: { kind: "none" } });
  });
});

describe("the cast zone: released over the table, past the margin", () => {
  it("is about 60px above the hand's top edge", () => {
    expect(CAST_ZONE_MARGIN_PX).toBe(60);
    expect(inCastZone(HAND_TOP - 61, HAND_TOP)).toBe(true);
    expect(inCastZone(HAND_TOP - 60, HAND_TOP)).toBe(false); // on the line: short
    expect(inCastZone(HAND_TOP - 10, HAND_TOP)).toBe(false);
    expect(inCastZone(HAND_TOP + 20, HAND_TOP)).toBe(false);
  });

  it("tracks whether the dragged card is past the line", () => {
    let res = run([down(), move(100, 620)]);
    expect(res.state.inCastZone).toBe(false);
    res = { ...res, ...step(res.state, move(100, TABLE_Y)) };
    expect(res.state.inCastZone).toBe(true);
    res = { ...res, ...step(res.state, move(100, 590)) };
    expect(res.state.inCastZone).toBe(false);
  });

  it("casts when released past the line while castable", () => {
    const { outcome, state } = run([down(), move(100, 400), up(100, TABLE_Y)]);
    expect(outcome).toEqual({ kind: "cast", cardID: "bolt" });
    expect(state).toEqual(IDLE);
  });

  it("snaps back, with no message, when released short of the line", () => {
    const { outcome } = run([down(), move(100, 600), up(100, 600)]);
    expect(outcome).toEqual({ kind: "snapBack" });
  });

  it("the release point decides, not the furthest point reached", () => {
    const { outcome } = run([down(), move(100, 200), move(100, 640), up(100, 640)]);
    expect(outcome).toEqual({ kind: "snapBack" });
  });
});

describe("castable vs not", () => {
  it("a red card released on the table snaps back with its reason", () => {
    const verdict = { castable: false, reason: "Not your priority" };
    const { outcome } = run([down(), move(100, 300), up(100, 300, verdict)]);
    expect(outcome).toEqual({ kind: "snapBack", reason: "Not your priority" });
  });

  it("a red card released short of the line snaps back silently", () => {
    const verdict = { castable: false, reason: "Not your priority" };
    const { outcome } = run([down(), move(100, 620), up(100, 620, verdict)]);
    expect(outcome).toEqual({ kind: "snapBack" });
  });

  it("release refuses a state with no card", () => {
    expect(release(IDLE, -1000, OK)).toEqual({ kind: "snapBack" });
  });
});

describe("cancel", () => {
  it("Escape / pointercancel mid-drag snaps the card back", () => {
    const { state, outcome } = run([down(), move(100, 300), { type: "cancel" }]);
    expect(outcome).toEqual({ kind: "snapBack" });
    expect(state).toEqual(IDLE);
  });

  it("cancelling a press that never became a drag is silent", () => {
    const { state, outcome } = run([down(), { type: "cancel" }]);
    expect(outcome).toEqual({ kind: "none" });
    expect(state).toEqual(IDLE);
  });

  it("a release after a cancel does nothing", () => {
    const { outcome } = run([down(), move(100, 300), { type: "cancel" }]);
    expect(outcome.kind).toBe("snapBack");
    expect(step(IDLE, up(100, TABLE_Y)).outcome).toEqual({ kind: "none" });
  });
});

describe("dragVerdict — the live gold / red answer", () => {
  it("is gold for a legal card with no preview yet", () => {
    expect(dragVerdict(card(), { legal: true }, null)).toEqual({ castable: true });
  });

  it("is red with the legality reason the click path uses", () => {
    expect(dragVerdict(card(), { legal: false, reason: "Not your priority" }, null)).toEqual({
      castable: false,
      reason: "Not your priority",
    });
  });

  it("is red with the preview's missing mana when the board can't pay", () => {
    const preview: AutoTapPreview = { ok: false, missing: ["{2}", "{U}"], cost: "{2}{U}" };
    expect(dragVerdict(card({ mana_cost: "{2}{U}" }), { legal: true }, preview)).toEqual({
      castable: false,
      reason: "Can't pay {2}{U}",
    });
  });

  it("is gold when the preview can pay", () => {
    const preview: AutoTapPreview = { ok: true, plan: ["mountain"], cost: "{R}" };
    expect(dragVerdict(card(), { legal: true }, preview)).toEqual({ castable: true });
  });

  it("ignores a failing preview when a later choice can change the price", () => {
    const preview: AutoTapPreview = { ok: false, missing: ["{U}"], cost: "{3}{U}{U}" };
    const fow = card({
      alternative_costs: [{ key: "pitch", label: "Exile a blue card" }],
    } as Partial<CardView>);
    expect(dragVerdict(fow, { legal: true }, preview)).toEqual({ castable: true });
  });

  it("the timing reason wins over the mana reason", () => {
    const preview: AutoTapPreview = { ok: false, missing: ["{R}"], cost: "{R}" };
    expect(
      dragVerdict(card(), { legal: false, reason: "Split second on the stack" }, preview).reason,
    ).toBe("Split second on the stack");
  });

  it("has a fallback reason for a denial that carries none", () => {
    expect(dragVerdict(card(), { legal: false }, null).reason).toBe("Can't cast now");
  });

  it("cantPayReason words an empty missing list too", () => {
    expect(cantPayReason({ ok: false, cost: "{1}" })).toBe("Can't pay this cost");
  });
});

describe("lands are not drag-played", () => {
  const forest = card({ instance_id: "forest", name: "Forest", type_line: "Basic Land — Forest" });

  it("a land is always red with 'Play lands by clicking'", () => {
    expect(LAND_DRAG_REASON).toBe("Play lands by clicking");
    expect(dragVerdict(forest, { legal: true }, null)).toEqual({
      castable: false,
      reason: LAND_DRAG_REASON,
    });
  });

  it("so a land dragged onto the table snaps back with that reason", () => {
    const verdict = dragVerdict(forest, { legal: true }, null);
    const { outcome } = run([down(), move(100, 300), up(100, 300, verdict)]);
    expect(outcome).toEqual({ kind: "snapBack", reason: LAND_DRAG_REASON });
  });

  it("but a land is still an ordinary click", () => {
    const verdict = dragVerdict(forest, { legal: true }, null);
    const { outcome } = run([down(), up(100, 650, verdict)]);
    expect(outcome).toEqual({ kind: "click" });
  });

  it("asks for no auto-tap preview", () => {
    expect(wantsPreview(forest, { legal: true })).toBe(false);
  });
});

describe("the auto-tap preview", () => {
  it("is fetched for a legal spell and not for a refused one", () => {
    expect(wantsPreview(card(), { legal: true })).toBe(true);
    expect(wantsPreview(card(), { legal: false, reason: "Not your priority" })).toBe(false);
  });

  it("decides mana only for a card whose price nothing later can change", () => {
    expect(previewDecidesMana(card())).toBe(true);
    expect(
      previewDecidesMana(
        card({ alternative_costs: [{ key: "evoke" }] } as unknown as Partial<CardView>),
      ),
    ).toBe(false);
    expect(previewDecidesMana(card({ alternative_cost_required: true }))).toBe(false);
    // A modal DFC has not picked its face: the other half may be cheap.
    expect(
      previewDecidesMana(
        card({ layout: "modal_dfc", faces: [{}, {}] } as unknown as Partial<CardView>),
      ),
    ).toBe(false);
    expect(previewDecidesMana(card({ phyrexian_symbols: 1 }))).toBe(false);
    expect(
      previewDecidesMana(card({ tap_cost: { kind: "convoke" } } as unknown as Partial<CardView>)),
    ).toBe(false);
  });

  it("highlights the described sources, falling back to the bare plan", () => {
    expect(
      previewSourceIDs({
        ok: true,
        cost: "{R}",
        plan: ["a"],
        sources: [{ card_id: "guide", zone: "hand", exile: true }],
      }),
    ).toEqual(["guide"]);
    expect(previewSourceIDs({ ok: true, cost: "{R}", plan: ["m1", "m2"] })).toEqual(["m1", "m2"]);
  });

  it("highlights nothing for a plan that can't pay, or no preview", () => {
    expect(previewSourceIDs({ ok: false, cost: "{R}", plan: ["m1"] })).toEqual([]);
    expect(previewSourceIDs(null)).toEqual([]);
  });

  it("the highlight store clears", () => {
    autoTapHighlight.set(new Set(["m1"]));
    clearAutoTapHighlight();
    expect(get(autoTapHighlight).size).toBe(0);
  });
});
