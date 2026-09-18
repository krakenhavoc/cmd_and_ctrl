import { describe, expect, it } from "vitest";

import {
  LIFE_POPUP_HOLD_MAX_MS,
  LIFE_POPUP_HOLD_MS,
  LIFE_POPUP_HOLD_PER_EXTRA_MS,
  LIFE_POPUP_MAX_LINES,
  emptyLifeTracker,
  lifePopupLabel,
  lifePopupView,
  trackLife,
} from "./lifePopup";
import type { LifeChangeView } from "./protocol";

// history builds a life log from a list of deltas, the way the server
// does: a running total, one seq per entry starting at `firstSeq`, and
// only the newest `cap` entries kept (the MaxLifeHistoryEntries trim).
function history(
  deltas: readonly number[],
  opts: { startLife?: number; firstSeq?: number; cap?: number } = {},
): LifeChangeView[] {
  const { startLife = 40, firstSeq = 1, cap = Infinity } = opts;
  let life = startLife;
  const out: LifeChangeView[] = deltas.map((delta, i) => {
    life += delta;
    return {
      delta,
      new_total: life,
      // RFC3339 SECONDS, as the wire really carries it: every entry
      // here shares a timestamp on purpose.
      at: "2026-09-16T12:00:00Z",
      seq: firstSeq + i,
    };
  });
  return cap === Infinity ? out : out.slice(Math.max(0, out.length - cap));
}

describe("trackLife", () => {
  it("primes on the first frame: no popup for changes that already happened", () => {
    const frame = trackLife(emptyLifeTracker(), history([-5, +2, -3]));
    expect(frame.primed).toBe(true);
    expect(frame.popup).toBeNull();
    expect(frame.tracker.maxSeq).toBe(3);
  });

  it("primes on an empty log, then pops the first real change", () => {
    const primed = trackLife(emptyLifeTracker(), []);
    expect(primed.popup).toBeNull();
    expect(primed.tracker.maxSeq).toBe(0);

    const next = trackLife(primed.tracker, history([-4]));
    expect(next.popup?.deltas).toEqual([-4]);
    expect(next.popup?.key).toBe(1);
  });

  it("primes on a null / undefined log without throwing", () => {
    expect(trackLife(emptyLifeTracker(), null).popup).toBeNull();
    expect(trackLife(emptyLifeTracker(), undefined).primed).toBe(true);
  });

  it("shows EVERY delta a frame carries, not just the last (double strike)", () => {
    // A 3/3 double striker hits an open player: both combat damage
    // steps land in one frame. The old length-diff code showed "-3".
    const primed = trackLife(emptyLifeTracker(), history([-1]));
    const frame = trackLife(primed.tracker, history([-1, -3, -3]));
    expect(frame.popup?.deltas).toEqual([-3, -3]);
    expect(frame.popup?.total).toBe(-6);
    expect(frame.popup?.newTotal).toBe(33);
    expect(frame.popup?.gap).toBe(false);
  });

  it("keeps popping once the log is trimmed at its cap", () => {
    const cap = 50;
    // 50 changes have happened and the log is full.
    const full = history(Array(cap).fill(-1), { cap });
    const primed = trackLife(emptyLifeTracker(), full);
    expect(primed.tracker.maxSeq).toBe(cap);

    // Two more land. The log LENGTH is unchanged — the old code's
    // `count <= baselineCount` was true here and nothing ever popped
    // again — but the seqs moved.
    const after = history([...Array(cap).fill(-1), -2, -4], { cap });
    expect(after).toHaveLength(cap);
    const frame = trackLife(primed.tracker, after);
    expect(frame.popup?.deltas).toEqual([-2, -4]);
    expect(frame.popup?.key).toBe(cap + 2);
  });

  it("flags a gap when the changes it missed were trimmed off the front", () => {
    const primed = trackLife(emptyLifeTracker(), history([-1, -1]));
    // The client missed frames; entries 3-5 have rolled off and only
    // 6 and 7 survive.
    const frame = trackLife(primed.tracker, history([-3, -4], { firstSeq: 6 }));
    expect(frame.popup?.deltas).toEqual([-3, -4]);
    expect(frame.popup?.gap).toBe(true);
    expect(lifePopupLabel(frame.popup!)).toContain("earlier changes not shown");
  });

  it("does not flag a gap when the frame carries the very next change", () => {
    const primed = trackLife(emptyLifeTracker(), history([-1, -1]));
    const frame = trackLife(primed.tracker, history([-1, -1, -2], { cap: 1 }));
    expect(frame.popup?.deltas).toEqual([-2]);
    expect(frame.popup?.gap).toBe(false);
  });

  it("raises no popup for a frame that carries no new change", () => {
    const log = history([-1, -2]);
    const primed = trackLife(emptyLifeTracker(), log);
    const again = trackLife(primed.tracker, log);
    expect(again.popup).toBeNull();
    expect(again.tracker.maxSeq).toBe(2);
  });

  it("treats a shrinking seq stream as a rewind and pops nothing", () => {
    const primed = trackLife(emptyLifeTracker(), history([-1, -2, -3]));
    const undone = trackLife(primed.tracker, history([-1, -2]));
    expect(undone.rewound).toBe(true);
    expect(undone.popup).toBeNull();
    expect(undone.tracker.maxSeq).toBe(2);

    // The replayed change pops when it comes back.
    const redone = trackLife(undone.tracker, history([-1, -2, -3]));
    expect(redone.popup?.deltas).toEqual([-3]);
  });

  it("ignores zero deltas and orders a frame's deltas oldest first", () => {
    const primed = trackLife(emptyLifeTracker(), []);
    const shuffled: LifeChangeView[] = [
      { delta: +2, new_total: 42, at: "t", seq: 3 },
      { delta: 0, new_total: 40, at: "t", seq: 2 },
      { delta: -5, new_total: 35, at: "t", seq: 1 },
    ];
    const frame = trackLife(primed.tracker, shuffled);
    expect(frame.popup?.deltas).toEqual([-5, +2]);
    expect(frame.popup?.key).toBe(3);
    expect(frame.tracker.maxSeq).toBe(3);
  });

  it("shows nothing for a pre-#703 log whose entries carry no seq", () => {
    // Every entry reads seq 0, so priming takes the watermark to 0
    // and nothing re-pops; the first change stamped after the upgrade
    // gets seq 1 and pops normally.
    const legacy: LifeChangeView[] = [
      { delta: -3, new_total: 37, at: "t", seq: 0 },
      { delta: -3, new_total: 34, at: "t", seq: 0 },
    ];
    const primed = trackLife(emptyLifeTracker(), legacy);
    expect(primed.tracker.maxSeq).toBe(0);
    expect(trackLife(primed.tracker, legacy).popup).toBeNull();

    const upgraded = [...legacy, { delta: -2, new_total: 32, at: "t", seq: 1 }];
    expect(trackLife(primed.tracker, upgraded).popup?.deltas).toEqual([-2]);
  });
});

describe("lifePopupView", () => {
  const popup = (deltas: number[], gap = false) => {
    const primed = trackLife(emptyLifeTracker(), []);
    const frame = trackLife(primed.tracker, history(deltas, { firstSeq: gap ? 5 : 1 }));
    return frame.popup!;
  };

  it("renders one line per delta, oldest first", () => {
    const view = lifePopupView(popup([-3, -3]));
    expect(view.lines).toEqual([
      { text: "-3", tone: "loss" },
      { text: "-3", tone: "loss" },
    ]);
    expect(view.collapsed).toBe(false);
    expect(view.count).toBe(2);
  });

  it("signs a gain and tones it apart from a loss", () => {
    const view = lifePopupView(popup([-2, +5]));
    expect(view.lines).toEqual([
      { text: "-2", tone: "loss" },
      { text: "+5", tone: "gain" },
    ]);
  });

  it("collapses to the net total past LIFE_POPUP_MAX_LINES", () => {
    const deltas = Array(LIFE_POPUP_MAX_LINES + 1).fill(-2);
    const view = lifePopupView(popup(deltas));
    expect(view.collapsed).toBe(true);
    expect(view.count).toBe(deltas.length);
    expect(view.lines).toEqual([{ text: `${-2 * deltas.length}`, tone: "loss" }]);
  });

  it("stacks right up to the limit without collapsing", () => {
    const view = lifePopupView(popup(Array(LIFE_POPUP_MAX_LINES).fill(-1)));
    expect(view.collapsed).toBe(false);
    expect(view.lines).toHaveLength(LIFE_POPUP_MAX_LINES);
  });

  it("plays the damage cue whenever anything was lost, the heal cue otherwise", () => {
    expect(lifePopupView(popup([-3])).sound).toBe("damage");
    expect(lifePopupView(popup([+3, -1])).sound).toBe("damage");
    expect(lifePopupView(popup([+3, +4])).sound).toBe("heal");
  });

  it("holds longer for more lines, up to the cap", () => {
    expect(lifePopupView(popup([-1])).holdMs).toBe(LIFE_POPUP_HOLD_MS);
    expect(lifePopupView(popup([-1, -1])).holdMs).toBe(
      LIFE_POPUP_HOLD_MS + LIFE_POPUP_HOLD_PER_EXTRA_MS,
    );
    // A collapsed popup is one line again, so it holds the base time.
    expect(lifePopupView(popup(Array(20).fill(-1))).holdMs).toBe(LIFE_POPUP_HOLD_MS);
    expect(lifePopupView(popup([-1, -1, -1, -1, -1, -1, -1, -1]), 8).holdMs).toBeLessThanOrEqual(
      LIFE_POPUP_HOLD_MAX_MS,
    );
  });

  it("announces the whole group as one sentence", () => {
    expect(lifePopupView(popup([-3])).label).toBe("lost 3 life, now 37");
    expect(lifePopupView(popup([-3, -3])).label).toBe(
      "lost 3, lost 3 — lost 6 life in all, now 34",
    );
    expect(lifePopupView(popup([-3, +3])).label).toBe("lost 3, gained 3 — no net change, now 40");
  });
});
