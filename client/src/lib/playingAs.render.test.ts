// @vitest-environment jsdom
//
// playingAs.render.test.ts — ADR 0167 §9 on the REAL Board: while Alice
// plays Bob's turn (Emrakul, the Promised End), the board Game.svelte
// mounts for her acting seat puts Bob at the bottom with his hand face
// up, lights his moves from his own digest, and sends his play with
// `player` set to him. Switched back, it is her own board again. The
// banner draws the toggle for her and the text alone for everyone else.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Board from "./components/board/Board.svelte";
import PlayingAsBanner from "./components/board/PlayingAsBanner.svelte";
import { legalActionsOf, visibleHighlights } from "./legalActions";
import { actingSeatID, playingAsBanner, viewForSeat, type PlayingAsPref } from "./playingAs";
import { _resetForTests as resetDock } from "./dock";
import type { ActionType, CardView, GameView, LegalMoveView, PlayerView } from "./protocol";
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
  resetDock();
  localStorage.clear();
});

const ALICE = "alice";
const BOB = "bob";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const card = (id: string, name: string, owner: string, typeLine: string): CardView =>
  ({
    instance_id: id,
    name,
    owner,
    controller: owner,
    known_by_you: true,
    type_line: typeLine,
    mana_cost: typeLine.includes("Land") ? "" : "{1}{G}",
  }) as CardView;

const seat = (id: string, name: string, n: number, hand: CardView[]): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
    land_drops_per_turn: 1,
    lands_played_this_turn: 0,
  }) as unknown as PlayerView;

const bobsLand: LegalMoveView = {
  type: "cast_spell",
  player: BOB,
  kind: "land",
  label: "Play Forest",
  source: "forest",
  params: { instance_id: "forest", from_zone: "hand" },
};
const bobsPass: LegalMoveView = {
  type: "pass_priority",
  player: BOB,
  kind: "pass",
  label: "Pass",
  source: "00000000-0000-0000-0000-000000000000",
};

// Alice's frame on Bob's controlled turn, Bob holding priority in his
// main phase: she knows his hand (CR 723.4), her own digest is empty,
// and his is under his ID (ADR 0167 §7).
function alicesFrame(): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [
      seat(ALICE, "Alice", 0, [card("island", "Island", ALICE, "Basic Land — Island")]),
      seat(BOB, "Bob", 1, [
        card("forest", "Forest", BOB, "Basic Land — Forest"),
        card("bears", "Grizzly Bears", BOB, "Creature — Bear"),
      ]),
    ],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { seq: 9, number: 9, active_seat: 1, priority_holder: 1, phase: "main1", step: "main" },
    mulligans_open: false,
    player_controls: [
      {
        controller: ALICE,
        controlled: BOB,
        source_name: "Emrakul, the Promised End",
        until: "turn",
      },
    ],
    legal_moves: [bobsPass, bobsLand],
    legal_actions: { pass: false, sources: {} },
    controlled_legal_actions: {
      [BOB]: { pass: true, sources: { forest: { kinds: ["land"], moves: 1, zones: ["hand"] } } },
    },
  } as unknown as GameView;
}

interface Sent {
  type: ActionType;
  params?: unknown;
  player?: string;
}

// mountAs mounts the Board the way Game.svelte does for Alice with the
// toggle at `pref`.
function mountAs(pref: PlayingAsPref) {
  const frame = alicesFrame();
  const actingID = actingSeatID(frame, ALICE, pref);
  const legal = legalActionsOf(frame, actingID);
  const sent: Sent[] = [];
  const r = render(
    Board as never,
    {
      view: viewForSeat(frame, actingID, ALICE),
      viewerID: actingID,
      isAdmin: false,
      sendAction: (type: ActionType, params?: unknown, player?: string) =>
        sent.push({ type, params, player }),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      legal: visibleHighlights(legal, true),
      legalGate: legal,
    } as never,
  );
  const self = () => r.container.querySelector<HTMLElement>(".seat-panel.seat-self")!;
  const selfHandNames = () =>
    [...self().querySelectorAll<HTMLElement>(".hand .card")].map(
      (el) => (el.getAttribute("aria-label") ?? "").split(",")[0],
    );
  const selfCard = (id: string) =>
    self().querySelector<HTMLElement>(`.hand .card[data-instance-id="${id}"]`);
  return { ...r, sent, actingID, self, selfHandNames, selfCard };
}

describe("the swapped board", () => {
  it("puts Bob at the bottom with his hand face up, and not Alice's", () => {
    const b = mountAs("controlled");
    expect(b.actingID).toBe(BOB);
    expect(b.selfHandNames().sort()).toEqual(["Forest", "Grizzly Bears"]);
    expect(b.selfCard("island")).toBeNull();
  });

  it("lights Bob's move from his own digest", () => {
    const b = mountAs("controlled");
    expect(b.selfCard("forest")?.classList.contains("ready")).toBe(true);
    expect(b.selfCard("bears")?.classList.contains("ready")).toBe(false);
  });

  it("plays Bob's land as Bob", () => {
    const b = mountAs("controlled");
    click(b.selfCard("forest")!);
    flushSync();
    const play = b.sent.find((s) => s.type === "cast_spell");
    expect(play, JSON.stringify(b.sent)).toBeTruthy();
    expect(play!.player).toBe(BOB);
    expect((play!.params as { instance_id?: string }).instance_id).toBe("forest");
  });

  it("is Alice's own board again on View my board", () => {
    const b = mountAs("self");
    expect(b.actingID).toBe(ALICE);
    expect(b.selfHandNames()).toEqual(["Island"]);
    // Her own digest offers nothing: Bob holds priority.
    expect(b.selfCard("island")?.classList.contains("ready")).toBe(false);
  });
});

describe("PlayingAsBanner", () => {
  function mountBanner(viewer: string | null, acting: string | null, onToggle?: () => void) {
    const frame = alicesFrame();
    const line = playingAsBanner(frame, viewer, acting)!;
    return render(PlayingAsBanner as never, { line, seats: frame.seats, onToggle } as never);
  }

  it("gives Alice the toggle", () => {
    const toggle = vi.fn();
    const r = mountBanner(ALICE, BOB, toggle);
    expect(r.container.textContent).toContain("Playing as Bob");
    expect(r.container.textContent).toContain("Emrakul, the Promised End");
    const button = r.container.querySelector<HTMLButtonElement>("button")!;
    expect(button.textContent?.trim()).toBe("View my board");
    click(button);
    expect(toggle).toHaveBeenCalledOnce();
  });

  it("gives a spectator the text only", () => {
    const r = mountBanner(null, null);
    expect(r.container.textContent).toContain("Alice controls Bob this turn");
    expect(r.container.querySelector("button")).toBeNull();
  });

  it("tells Bob who controls him", () => {
    const r = mountBanner(BOB, BOB);
    expect(r.container.textContent).toContain("Alice controls you this turn");
    expect(r.container.querySelector("button")).toBeNull();
  });
});
