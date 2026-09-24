// @vitest-environment jsdom
//
// #1467 — the fan stack lane's target arrows. The geometry is pure and
// pinned here; `measureStackArrows` is exercised against a hand-built
// DOM whose rects are stubbed, since jsdom does no layout.

import { describe, expect, it } from "vitest";

import type { CardView, PlayerView, StackItemView, ZoneView } from "./protocol";
import { seatColor } from "./colors";
import { buildStackLane } from "./stackLane";
import {
  TARGET_ATTR,
  TARGET_GAP,
  TARGET_RING_VAR,
  TOP_ARROW_COLOR,
  boardCurve,
  boxCentre,
  controlTowardCentre,
  curvePath,
  edgePoint,
  insideBox,
  laneArc,
  laneExit,
  measureStackArrows,
  planArrows,
  relativeTo,
  syncTargetMarks,
  targetSelector,
  type Box,
} from "./stackArrows";

const ME = "me";
const BOT = "bot";

const zone = (kind: string, cards: CardView[] = []): ZoneView =>
  ({ kind, count: cards.length, cards }) as ZoneView;
const seat = (id: string, name: string, n: number, graveyard: CardView[] = []): PlayerView =>
  ({ id, name, seat: n, life: 40, graveyard: zone("graveyard", graveyard) }) as PlayerView;
const card = (id: string, name: string, controller: string) =>
  ({ instance_id: id, name, owner: controller, controller, known_by_you: true }) as CardView;
const spell = (id: string, controller: string, targets: StackItemView["targets"] = []) =>
  ({
    id,
    kind: "spell",
    controller,
    owner: controller,
    source_card_id: id,
    targets,
  }) as StackItemView;

const bolt = card("bolt", "Lightning Bolt", ME);
const counterspell = card("cs", "Counterspell", BOT);
const birds = card("birds", "Birds of Paradise", BOT);
const buried = card("buried", "Eternal Witness", ME);

// Bottom first, as the wire sends it: Bolt aimed at Birds, then
// Counterspell aimed at Bolt, then a Bolt-like spell at the bot and at
// a graveyard card.
function model() {
  const aimAtPlayer = card("shock", "Shock", ME);
  return buildStackLane({
    stack: zone("stack", [bolt, counterspell, aimAtPlayer]),
    stackItems: [
      spell("bolt", ME, [{ kind: "card", id: "birds" }]),
      spell("cs", BOT, [{ kind: "card", id: "bolt" }]),
      spell("shock", ME, [
        { kind: "player", id: BOT },
        { kind: "card", id: "buried" },
      ]),
    ],
    pendingTriggers: [],
    seats: [seat(ME, "Me", 0, [buried]), seat(BOT, "Bot 2", 1)],
    battlefield: zone("battlefield", [birds]),
    viewerID: ME,
    priorityHolder: ME,
    splitSecondActive: false,
  });
}

describe("planArrows", () => {
  it("asks for one arrow per permanent, player and stack target, and none for a card", () => {
    const plans = planArrows(model().items);
    expect(plans.map((p) => [p.itemID, p.targetKind, p.targetID, p.layer])).toEqual([
      ["shock", "player", BOT, "board"],
      ["cs", "stack", "bolt", "lane"],
      ["bolt", "permanent", "birds", "board"],
    ]);
  });

  it("colours the top item's arrows with the flag tone and the rest by caster", () => {
    const plans = planArrows(model().items);
    const shock = plans.find((p) => p.itemID === "shock")!;
    expect(shock.fromTop).toBe(true);
    expect(shock.color).toBe(TOP_ARROW_COLOR);
    const cs = plans.find((p) => p.itemID === "cs")!;
    expect(cs.fromTop).toBe(false);
    expect(cs.color).toBe(seatColor(1));
  });

  it("draws one arrow for a thing targeted twice by the same item", () => {
    const m = buildStackLane({
      stack: zone("stack", [bolt]),
      stackItems: [
        spell("bolt", ME, [
          { kind: "card", id: "birds" },
          { kind: "card", id: "birds" },
        ]),
      ],
      pendingTriggers: [],
      seats: [seat(ME, "Me", 0), seat(BOT, "Bot 2", 1)],
      battlefield: zone("battlefield", [birds]),
      viewerID: ME,
      priorityHolder: ME,
      splitSecondActive: false,
    });
    expect(planArrows(m.items)).toHaveLength(1);
  });

  it("selects each kind by its board attribute", () => {
    expect(targetSelector("permanent", "a")).toBe('[data-instance-id="a"]');
    expect(targetSelector("player", "p")).toBe('[data-seat-id="p"]');
    expect(targetSelector("stack", "s")).toBe('[data-stack-item-id="s"]');
  });
});

describe("geometry", () => {
  const box = (left: number, top: number, width: number, height: number): Box => ({
    left,
    top,
    width,
    height,
  });

  it("translates a viewport rect into board-relative coordinates", () => {
    expect(relativeTo(box(130, 250, 40, 60), box(100, 200, 900, 700))).toEqual(box(30, 50, 40, 60));
  });

  it("puts an end on the edge of its rect facing the other end", () => {
    const b = box(0, 0, 100, 140); // centre 50,70
    expect(edgePoint(b, { x: 50, y: -500 })).toEqual({ x: 50, y: 0 });
    expect(edgePoint(b, { x: 500, y: 70 })).toEqual({ x: 100, y: 70 });
    // Along a diagonal it meets the nearer side.
    const p = edgePoint(b, { x: 150, y: 170 });
    expect(p.x).toBeCloseTo(100);
    expect(p.y).toBeCloseTo(120);
  });

  it("stands an arrowhead off its rect by the gap", () => {
    expect(edgePoint(box(0, 0, 100, 140), { x: 50, y: -500 }, 6)).toEqual({ x: 50, y: -6 });
  });

  it("answers the centre for a point inside the rect", () => {
    expect(edgePoint(box(0, 0, 100, 100), { x: 60, y: 60 })).toEqual({ x: 50, y: 50 });
  });

  it("bows the control point toward the table centre", () => {
    // A horizontal chord along the top of a 1000×1000 table bows down.
    const c = controlTowardCentre({ x: 100, y: 100 }, { x: 500, y: 100 }, { x: 500, y: 500 });
    expect(c.x).toBeCloseTo(300);
    expect(c.y).toBeGreaterThan(100);
    // The same chord along the bottom bows up.
    const d = controlTowardCentre({ x: 100, y: 900 }, { x: 500, y: 900 }, { x: 500, y: 500 });
    expect(d.y).toBeLessThan(900);
  });

  it("runs a board arrow from the tile's edge to a gap short of the target", () => {
    const tile = box(400, 400, 100, 140); // centre 450,470
    const target = box(420, 60, 70, 98); // above it; centre 455,109
    const c = boardCurve(tile, target, { width: 1000, height: 800 });
    expect(c.y1).toBeCloseTo(400); // the tile's top edge
    expect(c.y2).toBeGreaterThan(158); // below the target's bottom edge…
    expect(c.y2).toBeLessThan(158 + TARGET_GAP + 1); // …by the gap
  });

  it("leaves the lane on the border straight above or below its tile", () => {
    const lane = box(100, 300, 800, 300); // y 300..600
    const tile = box(400, 360, 100, 140); // centre 450,430
    expect(laneExit(tile, { x: 200, y: 100 }, lane)).toEqual({ x: 450, y: 300 });
    expect(laneExit(tile, { x: 800, y: 750 }, lane)).toEqual({ x: 450, y: 600 });
    // Beside the lane: out of its side at the tile's height.
    expect(laneExit(tile, { x: 20, y: 450 }, lane)).toEqual({ x: 100, y: 430 });
    expect(laneExit(tile, { x: 980, y: 450 }, lane)).toEqual({ x: 900, y: 430 });
  });

  it("starts a board arrow on the lane's border when given the lane", () => {
    const lane = box(100, 300, 800, 300);
    const tile = box(400, 360, 100, 140);
    const target = box(420, 60, 70, 98); // above the lane
    const c = boardCurve(tile, target, { width: 1000, height: 800 }, lane);
    expect([c.x1, c.y1]).toEqual([450, 300]);
    expect(c.y2).toBeGreaterThan(158);
    expect(c.y2).toBeLessThan(158 + TARGET_GAP + 1);
  });

  it("never bows a board arrow back into the lane", () => {
    // The lane covers the table's centre; its top border is y = 300.
    const lane = box(100, 300, 800, 300);
    const from = { x: 800, y: 300 };
    const to = { x: 150, y: 100 };
    const centre = { x: 500, y: 450 };
    // Toward the centre alone, it would dip into the lane…
    expect(insideBox(controlTowardCentre(from, to, centre), lane)).toBe(true);
    // …so with the lane to avoid it bows the other way, above it.
    const c = controlTowardCentre(from, to, centre, undefined, lane);
    expect(insideBox(c, lane)).toBe(false);
    expect(c.y).toBeLessThan(300);
    const curve = boardCurve(
      box(750, 360, 100, 140),
      box(120, 60, 60, 60),
      {
        width: 1000,
        height: 900,
      },
      lane,
    );
    expect(curve.cy).toBeLessThan(300);
  });

  it("knows a point inside a rect", () => {
    expect(insideBox({ x: 5, y: 5 }, box(0, 0, 10, 10))).toBe(true);
    expect(insideBox({ x: 11, y: 5 }, box(0, 0, 10, 10))).toBe(false);
  });

  it("arcs from one tile to another over the lane", () => {
    const counterspell = box(0, 30, 180, 252);
    const bolt = box(380, 65, 130, 182);
    const c = laneArc(counterspell, bolt);
    // Leaves the source's top edge on the side facing the target…
    expect(c.y1).toBe(30);
    expect(c.x1).toBeGreaterThan(boxCentre(counterspell).x);
    // …lands a gap above the target's top edge, at its middle…
    expect(c.x2).toBe(boxCentre(bolt).x);
    expect(c.y2).toBe(65 - TARGET_GAP);
    // …and rises above the higher tile.
    expect(c.cy).toBeLessThan(30);
  });

  it("arcs leftward when the target tile is to the left", () => {
    const c = laneArc(box(400, 60, 130, 182), box(0, 30, 180, 252));
    expect(c.x1).toBeLessThan(465);
    expect(c.x2).toBe(90);
  });

  it("writes a quadratic path", () => {
    expect(curvePath({ x1: 1, y1: 2, cx: 3.333, cy: 4, x2: 5, y2: 6 })).toBe("M 1 2 Q 3.3 4 5 6");
  });
});

describe("measureStackArrows", () => {
  function stubRect(el: Element, left: number, top: number, width: number, height: number) {
    el.getBoundingClientRect = () =>
      ({
        left,
        top,
        width,
        height,
        right: left + width,
        bottom: top + height,
        x: left,
        y: top,
        toJSON: () => ({}),
      }) as DOMRect;
  }

  // A board at (100, 50) with the bot's header and Birds above a lane
  // holding three tiles, and the Eternal Witness card in a graveyard
  // pile (which is never an arrow target).
  function dom() {
    const board = document.createElement("div");
    board.className = "board";
    board.innerHTML = `
      <div class="seat">
        <div data-seat-id="${BOT}"></div>
        <div data-instance-id="birds"></div>
        <div data-instance-id="buried"></div>
      </div>
      <section class="stack-lane">
        <div class="track">
          <div data-stack-item-id="shock"></div>
          <div data-stack-item-id="cs"></div>
          <div data-stack-item-id="bolt"></div>
        </div>
      </section>`;
    document.body.appendChild(board);
    const q = (s: string) => board.querySelector(s)!;
    stubRect(board, 100, 50, 1200, 800);
    stubRect(q(`[data-seat-id="${BOT}"]`), 700, 70, 60, 60);
    stubRect(q('[data-instance-id="birds"]'), 800, 150, 70, 98);
    stubRect(q('[data-instance-id="buried"]'), 200, 150, 70, 98);
    const lane = q(".stack-lane");
    const track = q(".track") as HTMLElement;
    stubRect(lane, 180, 380, 940, 340);
    stubRect(track, 200, 420, 900, 300);
    stubRect(q('[data-stack-item-id="shock"]'), 204, 450, 180, 252);
    stubRect(q('[data-stack-item-id="cs"]'), 406, 471, 150, 210);
    stubRect(q('[data-stack-item-id="bolt"]'), 578, 485, 130, 182);
    return { board, lane, track, q };
  }

  it("measures board arrows in board coordinates and lane arcs in track coordinates", () => {
    const { board, lane, track } = dom();
    const m = measureStackArrows({ board, lane, track, items: model().items });

    expect(m.board.map((a) => [a.itemID, a.targetID])).toEqual([
      ["shock", BOT],
      ["bolt", "birds"],
    ]);
    const toBirds = m.board.find((a) => a.targetID === "birds")!;
    // Lands a gap short of Birds (board-relative 700,100 70×98), on the
    // edge facing the lane: outside the card, and within the gap of it.
    const birdsBox = { left: 700, top: 100, width: 70, height: 98 };
    const end = { x: toBirds.x2, y: toBirds.y2 };
    expect(insideBox(end, birdsBox)).toBe(false);
    const dx = Math.max(birdsBox.left - end.x, 0, end.x - (birdsBox.left + birdsBox.width));
    const dy = Math.max(birdsBox.top - end.y, 0, end.y - (birdsBox.top + birdsBox.height));
    expect(Math.hypot(dx, dy)).toBeLessThanOrEqual(TARGET_GAP + 0.01);
    // Leaves the lane's top border (380-50) straight above Bolt's tile.
    expect([toBirds.x1, toBirds.y1]).toEqual([578 + 65 - 100, 380 - 50]);

    expect(m.lane).toHaveLength(1);
    const arc = m.lane[0];
    expect([arc.itemID, arc.targetID]).toEqual(["cs", "bolt"]);
    // Relative to the track: Bolt's tile is at 578-200 = 378 across.
    expect(arc.x2).toBe(378 + 65);
    expect(arc.y2).toBe(485 - 420 - TARGET_GAP);

    board.remove();
  });

  it("follows the track's scroll for lane arcs", () => {
    const { board, lane, track } = dom();
    // jsdom does no layout, so it would clamp a real scroll to 0.
    Object.defineProperty(track, "scrollLeft", { value: 40 });
    const m = measureStackArrows({ board, lane, track, items: model().items });
    // The tile's viewport rect is unchanged, so in content coordinates
    // it sits 40px further along.
    expect(m.lane[0].x2).toBe(378 + 65 + 40);
    board.remove();
  });

  it("draws no board arrow from a tile scrolled out of the lane", () => {
    const { board, lane, track, q } = dom();
    // Bolt's tile has scrolled off the track's right edge (200 + 900).
    stubRect(q('[data-stack-item-id="bolt"]'), 1180, 485, 130, 182);
    const m = measureStackArrows({ board, lane, track, items: model().items });
    expect(m.board.map((a) => a.targetID)).toEqual([BOT]);
    // Birds is still what the Bolt targets, so it is still lit.
    expect(m.targets.map((t) => t.el)).toContain(q('[data-instance-id="birds"]'));
    board.remove();
  });

  it("draws no arrow to a target hidden under the lane, but still lights it", () => {
    const { board, lane, track, q } = dom();
    stubRect(q('[data-instance-id="birds"]'), 800, 600, 70, 98);
    const m = measureStackArrows({ board, lane, track, items: model().items });
    expect(m.board.map((a) => a.targetID)).toEqual([BOT]);
    expect(m.targets.map((t) => t.el)).toContain(q('[data-instance-id="birds"]'));
    board.remove();
  });

  it("draws no arrow to a target that is not on the board", () => {
    const { board, lane, track, q } = dom();
    q('[data-instance-id="birds"]').remove();
    const m = measureStackArrows({ board, lane, track, items: model().items });
    expect(m.board.map((a) => a.targetID)).toEqual([BOT]);
    board.remove();
  });

  it("draws no arrow to a target with no size on screen", () => {
    const { board, lane, track, q } = dom();
    stubRect(q(`[data-seat-id="${BOT}"]`), 0, 0, 0, 0);
    const m = measureStackArrows({ board, lane, track, items: model().items });
    expect(m.board.map((a) => a.targetID)).toEqual(["birds"]);
    board.remove();
  });

  it("marks each reached element once, and the ring follows the top item", () => {
    const { board, lane, track, q } = dom();
    const m = measureStackArrows({ board, lane, track, items: model().items });
    const seatEl = q(`[data-seat-id="${BOT}"]`) as HTMLElement;
    const birdsEl = q('[data-instance-id="birds"]') as HTMLElement;
    expect(m.targets.map((t) => t.el)).toEqual([seatEl, birdsEl]);

    let marked = syncTargetMarks(new Set(), m.targets);
    expect(seatEl.getAttribute(TARGET_ATTR)).toBe("next");
    expect(seatEl.style.getPropertyValue(TARGET_RING_VAR)).toBe(TOP_ARROW_COLOR);
    expect(birdsEl.getAttribute(TARGET_ATTR)).toBe("");
    expect(birdsEl.style.getPropertyValue(TARGET_RING_VAR)).toBe(seatColor(0));

    // Birds is no longer targeted: its mark comes off, the seat's stays.
    marked = syncTargetMarks(marked, [m.targets[0]]);
    expect(birdsEl.hasAttribute(TARGET_ATTR)).toBe(false);
    expect(birdsEl.style.getPropertyValue(TARGET_RING_VAR)).toBe("");
    expect(seatEl.hasAttribute(TARGET_ATTR)).toBe(true);

    // Teardown clears everything.
    syncTargetMarks(marked, []);
    expect(seatEl.hasAttribute(TARGET_ATTR)).toBe(false);
    board.remove();
  });
});
