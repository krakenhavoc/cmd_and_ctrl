// seatNotch.test.ts — ADR 0204: what a seated avatar keeps clear.

import { describe, expect, it } from "vitest";
import {
  SEAT_BACK_GAP,
  SEAT_CLEARANCE,
  PANEL_PAD,
  backRowSplit,
  partBackRow,
  handNotch,
  seatOverhang,
  slotDips,
} from "./seatNotch";

describe("seatOverhang (§2, §3, call 3)", () => {
  it("a top-row identity reaching past its hand's row keeps that depth clear below it", () => {
    // Hand row 8–60, gap 6: the back row starts at 66. The life chip
    // ends at 110, so 110 + clearance − 66.
    const v = seatOverhang({
      identity: { top: 2, bottom: 110 },
      handRow: { top: 8, bottom: 60 },
      rowGap: 6,
      handOnTop: true,
    });
    expect(v).toBe(110 + SEAT_CLEARANCE - 66);
  });

  it("is zero when the identity stays inside its hand's row", () => {
    expect(
      seatOverhang({
        identity: { top: 2, bottom: 50 },
        handRow: { top: 8, bottom: 70 },
        rowGap: 6,
        handOnTop: true,
      }),
    ).toBe(0);
  });

  it("an upright identity reaching up past its hand's row keeps that depth clear above it", () => {
    // Hand row 300–360, gap 6: the back row ends at 294. The name starts at 240.
    const v = seatOverhang({
      identity: { top: 240, bottom: 356 },
      handRow: { top: 300, bottom: 360 },
      rowGap: 6,
      handOnTop: false,
    });
    expect(v).toBe(294 - (240 - SEAT_CLEARANCE));
  });
});

describe("handNotch (§3, call 4)", () => {
  it("puts the avatar in the strip's x and dips a card below its lowest point", () => {
    const n = handNotch({
      avatar: { left: 500, right: 584 },
      avatarBottom: 812,
      handTop: 800,
      handLeft: 180,
    });
    expect(n).toEqual({
      left: 500 - 180 - SEAT_CLEARANCE,
      right: 584 - 180 + SEAT_CLEARANCE,
      depth: 812 + SEAT_CLEARANCE - 800,
    });
  });

  it("is null when the avatar ends above the hand", () => {
    expect(
      handNotch({
        avatar: { left: 500, right: 584 },
        avatarBottom: 760,
        handTop: 800,
        handLeft: 0,
      }),
    ).toBeNull();
  });
});

describe("slotDips", () => {
  const notch = { left: 300, right: 400, depth: 20 };

  it("dips exactly the cards whose slot reaches under the avatar", () => {
    const slots = [
      { left: 100, right: 200 },
      { left: 180, right: 280 },
      { left: 260, right: 360 },
      { left: 340, right: 440 },
      { left: 420, right: 520 },
    ];
    // The second card's edge is within the tilt allowance (6px) of the notch.
    expect(slotDips(slots, notch)).toEqual([0, 0, 20, 20, 0]);
    expect(slotDips(slots, notch, 30)).toEqual([0, 20, 20, 20, 20]);
  });

  it("dips nothing without a notch", () => {
    expect(slotDips([{ left: 0, right: 1000 }], null)).toEqual([0]);
  });
});

describe("backRowSplit (§3, call 3)", () => {
  it("parts the back row around the avatar: a part each side, the right one up to the dock", () => {
    // A 1008px board: the avatar and its steppers at 312..471, the dock at 684.
    const s = backRowSplit({
      panelW: 1008,
      start: 8,
      avatar: { left: 312, right: 471 },
      dockLeft: 684,
    });
    expect(s).toEqual({
      left: 312 - SEAT_BACK_GAP - 8,
      notch: 471 + SEAT_BACK_GAP - (312 - SEAT_BACK_GAP),
      zones: 684 - PANEL_PAD - (471 + SEAT_BACK_GAP),
      right: 1008 - 684,
    });
  });

  it("starts the left part after the coach card's room", () => {
    const s = backRowSplit({
      panelW: 1008,
      start: 8 + 200,
      avatar: { left: 312, right: 471 },
      dockLeft: 684,
    });
    expect(s.left).toBe(312 - SEAT_BACK_GAP - 208);
  });

  it("never sizes a part below zero", () => {
    const s = backRowSplit({
      panelW: 600,
      start: 200,
      avatar: { left: 100, right: 250 },
      dockLeft: 240,
    });
    expect(s.left).toBe(0);
    expect(s.notch).toBe(250 + SEAT_BACK_GAP - 200);
    expect(s.zones).toBe(0);
  });
});

describe("partBackRow (§3, call 3)", () => {
  const zones = [
    { zone: "planeswalker", weight: 8 },
    { zone: "battle", weight: 2 },
    { zone: "artifact", weight: 3 },
    { zone: "enchantment", weight: 3 },
  ];

  it("shares the groups by room, in order, the lands always left", () => {
    // 1 land pile of 17 in all; the left part has 54% of the room.
    const p = partBackRow(1, zones, 430, 362);
    expect(p).toEqual({ left: ["planeswalker"], right: ["battle", "artifact", "enchantment"] });
  });

  it("keeps everything left when the right part has no room", () => {
    expect(partBackRow(1, zones, 400, 0)).toEqual({
      left: ["planeswalker", "battle", "artifact", "enchantment"],
      right: [],
    });
  });

  it("sends everything right when the left part has none", () => {
    expect(partBackRow(0, zones, 0, 400)).toEqual({
      left: [],
      right: ["planeswalker", "battle", "artifact", "enchantment"],
    });
  });
});
