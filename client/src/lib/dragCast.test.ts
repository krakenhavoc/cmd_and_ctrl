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
  dragMode,
  dragVerdict,
  inCastZone,
  inReorderBand,
  insertionIndex,
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

// ---- #1524: one gesture, two outcomes --------------------------------
//
// The hand's band is its own top to bottom edge. Sideways inside it is
// a reorder; up past the line is #1508's cast; released between the
// two the card snaps back.

describe("the reorder band (#1524)", () => {
  const HAND_BOTTOM = 760;
  // The OTHER cards' centres; the dragged card is index 1 of four.
  const CENTERS = [100, 300, 400];
  const rdown = (x = 200, y = 650): DragInput => ({
    type: "down",
    cardID: "bolt",
    x,
    y,
    handTop: HAND_TOP,
    handBottom: HAND_BOTTOM,
    slotCenters: CENTERS,
    fromIndex: 1,
  });

  it("the mode follows the pointer's height", () => {
    expect(dragMode(HAND_TOP, HAND_TOP, HAND_BOTTOM)).toBe("reorder");
    expect(dragMode(700, HAND_TOP, HAND_BOTTOM)).toBe("reorder");
    expect(dragMode(HAND_BOTTOM, HAND_TOP, HAND_BOTTOM)).toBe("reorder");
    expect(dragMode(HAND_TOP - 1, HAND_TOP, HAND_BOTTOM)).toBe("none");
    expect(dragMode(HAND_TOP - CAST_ZONE_MARGIN_PX, HAND_TOP, HAND_BOTTOM)).toBe("none");
    expect(dragMode(TABLE_Y, HAND_TOP, HAND_BOTTOM)).toBe("cast");
    expect(dragMode(HAND_BOTTOM + 1, HAND_TOP, HAND_BOTTOM)).toBe("none");
  });

  it("a hand that gave no bottom edge has no reorder band", () => {
    expect(inReorderBand(650, HAND_TOP, null)).toBe(false);
    expect(dragMode(650, HAND_TOP, null)).toBe("none");
    expect(dragMode(TABLE_Y, HAND_TOP, null)).toBe("cast");
  });

  it("the state machine tracks the mode and the insertion index while dragging", () => {
    let res = run([rdown(), move(40, 650)]);
    expect(res.state.mode).toBe("reorder");
    expect(res.state.insertAt).toBe(0);
    res = { ...res, ...step(res.state, move(350, 700)) };
    expect(res.state.insertAt).toBe(2);
    res = { ...res, ...step(res.state, move(350, 580)) };
    expect(res.state.mode).toBe("none");
    expect(res.state.insertAt).toBe(-1);
    res = { ...res, ...step(res.state, move(350, TABLE_Y)) };
    expect(res.state.mode).toBe("cast");
    expect(res.state.inCastZone).toBe(true);
  });

  it("a press is still a click, and has no mode", () => {
    const { outcome, state } = run([rdown(), up(203, 652)]);
    expect(outcome).toEqual({ kind: "click" });
    expect(state).toEqual(IDLE);
  });

  it("released sideways inside the band, the card moves to the insertion index", () => {
    expect(run([rdown(), move(450, 650), up(450, 650)]).outcome).toEqual({
      kind: "reorder",
      cardID: "bolt",
      from: 1,
      to: 3,
    });
    expect(run([rdown(), move(20, 700), up(20, 700)]).outcome).toEqual({
      kind: "reorder",
      cardID: "bolt",
      from: 1,
      to: 0,
    });
  });

  it("released back where it started, it snaps back rather than reordering", () => {
    // Between the first and second other card: index 1, where it was.
    const { outcome } = run([rdown(), move(150, 700), up(250, 700)]);
    expect(outcome).toEqual({ kind: "snapBack" });
  });

  it("released up past the line, it casts exactly as before", () => {
    const { outcome } = run([rdown(), move(450, 650), up(450, TABLE_Y)]);
    expect(outcome).toEqual({ kind: "cast", cardID: "bolt" });
  });

  it("released between the band and the line, it snaps back silently", () => {
    const red = { castable: false, reason: "Not your priority" };
    expect(run([rdown(), move(450, 650), up(450, HAND_TOP - 20)]).outcome).toEqual({
      kind: "snapBack",
    });
    expect(run([rdown(), move(450, 650), up(450, HAND_TOP - 20, red)]).outcome).toEqual({
      kind: "snapBack",
    });
  });

  it("the release point decides, not the band the drag passed through", () => {
    const { outcome } = run([rdown(), move(450, TABLE_Y), move(450, 700), up(450, 700)]);
    expect(outcome.kind).toBe("reorder");
  });

  it("a reorder ignores the cast verdict", () => {
    const red = { castable: false, reason: "Not your priority" };
    const { outcome } = run([rdown(), move(450, 650), up(450, 650, red)]);
    expect(outcome.kind).toBe("reorder");
  });

  it("a land can be reordered, but is still not cast by drag", () => {
    const forest = card({ instance_id: "forest", type_line: "Basic Land — Forest" });
    const verdict = dragVerdict(forest, { legal: true }, null);
    const land = (): DragInput => ({ ...rdown(), cardID: "forest" }) as DragInput;
    expect(run([land(), move(450, 650), up(450, 650, verdict)]).outcome).toEqual({
      kind: "reorder",
      cardID: "forest",
      from: 1,
      to: 3,
    });
    expect(run([land(), move(450, TABLE_Y), up(450, TABLE_Y, verdict)]).outcome).toEqual({
      kind: "snapBack",
      reason: LAND_DRAG_REASON,
    });
  });

  it("a gesture that gave no reorder geometry never reorders (#1508's drag alone)", () => {
    const { outcome } = run([down(), move(450, 650), up(450, 650)]);
    expect(outcome).toEqual({ kind: "snapBack" });
  });
});

describe("insertionIndex", () => {
  it("counts the other cards' centres to the left of the pointer", () => {
    const centers = [100, 200, 300];
    expect(insertionIndex(50, centers)).toBe(0);
    expect(insertionIndex(150, centers)).toBe(1);
    expect(insertionIndex(250, centers)).toBe(2);
    expect(insertionIndex(999, centers)).toBe(3);
  });

  it("a pointer exactly on a centre goes before that card", () => {
    expect(insertionIndex(200, [100, 200, 300])).toBe(1);
  });

  it("with no other cards, the only place is 0", () => {
    expect(insertionIndex(500, [])).toBe(0);
  });
});
