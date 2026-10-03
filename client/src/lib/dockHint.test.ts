// dockHint.test.ts — ADR 0111 §1: the action dock's "what does next
// do" line. Public state only: the top of the stack and the step.

import { describe, it, expect } from "vitest";

import { passHint } from "./dockHint";
import type { CardView, GameView, StackItemView } from "./protocol";

const zone = (kind: string, cards: CardView[] = []) => ({ kind, count: cards.length, cards });

function view(step: string, over: Partial<GameView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [{ id: "a", name: "Alice", seat: 0 }],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    turn: { number: 2, active_seat: 0, priority_holder: 0, phase: step, step },
    mulligans_open: false,
    ...over,
  } as unknown as GameView;
}

const bolt: CardView = {
  instance_id: "bolt",
  name: "Lightning Bolt",
  owner: "a",
  controller: "a",
  known_by_you: true,
} as CardView;

const boltItem: StackItemView = {
  id: "bolt",
  kind: "spell",
  controller: "a",
  owner: "a",
  source_card_id: "bolt",
};

describe("passHint", () => {
  it("says nothing without priority, during mulligans, or with no view", () => {
    expect(passHint(view("upkeep"), false)).toBe("");
    expect(passHint(view("upkeep", { mulligans_open: true }), true)).toBe("");
    expect(passHint(null, true)).toBe("");
  });

  it("names the top of the stack", () => {
    const v = view("precombat_main", {
      stack: zone("stack", [bolt]) as GameView["stack"],
      stack_items: [boltItem],
    });
    expect(passHint(v, true)).toBe("passing lets Lightning Bolt resolve");
  });

  it.each([
    ["upkeep", "passing moves to Draw"],
    ["draw", "passing moves to Main 1"],
    ["precombat_main", "passing moves to Begin Combat"],
    ["begin_combat", "passing moves to Declare Attackers"],
    ["declare_blockers", "passing moves to Combat Damage"],
    ["first_strike_damage", "passing moves to Combat Damage"],
    ["combat_damage", "passing moves to End Combat"],
    ["end_combat", "passing moves to Main 2"],
    ["postcombat_main", "passing moves to End"],
    ["end", "passing ends the turn"],
  ])("on an empty stack at %s: %s", (step, want) => {
    expect(passHint(view(step), true)).toBe(want);
  });

  it("skips to end of combat when nothing attacked (CR 508.8)", () => {
    expect(passHint(view("declare_attackers"), true)).toBe("passing moves to End Combat");
    const bear = { instance_id: "bear", name: "Bear", attacking_target: "b" } as CardView;
    const attacking = view("declare_attackers", {
      battlefield: zone("battlefield", [bear]) as GameView["battlefield"],
    });
    expect(passHint(attacking, true)).toBe("passing moves to Declare Blockers");
  });

  // #1501: nobody holds priority while defenders declare blockers, so
  // the line names who the table is waiting on — for every viewer
  // without priority, which in that window is every viewer.
  it("names the defenders still declaring while priority is parked", () => {
    const seats = [
      { id: "a", name: "Alice", seat: 0 },
      { id: "b", name: "Bob", seat: 1 },
      { id: "c", name: "Cara", seat: 2 },
      { id: "d", name: "Dev", seat: 3 },
    ] as unknown as GameView["seats"];
    const parked = (pending: number[]) =>
      view("declare_blockers", {
        seats,
        turn: {
          number: 2,
          active_seat: 0,
          priority_holder: -1,
          phase: "combat",
          step: "declare_blockers",
          block_pending_seats: pending,
        },
      } as unknown as Partial<GameView>);
    expect(passHint(parked([1]), false)).toBe("waiting for Bob to declare blockers");
    expect(passHint(parked([1, 3]), false)).toBe("waiting for Bob and Dev to declare blockers");
    expect(passHint(parked([1, 2, 3]), false)).toBe(
      "waiting for Bob, Cara and Dev to declare blockers",
    );
    // Once the last declaration is in, the active player holds
    // priority again and the line goes back to what `next` does.
    const done = view("declare_blockers", { seats });
    expect(passHint(done, false)).toBe("");
    expect(passHint(done, true)).toBe("passing moves to Combat Damage");
    expect(passHint(parked([]), false)).toBe("");
  });

  it("says nothing on a step with no priority or one it does not know", () => {
    expect(passHint(view("untap"), true)).toBe("");
    expect(passHint(view("cleanup"), true)).toBe("");
    expect(passHint(view("some_future_step"), true)).toBe("");
  });
});
