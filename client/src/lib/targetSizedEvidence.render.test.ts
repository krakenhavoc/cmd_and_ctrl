// @vitest-environment jsdom
//
// targetSizedEvidence.render.test.ts — #5240, on the REAL Board. Urgent
// Necropsy's "collect evidence X, where X is the total mana value of the
// permanents this spell targets": the cast chain opens targeting first
// (CR 601.2c), and the graveyard picker only after the targets are picked,
// with their total as its floor (CR 601.2f). One cast_spell carries both.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));
vi.mock("./api", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  return {
    ...actual,
    fetchAutoTapPreview: () => Promise.resolve({ ok: true, cost: "{2}{B}{G}", plan: [] }),
  };
});

import ActionDock from "./components/board/ActionDock.svelte";
import Board from "./components/board/Board.svelte";
import { activeDockRequest, _resetForTests as resetDock } from "./dock";
import { resetSettings } from "./settings";
import { confirm, targeting } from "./targeting";
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
});

afterEach(() => {
  cleanup();
  targeting.set(null);
  resetDock();
  resetSettings();
});

const ME = "me";
const THEM = "them";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const clause = (label: string, cards: string[]) => ({ cards, min: 0, max: 1, label });

// necropsy is the hand card as the server projects it: four "up to one"
// clauses (only the creature one has a legal target here) and the
// target-sized collect evidence cost.
const necropsy = (): CardView =>
  ({
    instance_id: "necropsy",
    name: "Urgent Necropsy",
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Instant",
    mana_cost: "{2}{B}{G}",
    castable_here: true,
    target_mode: "permanent",
    legal_targets: clause("up to one target artifact", []),
    clauses: [
      clause("up to one target artifact", []),
      clause("up to one target creature", ["bear"]),
      clause("up to one target enchantment", []),
      clause("up to one target planeswalker", []),
    ],
    additional_cost: {
      label:
        "Collect evidence X, where X is the total mana value of the permanents this spell targets",
      exile_cost_label: "cards from your graveyard with total mana value X or greater",
      exile_cost_options: ["one", "two"],
      exile_cost_zone: "graveyard",
      exile_cost_from_targets: true,
      exile_cost_target_mana_values: { bear: 3 },
    },
  }) as unknown as CardView;

const gyCard = (id: string, cost: string): CardView =>
  ({
    instance_id: id,
    name: `Evidence ${id}`,
    owner: ME,
    controller: ME,
    type_line: "Sorcery",
    mana_cost: cost,
  }) as CardView;

const bear = (): CardView =>
  ({
    instance_id: "bear",
    name: "Bear",
    owner: THEM,
    controller: THEM,
    type_line: "Creature — Bear",
    mana_cost: "{2}{G}",
  }) as CardView;

const seat = (id: string, n: number, hand: CardView[], graveyard: CardView[]): PlayerView =>
  ({
    id,
    name: id === ME ? "Me" : "Them",
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id, graveyard),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const gameView = (): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [
      seat(ME, 0, [necropsy()], [gyCard("one", "{1}"), gyCard("two", "{2}")]),
      seat(THEM, 1, [], []),
    ],
    battlefield: zone("battlefield", undefined, [bear()]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    legal_moves: [
      {
        type: "cast_spell",
        player: ME,
        kind: "cast",
        source: "necropsy",
        label: "Cast Urgent Necropsy",
        params: { instance_id: "necropsy", from_zone: "hand" },
      },
    ],
  }) as unknown as GameView;

function mountTable() {
  const view = gameView();
  const sent: { type: ActionType; params?: unknown }[] = [];
  const board = render(
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
  const handCard = (name: string) =>
    board.container.querySelector<HTMLElement>(`.hand .card[aria-label='${name}']`);
  const option = (name: string) =>
    [...dock.container.querySelectorAll<HTMLButtonElement>("button.prompt-opt")].find((b) =>
      b.textContent?.includes(name),
    );
  return { sent, handCard, option };
}

async function settle(): Promise<void> {
  await Promise.resolve();
  await Promise.resolve();
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
}

describe("Urgent Necropsy's evidence is asked after its targets (#5240)", () => {
  it("opens targeting first, then the graveyard picker with the targets' total", async () => {
    const t = mountTable();
    click(t.handCard("Urgent Necropsy")!);
    await settle();
    // Targeting is open on the creature clause; no cost picker yet.
    expect(get(targeting)?.label).toBe("up to one target creature");
    expect(get(activeDockRequest)?.label).not.toBe("Urgent Necropsy");
    expect(t.sent).toEqual([]);

    targeting.set({ ...get(targeting)!, picked: [{ kind: "card", id: "bear" }] });
    confirm();
    await settle();
    // The Bear is worth 3: the picker asks for that much, and nothing is
    // sent until it is paid.
    expect(t.sent).toEqual([]);
    expect(get(targeting)).toBeNull();
    const req = get(activeDockRequest);
    expect(req?.label).toBe("Urgent Necropsy");
    expect(req?.sheet?.count).toBe("0 picked · mana value 0 / 3");
    expect(req?.primary?.disabled).toBe(true);

    click(t.option("Evidence one")!);
    click(t.option("Evidence two")!);
    await settle();
    expect(get(activeDockRequest)?.primary?.disabled).toBe(false);
    get(activeDockRequest)!.primary!.onPress();
    await settle();
    expect(t.sent).toHaveLength(1);
    expect(t.sent[0].type).toBe("cast_spell");
    const params = t.sent[0].params as Record<string, unknown>;
    expect(params.exile_ids).toEqual(["one", "two"]);
    expect(params.targets).toEqual([{ kind: "card", id: "bear", mode: 0, slot: 1 }]);
  });

  it("casts with no evidence asked when no target is picked (X = 0)", async () => {
    const t = mountTable();
    click(t.handCard("Urgent Necropsy")!);
    await settle();
    expect(get(targeting)?.label).toBe("up to one target creature");
    confirm();
    await settle();
    expect(get(activeDockRequest)?.label).not.toBe("Urgent Necropsy");
    expect(t.sent).toHaveLength(1);
    const params = t.sent[0].params as Record<string, unknown>;
    expect(params).not.toHaveProperty("exile_ids");
  });
});
