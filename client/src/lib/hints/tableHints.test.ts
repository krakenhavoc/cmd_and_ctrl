// tableHints.test.ts — the table's first-use hints (ADR 0125 §3.7,
// Delivery PR 7): the nine ids the tutorial will name, each hint's
// condition and anchor over the fixture boards, that none is offered
// while the table's moment is not quiet, and that a hint whose feature
// is not on screen is never drawn. The DOM half, that each anchor
// resolves on a real board, is tableHints.render.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { HINTS } from "./index";
import { anchorOf, emptyContext, holds, type Hint, type HintContext } from "./hint";
import { createHintEngine } from "./queue";
import { TABLE_GAP_MS, TABLE_SETTLE_MS, notQuietReason } from "./tableMoment";
import { atViewersSecondTurn, firstOpponent } from "./tableWhen";
import { COPY_CONTEXTS } from "./rules";
import { copyText } from "../tutorial";
import { L } from "../labels";
import type { CardView, GameView, PlayerView } from "../protocol";
import { moment } from "../test/hintContexts";
import { BOT, ME, board, elves, forest } from "../test/tutorialBoards";

const TABLE = HINTS.filter((h) => h.place === "table");
const byID = (id: string): Hint => {
  const h = TABLE.find((x) => x.id === id);
  if (!h) throw new Error(`no hint ${id}`);
  return h;
};

/** A seat shaped like the wire's, for the boards below. */
function seat(id: string, name: string, n: number, commander = false): PlayerView {
  const zone = (kind: string, cards: CardView[] = []) => ({ kind, count: cards.length, cards });
  return {
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library"),
    hand: zone("hand"),
    graveyard: zone("graveyard"),
    command: zone("command", commander ? [elves({ instance_id: `cmd-${id}` })] : []),
  } as unknown as PlayerView;
}

const withAbility = (): CardView =>
  forest({
    controller: ME,
    activated_abilities: [{ index: 0, ref: "own:0", label: "{T}: Add {G}." }],
  } as Partial<CardView>);

/** A four-seat table at the viewer's third turn, with something for every hint. */
function fullView(over: Partial<GameView> = {}): GameView {
  const v = board({ mine: [withAbility()] });
  v.seats = [
    seat(ME, "Player", 0, true),
    seat(BOT, "Practice Bot", 1),
    seat("c", "Cat", 2),
    seat("d", "Dee", 3),
  ];
  v.turn.number = 3;
  return { ...v, ...over };
}

function ctxFor(view: GameView | null, over: Partial<HintContext> = {}): HintContext {
  return {
    ...emptyContext("table", "game"),
    moment: moment(),
    view,
    viewerID: ME,
    ...over,
  };
}

const item = (controller: string) => ({
  id: "i1",
  kind: "spell" as const,
  controller,
  owner: controller,
  source_card_id: "c1",
});

const withStack = (controller: string): GameView => {
  const v = fullView();
  v.stack_items = [item(controller)];
  v.stack = { kind: "stack", count: 1, cards: [elves()] };
  return v;
};

describe("the table's hints", () => {
  it("are the nine the ADR names, at version 1, under the ids the tutorial will teach", () => {
    expect(TABLE.map((h) => h.id).sort()).toEqual(
      [
        "table.attention",
        "table.commander",
        "table.dock",
        "table.expand",
        "table.more",
        "table.opening-roll",
        "table.right-click",
        "table.shortcuts",
        "table.stack",
      ].sort(),
    );
    for (const h of TABLE) {
      expect(h.version, h.id).toBe(1);
      expect(h.place, h.id).toBe("table");
    }
    expect(new Set(TABLE.map((h) => h.order)).size).toBe(TABLE.length);
  });

  it("are ordered dock first and shortcuts last, so the dock's has been shown first", () => {
    const ids = [...TABLE].sort((a, b) => a.order - b.order).map((h) => h.id);
    expect(ids[0]).toBe("table.dock");
    expect(ids.at(-1)).toBe("table.shortcuts");
    expect(ids.indexOf("table.more")).toBeLessThan(ids.indexOf("table.shortcuts"));
  });

  it("anchor on the contract labels the table renders", () => {
    const c = ctxFor(fullView());
    expect(anchorOf(byID("table.dock"), c)).toEqual({ label: L.actions });
    expect(anchorOf(byID("table.opening-roll"), c)).toEqual({ label: L.openingRoll });
    expect(anchorOf(byID("table.stack"), c)).toEqual({ label: L.stackPile.any });
    expect(anchorOf(byID("table.right-click"), c)).toEqual({ label: L.yourBoard });
    expect(anchorOf(byID("table.commander"), c)).toEqual({
      label: L.commandZone.any,
      within: L.yourBoard,
    });
    expect(anchorOf(byID("table.attention"), c)).toEqual({ label: L.attention });
    expect(anchorOf(byID("table.more"), c)).toEqual({ label: L.moreActions });
    expect(anchorOf(byID("table.expand"), c)).toEqual({ label: L.seatBoard("Practice Bot") });
    expect(anchorOf(byID("table.shortcuts"), c)).toEqual({ label: L.actions });
  });

  it("name the first opponent still in the game for the expanded board", () => {
    const v = fullView();
    v.seats[1] = { ...v.seats[1], eliminated: true };
    expect(anchorOf(byID("table.expand"), ctxFor(v))).toEqual({ label: L.seatBoard("Cat") });
    expect(firstOpponent(ctxFor(null))).toBeNull();
    expect(anchorOf(byID("table.expand"), ctxFor(null))).toBeNull();
  });
});

describe("each hint's own condition", () => {
  beforeEach(() => {
    vi.stubGlobal("matchMedia", (q: string) => ({ matches: q.includes("hover: hover") }));
  });
  afterEach(() => vi.unstubAllGlobals());

  const when = (id: string, view: GameView | null, over: Partial<HintContext> = {}) =>
    holds(byID(id), ctxFor(view, over));

  it("table.dock is always eligible", () => {
    expect(when("table.dock", fullView())).toBe(true);
  });

  it("table.opening-roll waits until the viewer has rolled and someone has not", () => {
    const roll = (rolls: { seat: number; result: number }[], chooser?: number): GameView => {
      const v = fullView();
      v.opening_roll = {
        rounds: [{ seats: [0, 1, 2, 3], rolls }],
        ...(chooser === undefined ? {} : { chooser }),
      } as never;
      return v;
    };
    expect(when("table.opening-roll", fullView())).toBe(false);
    expect(when("table.opening-roll", roll([]))).toBe(false);
    expect(when("table.opening-roll", roll([{ seat: 1, result: 9 }]))).toBe(false);
    expect(when("table.opening-roll", roll([{ seat: 0, result: 9 }]))).toBe(true);
    const all = [0, 1, 2, 3].map((seat) => ({ seat, result: 5 + seat }));
    expect(when("table.opening-roll", roll(all, 3))).toBe(false);
    expect(when("table.opening-roll", roll([{ seat: 0, result: 9 }]), { viewerID: null })).toBe(
      false,
    );
  });

  it("table.stack needs an item on the stack", () => {
    expect(when("table.stack", fullView())).toBe(false);
    expect(when("table.stack", withStack(ME))).toBe(true);
  });

  it("table.right-click needs a permanent of the viewer's with an ability", () => {
    expect(when("table.right-click", fullView())).toBe(true);
    expect(when("table.right-click", board({ mine: [forest()] }))).toBe(false);
    const theirs = board({ theirs: [withAbility()] });
    expect(when("table.right-click", theirs)).toBe(false);
  });

  it("table.commander needs a commander in the zone", () => {
    expect(when("table.commander", fullView())).toBe(true);
    expect(when("table.commander", board())).toBe(false);
  });

  it("table.attention waits for an empty stack, and the strip's own area", () => {
    expect(when("table.attention", fullView())).toBe(true);
    expect(when("table.attention", withStack(BOT))).toBe(false);
  });

  it("table.more and table.shortcuts wait for the viewer's second turn", () => {
    const at = (number: number, active: number): GameView => {
      const v = fullView();
      v.turn.number = number;
      v.turn.active_seat = active;
      return v;
    };
    for (const id of ["table.more", "table.shortcuts"]) {
      expect(when(id, at(1, 0)), id).toBe(false);
      expect(when(id, at(2, 0)), id).toBe(true);
      expect(when(id, at(3, 0)), id).toBe(true);
    }
    // Seat 3's second turn is the end of round two, not its start.
    const v = at(2, 1);
    v.seats[0] = { ...v.seats[0], seat: 3 };
    expect(atViewersSecondTurn(ctxFor(v))).toBe(false);
  });

  it("table.shortcuts needs a hover-capable fine pointer", () => {
    expect(when("table.shortcuts", fullView())).toBe(true);
    vi.stubGlobal("matchMedia", () => ({ matches: false }));
    expect(when("table.shortcuts", fullView())).toBe(false);
  });

  it("table.shortcuts names the player's own keys", () => {
    const [d, rebound, none] = COPY_CONTEXTS.map((k) => copyText(byID("table.shortcuts").body, k));
    expect(d).toBe("Press ? to see every keyboard shortcut. Space presses next.");
    expect(rebound).toContain("Ctrl+Shift+F12");
    expect(rebound).toContain("Ctrl+Shift+Enter presses next.");
    expect(none).toMatch(/Settings/);
  });

  it("table.expand needs three or more seats", () => {
    expect(when("table.expand", fullView())).toBe(true);
    expect(when("table.expand", board())).toBe(false);
  });
});

describe("the table's hints and the quiet moment", () => {
  beforeEach(() => {
    vi.stubGlobal("matchMedia", (q: string) => ({ matches: q.includes("hover: hover") }));
  });
  afterEach(() => vi.unstubAllGlobals());

  const NOW = 10 * TABLE_SETTLE_MS;

  /** One engine poll over a table where every anchor is on screen. */
  function offered(c: HintContext, lastClosed: number | null = null): string | null {
    const engine = createHintEngine({ log: () => {} });
    if (lastClosed !== null) {
      engine.tick({
        hints: TABLE,
        ctx: ctxFor(fullView()),
        visit: "g",
        seen: {},
        tipsOff: false,
        now: lastClosed - 1,
        resolves: () => true,
      });
      engine.dismiss(lastClosed);
    }
    return (
      engine.tick({
        hints: TABLE,
        ctx: c,
        visit: "g",
        seen: {},
        tipsOff: false,
        now: NOW,
        resolves: () => true,
      })?.id ?? null
    );
  }

  it("offers the dock's hint first, at the first quiet moment", () => {
    expect(offered(ctxFor(fullView()))).toBe("table.dock");
  });

  it("offers nothing while the moment is not quiet", () => {
    const view = fullView();
    const loud: Record<string, HintContext> = {
      "dock request": ctxFor(view, { moment: moment({ dockRequest: true }) }),
      dialog: ctxFor(view, { moment: moment({ dialog: true }) }),
      gesture: ctxFor(view, { moment: moment({ gesture: true }) }),
      coach: ctxFor(view, { moment: moment({ coachVisible: true }) }),
      settling: ctxFor(view, { moment: moment({ since: NOW - TABLE_SETTLE_MS + 1 }) }),
      "an opponent's item with priority": ctxFor(withStack(BOT), {
        moment: moment({ viewerHasPriority: true, stackTopController: BOT }),
      }),
      "no table moment at all": ctxFor(view, { moment: null }),
    };
    for (const [name, c] of Object.entries(loud)) {
      expect(offered(c), name).toBeNull();
    }
    expect(notQuietReason(loud["dock request"].moment!, NOW, null)).toBe("dock-request");
  });

  it("offers nothing within the gap after a hint closed", () => {
    expect(offered(ctxFor(fullView()), NOW - TABLE_GAP_MS + 1_000)).toBeNull();
    expect(offered(ctxFor(fullView()), NOW - TABLE_GAP_MS - 1_000)).toBe("table.dock");
  });

  it("offers the stack's hint over the viewer's own item, not over an opponent's", () => {
    const pick = (c: HintContext) =>
      createHintEngine({ log: () => {} }).tick({
        hints: TABLE.filter((h) => h.id === "table.stack"),
        ctx: c,
        visit: "g",
        seen: {},
        tipsOff: false,
        now: NOW,
        resolves: () => true,
      })?.id ?? null;
    const mine = withStack(ME);
    expect(
      pick(ctxFor(mine, { moment: moment({ viewerHasPriority: true, stackTopController: ME }) })),
    ).toBe("table.stack");
    const theirs = withStack(BOT);
    expect(
      pick(
        ctxFor(theirs, { moment: moment({ viewerHasPriority: true, stackTopController: BOT }) }),
      ),
    ).toBeNull();
  });

  it("never offers a hint whose anchor is not on screen", () => {
    const engine = createHintEngine({ log: () => {} });
    const got = engine.tick({
      hints: TABLE,
      ctx: ctxFor(fullView()),
      visit: "g",
      seen: {},
      tipsOff: false,
      now: NOW,
      resolves: (h) => h.id !== "table.dock" && h.id !== "table.opening-roll",
    });
    expect(got?.id).toBe("table.right-click");
  });

  it("logs the missing anchor and keeps the hint unseen", () => {
    const log = vi.fn();
    const engine = createHintEngine({ log });
    engine.tick({
      hints: [byID("table.dock")],
      ctx: ctxFor(fullView()),
      visit: "g",
      seen: {},
      tipsOff: false,
      now: NOW,
      resolves: () => false,
    });
    expect(log).toHaveBeenCalledWith("hint: table.dock has no anchor on the page");
  });
});
