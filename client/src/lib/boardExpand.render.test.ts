// @vitest-environment jsdom
//
// boardExpand.render.test.ts — ADR 0120 PR 3, the expanded board on a
// real Board with four seats: what is named what, that a click in the
// overlay goes where the same click on the table goes, that an avatar
// click still targets before it pins, who owns Escape, one popover per
// card, the expand button's focus, and that the overlay never covers
// the avatar it opened from. The transitions themselves are pinned in
// boardExpand.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { readFileSync } from "node:fs";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import { abilityPopover } from "./abilityPopover";
import { begin, targeting } from "./targeting";
import { _resetForTests as resetDock, cancelAction, pushDockRequest } from "./dock";
import { resetSettings, updateSettings } from "./settings";
import type { ActionType, CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

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

// Two usable rows, one not mana: a left-click opens the popover.
const relic = (): CardView =>
  ({
    instance_id: "relic",
    name: "Relic of Sauron",
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Legendary Artifact",
    activated_abilities: [
      { index: 0, ref: "own:0", label: "{3}, {T}: Draw two cards.", tap_cost: true },
      { index: 1, ref: "own:1", label: "{1}: Scry 1.", mana_cost: "{1}" },
    ],
  }) as unknown as CardView;

const shock = (): CardView =>
  ({
    instance_id: "shock",
    name: "Shock",
    owner: ME,
    controller: ME,
    type_line: "Instant",
    legal_targets: { min: 1, max: 1, players: [BOB, CAT, DEE], cards: ["bob-bear"] },
  }) as unknown as CardView;

const gameView = (): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(BOB, "Bob", 1), seat(CAT, "Cat", 2), seat(DEE, "Dee", 3)],
    battlefield: zone("battlefield", undefined, [relic(), bear("bob-bear", BOB)]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "precombat_main",
    },
    mulligans_open: false,
  }) as unknown as GameView;

function mountBoard(view = gameView()) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const r = render(
    Board as never,
    {
      view,
      viewerID: ME,
      isAdmin: false,
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
  return { ...r, sent };
}

const overlay = () => document.querySelector<HTMLElement>("[data-board-expanded]");
const tableAvatar = (seatID: string) =>
  document.querySelector<HTMLElement>(`.board > .slot [data-seat-id="${seatID}"]`)!;
const regionsNamed = (name: string) =>
  [...document.querySelectorAll('[role="region"][aria-label]')].filter((e) =>
    e.getAttribute("aria-label")!.includes(name),
  );

function pointer(el: Element, type: string, init: MouseEventInit = {}): void {
  el.dispatchEvent(new MouseEvent(type, { bubbles: type === "pointerdown", ...init }));
  flushSync();
}
function hoverOpen(el: Element): void {
  pointer(el, "pointerenter");
  vi.advanceTimersByTime(400);
  flushSync();
}
function escape(): KeyboardEvent {
  const e = new KeyboardEvent("keydown", { key: "Escape", bubbles: true, cancelable: true });
  window.dispatchEvent(e);
  flushSync();
  return e;
}

describe("hover and click on an avatar", () => {
  it("hovering opens Bob's board, named for him, and the table's name stays unique", () => {
    mountBoard();
    pointer(tableAvatar(BOB), "pointerenter");
    vi.advanceTimersByTime(399);
    flushSync();
    expect(overlay()).toBeNull();
    vi.advanceTimersByTime(1);
    flushSync();
    const o = overlay()!;
    expect(o.getAttribute("role")).toBe("region");
    expect(o.getAttribute("aria-label")).toBe("Bob's board, expanded");
    expect(regionsNamed("Bob board")).toHaveLength(1);
    // Its panel is an unnamed group, larger and upright.
    const panel = o.querySelector(".panel")!;
    expect(panel.getAttribute("role")).toBe("group");
    expect(panel.classList.contains("expanded")).toBe(true);
    expect(panel.classList.contains("flipped")).toBe(false);
    // And it draws Bob's cards.
    expect(o.querySelector('[data-instance-id="bob-bear"]')).not.toBeNull();
  });

  it("closes after the pointer leaves, unless it crosses into the overlay", () => {
    mountBoard();
    hoverOpen(tableAvatar(BOB));
    pointer(tableAvatar(BOB), "pointerleave");
    pointer(overlay()!, "pointerenter");
    vi.advanceTimersByTime(1000);
    flushSync();
    expect(overlay()).not.toBeNull();
    pointer(overlay()!, "pointerleave");
    vi.advanceTimersByTime(250);
    flushSync();
    expect(overlay()).toBeNull();
  });

  it("a click pins it; the pin button says so; a second click closes it", () => {
    mountBoard();
    const av = tableAvatar(CAT);
    pointer(av, "pointerenter");
    pointer(av, "pointerdown");
    click(av);
    expect(overlay()?.getAttribute("aria-label")).toBe("Cat's board, expanded");
    const pin = overlay()!.querySelector<HTMLElement>(".pin-board")!;
    expect(pin.getAttribute("aria-label")).toBe("Pin Cat's expanded board");
    expect(pin.getAttribute("aria-pressed")).toBe("true");
    pointer(av, "pointerleave");
    vi.advanceTimersByTime(2000);
    flushSync();
    expect(overlay()).not.toBeNull();
    pointer(av, "pointerenter");
    pointer(av, "pointerdown");
    click(av);
    expect(overlay()).toBeNull();
  });

  it("works on the viewer's own seat too", () => {
    mountBoard();
    hoverOpen(tableAvatar(ME));
    expect(overlay()?.getAttribute("aria-label")).toBe("your own board, expanded");
    expect(regionsNamed("your board")).toHaveLength(1);
  });

  it("a held button opens nothing (a card dragged across the avatar)", () => {
    mountBoard();
    pointer(tableAvatar(BOB), "pointerenter", { buttons: 1 });
    vi.advanceTimersByTime(1000);
    flushSync();
    expect(overlay()).toBeNull();
  });
});

describe("act through it (ADR 0120 §3)", () => {
  it("a card click in the overlay during targeting sends the target", () => {
    const b = mountBoard();
    click(tableAvatar(BOB));
    begin(shock(), "any");
    flushSync();
    const tile = overlay()!.querySelector<HTMLElement>('[data-instance-id="bob-bear"]')!;
    click(tile);
    expect(b.sent.map((s) => s.type)).toContain("cast_spell");
    expect(JSON.stringify(b.sent)).toContain("bob-bear");
  });

  it("an avatar click during targeting targets the player and does not pin", () => {
    const b = mountBoard();
    begin(shock(), "any");
    flushSync();
    const av = tableAvatar(DEE);
    pointer(av, "pointerdown");
    click(av);
    expect(b.sent.map((s) => s.type)).toContain("cast_spell");
    expect(JSON.stringify(b.sent)).toContain(DEE);
    expect(overlay()).toBeNull();
  });

  it("the overlay's ability popover is drawn once, in the overlay", () => {
    mountBoard();
    click(document.querySelector<HTMLElement>('button[aria-label="Expand your own board"]')!);
    const tile = overlay()!.querySelector<HTMLElement>('.card[data-instance-id="relic"]')!;
    click(tile);
    expect(get(abilityPopover)).toEqual({ cardID: "relic", surface: "expanded" });
    expect(document.querySelectorAll(".mana-menu")).toHaveLength(1);
    // #2960: drawn in the board's popover host, out of the card and the
    // overlay's clipping, and from the overlay's copy only.
    expect(document.querySelectorAll("[data-popover-host] .mana-menu")).toHaveLength(1);
    expect(overlay()!.querySelectorAll(".mana-menu")).toHaveLength(0);
    // Placed against the viewport (menuPlacement.ts), above or below.
    const anchor = document.querySelector<HTMLElement>("[data-popover-host] .mana-menu-anchor")!;
    expect(["above", "below"]).toContain(anchor.dataset.side);
    expect(anchor.style.left).toMatch(/px$/);
    expect(anchor.style.maxHeight).toMatch(/px$/);
  });
});

describe("Escape (ADR 0120 §4)", () => {
  it("closes a pinned board when nothing else owns it", () => {
    mountBoard();
    click(tableAvatar(BOB));
    const e = escape();
    expect(overlay()).toBeNull();
    expect(e.defaultPrevented).toBe(true);
  });

  it("is the targeting prompt's while one is live, and leaves the overlay", () => {
    mountBoard();
    click(tableAvatar(BOB));
    begin(shock(), "any");
    flushSync();
    const e = escape();
    expect(overlay()).not.toBeNull();
    expect(e.defaultPrevented).toBe(false);
    expect(get(targeting)).not.toBeNull();
  });

  it("is the dock's when its request advertises Escape", () => {
    mountBoard();
    click(tableAvatar(BOB));
    const onCancel = vi.fn();
    pushDockRequest({
      rank: "flow",
      label: "a cost",
      primary: null,
      secondary: [cancelAction(onCancel)],
    });
    escape();
    expect(overlay()).not.toBeNull();
  });

  it("is the pile's '+N more' list's while it is open, then the overlay's", () => {
    updateSettings("display", "stackStyle", "pile");
    const items = Array.from({ length: 7 }, (_, i) => ({
      id: `s${i}`,
      kind: "spell",
      controller: BOB,
      owner: BOB,
      source_card_id: `s${i}`,
      targets: [],
    })) as unknown as StackItemView[];
    const cards = items.map((it) => ({
      instance_id: it.id,
      name: `Spell ${it.id}`,
      owner: BOB,
      controller: BOB,
      type_line: "Instant",
    })) as unknown as CardView[];
    mountBoard({
      ...gameView(),
      stack: zone("stack", undefined, cards),
      stack_items: items,
    } as GameView);
    click(tableAvatar(BOB));
    click(document.querySelector<HTMLElement>(".more-chip")!);
    expect(document.querySelector(".all-items")).not.toBeNull();
    escape();
    expect(document.querySelector(".all-items")).toBeNull();
    expect(overlay()).not.toBeNull();
    escape();
    expect(overlay()).toBeNull();
  });

  it("is the open popover's, and leaves the overlay", () => {
    mountBoard();
    click(document.querySelector<HTMLElement>('button[aria-label="Expand your own board"]')!);
    click(overlay()!.querySelector<HTMLElement>('.card[data-instance-id="relic"]')!);
    escape();
    expect(overlay()).not.toBeNull();
  });
});

describe("the expand button (ADR 0120 §4)", () => {
  it("opens the board pinned and moves focus to its pin; closing gives focus back", async () => {
    mountBoard({
      ...gameView(),
      // A full panel for Bob (the active seat), so the slot has the button.
      turn: { ...gameView().turn, active_seat: 1 },
    } as GameView);
    const btn = document.querySelector<HTMLButtonElement>(
      'button[aria-label="Expand Bob\'s board"]',
    )!;
    expect(btn).not.toBeNull();
    btn.focus();
    click(btn);
    await vi.advanceTimersByTimeAsync(0);
    flushSync();
    const pin = overlay()!.querySelector<HTMLButtonElement>(".pin-board")!;
    expect(pin.getAttribute("aria-pressed")).toBe("true");
    expect(document.activeElement).toBe(pin);
    click(pin);
    expect(overlay()).toBeNull();
    expect(document.activeElement).toBe(btn);
  });

  it("a summary's ⤢ opens the overlay pinned, and the summary stays a summary", () => {
    // Full boards are the default since #2433; a summary is chosen.
    updateSettings("display", "opponentDetail", "summary");
    mountBoard();
    const show = document.querySelector<HTMLElement>(
      'button[aria-label="Show Cat\'s full board"]',
    )!;
    click(show);
    expect(overlay()?.getAttribute("aria-label")).toBe("Cat's board, expanded");
    // ADR 0120 §6: the table does not move; the in-place pin is gone.
    expect(document.querySelectorAll(".board > .slot .summary").length).toBeGreaterThan(0);
    expect(document.querySelector(".unpin")).toBeNull();
  });

  it("a seat that leaves the table takes its overlay with it", () => {
    const b = mountBoard();
    click(tableAvatar(BOB));
    expect(overlay()).not.toBeNull();
    const v = gameView();
    b.setProps({ view: { ...v, seats: v.seats.filter((s) => s.id !== BOB) } } as never);
    expect(overlay()).toBeNull();
  });
});

describe("placement (ADR 0120 §2)", () => {
  const rect = (left: number, top: number, width: number, height: number): DOMRect =>
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

  it("never covers the avatar's box", () => {
    mountBoard();
    const board = document.querySelector<HTMLElement>(".board")!;
    board.getBoundingClientRect = () => rect(0, 0, 1600, 900);
    const boxes: Record<string, DOMRect> = {
      [BOB]: rect(680, 30, 84, 84),
      [CAT]: rect(1480, 30, 84, 84),
      [DEE]: rect(680, 420, 84, 84),
      [ME]: rect(1480, 420, 84, 84),
    };
    for (const id of [BOB, CAT, DEE, ME]) {
      tableAvatar(id).getBoundingClientRect = () => boxes[id];
    }
    for (const id of [BOB, CAT, DEE, ME]) {
      click(tableAvatar(id));
      const o = overlay()!;
      const left = parseFloat(o.style.left);
      const width = parseFloat(o.style.width);
      // Full height from the board's 10px inset, unless it opens below
      // the avatar (ADR 0204 §4).
      const top = o.style.top ? parseFloat(o.style.top) : 10;
      const a = boxes[id];
      expect(width, id).toBeGreaterThan(0);
      const overlaps = left < a.left + a.width && a.left < left + width && top < a.top + a.height;
      expect(overlaps, `overlay covers ${id}'s avatar`).toBe(false);
      click(tableAvatar(id));
      expect(overlay()).toBeNull();
    }
  });

  // ADR 0204 §4 (call 5): a top-row seat's avatar is at the top centre of
  // its panel, so its board opens below it, toward the table's centre, in
  // the wider half, 8px under it. Your own (and an upright seat's) still
  // opens beside it.
  it("opens below a top-row avatar and beside your own", () => {
    mountBoard();
    const board = document.querySelector<HTMLElement>(".board")!;
    board.getBoundingClientRect = () => rect(0, 0, 1600, 900);
    const boxes: Record<string, DOMRect> = {
      [BOB]: rect(230, 10, 84, 84),
      [CAT]: rect(758, 10, 84, 84),
      [DEE]: rect(1290, 10, 84, 84),
      [ME]: rect(500, 700, 84, 84),
    };
    for (const id of [BOB, CAT, DEE, ME]) {
      tableAvatar(id).getBoundingClientRect = () => boxes[id];
    }
    for (const id of [BOB, CAT, DEE]) {
      const panel = tableAvatar(id).closest(".panel")!;
      expect(panel.classList.contains("flipped"), id).toBe(true);
      click(tableAvatar(id));
      const o = overlay()!;
      const a = boxes[id];
      expect(parseFloat(o.style.top), id).toBe(a.top + a.height + 8);
      const left = parseFloat(o.style.left);
      const width = parseFloat(o.style.width);
      // Under the avatar's own column, in the wider half beside it.
      expect(left, id).toBeLessThanOrEqual(a.left);
      expect(left + width, id).toBeGreaterThanOrEqual(a.left + a.width);
      click(tableAvatar(id));
    }
    // Bob is left of centre, so his opens rightward; Dee's leftward.
    click(tableAvatar(BOB));
    expect(parseFloat(overlay()!.style.left)).toBe(230);
    click(tableAvatar(BOB));
    click(tableAvatar(DEE));
    expect(parseFloat(overlay()!.style.left) + parseFloat(overlay()!.style.width)).toBe(1374);
    click(tableAvatar(DEE));
    click(tableAvatar(ME));
    expect(overlay()!.style.top).toBe("");
    const left = parseFloat(overlay()!.style.left);
    const width = parseFloat(overlay()!.style.width);
    expect(left + width <= 500 - 8 || left >= 584 + 8).toBe(true);
  });
});

// ADR 0120 as amended 2026-10-18 (#4122): the overlay's own layout. A
// header names the seat and holds its one identity, the command zone is
// a back-row group, and the strip holds the piles and the hand.
describe("the overlay's layout (#4122)", () => {
  const commander = (owner: string): CardView =>
    ({
      instance_id: `${owner}-cmdr`,
      name: "Adeline, Resplendent Cathar",
      owner,
      controller: owner,
      type_line: "Legendary Creature — Human Knight",
    }) as unknown as CardView;

  const withCommanderAndHand = (view: GameView): GameView => {
    const bob = view.seats[1] as PlayerView;
    (bob as { command: unknown }).command = zone("command", BOB, [commander(BOB)]);
    (bob as { hand: unknown }).hand = { kind: "hand", owner: BOB, count: 5, cards: [] };
    return view;
  };

  it("names the seat in a header that holds its one avatar", () => {
    mountBoard();
    click(tableAvatar(BOB));
    const o = overlay()!;
    const head = o.querySelector<HTMLElement>(".head")!;
    expect(head).not.toBeNull();
    expect(head.querySelector(".name")?.textContent).toBe("Bob");
    expect(head.querySelector(".life")?.textContent).toBe("40");
    // One identity for the seat in the overlay, and it is the header's,
    // so a pick of the player or an arrow to them lands there.
    const avatars = o.querySelectorAll(`[data-seat-id="${BOB}"]`);
    expect(avatars).toHaveLength(1);
    expect(head.contains(avatars[0])).toBe(true);
    // The header is not a second name for the board.
    expect(regionsNamed("Bob board")).toHaveLength(1);
    // A full board's seat has no turn chip.
    expect(head.querySelector(".turn-chip")).toBeNull();
  });

  it("draws the command zone as a back-row group, and the hand count in the strip", () => {
    mountBoard(withCommanderAndHand(gameView()));
    click(tableAvatar(BOB));
    const o = overlay()!;
    const cmd = o.querySelector<HTMLElement>('[role="group"][aria-label^="Bob command zone"]')!;
    expect(cmd).not.toBeNull();
    expect(cmd.closest(".grid-middle")).not.toBeNull();
    expect(cmd.querySelector(".row-label")?.textContent).toContain("Command");
    expect(o.querySelector(".grid-bottom .command-strip")).toBeNull();
    const count = o.querySelector<HTMLElement>(".grid-bottom .strip-hand-count")!;
    expect(count.getAttribute("aria-label")).toBe("5 cards in hand");
    // The piles leave the rail for the strip; the rail holds no identity.
    expect(o.querySelector(".rail > .identity")).toBeNull();
    expect(o.querySelector(".rail > .pile-bar")).not.toBeNull();
  });

  it("shows the reflow turn chip of a seat that is an avatar card", () => {
    updateSettings("display", "expandStyle", "reflow");
    const view = gameView();
    (view.turn as { active_seat: number }).active_seat = 1;
    mountBoard(view);
    // Bob's turn: Cat plays next and is an avatar card.
    click(tableAvatar(CAT));
    const chip = overlay()!.querySelector(".head .turn-chip");
    expect(chip?.textContent?.trim()).toBe("Next up");
    click(tableAvatar(CAT));
    // Bob, the active seat, is a full board: no chip.
    click(tableAvatar(BOB));
    expect(overlay()!.querySelector(".head .turn-chip")).toBeNull();
  });

  // jsdom has no layout, so the sizes are read off the stylesheet, as
  // boardArtPip.test.ts reads BattlefieldRow's (call 9).
  it("sizes every group at one card size, capped at 240px times the setting", () => {
    const src = readFileSync("src/lib/components/board/PlayerPanel.svelte", "utf8");
    const at = src.indexOf("  .panel.expanded:not(.flipped) {");
    expect(at).toBeGreaterThan(0);
    const block = src.slice(at, src.indexOf("\n  }", at));
    expect(block).toContain("--card-h-sm: var(--card-h);");
    expect(block).toContain("calc(240px * var(--card-scale, 1))");
    expect(block).not.toContain("--card-scale-opponent");
  });

  // #4122's branch E2E: with a dock request open the overlay is short,
  // and a 96px floor made the rows taller than the panel. The back row
  // ran down under the strip and the opponent's face-down fan took the
  // click meant for a card in it. The floor is rowFit's MIN_CARD_H, the
  // back row's track may shrink to nothing, and the back row is cut at
  // its track rather than reaching under the strip.
  it("never lets the back row reach under the strip", () => {
    const src = readFileSync("src/lib/components/board/PlayerPanel.svelte", "utf8");
    const at = src.indexOf("  .panel.expanded:not(.flipped) {");
    const block = src.slice(at, src.indexOf("\n  }", at));
    expect(block).toMatch(/--card-h: clamp\(\s*56px,/);
    expect(block).toContain(
      "grid-template-rows: auto minmax(var(--creature-min, 0px), auto) minmax(0, 1fr) auto;",
    );
    const mid = src.indexOf("  .panel.expanded:not(.flipped) .grid-middle {");
    expect(mid).toBeGreaterThan(0);
    expect(src.slice(mid, src.indexOf("\n  }", mid))).toContain("min-height: 0;");
    // .grid-middle itself clips what does not fit, vertically.
    const base = src.indexOf("  .grid-middle {");
    expect(src.slice(base, src.indexOf("\n  }", base))).toContain("overflow-y: hidden;");
  });
});
