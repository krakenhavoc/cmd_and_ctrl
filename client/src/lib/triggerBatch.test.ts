// triggerBatch.test.ts — #3817: twenty Blood Artists and a wrath are
// dozens of target prompts that look alike. What the table is told about
// the batch: how many, how many still need a target, and which one a
// prompt answers.

import { describe, expect, it } from "vitest";

import type { CardView, GameView, PendingChoiceView, StackItemView } from "./protocol";
import { batchPosition, cardNames, waitingLine, waitingTriggers } from "./triggerBatch";

const ME = "me";
const BOT = "bot";

const artist = (id: string): CardView =>
  ({ instance_id: id, name: "Blood Artist", owner: ME, controller: ME }) as CardView;

const prompt = (id: string, source: string, chooser = ME): PendingChoiceView =>
  ({
    id,
    kind: "pick_target",
    chooser,
    from_player: chooser,
    count: 0,
    source,
    reason: "target player",
    pick_target: { players: [ME, BOT], min: 1, max: 1 },
  }) as PendingChoiceView;

const answered = (id: string, source: string, controller = ME): StackItemView =>
  ({
    id,
    kind: "triggered",
    controller,
    owner: controller,
    source_card_id: source,
    label: "Blood Artist — drain 1",
  }) as StackItemView;

function view(choices: PendingChoiceView[], pending: StackItemView[] = []): GameView {
  return {
    battlefield: { kind: "battlefield", count: 3, cards: [artist("a1"), artist("a2")] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    seats: [
      { id: ME, name: "LUke", graveyard: { kind: "graveyard", count: 1, cards: [artist("a3")] } },
      { id: BOT, name: "Bot 1" },
    ],
    pending_choices: choices,
    pending_triggers: pending,
  } as unknown as GameView;
}

describe("batchPosition (#3817)", () => {
  it("counts copies of one card as one batch, the answered ones first", () => {
    // Two answered (one from a Blood Artist that died in the wrath too),
    // four still asking.
    const v = view(
      [prompt("p1", "a1"), prompt("p2", "a2"), prompt("p3", "a3"), prompt("p4", "a1")],
      [answered("t1", "a2"), answered("t2", "a3")],
    );
    expect(batchPosition(v, "p1")).toEqual({ index: 3, total: 6 });
    expect(batchPosition(v, "p4")).toEqual({ index: 6, total: 6 });
  });

  it("is null for a lone trigger, an unknown choice or another kind", () => {
    expect(batchPosition(view([prompt("p1", "a1")]), "p1")).toBeNull();
    expect(batchPosition(view([prompt("p1", "a1")]), "nope")).toBeNull();
    expect(batchPosition(view([prompt("p1", "a1")]), undefined)).toBeNull();
    const order = { ...prompt("o", "a1"), kind: "trigger_order" } as PendingChoiceView;
    expect(batchPosition(view([order, prompt("p1", "a2")]), "o")).toBeNull();
  });

  it("does not mix another player's batch into yours", () => {
    const v = view([prompt("p1", "a1"), prompt("p2", "a2", BOT)], [answered("t1", "a1", BOT)]);
    expect(batchPosition(v, "p1")).toBeNull();
    expect(batchPosition(v, "p2")).toEqual({ index: 2, total: 2 });
  });
});

describe("waitingTriggers and waitingLine (#3817)", () => {
  it("groups by player and name, counting the ones still asking for a target", () => {
    const v = view(
      [prompt("p1", "a1"), prompt("p2", "a2"), prompt("p3", "a3")],
      [answered("t1", "a1")],
    );
    const groups = waitingTriggers(v.pending_choices, v.pending_triggers, cardNames(v));
    expect(groups).toEqual([{ name: "Blood Artist", player: ME, count: 4, needTarget: 3 }]);
    expect(waitingLine(groups, ME, () => undefined)).toBe(
      "Blood Artist ×4 waiting to go on the stack · 3 need your target",
    );
  });

  it("names the player a spectator or an opponent waits on, and drops the count once answered", () => {
    const groups = [
      { name: "Blood Artist", player: BOT, count: 5, needTarget: 1 },
      { name: "Zulaport Cutthroat", player: ME, count: 2, needTarget: 0 },
    ];
    expect(waitingLine(groups, ME, (id) => (id === BOT ? "Bot 1" : undefined))).toBe(
      "Blood Artist ×5 waiting to go on the stack · 1 needs Bot 1's target; " +
        "Zulaport Cutthroat ×2 waiting to go on the stack",
    );
  });

  it("falls back to the trigger's label for a source no longer in view (a token)", () => {
    const groups = waitingTriggers([], [answered("t1", "gone"), answered("t2", "gone")], new Map());
    expect(groups).toEqual([{ name: "Blood Artist", player: ME, count: 2, needTarget: 0 }]);
  });
});
