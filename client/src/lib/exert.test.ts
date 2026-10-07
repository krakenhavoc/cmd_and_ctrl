// exert.test.ts — ADR 0130 §7 (owner decision 1: every path asks).
// The pure half of the client's exert choice: the click flow's
// selection and payload, the attack-with-all plan, the dock's question,
// the card menu's rows and the refusal sentence. The picker's toggle is
// exertPicker.render.test.ts.

import { describe, expect, it } from "vitest";

import { attackAllParams, exertChoiceAt, offersAttackPicker, planAttackAll } from "./attackAll";
import { EXERT_NOTE, combatSelectionRequest } from "./combatDock";
import { buildMenuSections, type MenuItem } from "./contextMenu.logic";
import {
  CANT_EXERT_SENTENCE,
  declareAttackerParams,
  exertUnanswered,
  isCantExertError,
  selectAttackerFor,
} from "./exert";
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

const creature = (id: string, name: string, over: Partial<CardView> = {}): CardView => ({
  instance_id: id,
  name,
  owner: "me",
  controller: "me",
  type_line: "Creature — Human Warrior",
  ...over,
});

// A duel in declare attackers: Oketra's Avenger may be exerted, the
// Bear may not, and the server lists both as attackers of Bob.
function snap(withAvenger = true, avenger: Partial<CardView> = {}): GameView {
  const cards = [creature("bear", "Bear")];
  if (withAvenger) cards.unshift(creature("avenger", "Oketra's Avenger", avenger));
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
    battlefield: zone("battlefield", undefined, cards),
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
        ...(withAvenger && !avenger.attacking_target
          ? {
              avenger: {
                kinds: ["attack" as const],
                moves: 2,
                attack_targets: ["bob"],
                exert_on_attack: true,
              },
            }
          : {}),
        bear: { kinds: ["attack" as const], moves: 1, attack_targets: ["bob"] },
      },
    },
  };
}

describe("the lookup", () => {
  it("reads exert_on_attack off the digest", () => {
    const gate = legalActionsOf(snap());
    expect(gate.canExertOnAttack("avenger")).toBe(true);
    expect(gate.canExertOnAttack("bear")).toBe(false);
    expect(NO_LEGAL_ACTIONS.canExertOnAttack("avenger")).toBe(false);
  });

  it("agrees when folded from the move list instead of the digest", () => {
    const v = snap();
    delete v.legal_actions;
    v.legal_moves = [
      {
        type: "declare_attacker",
        player: "me",
        kind: "attack",
        label: "Attack Bob with Oketra's Avenger",
        source: "avenger",
        params: { attacker: "avenger", target: "bob" },
      },
      {
        type: "declare_attacker",
        player: "me",
        kind: "attack",
        label: "Attack Bob with Oketra's Avenger and exert it",
        source: "avenger",
        params: { attacker: "avenger", target: "bob", exert: true },
      },
    ];
    expect(legalActionsOf(v).canExertOnAttack("avenger")).toBe(true);
  });
});

describe("the click flow", () => {
  const gate = legalActionsOf(snap());

  it("asks before an exertable creature's attack is sent: a seat click commits nothing", () => {
    const sel = selectAttackerFor("avenger", gate);
    expect(exertUnanswered(sel)).toBe(true);
    expect(declareAttackerParams(sel, "bob")).toBeNull();
  });

  it("sends exert only when chosen", () => {
    const sel = selectAttackerFor("avenger", gate);
    expect(declareAttackerParams({ ...sel, exert: false }, "bob")).toEqual({
      attacker: "avenger",
      target: "bob",
      auto_tap: true,
    });
    expect(declareAttackerParams({ ...sel, exert: true }, "bob")).toEqual({
      attacker: "avenger",
      target: "bob",
      auto_tap: true,
      exert: true,
    });
  });

  it("asks nothing for a creature that can't be exerted", () => {
    const sel = selectAttackerFor("bear", gate);
    expect(sel.exert).toBeUndefined();
    expect(declareAttackerParams(sel, "bob")).toEqual({
      attacker: "bear",
      target: "bob",
      auto_tap: true,
    });
  });

  it("re-points a declared attacker with no exert question and no exert sent (the server keeps a staged one)", () => {
    const declared = legalActionsOf(snap(true, { attacking_target: "bob", tapped: true }));
    const sel = selectAttackerFor("avenger", declared);
    expect(sel.exert).toBeUndefined();
    expect(declareAttackerParams(sel, "bob")).toEqual({
      attacker: "avenger",
      target: "bob",
      auto_tap: true,
    });
  });

  it("an exertable creature's two moves never make a lone-move auto-pick", () => {
    // The digest counts both moves; nothing reads that count to act, and
    // the selection waits on the question whatever it says.
    expect(snap().legal_actions!.sources!.avenger!.moves).toBe(2);
    expect(exertUnanswered(selectAttackerFor("avenger", gate))).toBe(true);
  });
});

describe("the dock's question", () => {
  const noop = () => {};

  it("offers Attack and Attack and exert, with the reminder, while unanswered", () => {
    let chose: boolean | null = null;
    const req = combatSelectionRequest("attacker", "Oketra's Avenger", noop, {
      chosen: null,
      onChoose: (e) => (chose = e),
    });
    expect(req.question).toBe("Attack with Oketra's Avenger: exert it as it attacks?");
    expect(req.detail).toBe(EXERT_NOTE);
    expect(req.row?.map((a) => a.label)).toEqual([L.attackPlain, L.attackAndExert]);
    req.row![1]!.onPress();
    expect(chose).toBe(true);
  });

  it("says which was chosen once answered", () => {
    const req = combatSelectionRequest("attacker", "Oketra's Avenger", noop, {
      chosen: true,
      onChoose: noop,
    });
    expect(req.question).toMatch(/^Attacking and exerting with Oketra's Avenger/);
    expect(req.row?.map((a) => a.pressed)).toEqual([false, true]);
  });

  it("is unchanged for a creature that can't be exerted", () => {
    const req = combatSelectionRequest("attacker", "Bear", noop);
    expect(req.row).toBeUndefined();
    expect(req.question).toMatch(/^Attacking with Bear/);
  });
});

describe("attack with all", () => {
  it("lists the exertable attackers and opens the picker for them", () => {
    const v = snap();
    const plan = planAttackAll(v, "me", legalActionsOf(v));
    expect(plan.exertable).toEqual(["avenger"]);
    expect(exertChoiceAt(plan, "bob")).toBe(true);
    expect(offersAttackPicker(v, plan, "bob")).toBe(true);
  });

  it("is one click with no exertable creature", () => {
    const v = snap(false);
    const plan = planAttackAll(v, "me", legalActionsOf(v));
    expect(exertChoiceAt(plan, "bob")).toBe(false);
    // A lone attacker with nothing to choose is the per-creature click's.
    expect(offersAttackPicker(v, plan, "bob")).toBe(false);
  });

  it("sends exert only for chosen, exertable attackers", () => {
    const v = snap();
    const plan = planAttackAll(v, "me", legalActionsOf(v));
    expect(attackAllParams(plan, "bob", { exert: ["avenger", "bear"] })?.attackers).toEqual([
      { attacker: "avenger", target: "bob", exert: true },
      { attacker: "bear", target: "bob" },
    ]);
    expect(attackAllParams(plan, "bob")?.attackers).toEqual([
      { attacker: "avenger", target: "bob" },
      { attacker: "bear", target: "bob" },
    ]);
  });
});

describe("the card menu", () => {
  const combat = (v: GameView, id: string, viewer = "me", admin = false): MenuItem[] => {
    const gate = legalActionsOf(v);
    const card = v.battlefield.cards.find((c) => c.instance_id === id)!;
    return (
      buildMenuSections(v, card, viewer, admin, gate, gate).find((s) => s.id === "combat")?.items ??
      []
    );
  };

  it("has Declare attacker and exert beside Declare attacker", () => {
    const items = combat(snap(), "avenger");
    const ids = items.map((i) => i.id);
    expect(ids.indexOf("combat-attack-exert")).toBe(ids.indexOf("combat-attack") + 1);
    const row = items.find((i) => i.id === "combat-attack-exert")!;
    expect(row.label).toBe(L.declareAttackerAndExert);
    expect(row.items?.[0]?.action?.params).toEqual({
      attacker: "avenger",
      target: "bob",
      auto_tap: true,
      exert: true,
    });
    const plain = items.find((i) => i.id === "combat-attack")!;
    expect(plain.items?.[0]?.action?.params).toEqual({
      attacker: "avenger",
      target: "bob",
      auto_tap: true,
    });
  });

  it("has no exert row for a creature that can't be exerted, or a re-point", () => {
    expect(combat(snap(), "bear").some((i) => i.id === "combat-attack-exert")).toBe(false);
    const declared = snap(true, { attacking_target: "bob", tapped: true });
    expect(combat(declared, "avenger").some((i) => i.id === "combat-attack-exert")).toBe(false);
  });

  it("does not send Attack with all blind past an exertable creature", () => {
    const all = combat(snap(), "bear").find((i) => i.id === "combat-attack-all")!;
    const bob = all.items!.find((i) => i.id === "combat-attack-all-bob")!;
    expect(bob.disabled).toBe(true);
    expect(bob.action).toBeUndefined();
  });

  it("an admin driving another seat gets the plain rows", () => {
    const items = combat(snap(), "avenger", "admin", true);
    expect(items.some((i) => i.id === "combat-attack-exert")).toBe(false);
  });
});

describe("the refusal", () => {
  it("recognises ErrCantExert and says it plainly", () => {
    expect(isCantExertError("this creature can't be exerted as it attacks")).toBe(true);
    expect(isCantExertError("card is not a creature")).toBe(false);
    expect(isCantExertError(undefined)).toBe(false);
    expect(CANT_EXERT_SENTENCE).toContain("nothing was declared");
  });
});
