// boardExpand.test.ts — ADR 0120 §1 and §2 as pure functions: every
// transition of the expanded board, and where it sits.

import { describe, expect, it } from "vitest";

import {
  initialExpandState,
  liveExpanded,
  parsePx,
  PEEK_CLOSE_GRACE_MS,
  peekOpenDelay,
  placeOverlay,
  reduceExpand,
  type ExpandEvent,
  type ExpandState,
} from "./boardExpand";

const OPEN = 400;

// A tiny driver: feed events, and fire the timer the state asks for.
function run(...events: ExpandEvent[]): ExpandState {
  return events.reduce(reduceExpand, initialExpandState());
}
function fire(s: ExpandState): ExpandState {
  expect(s.timer, "no timer to fire").not.toBeNull();
  return reduceExpand(s, { type: "elapsed", seq: s.timer!.seq });
}
const enter = (seatID: string): ExpandEvent => ({
  type: "avatar-enter",
  seatID,
  openDelayMs: OPEN,
});
const leave = (seatID: string): ExpandEvent => ({ type: "avatar-leave", seatID });
const press = (seatID: string): ExpandEvent => ({ type: "avatar-press", seatID });
const click = (seatID: string): ExpandEvent => ({ type: "avatar-click", seatID });

describe("the open delay", () => {
  it("is 400 ms at the default zoom delay, and never less", () => {
    expect(peekOpenDelay(300)).toBe(400);
    expect(peekOpenDelay(0)).toBe(400);
  });

  it("follows a slower hover zoom", () => {
    expect(peekOpenDelay(650)).toBe(650);
  });
});

describe("hover opens a peek", () => {
  it("after the delay, not before", () => {
    const s = run(enter("bob"));
    expect(s.expanded).toBeNull();
    expect(s.timer).toMatchObject({ kind: "open", ms: OPEN });
    expect(fire(s).expanded).toEqual({ seatID: "bob", pinned: false });
  });

  it("not when the pointer leaves first", () => {
    const s = run(enter("bob"), leave("bob"));
    expect(s.timer).toBeNull();
    expect(s.expanded).toBeNull();
  });

  it("replaces an open peek for another seat", () => {
    const s = fire(reduceExpand(fire(run(enter("bob"))), enter("cat")));
    expect(s.expanded).toEqual({ seatID: "cat", pinned: false });
  });

  it("does not open while a board is pinned: one overlay at a time", () => {
    const s = run(click("bob"), leave("bob"), enter("cat"));
    expect(s.timer).toBeNull();
    expect(s.expanded).toEqual({ seatID: "bob", pinned: true });
  });
});

describe("a peek closes", () => {
  const peeked = () => fire(run(enter("bob")));

  it("once the pointer has been outside the avatar and the overlay for the grace", () => {
    const s = reduceExpand(peeked(), leave("bob"));
    expect(s.expanded).not.toBeNull();
    expect(s.timer).toMatchObject({ kind: "close", ms: PEEK_CLOSE_GRACE_MS });
    expect(fire(s).expanded).toBeNull();
  });

  it("not when the pointer crosses the gap into the overlay inside the grace", () => {
    let s = reduceExpand(peeked(), leave("bob"));
    const stale = s.timer!.seq;
    s = reduceExpand(s, { type: "overlay-enter" });
    expect(s.timer).toBeNull();
    // The old timer firing late changes nothing.
    s = reduceExpand(s, { type: "elapsed", seq: stale });
    expect(s.expanded).toEqual({ seatID: "bob", pinned: false });
  });

  it("not when the pointer comes back to the avatar inside the grace", () => {
    const s = [leave("bob"), enter("bob")].reduce(reduceExpand, peeked());
    expect(s.timer).toBeNull();
    expect(s.expanded?.seatID).toBe("bob");
  });

  it("when the pointer leaves the overlay too", () => {
    const s = [
      leave("bob"),
      { type: "overlay-enter" } as const,
      { type: "overlay-leave" } as const,
    ].reduce(reduceExpand, peeked());
    expect(fire(s).expanded).toBeNull();
  });

  it("on Escape", () => {
    expect(reduceExpand(peeked(), { type: "escape" }).expanded).toBeNull();
  });
});

describe("a press consumes the hover", () => {
  it("cancels the pending peek", () => {
    const s = run(enter("bob"), press("bob"));
    expect(s.timer).toBeNull();
    expect(s.pendingSeat).toBeNull();
  });

  it("and the avatar starts no other until the pointer has left it", () => {
    let s = run(enter("bob"), press("bob"), leave("bob"));
    expect(s.consumed).toBeNull();
    s = reduceExpand(s, enter("bob"));
    expect(s.timer).toMatchObject({ kind: "open" });
  });

  it("a re-entry without leaving does not peek", () => {
    // Pointer events can repeat an enter without a leave (a child
    // re-rendering under the pointer); a consumed avatar ignores it.
    const s = run(enter("bob"), press("bob"), enter("bob"));
    expect(s.timer).toBeNull();
  });
});

describe("click pins", () => {
  it("pins the seat clicked, at once", () => {
    expect(run(enter("bob"), press("bob"), click("bob")).expanded).toEqual({
      seatID: "bob",
      pinned: true,
    });
  });

  it("pins an open peek in place", () => {
    const s = reduceExpand(fire(run(enter("bob"))), click("bob"));
    expect(s.expanded).toEqual({ seatID: "bob", pinned: true });
  });

  it("stays open when the pointer leaves", () => {
    const s = run(enter("bob"), click("bob"), leave("bob"), { type: "overlay-leave" });
    expect(s.timer).toBeNull();
    expect(s.expanded?.pinned).toBe(true);
  });

  it("clicking the pinned seat again closes it, and its hover stays used up", () => {
    let s = run(enter("bob"), press("bob"), click("bob"), press("bob"), click("bob"));
    expect(s.expanded).toBeNull();
    s = reduceExpand(s, enter("bob"));
    expect(s.timer).toBeNull();
    s = [leave("bob"), enter("bob")].reduce(reduceExpand, s);
    expect(s.timer).toMatchObject({ kind: "open" });
  });

  it("another seat's click moves the pin", () => {
    const s = run(click("bob"), click("cat"));
    expect(s.expanded).toEqual({ seatID: "cat", pinned: true });
  });

  it("an intercept beats the pin: only the click it did not take reaches the reducer", () => {
    // PlayerIdentity runs the target and attack intercepts first and
    // sends avatar-click only when neither applies. A press alone (the
    // pointerdown of a click that targeted the player) pins nothing.
    const s = run(enter("bob"), press("bob"));
    expect(s.expanded).toBeNull();
    expect(fire(run(enter("cat")))).toMatchObject({ expanded: { seatID: "cat" } });
  });
});

describe("the pin button and the expand button", () => {
  it("the pin button pins a peek", () => {
    const s = reduceExpand(fire(run(enter("bob"))), { type: "pin-toggle" });
    expect(s.expanded).toEqual({ seatID: "bob", pinned: true });
    expect(s.timer).toBeNull();
  });

  it("the pin button on a pinned board unpins, which closes it", () => {
    const s = run({ type: "expand", seatID: "bob" }, { type: "pin-toggle" });
    expect(s.expanded).toBeNull();
  });

  it("the expand button opens pinned", () => {
    expect(run({ type: "expand", seatID: "bob" }).expanded).toEqual({
      seatID: "bob",
      pinned: true,
    });
  });

  it("Escape closes a pinned board too", () => {
    expect(run({ type: "expand", seatID: "bob" }, { type: "escape" }).expanded).toBeNull();
  });

  it("Escape with the pointer on the avatar uses its hover up", () => {
    const s = run(enter("bob"), click("bob"), leave("bob"), enter("bob"), { type: "escape" });
    expect(s.expanded).toBeNull();
    expect(reduceExpand(s, enter("bob")).timer).toBeNull();
  });
});

describe("a seat that leaves drops the state", () => {
  it("is a derivation over the seats at the table", () => {
    const e = { seatID: "bob", pinned: true };
    expect(liveExpanded(e, ["me", "bob"])).toBe(e);
    expect(liveExpanded(e, ["me", "cat"])).toBeNull();
    expect(liveExpanded(null, ["me"])).toBeNull();
  });
});

describe("placement (ADR 0120 §2)", () => {
  const avatar = (left: number) => ({ left, top: 40, width: 84, height: 84 });
  const overlaps = (span: { left: number; width: number }, a: { left: number; width: number }) =>
    span.left < a.left + a.width && a.left < span.left + span.width;

  it("takes the wider side, hugging the avatar, with an 8px gap", () => {
    // A left-hand seat's rail: the avatar sits mid-board.
    const a = avatar(800);
    const s = placeOverlay(a, { boardWidth: 1800 });
    expect(s.side).toBe("right");
    expect(s.left).toBe(800 + 84 + 8);
    expect(s.width).toBe(1800 - (800 + 84 + 8));
  });

  it("opens to the left for a right-hand seat, capped at 1200px", () => {
    const a = avatar(1700);
    const s = placeOverlay(a, { boardWidth: 1800 });
    expect(s.side).toBe("left");
    expect(s.width).toBe(1200);
    expect(s.left + s.width).toBe(1700 - 8);
  });

  it("never covers the avatar's box", () => {
    for (const left of [0, 50, 400, 858, 900, 1300, 1716]) {
      const a = avatar(left);
      const s = placeOverlay(a, { boardWidth: 1800, edge: 6 });
      expect(overlaps(s, a), `avatar at ${left}`).toBe(false);
      expect(s.left).toBeGreaterThanOrEqual(6);
      expect(s.left + s.width).toBeLessThanOrEqual(1800 - 6);
    }
  });

  it("starts right of the stack pile's clearance", () => {
    const s = placeOverlay(avatar(900), { boardWidth: 1000, leftClear: 300 });
    expect(s.side).toBe("left");
    expect(s.left).toBe(300);
    expect(s.width).toBe(900 - 8 - 300);
  });

  it("reads the pile's clearance as px, and unset as 0", () => {
    expect(parsePx("212px")).toBe(212);
    expect(parsePx(" 0 ")).toBe(0);
    expect(parsePx("")).toBe(0);
    expect(parsePx(undefined)).toBe(0);
  });
});
