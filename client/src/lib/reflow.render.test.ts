// @vitest-environment jsdom
//
// reflow.render.test.ts — ADR 0077 §6 as amended 2026-10-16 (#4065):
// the "reflow" expand style on a real Board with four seats in the row
// layout. On an opponent's turn their board takes most of the row and
// the other two are avatar cards; on your own turn, with the overlay
// style, or with the active-player expansion off, the row is three
// equal boards; a seat a live prompt needs is never an avatar card; and
// an avatar card's avatar still opens the ADR 0120 overlay. The decision
// itself is pinned in expansion.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import { begin, targeting } from "./targeting";
import { _resetForTests as resetDock } from "./dock";
import { resetSettings, updateSettings } from "./settings";
import { findStandIn } from "./boardAnchor";
import type { CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { cleanup, flushSync, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  vi.useFakeTimers();
  resetSettings();
  updateSettings("display", "expandStyle", "reflow");
});

afterEach(() => {
  cleanup();
  targeting.set(null);
  resetDock();
  resetSettings();
  vi.useRealTimers();
});

const ME = "me";
const BOB = "bob";
const CAT = "cat";
const DEE = "dee";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, name: string, n: number): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const bear = (id: string, controller: string): CardView =>
  ({
    instance_id: id,
    name: "Grizzly Bears",
    owner: controller,
    controller,
    type_line: "Creature — Bear",
    power: 2,
    toughness: 2,
  }) as unknown as CardView;

// A spell that can target Cat's bear and nothing else.
const pinpoint = (): CardView =>
  ({
    instance_id: "pinpoint",
    name: "Pinpoint",
    owner: ME,
    controller: ME,
    type_line: "Instant",
    legal_targets: { min: 1, max: 1, players: [], cards: ["cat-bear"] },
  }) as unknown as CardView;

// Seats in turn order: Bob sits next (left), Cat across, Dee across_next.
const gameView = (activeSeat: number): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(BOB, "Bob", 1), seat(CAT, "Cat", 2), seat(DEE, "Dee", 3)],
    battlefield: zone("battlefield", undefined, [
      bear("bob-bear", BOB),
      bear("cat-bear", CAT),
      bear("dee-bear", DEE),
    ]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      // Turn n+1 is seat n's, so consecutive seats' turns differ in parity.
      number: activeSeat + 1,
      active_seat: activeSeat,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
  }) as unknown as GameView;

function mountBoard(activeSeat: number, view: GameView = gameView(activeSeat)) {
  return render(
    Board as never,
    {
      view,
      viewerID: ME,
      isAdmin: false,
      sendAction: () => {},
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
}

const board = () => document.querySelector<HTMLElement>(".board")!;
const region = (name: string) =>
  document.querySelector<HTMLElement>(`.board > .slot [role="region"][aria-label="${name} board"]`);
const isAvatarCard = (name: string) => region(name)?.classList.contains("avatar-card") ?? false;
const isFullBoard = (name: string) => region(name)?.classList.contains("panel") ?? false;
const avatarCards = () => [...document.querySelectorAll(".board > .slot .avatar-card")];

describe("reflow on an opponent's turn", () => {
  it("gives Bob most of the row and stacks Cat and Dee in one column at its left", () => {
    mountBoard(1);
    expect(isFullBoard("Bob")).toBe(true);
    expect(isAvatarCard("Cat")).toBe(true);
    expect(isAvatarCard("Dee")).toBe(true);
    // One fixed column for both cards, then Bob's two shares.
    const b = board().style;
    expect(b.gridTemplateColumns).toBe("var(--reflow-avatar-w) minmax(0, 2fr)");
    // Each card has its own row, half the opponents' row.
    expect(b.gridTemplateRows).toBe("minmax(0, 0.5fr) minmax(0, 0.5fr) minmax(0, 1fr)");
    expect(b.gridTemplateAreas).toBe('"across next" "across_next next" "self self"');
    expect(board().hasAttribute("data-reflow")).toBe(true);
    // Your own board is untouched.
    expect(document.querySelector(".board > .slot[data-pos='self'] .avatar-card")).toBeNull();
  });

  it("follows the turn to Dee, and the column stays at the left", () => {
    mountBoard(3);
    expect(isFullBoard("Dee")).toBe(true);
    expect(isAvatarCard("Bob")).toBe(true);
    expect(isAvatarCard("Cat")).toBe(true);
    expect(board().style.gridTemplateColumns).toBe("var(--reflow-avatar-w) minmax(0, 2fr)");
    expect(board().style.gridTemplateAreas).toBe(
      '"next across_next" "across across_next" "self self"',
    );
  });

  // The column reads in the order the seats play after the active one,
  // and the DOM order is the visual order.
  it.each([
    [1, ["Cat", "Dee"], ["Next up", "After Cat"]],
    [2, ["Dee", "Bob"], ["Next up", "After you"]],
    [3, ["Bob", "Cat"], ["After you", "After Bob"]],
  ])("on seat %i's turn the cards read %j with chips %j", (active, names, chips) => {
    mountBoard(active);
    const cards = avatarCards();
    expect(cards.map((c) => c.getAttribute("aria-label"))).toEqual(names.map((n) => `${n} board`));
    expect(cards.map((c) => c.querySelector(".turn-chip")?.textContent?.trim())).toEqual(chips);
    // DOM order: the cards top to bottom, then the focus board, then yours.
    const order = [...document.querySelectorAll<HTMLElement>(".board > .slot")].map(
      (s) => s.dataset.pos,
    );
    expect(order.at(-1)).toBe("self");
    expect(order.slice(0, 2)).toEqual(
      cards.map((c) => (c.closest(".slot") as HTMLElement).dataset.pos),
    );
  });

  it("reorders the column at once when the turn moves, restarting the fade", () => {
    const r = mountBoard(1);
    const slotOf = (name: string) => region(name)!.closest<HTMLElement>(".slot")!;
    const deeBefore = slotOf("Dee").className;
    r.setProps({ view: gameView(2) } as never);
    expect(avatarCards().map((c) => c.getAttribute("aria-label"))).toEqual([
      "Dee board",
      "Bob board",
    ]);
    // The turn's parity names the fade's keyframes, so it starts again.
    expect(slotOf("Dee").className).not.toBe(deeBefore);
  });

  it("sorts an eliminated opponent last, with no chip, reading out", () => {
    const view = gameView(1);
    (view.seats[2] as { eliminated?: boolean }).eliminated = true;
    mountBoard(1, view);
    const cards = avatarCards();
    expect(cards.map((c) => c.getAttribute("aria-label"))).toEqual(["Dee board", "Cat board"]);
    // With Cat out, Dee plays straight after Bob.
    expect(cards[0].querySelector(".turn-chip")?.textContent?.trim()).toBe("Next up");
    expect(cards[1].querySelector(".turn-chip")).toBeNull();
    expect(cards[1].querySelector(".out")?.textContent).toBe("out");
    expect(cards[1].querySelector(".readout")).toBeNull();
  });

  it("keeps each avatar card the seat's board region, named, with its life and avatar", () => {
    mountBoard(1);
    const cat = region("Cat")!;
    expect(cat.getAttribute("role")).toBe("region");
    expect(cat.querySelector(`[data-seat-id="${CAT}"]`)).not.toBeNull();
    expect(cat.textContent).toContain("Cat");
    expect(cat.textContent).toContain("40");
    // The keyboard route to the overlay is still there.
    expect(cat.querySelector('button[aria-label="Show Cat\'s full board"]')).not.toBeNull();
    // It is the stand-in an arrow to one of Cat's cards ends on.
    expect(findStandIn(document, CAT)).toBe(cat);
    expect(findStandIn(document, BOB)).toBeNull();
  });

  // #4102: the card summarises the seat under its identity.
  // #4118: as stat tiles over a footer of the piles.
  it("reads out Cat's hand, creatures, power, mana and piles, in that order", () => {
    const view = gameView(1);
    const cat = view.seats[2] as unknown as Record<string, unknown>;
    cat.hand = { kind: "hand", owner: CAT, count: 4, cards: [] };
    cat.library = { kind: "library", owner: CAT, count: 87, cards: [] };
    cat.graveyard = zone("graveyard", CAT, [bear("cat-dead", CAT)]);
    const forest = {
      instance_id: "cat-forest",
      name: "Forest",
      owner: CAT,
      controller: CAT,
      type_line: "Basic Land — Forest",
      mana_abilities: [{ index: 0, label: "Add {G}", produced: "{G}", tap_cost: true }],
    } as unknown as CardView;
    view.battlefield.cards.push(forest);
    mountBoard(1, view);
    const card = region("Cat")!;
    const tiles = [...card.querySelectorAll<HTMLElement>(".readout .tiles > .tile")];
    expect(tiles.map((t) => t.classList[1])).toEqual(["hand", "creatures", "power", "mana"]);
    const text = (el: Element) => el.textContent!.replace(/\s+/g, " ").trim();
    const num = (el: Element) => text(el.querySelector(".num")!);
    const spoken = (el: Element) => text(el.querySelector(".sr-only")!);
    expect(num(tiles[0])).toBe("4");
    expect(spoken(tiles[0])).toBe("4 in hand");
    expect(num(tiles[1])).toBe("1/1");
    expect(spoken(tiles[1])).toBe("1 of 1 creatures untapped");
    expect(num(tiles[2])).toBe("2");
    expect(spoken(tiles[2])).toBe("2 untapped power");
    expect(tiles[3].getAttribute("aria-label")).toBe("1 mana from 1 source");
    expect(tiles[3].querySelectorAll(".mana-pip")).toHaveLength(1);
    const zones = card.querySelector<HTMLElement>(".readout > .zones")!;
    expect(text(zones.querySelector(".deck")!)).toBe("deck 87");
    // The piles are buttons with PileButton's names.
    expect(zones.querySelector('button[aria-label="grave: 1 card"]')).not.toBeNull();
    expect(zones.querySelector('button[aria-label="exile: 0 cards"]')).not.toBeNull();
    // Dee has no mana: a 0, and the summary's spoken "no mana" form.
    const deeMana = region("Dee")!.querySelector(".tile.mana")!;
    expect(num(deeMana)).toBe("0");
    expect(deeMana.querySelectorAll(".mana-pip")).toHaveLength(0);
  });

  // The card is a size container: jsdom lays nothing out, so this pins
  // the classes the container queries step and the rules that step them.
  it("carries the classes its container steps act on", async () => {
    mountBoard(1);
    const card = region("Cat")!;
    expect(card.querySelector(".card-head .turn-chip")).not.toBeNull();
    expect(card.querySelector(".card-identity [data-seat-id]")).not.toBeNull();
    expect(card.querySelector(".readout > .tiles")).not.toBeNull();
    expect(card.querySelector(".readout > .zones .deck")).not.toBeNull();
    const src = (await import("./components/board/SeatSummary.svelte?raw")).default as string;
    expect(src).toMatch(/\.avatar-card \{[^}]*container: avatar-card \/ size;/);
    expect(src).toMatch(/--avatar-size-base: 84px;/);
    expect(src).toMatch(/\.tiles \{\s*grid-template-columns: 1fr 1fr;/);
    const short = src.slice(src.indexOf("(max-height: 239px)"));
    expect(short).toMatch(/^[^@]*\.card-identity \{\s*--avatar-size-base: 56px;/);
    // A short card lays the four tiles in one row.
    expect(short).toMatch(/^[^@]*\.tiles \{\s*grid-template-columns: repeat\(3,/);
  });

  it("applies under summaries too: the active opponent is a full board", () => {
    updateSettings("display", "opponentDetail", "summary");
    mountBoard(1);
    expect(isFullBoard("Bob")).toBe(true);
    expect(avatarCards()).toHaveLength(2);
  });

  it("hovering an avatar card's avatar opens that seat's board in the overlay", () => {
    mountBoard(1);
    const av = region("Cat")!.querySelector<HTMLElement>(`[data-seat-id="${CAT}"]`)!;
    av.dispatchEvent(new MouseEvent("pointerenter"));
    flushSync();
    vi.advanceTimersByTime(400);
    flushSync();
    const o = document.querySelector<HTMLElement>("[data-board-expanded]")!;
    expect(o.getAttribute("aria-label")).toBe("Cat's board, expanded");
    expect(o.querySelector('[data-instance-id="cat-bear"]')).not.toBeNull();
  });

  it("never shrinks a seat a live prompt needs", () => {
    mountBoard(1);
    begin(pinpoint(), "any");
    flushSync();
    expect(isFullBoard("Cat")).toBe(true);
    expect(region("Cat")!.parentElement!.querySelector('[data-instance-id="cat-bear"]')).not.toBe(
      null,
    );
    expect(isAvatarCard("Dee")).toBe(true);
    targeting.set(null);
    flushSync();
    expect(isAvatarCard("Cat")).toBe(true);
  });
});

describe("no reflow", () => {
  function expectEqualRow(): void {
    expect(avatarCards()).toHaveLength(0);
    expect(board().style.gridTemplateColumns).toBe("");
    expect(board().hasAttribute("data-reflow")).toBe(false);
    for (const name of ["Bob", "Cat", "Dee"]) expect(isFullBoard(name)).toBe(true);
  }

  it("on your own turn: three equal boards", () => {
    mountBoard(0);
    expectEqualRow();
  });

  it("with the overlay style, which keeps the boards equal", () => {
    updateSettings("display", "expandStyle", "overlay");
    mountBoard(1);
    expectEqualRow();
  });

  it("with the active-player expansion off", () => {
    updateSettings("display", "expandActivePlayer", false);
    mountBoard(1);
    expectEqualRow();
  });

  it("at a four-player quadrant table, which has no one opponents' row", () => {
    updateSettings("display", "tableLayout", "quadrant");
    mountBoard(1);
    expectEqualRow();
  });
});

// The stack pile docks on a seam of the top row (#3380). With the
// reflow's avatar cards in the row it stays over the full board instead
// (stackPile.ts, tableSeams().dock), and measuring that span must not
// feed back into the effect that measures it.
describe("the stack pile beside the avatar cards", () => {
  const BOARD = { w: 1400, h: 800 };
  const AV = 176;
  let restore: (() => void) | null = null;
  afterEach(() => {
    restore?.();
    restore = null;
  });

  // jsdom lays nothing out, so the board and its slots report the boxes
  // the row layout gives them on Dee's turn: Bob's card over Cat's in
  // one column, then Dee.
  function layOut(): void {
    const proto = HTMLElement.prototype;
    const original = proto.getBoundingClientRect;
    proto.getBoundingClientRect = function (this: HTMLElement): DOMRect {
      const box = (left: number, top: number, width: number, height: number) =>
        ({
          left,
          top,
          width,
          height,
          right: left + width,
          bottom: top + height,
          x: left,
          y: top,
        }) as DOMRect;
      if (this.classList.contains("board")) return box(0, 0, BOARD.w, BOARD.h);
      if (this.classList.contains("slot")) {
        const half = BOARD.h / 2;
        const card = (half - 8) / 2;
        switch (this.dataset.pos) {
          case "next":
            return box(0, 0, AV, card);
          case "across":
            return box(0, card + 8, AV, card);
          case "across_next":
            return box(AV + 8, 0, BOARD.w - (AV + 8), half);
          case "self":
            return box(0, half + 8, BOARD.w, half - 8);
        }
      }
      // An avatar card and its disc: the card fills its slot, the disc
      // sits in its middle.
      const slot = this.closest<HTMLElement>(".slot");
      if (slot && this.matches("[data-seat-stand-in], [data-seat-id]")) {
        const s = slot.getBoundingClientRect();
        if (this.hasAttribute("data-seat-stand-in")) return box(s.left, s.top, s.width, s.height);
        return box(s.left + (s.width - 84) / 2, s.top + 30, 84, 84);
      }
      return original.call(this);
    };
    restore = () => {
      proto.getBoundingClientRect = original;
    };
  }

  it("mounts with a spell on the stack and docks clear of the avatar cards", () => {
    updateSettings("display", "stackStyle", "pile");
    layOut();
    const item = {
      id: "s1",
      kind: "spell",
      controller: DEE,
      owner: DEE,
      source_card_id: "s1",
      targets: [],
    } as unknown as StackItemView;
    const spell = {
      instance_id: "s1",
      name: "Lightning Bolt",
      owner: DEE,
      controller: DEE,
      type_line: "Instant",
    } as unknown as CardView;
    const view = { ...gameView(3), stack: zone("stack", undefined, [spell]), stack_items: [item] };
    expect(() => {
      mountBoard(3, view as GameView);
      flushSync();
    }).not.toThrow();
    expect(isAvatarCard("Bob")).toBe(true);
    expect(isAvatarCard("Cat")).toBe(true);
    const lane = document.querySelector<HTMLElement>(".stack-lane");
    expect(lane).not.toBeNull();
  });

  // Call 8: the overlay opens to the card's right and never covers it.
  it("opens an avatar card's board to the card's right, clear of the whole card", () => {
    layOut();
    mountBoard(3);
    const av = region("Cat")!.querySelector<HTMLElement>(`[data-seat-id="${CAT}"]`)!;
    av.click();
    flushSync();
    const o = document.querySelector<HTMLElement>("[data-board-expanded]")!;
    expect(o).not.toBeNull();
    // The card is 176px wide from the board's left edge; the disc is in
    // its middle, so a placement off the disc alone would start at 142.
    expect(parseFloat(o.style.left)).toBeGreaterThanOrEqual(AV + 8);
    expect(parseFloat(o.style.width)).toBeGreaterThan(0);
  });
});
