// @vitest-environment jsdom
//
// commanderCastChain.render.test.ts — #2202, on the REAL Board. Before
// #2202 the command zone panel's cast sent a bare cast_spell
// `{instance_id, from_zone: "command"}` straight to the server, so an
// X-cost commander was cast with no X picker at all. It now starts the
// Board's one cast chain, the one the castable-from-other-zones strip
// beside the hand also starts, so:
//
//   - an X-cost commander cast from the command zone opens the chain's
//     X picker before anything is sent;
//   - a plain commander sends the bare cast from the command zone. Since
//     #2349 the strip is the only surface: the tile is gone.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./api", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  return {
    ...actual,
    fetchAutoTapPreview: () => Promise.resolve({ ok: true, cost: "{G}{U}", plan: [] }),
  };
});

import Board from "./components/board/Board.svelte";
import { activeDockRequest, _resetForTests as resetDock } from "./dock";
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

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const commander = (id: string, name: string, cost: string): CardView =>
  ({
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    known_by_you: true,
    is_commander: true,
    type_line: "Legendary Creature — Elemental",
    mana_cost: cost,
  }) as CardView;

const seat = (command: CardView[]): PlayerView =>
  ({
    id: ME,
    name: "Me",
    seat: 0,
    life: 40,
    library: zone("library", ME),
    hand: zone("hand", ME),
    graveyard: zone("graveyard", ME),
    command: zone("command", ME, command),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const gameView = (command: CardView[]): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(command)],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountBoard(command: CardView[]) {
  const sent: Sent[] = [];
  const r = render(
    Board as never,
    {
      view: gameView(command),
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
  const stripCard = (name: string) =>
    r.container.querySelector<HTMLElement>(`.exile-strip .card[aria-label='${name}']`)!;
  return { ...r, sent, stripCard };
}

// #2349: the strip beside the hand is the only place a commander is cast
// from now; the command zone tile and its cast button are gone.
describe("the command zone's cast walks the Board's cast chain — #2202", () => {
  it("an X-cost commander clicked in the strip opens the X picker first", () => {
    const b = mountBoard([commander("verazol", "Verazol", "{X}{G}{U}")]);
    click(b.stripCard("Verazol"));
    expect(b.sent).toEqual([]);
    expect(get(activeDockRequest)?.label).toBe("Choose X for Verazol");
  });

  it("a plain commander sends a cast from the command zone", () => {
    const b = mountBoard([commander("kenrith", "Kenrith", "{4}{W}")]);
    click(b.stripCard("Kenrith"));
    expect(b.sent).toEqual([
      { type: "cast_spell", params: { instance_id: "kenrith", from_zone: "command" } },
    ]);
  });

  it("the board has no command zone tile left to cast from", () => {
    const b = mountBoard([commander("kenrith", "Kenrith", "{4}{W}")]);
    expect(b.container.querySelector(".cmd-zone")).toBeNull();
    expect(b.container.querySelector("[aria-label='Me command zone, 1 card']")).not.toBeNull();
  });
});
