// #4552: where a tombstone sits on its seat's panel. One rule for every
// seat: centred in the part of the panel the identity leaves free.

import { describe, expect, it } from "vitest";
import {
  TOMBSTONE_FULL_HEIGHT,
  TOMBSTONE_GAP,
  TOMBSTONE_MARGIN,
  tombstoneBand,
  tombstoneFits,
} from "./tombstoneAnchor";

// A panel from y=100 to y=700 (600 tall).
const panel = { top: 100, bottom: 700 };

describe("tombstoneBand", () => {
  it("is the whole panel when there is no identity to keep clear", () => {
    expect(tombstoneBand(panel, [])).toEqual({ top: 0, bottom: 0 });
    // A box with no size is not there.
    expect(tombstoneBand(panel, [{ top: 120, bottom: 120 }])).toEqual({ top: 0, bottom: 0 });
  });

  it("starts under an identity against the top edge (a top-row opponent)", () => {
    // Avatar and badges from 108 to 230.
    const b = tombstoneBand(panel, [{ top: 108, bottom: 230 }]);
    expect(b).toEqual({ top: 230 + TOMBSTONE_GAP - 100, bottom: TOMBSTONE_MARGIN });
  });

  it("ends above an identity against the bottom edge (a bottom-row seat or the viewer)", () => {
    // Avatar and badges from 560 to 692.
    const b = tombstoneBand(panel, [{ top: 560, bottom: 692 }]);
    expect(b).toEqual({ top: TOMBSTONE_MARGIN, bottom: 700 - (560 - TOMBSTONE_GAP) });
  });

  it("keeps clear of every part of the identity: the union of the boxes", () => {
    // The name above the avatar, and a column of counters low on its side.
    const b = tombstoneBand(panel, [
      { top: 104, bottom: 120 },
      { top: 120, bottom: 210 },
      { top: 150, bottom: 260 },
    ]);
    expect(b.top).toBe(260 + TOMBSTONE_GAP - 100);
  });

  it("puts the band on the same side of the identity whatever the panel's height", () => {
    // The short panel of a top-row seat and a tall one: the identity is at the
    // top of both, so the stone is under it in both, never over it.
    for (const height of [220, 300, 600, 900]) {
      const p = { top: 0, bottom: height };
      const b = tombstoneBand(p, [{ top: 8, bottom: 130 }]);
      expect(b.top).toBe(130 + TOMBSTONE_GAP);
      expect(b.top).toBeGreaterThan(b.bottom);
    }
  });

  it("never gives a negative inset or one past the panel", () => {
    const high = tombstoneBand(panel, [{ top: 90, bottom: 400 }]);
    expect(high.top).toBeGreaterThanOrEqual(0);
    expect(high.bottom).toBeGreaterThanOrEqual(0);
    const tall = tombstoneBand({ top: 0, bottom: 100 }, [{ top: 0, bottom: 5000 }]);
    // An identity taller than the panel leaves no room: the whole panel is its.
    expect(tall.top + tall.bottom).toBe(100);
    expect(tombstoneBand({ top: 0, bottom: 0 }, [{ top: 0, bottom: 10 }])).toEqual({
      top: 0,
      bottom: 0,
    });
  });
});

describe("tombstoneFits", () => {
  it("is true when the band is as tall as the stone, false below it", () => {
    const p = { top: 0, bottom: 400 };
    expect(tombstoneFits(p, { top: 400 - TOMBSTONE_FULL_HEIGHT, bottom: 0 })).toBe(true);
    expect(tombstoneFits(p, { top: 400 - TOMBSTONE_FULL_HEIGHT + 1, bottom: 0 })).toBe(false);
    expect(tombstoneFits(p, { top: 0, bottom: 0 })).toBe(true);
    // A panel not laid out yet has no height to compare.
    expect(tombstoneFits({ top: 0, bottom: 0 }, { top: 0, bottom: 0 })).toBe(true);
  });
});
