// tableMoment.test.ts — the quiet moment at the table (ADR 0125 §3.5,
// §8): each rule against stub moments, including an opponent's item on
// top while the viewer holds priority (not quiet) and the viewer's own
// item on top (quiet).

import { describe, expect, it } from "vitest";
import { get } from "svelte/store";

import {
  TABLE_GAP_MS,
  TABLE_SETTLE_MS,
  _resetTableMomentForTests,
  isQuiet,
  notQuietReason,
  publishTableMoment,
  stackTopController,
  tableState,
} from "./tableMoment";
import { moment } from "../test/hintContexts";
import { BOT, ME, board, elves } from "../test/tutorialBoards";

// Long after the table came on screen, and no hint has closed.
const NOW = 1_000_000;

describe("the quiet moment", () => {
  it("is quiet with nothing open, nothing held and an empty stack", () => {
    expect(notQuietReason(moment(), NOW, null)).toBeNull();
    expect(isQuiet(moment(), NOW, null)).toBe(true);
  });

  it("rule 1: not while a dock request is open for the viewer", () => {
    expect(notQuietReason(moment({ dockRequest: true }), NOW, null)).toBe("dock-request");
  });

  it("rule 1: not while a dialog or a sheet is open", () => {
    expect(notQuietReason(moment({ dialog: true }), NOW, null)).toBe("dialog");
  });

  it("rule 2: not mid-gesture (targeting, a drag, an open menu)", () => {
    expect(notQuietReason(moment({ gesture: true }), NOW, null)).toBe("gesture");
  });

  it("rule 3: not while an opponent's item is on top and the viewer holds priority", () => {
    const m = moment({ viewerHasPriority: true, stackTopController: BOT });
    expect(notQuietReason(m, NOW, null)).toBe("opponent-stack");
  });

  it("rule 3: the viewer's own item on top is quiet", () => {
    const m = moment({ viewerHasPriority: true, stackTopController: ME });
    expect(isQuiet(m, NOW, null)).toBe(true);
  });

  it("rule 3: an opponent's item while someone else holds priority is quiet", () => {
    const m = moment({ viewerHasPriority: false, stackTopController: BOT });
    expect(isQuiet(m, NOW, null)).toBe(true);
  });

  it("rule 4: the table has been on screen for five seconds", () => {
    const m = moment({ since: NOW - TABLE_SETTLE_MS + 1 });
    expect(notQuietReason(m, NOW, null)).toBe("settling");
    expect(isQuiet(moment({ since: NOW - TABLE_SETTLE_MS }), NOW, null)).toBe(true);
  });

  it("rule 4: no hint has closed in the last twenty seconds", () => {
    expect(notQuietReason(moment(), NOW, NOW - TABLE_GAP_MS + 1)).toBe("recent-hint");
    expect(isQuiet(moment(), NOW, NOW - TABLE_GAP_MS)).toBe(true);
  });

  it("is off while the tutorial coach is visible", () => {
    expect(notQuietReason(moment({ coachVisible: true }), NOW, null)).toBe("coach");
  });
});

describe("stackTopController", () => {
  it("reads the top item's controller, or null for an empty stack", () => {
    const v = board({ stack: [elves(), elves()] });
    expect(stackTopController(v)).toBeNull();
    v.stack_items = [
      { id: "a", kind: "spell", controller: ME, owner: ME, source_card_id: "x" },
      { id: "b", kind: "triggered", controller: BOT, owner: BOT, source_card_id: "y" },
    ];
    expect(stackTopController(v)).toBe(BOT);
    expect(stackTopController(null)).toBeNull();
  });
});

describe("publishTableMoment", () => {
  it("publishes, updates and takes down a table, and an older table cannot close a newer one", () => {
    _resetTableMomentForTests();
    const first = publishTableMoment({ moment: moment(), view: null, viewerID: ME });
    expect(get(tableState)?.moment.dockRequest).toBe(false);
    first.update({ moment: moment({ dockRequest: true }), view: null, viewerID: ME });
    expect(get(tableState)?.moment.dockRequest).toBe(true);

    // A keyed remount: the new table registers before the old one's cleanup.
    const second = publishTableMoment({
      moment: moment({ gesture: true }),
      view: null,
      viewerID: ME,
    });
    first.close();
    first.update({ moment: moment(), view: null, viewerID: ME });
    expect(get(tableState)?.moment.gesture).toBe(true);
    second.close();
    expect(get(tableState)).toBeNull();
  });
});
