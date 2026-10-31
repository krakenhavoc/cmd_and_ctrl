// @vitest-environment jsdom
//
// skippedFirstTarget.render.test.ts — #4497. Meager Meal, the Adventure
// half of Gollum, Silent Slinker, reads "Put a +1/+1 counter on up to one
// target creature. Target player gains 2 life." Its first clause may be
// left empty (CR 601.2c: a variable count's number of targets is
// announced first, and "up to one" allows none), and then the one pick
// answers the SECOND clause. The
// server files a pick by the clause it names, so the cast_spell has to
// say `slot: 1` for the player; a pick that said nothing used to be filed
// under the creature clause and refused.
//
// The view below is the shape the server projects for that card in its
// owner's hand (skipped_first_target_4497_test.go pins the server side):
// faces[1] carries both clauses, the first 0..1.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import ActionDock from "./components/board/ActionDock.svelte";
import Board from "./components/board/Board.svelte";
import { _resetForTests as resetDock, pushDockRequest } from "./dock";
import { cardAsFace } from "./faces";
import { resetSettings } from "./settings";
import {
  advance,
  allPicks,
  begin,
  canConfirm,
  confirm,
  isLegalPlayerTarget,
  isMultiPick,
  pickPlayer,
  targeting,
  type TargetingMode,
} from "./targeting";
import { targetingRequest } from "./targetingDock";
import type { ActionType, CardView, GameView, PlayerView } from "./protocol";
import { nameOf } from "./test/dockView";
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

// gollum is the hand card as the server projects it: the creature face
// up, the Adventure's announce block on faces[1]. `creatures` is the
// first clause's legal set; empty is a table with no creature on it.
function gollum(creatures: string[] = ["bear"]): CardView {
  return {
    instance_id: "gollum",
    name: "Gollum, Silent Slinker",
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Legendary Creature — Halfling Horror",
    mana_cost: "{3}{B}",
    layout: "adventure",
    faces: [
      {
        name: "Gollum, Silent Slinker",
        type_line: "Legendary Creature — Halfling Horror",
        mana_cost: "{3}{B}",
      },
      {
        name: "Meager Meal",
        type_line: "Sorcery — Adventure",
        mana_cost: "{B}",
        target_mode: "creature",
        legal_targets: { cards: creatures, min: 0, max: 1 },
        clauses: [
          { cards: creatures, min: 0, max: 1, label: "up to one target creature" },
          { players: [ME, BOB], min: 1, max: 1, label: "target player" },
        ],
      },
    ],
  } as unknown as CardView;
}

function startMeal(creatures?: string[]): void {
  const meal = cardAsFace(gollum(creatures), 1);
  begin(meal, (meal.target_mode ?? "any") as TargetingMode, { face: 1 });
}

describe("the walk with the first clause left empty (#4497)", () => {
  it("stamps the player with the clause it answered", () => {
    startMeal();
    const first = get(targeting)!;
    expect(first.label).toBe("up to one target creature");
    // "Up to one" is a pick list, and Done is open with nothing picked.
    expect(isMultiPick(first)).toBe(true);
    expect(canConfirm(first)).toBe(true);

    const second = advance(first)!;
    expect(second.step).toBe(1);
    expect(second.label).toBe("target player");
    expect(second.done).toEqual([]);
    expect(isLegalPlayerTarget(second, BOB)).toBe(true);

    const last = { ...second, picked: [{ kind: "player" as const, id: BOB }] };
    expect(advance(last)).toBeNull();
    expect(allPicks(last)).toEqual([{ kind: "player", id: BOB, mode: 0, slot: 1 }]);
  });

  it("asks the empty clause even with no creature to choose, and still stamps slot 1", () => {
    startMeal([]);
    const first = get(targeting)!;
    expect(first.legal?.cards.size).toBe(0);
    expect(canConfirm(first)).toBe(true);
    const last = { ...advance(first)!, picked: [{ kind: "player" as const, id: ME }] };
    expect(allPicks(last)).toEqual([{ kind: "player", id: ME, mode: 0, slot: 1 }]);
  });

  it("both clauses answered keep their own slots", () => {
    startMeal();
    const first = { ...get(targeting)!, picked: [{ kind: "card" as const, id: "bear" }] };
    const last = { ...advance(first)!, picked: [{ kind: "player" as const, id: BOB }] };
    expect(allPicks(last)).toEqual([
      { kind: "card", id: "bear", mode: 0, slot: 0 },
      { kind: "player", id: BOB, mode: 0, slot: 1 },
    ]);
  });
});

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, name: string, n: number, hand: CardView[] = []): PlayerView =>
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
  }) as unknown as PlayerView;

const gameView = (): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0, [gollum()]), seat(BOB, "Bob", 1)],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "precombat_main" },
    mulligans_open: false,
  }) as unknown as GameView;

function mountTable() {
  const view = gameView();
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
  // What Game.svelte does on every step of the walk: the dock's request
  // built from the live targeting state, Done and the player row wired to
  // Board's handlers.
  const prompt = () => {
    resetDock();
    pushDockRequest(
      targetingRequest(get(targeting)!, view, {
        onDone: confirm,
        onCancel: () => {},
        onPlayer: pickPlayer,
      }),
    );
    flushSync();
  };
  const button = (name: string) =>
    [...dock.container.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => nameOf(b) === name,
    );
  return { sent, prompt, button };
}

describe("on the Board, Done on the empty clause then a player (#4497)", () => {
  it("sends one cast_spell of face 1 with the player at slot 1", () => {
    const t = mountTable();
    startMeal();
    flushSync();

    t.prompt();
    const done = t.button("Done")!;
    expect(done.disabled).toBe(false);
    click(done);
    expect(t.sent).toEqual([]);
    expect(get(targeting)?.label).toBe("target player");

    t.prompt();
    click(t.button("Target Bob")!);
    expect(t.sent).toEqual([
      {
        type: "cast_spell",
        params: {
          instance_id: "gollum",
          targets: [{ kind: "player", id: BOB, mode: 0, slot: 1 }],
          face: 1,
        },
      },
    ]);
    expect(get(targeting)).toBeNull();
  });
});
