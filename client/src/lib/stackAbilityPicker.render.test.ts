// @vitest-environment jsdom
//
// stackAbilityPicker.render.test.ts — #4043, on the REAL Board. A spell's
// own stack ability (Lightning Storm: "Discard a land card: ...") asks
// which land to discard in the Board's discard picker. If the storm
// resolves or is countered while the player is still choosing, the engine
// would refuse the activation, so the picker closes with a short notice
// (L.stackAbilityGone). A picker opened for a card that was never on the
// stack is left alone.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {}, arm: () => {}, preload: () => {} }));

import Board from "./components/board/Board.svelte";
import { get } from "svelte/store";
import { activeDockRequest, _resetForTests as resetDock } from "./dock";
import { L } from "./labels";
import type { ActionType, CardView, GameView, PlayerView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

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
});

const ME = "me";
const OPP = "opp";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const land = (id: string, name: string): CardView =>
  ({
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Basic Land",
  }) as CardView;

const STORM_ROW = {
  index: 0,
  ref: "own:0",
  label: "Discard a land card: Put two charge counters on Lightning Storm.",
  any_player: true,
  discard_cost_n: 1,
  discard_cost_label: "a land card",
  discard_cost_options: ["forest", "island"],
};

const storm = (): CardView =>
  ({
    instance_id: "storm",
    name: "Lightning Storm",
    owner: OPP,
    controller: OPP,
    known_by_you: true,
    type_line: "Instant",
    zone_abilities: [STORM_ROW],
  }) as unknown as CardView;

function seat(id: string, n: number, hand: CardView[]): PlayerView {
  return {
    id,
    name: id,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  } as unknown as PlayerView;
}

function view(onStack: boolean): GameView {
  const hand = [land("forest", "Forest"), land("island", "Island")];
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME, 0, hand), seat(OPP, 1, [])],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined, onStack ? [storm()] : []),
    exile: zone("exile", undefined),
    stack_items: onStack
      ? [
          {
            id: "storm",
            kind: "spell",
            controller: OPP,
            owner: OPP,
            source_card_id: "storm",
            targets: [],
          },
        ]
      : [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 1, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  } as unknown as GameView;
}

function mount() {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const r = render(
    Board as never,
    {
      view: view(true),
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

const NAME = L.stackAbility("Lightning Storm", "Discard a land card");

describe("the discard picker for a stack ability — #4043", () => {
  it("closes with a notice when the spell leaves the stack, and sends nothing", () => {
    const b = mount();
    const btn = b.container.querySelector<HTMLButtonElement>(`button[aria-label="${NAME}"]`);
    expect(btn).not.toBeNull();
    click(btn!);
    // Two lands in hand, one to discard: the picker asks.
    expect(get(activeDockRequest)).not.toBeNull();
    expect(b.container.querySelector(".stack-gone")).toBeNull();

    b.setProps({ view: view(false) } as never);

    expect(b.container.querySelector(".stack-gone")?.textContent).toBe(
      L.stackAbilityGone("Lightning Storm"),
    );
    expect(get(activeDockRequest)).toBeNull();
    expect(b.sent).toEqual([]);
  });
});
