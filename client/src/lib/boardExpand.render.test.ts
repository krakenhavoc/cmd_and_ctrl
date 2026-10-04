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
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import { abilityPopover } from "./abilityPopover";
import { begin, targeting } from "./targeting";
import { _resetForTests as resetDock, cancelAction, pushDockRequest } from "./dock";
import { resetSettings } from "./settings";
import type { ActionType, CardView, GameView, PlayerView } from "./protocol";
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
    expect(overlay()!.querySelectorAll(".mana-menu")).toHaveLength(1);
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
      const a = boxes[id];
      expect(width, id).toBeGreaterThan(0);
      const overlaps = left < a.left + a.width && a.left < left + width;
      expect(overlaps, `overlay covers ${id}'s avatar`).toBe(false);
      click(tableAvatar(id));
      expect(overlay()).toBeNull();
    }
  });
});
