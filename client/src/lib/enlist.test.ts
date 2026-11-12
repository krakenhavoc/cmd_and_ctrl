// enlist.test.ts — #5658, the pure half of the client's enlist choice
// (CR 702.154): the digest lookup, the click flow's selection and stage,
// the dock's question, the attack-with-all plan, the card menu's rows and
// the refusal sentence.

import { describe, expect, it } from "vitest";

import { attackAllParams, enlistChoiceAt, offersAttackPicker, planAttackAll } from "./attackAll";
import { joinsSelection } from "./attackSelect";
import { stageAt, stageMany, stagedParams } from "./attackStage";
import { combatSelectionRequest } from "./combatDock";
import { buildMenuSections, type MenuItem } from "./contextMenu.logic";
import {
  CANT_ENLIST_SENTENCE,
  ENLIST_NOTE,
  enlistCandidates,
  enlistOptionLabel,
  enlistOptions,
  enlistTaken,
  enlistUnanswered,
  isCantEnlistError,
  withEnlistQuestion,
} from "./enlist";
import { declareAttackerParams, selectAttackerFor } from "./exert";
import { L } from "./labels";
import { NO_LEGAL_ACTIONS, legalActionsOf } from "./legalActions";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string, n: number): PlayerView {
  return {
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

const creature = (
  id: string,
  name: string,
  power: number,
  over: Partial<CardView> = {},
): CardView => ({
  instance_id: id,
  name,
  owner: "me",
  controller: "me",
  type_line: "Creature — Elk",
  power,
  ...over,
});

// A duel in declare attackers: Barkweave Crusher has enlist and may
// enlist the Elk or either of two identical Soldier tokens; all four can
// attack Bob.
function snap(crusher: Partial<CardView> = {}): GameView {
  const attack = { kinds: ["attack" as const], moves: 1, attack_targets: ["bob"] };
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
    battlefield: zone("battlefield", undefined, [
      creature("crusher", "Barkweave Crusher", 2, crusher),
      creature("elk", "Elk", 3),
      creature("s1", "Soldier", 1),
      creature("s2", "Soldier", 1),
    ]),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
    },
    mulligans_open: false,
    legal_actions: {
      pass: true,
      sources: {
        ...(crusher.attacking_target
          ? {}
          : {
              crusher: {
                kinds: ["attack" as const],
                moves: 4,
                attack_targets: ["bob"],
                enlist_on_attack: ["elk", "s1", "s2"],
              },
            }),
        elk: attack,
        s1: attack,
        s2: attack,
      },
    },
  };
}

describe("the lookup", () => {
  it("reads enlist_on_attack off the digest", () => {
    const gate = legalActionsOf(snap());
    expect(gate.enlistOnAttack("crusher")).toEqual(["elk", "s1", "s2"]);
    expect(gate.enlistOnAttack("elk")).toEqual([]);
    expect(NO_LEGAL_ACTIONS.enlistOnAttack("crusher")).toEqual([]);
  });

  it("agrees when folded from the move list instead of the digest", () => {
    const v = snap();
    delete v.legal_actions;
    const move = (enlist?: string) => ({
      type: "declare_attacker" as const,
      player: "me",
      kind: "attack" as const,
      label: "Attack Bob with Barkweave Crusher",
      source: "crusher",
      params: enlist
        ? { attacker: "crusher", target: "bob", enlist }
        : { attacker: "crusher", target: "bob" },
    });
    v.legal_moves = [move(), move("elk"), move("s1")];
    const gate = legalActionsOf(v);
    expect(gate.enlistOnAttack("crusher")).toEqual(["elk", "s1"]);
    expect(gate.attackTargets("crusher")).toEqual(["bob"]);
  });

  it("leaves out what the stage has taken", () => {
    const gate = legalActionsOf(snap());
    const taken = enlistTaken([{ attacker: "elk" }, { attacker: "x", enlist: "s1" }]);
    expect(enlistCandidates("crusher", gate, taken)).toEqual(["s2"]);
  });
});

describe("the click flow", () => {
  const gate = legalActionsOf(snap());

  it("asks before an enlist creature's attack is staged or sent", () => {
    const sel = withEnlistQuestion(selectAttackerFor("crusher", gate), gate);
    expect(sel.enlist).toBeNull();
    expect(enlistUnanswered(sel)).toBe(true);
    expect(stageAt([], sel, "bob")).toBeNull();
    expect(declareAttackerParams(sel, "bob")).toBeNull();
  });

  it("stages the answer, and sends enlist only when a creature was chosen", () => {
    const sel = withEnlistQuestion(selectAttackerFor("crusher", gate), gate);
    const yes = stageAt([], { ...sel, enlist: "elk" }, "bob");
    expect(yes).toEqual([{ attacker: "crusher", target: "bob", enlist: "elk" }]);
    expect(stagedParams(yes!)?.attackers).toEqual([
      { attacker: "crusher", target: "bob", enlist: "elk" },
    ]);
    expect(stageAt([], { ...sel, enlist: "" }, "bob")).toEqual([
      { attacker: "crusher", target: "bob" },
    ]);
    expect(declareAttackerParams({ ...sel, enlist: "elk" }, "bob")).toEqual({
      attacker: "crusher",
      target: "bob",
      auto_tap: true,
      enlist: "elk",
    });
  });

  it("asks nothing for a creature with no one to enlist", () => {
    expect(withEnlistQuestion(selectAttackerFor("elk", gate), gate).enlist).toBeUndefined();
    const all = new Set(["elk", "s1", "s2"]);
    expect(
      withEnlistQuestion(selectAttackerFor("crusher", gate), gate, all).enlist,
    ).toBeUndefined();
  });

  it("keeps a creature with an enlist question alone in the selection", () => {
    const asking = withEnlistQuestion(selectAttackerFor("crusher", gate), gate);
    const elk = selectAttackerFor("elk", gate);
    expect(joinsSelection(asking, elk, () => false)).toBe(false);
    expect(joinsSelection(elk, asking, () => false)).toBe(false);
  });

  it("a creature staged to attack is no longer enlisted", () => {
    const stage = [{ attacker: "crusher", target: "bob", enlist: "elk" }];
    expect(stageAt(stage, { kind: "attacker", cardID: "elk" }, "bob")).toEqual([
      { attacker: "crusher", target: "bob" },
      { attacker: "elk", target: "bob" },
    ]);
    expect(stageMany(stage, [{ attacker: "elk", target: "bob" }])).toEqual([
      { attacker: "crusher", target: "bob" },
      { attacker: "elk", target: "bob" },
    ]);
  });
});

describe("the options", () => {
  it("names each candidate with its power, identical ones once", () => {
    const v = snap();
    const opts = enlistOptions(["elk", "s1", "s2"], v.battlefield.cards);
    expect(opts).toEqual([
      { id: "elk", name: "Elk", power: 3 },
      { id: "s1", name: "Soldier", power: 1 },
    ]);
    expect(opts.map(enlistOptionLabel)).toEqual(["Elk (+3)", "Soldier (+1)"]);
  });
});

describe("the dock's question", () => {
  const noop = () => {};
  const options = [
    { id: "elk", name: "Elk", power: 3 },
    { id: "s1", name: "Soldier", power: 1 },
  ];

  it("offers Attack and one Enlist per creature, with the reminder, while unanswered", () => {
    let chose: string | null = null;
    const req = combatSelectionRequest("attacker", "Barkweave Crusher", noop, null, false, {
      chosen: null,
      options,
      onChoose: (id) => (chose = id),
    });
    expect(req.question).toBe("Attack with Barkweave Crusher: enlist a creature as it attacks?");
    expect(req.detail).toBe(ENLIST_NOTE);
    expect(req.row?.map((a) => a.label)).toEqual([
      L.attackPlain,
      L.attackAndEnlist("Elk (+3)"),
      L.attackAndEnlist("Soldier (+1)"),
    ]);
    req.row![1]!.onPress();
    expect(chose).toBe("elk");
    req.row![0]!.onPress();
    expect(chose).toBe("");
  });

  it("says which was chosen once answered", () => {
    const req = combatSelectionRequest("attacker", "Barkweave Crusher", noop, null, false, {
      chosen: "elk",
      options,
      onChoose: noop,
    });
    expect(req.question).toMatch(/^Attacking and enlisting Elk with Barkweave Crusher/);
    expect(req.row?.map((a) => a.pressed)).toEqual([false, true, false]);
    const none = combatSelectionRequest("attacker", "Barkweave Crusher", noop, null, false, {
      chosen: "",
      options,
      onChoose: noop,
    });
    expect(none.question).toMatch(/^Attacking with Barkweave Crusher/);
    expect(none.row?.map((a) => a.pressed)).toEqual([true, false, false]);
  });
});

describe("attack with all", () => {
  it("lists who may enlist whom and opens the picker for them", () => {
    const v = snap();
    const plan = planAttackAll(v, "me", legalActionsOf(v));
    expect(plan.enlistable).toEqual({ crusher: ["elk", "s1", "s2"] });
    expect(enlistChoiceAt(plan, "bob")).toBe(true);
    expect(offersAttackPicker(v, plan, "bob")).toBe(true);
  });

  it("sends enlist only for a listed creature that is not attacking or already enlisted", () => {
    const v = snap();
    const plan = planAttackAll(v, "me", legalActionsOf(v));
    expect(
      attackAllParams(plan, "bob", { only: ["crusher", "s1"], enlist: { crusher: "elk" } })
        ?.attackers,
    ).toEqual([
      { attacker: "crusher", target: "bob", enlist: "elk" },
      { attacker: "s1", target: "bob" },
    ]);
    // The Elk is sent to attack too: it can't be enlisted.
    expect(
      attackAllParams(plan, "bob", { only: ["crusher", "elk"], enlist: { crusher: "elk" } })
        ?.attackers,
    ).toEqual([
      { attacker: "crusher", target: "bob" },
      { attacker: "elk", target: "bob" },
    ]);
    // A creature the server did not list, or an attacker without enlist.
    expect(
      attackAllParams(plan, "bob", {
        only: ["crusher", "elk"],
        enlist: { crusher: "x", elk: "s1" },
      })?.attackers,
    ).toEqual([
      { attacker: "crusher", target: "bob" },
      { attacker: "elk", target: "bob" },
    ]);
  });
});

describe("the card menu", () => {
  const combat = (v: GameView, id: string): MenuItem[] => {
    const gate = legalActionsOf(v);
    const card = v.battlefield.cards.find((c) => c.instance_id === id)!;
    return (
      buildMenuSections(v, card, "me", false, gate, gate).find((s) => s.id === "combat")?.items ??
      []
    );
  };

  it("has one Declare attacker and enlist row per creature it may enlist", () => {
    const items = combat(snap(), "crusher");
    const rows = items.filter((i) => i.id.startsWith("combat-attack-enlist-"));
    expect(rows.map((r) => r.label)).toEqual([
      L.declareAttackerAndEnlist("Elk (+3)"),
      L.declareAttackerAndEnlist("Soldier (+1)"),
    ]);
    expect(rows[0]!.items?.[0]?.action?.params).toEqual({
      attacker: "crusher",
      target: "bob",
      auto_tap: true,
      enlist: "elk",
    });
  });

  it("has no enlist row for a creature without candidates, or a re-point", () => {
    expect(combat(snap(), "elk").some((i) => i.id.startsWith("combat-attack-enlist-"))).toBe(false);
    const declared = snap({ attacking_target: "bob", tapped: true });
    expect(combat(declared, "crusher").some((i) => i.id.startsWith("combat-attack-enlist-"))).toBe(
      false,
    );
  });

  it("does not send Attack with all blind past a creature that may enlist", () => {
    const all = combat(snap(), "elk").find((i) => i.id === "combat-attack-all")!;
    const bob = all.items!.find((i) => i.id === "combat-attack-all-bob")!;
    expect(bob.disabled).toBe(true);
    expect(bob.action).toBeUndefined();
  });
});

describe("the refusal", () => {
  it("recognises the server's enlist refusals and says them plainly", () => {
    expect(isCantEnlistError("game: that creature can't be enlisted by this attacker")).toBe(true);
    expect(isCantEnlistError("game: this creature has been enlisted, so it can't attack")).toBe(
      true,
    );
    expect(isCantEnlistError("game: this creature can't be exerted as it attacks")).toBe(false);
    expect(isCantEnlistError(undefined)).toBe(false);
    expect(CANT_ENLIST_SENTENCE).toMatch(/nothing was declared/);
  });
});
