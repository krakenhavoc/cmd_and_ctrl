// @vitest-environment jsdom
//
// targetPlayerReach.render.test.ts — #4435. While a player is being
// chosen as a target, a dialog (the action dock, a sheet) or the hand
// could sit over that player's avatar on the board and hide it, so the
// player could not be clicked. jsdom has no layout, so "covered" cannot be
// measured here (target-player-reach-4435.spec.ts measures it in a real
// browser; targetReach.test.ts pins the cover rule Game.svelte uses to offer
// the row). What is pinned here is the row itself: with a targeting prompt
// open at a four-seat table whose request is given the player pick, the
// dock lists every legal player as a button, and pressing one picks that
// player exactly as a click on the avatar does.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import ActionDock from "./components/board/ActionDock.svelte";
import Board from "./components/board/Board.svelte";
import { _resetForTests as resetDock, pushDockRequest } from "./dock";
import { resetSettings } from "./settings";
import { BOARD_ACTION_SELECTOR } from "./sheetCover";
import { begin, pickPlayer, targeting } from "./targeting";
import { targetingRequest } from "./targetingDock";
import type { ActionType, CardView, GameView, PendingChoiceView, PlayerView } from "./protocol";
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
});

afterEach(() => {
  cleanup();
  targeting.set(null);
  resetDock();
  resetSettings();
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

const seat = (id: string, name: string, n: number, over: Partial<PlayerView> = {}): PlayerView =>
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
    ...over,
  }) as unknown as PlayerView;

// A burn spell that may aim at any of the three opponents, and not at
// the caster.
const shock = (players: string[] = [BOB, CAT, DEE]): CardView =>
  ({
    instance_id: "shock",
    name: "Shock",
    owner: ME,
    controller: ME,
    type_line: "Instant",
    legal_targets: { min: 1, max: 1, players, cards: [] },
  }) as unknown as CardView;

const gameView = (seats?: PlayerView[]): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: seats ?? [
      seat(ME, "Me", 0),
      seat(BOB, "Bob", 1),
      seat(CAT, "Cat", 2),
      seat(DEE, "Dee", 3),
    ],
    battlefield: zone("battlefield", undefined),
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

function mountTable(view = gameView()) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  render(
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
  const dock = render(
    ActionDock as never,
    {
      view,
      viewerHasPriority: true,
      viewerIsActive: true,
      autopassEnabled: false,
      onPassPriority: () => {},
      onPassTurn: () => {},
      onToggleAutopass: () => {},
      onSize: () => {},
    } as never,
  );
  return { sent, dock, view };
}

// openPrompt does what Game.svelte does: builds the dock's request from
// the live targeting state, with the player pick wired to Board's.
function openPrompt(view: GameView): void {
  pushDockRequest(
    targetingRequest(get(targeting)!, view, {
      onDone: () => {},
      onCancel: () => {},
      onPlayer: pickPlayer,
    }),
  );
  flushSync();
}

const dialog = (root: ParentNode) =>
  root.querySelector<HTMLElement>('[role="dialog"][aria-label^="Select target for"]');
const playerButtons = (root: ParentNode) =>
  [...(dialog(root)?.querySelectorAll<HTMLButtonElement>("button") ?? [])].filter((b) =>
    (b.getAttribute("aria-label") ?? "").startsWith("Target "),
  );

describe("a player target is reachable from the prompt (#4435)", () => {
  it("lists every legal player as a button, in seat order, each with its seat colour", () => {
    const t = mountTable();
    begin(shock(), "any");
    flushSync();
    openPrompt(t.view);

    const buttons = playerButtons(t.dock.container);
    expect(buttons.map((b) => b.getAttribute("aria-label"))).toEqual([
      "Target Bob",
      "Target Cat",
      "Target Dee",
    ]);
    // The visible text is the name; the dot is the seat's colour.
    expect(buttons.map((b) => b.textContent?.trim())).toEqual(["Bob", "Cat", "Dee"]);
    const dots = buttons.map((b) => b.querySelector<HTMLElement>(".seat-dot")?.style.background);
    expect(dots.every(Boolean)).toBe(true);
    expect(new Set(dots).size).toBe(3);
    expect(dialog(t.dock.container)!.textContent).toContain("Players");
  });

  it("every legal avatar on the board has a matching button", () => {
    const t = mountTable();
    begin(shock(), "any");
    flushSync();
    openPrompt(t.view);

    const onBoard = [...document.querySelectorAll<HTMLElement>("[data-seat-identity]")]
      .filter((id) => id.classList.contains("cast-targetable"))
      .map((id) => id.dataset.seatIdentity);
    expect(onBoard.sort()).toEqual([BOB, CAT, DEE]);
    const names = playerButtons(t.dock.container).map((b) => b.textContent?.trim());
    expect(names).toHaveLength(onBoard.length);
  });

  it("a button picks that player: the same cast a click on the avatar sends", () => {
    const t = mountTable();
    begin(shock(), "any");
    flushSync();
    openPrompt(t.view);

    const dee = playerButtons(t.dock.container).find((b) => b.textContent?.trim() === "Dee")!;
    click(dee);
    expect(t.sent.map((s) => s.type)).toContain("cast_spell");
    expect(JSON.stringify(t.sent)).toContain(DEE);
    expect(get(targeting)).toBeNull();
  });

  it("lists only the players the server says are legal, and never an eliminated seat", () => {
    const view = gameView([
      seat(ME, "Me", 0),
      seat(BOB, "Bob", 1),
      seat(CAT, "Cat", 2, { eliminated: true }),
      seat(DEE, "Dee", 3),
    ]);
    const t = mountTable(view);
    // Cat is eliminated; Bob is the only one the clause allows besides Cat.
    begin(shock([BOB, CAT]), "any");
    flushSync();
    openPrompt(t.view);
    expect(playerButtons(t.dock.container).map((b) => b.textContent?.trim())).toEqual(["Bob"]);
  });

  it("offers your own seat when the clause allows it", () => {
    const t = mountTable();
    begin(shock([ME, BOB, CAT, DEE]), "any");
    flushSync();
    openPrompt(t.view);
    expect(playerButtons(t.dock.container).map((b) => b.getAttribute("aria-label"))).toEqual([
      "Target Me",
      "Target Bob",
      "Target Cat",
      "Target Dee",
    ]);
  });

  it("a trigger's pick_target offers them too, and its answer goes out as the choice's", () => {
    const choice = {
      id: "c1",
      kind: "pick_target",
      chooser: ME,
      from_player: ME,
      count: 0,
      reason: "target player loses 1 life",
      pick_target: { players: [BOB, DEE], min: 1, max: 1 },
    } as PendingChoiceView;
    // Board opens the prompt itself for a pending choice addressed to the viewer.
    const view = { ...gameView(), pending_choices: [choice] } as GameView;
    const t = mountTable(view);
    flushSync();
    expect(get(targeting)?.choiceID).toBe("c1");
    openPrompt(t.view);
    const buttons = playerButtons(t.dock.container);
    expect(buttons.map((b) => b.textContent?.trim())).toEqual(["Bob", "Dee"]);
    click(buttons[1]!);
    expect(t.sent.map((s) => s.type)).toContain("resolve_choice");
    expect(JSON.stringify(t.sent)).toContain(DEE);
  });

  // A sheet that opens over an avatar folds down to its question line
  // (sheetCover.ts), as it already did over a targetable card.
  it("a ringed legal avatar counts as something to act on under an open sheet", () => {
    mountTable();
    begin(shock(), "any");
    flushSync();
    const acts = [...document.querySelectorAll(BOARD_ACTION_SELECTOR)];
    const seats = acts
      .map((e) => e.closest<HTMLElement>("[data-seat-identity]")?.dataset.seatIdentity)
      .filter(Boolean);
    expect(seats.sort()).toEqual([BOB, CAT, DEE]);
  });

  it("a prompt that takes only creatures lists no players", () => {
    const t = mountTable();
    begin(
      {
        instance_id: "doom",
        name: "Doom Blade",
        owner: ME,
        controller: ME,
        type_line: "Instant",
        legal_targets: { min: 1, max: 1, players: [], cards: ["x"] },
      } as unknown as CardView,
      "creature",
    );
    flushSync();
    openPrompt(t.view);
    expect(playerButtons(t.dock.container)).toHaveLength(0);
    expect(dialog(t.dock.container)!.textContent).not.toContain("Players");
  });
});
