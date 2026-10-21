// #4198 — paying a Phyrexian attack tax (Norn's Annex's {W/P}) with life.
// The pure half: the symbol count off the view, when the dock asks, the
// cap CR 119.4 puts on it, and the params that carry the answer. The
// markup half is attackLife.render.test.ts.

import { describe, expect, it } from "vitest";

import { attackAllParams, planAttackAll } from "./attackAll";
import { confirmAttacksParams } from "./attackHandoff";
import {
  attackLifeAsks,
  attackLifeCeiling,
  attackLifePrompt,
  attackLifeSentence,
  attackLifeSymbols,
  clampAttackLife,
  withAttackLife,
} from "./attackLife";
import { stagedParams } from "./attackStage";
import type { AttackTargetView, CardView, GameView, PlayerView, ZoneView } from "./protocol";

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string, n: number, life = 40): PlayerView {
  return {
    id,
    name,
    seat: n,
    life,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

function snap(rows: AttackTargetView[], bears = 3): GameView {
  const bf: CardView[] = Array.from({ length: bears }, (_, i) => ({
    instance_id: `bear${i + 1}`,
    name: `Bear ${i + 1}`,
    owner: "me",
    controller: "me",
    type_line: "Creature — Bear",
  }));
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1), seat("cara", "Cara", 2)],
    battlefield: zone("battlefield", undefined, bf),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
      attack_targets: rows,
    },
    mulligans_open: false,
  };
}

const NORN: AttackTargetView = { kind: "player", id: "bob", tax: "{W/P}", tax_phyrexian: 1 };
const FREE: AttackTargetView = { kind: "player", id: "cara" };

describe("attackLifeSymbols — the server's count, times the attackers", () => {
  it("multiplies one creature's symbols by the attackers aimed at each target", () => {
    const v = snap([NORN, FREE]);
    expect(attackLifeSymbols(v, ["bob", "bob", "bob"])).toEqual({ total: 3, granted: 0 });
    // A target with no symbols adds none, a split attack adds its taxed half.
    expect(attackLifeSymbols(v, ["bob", "cara", "cara"]).total).toBe(1);
    expect(attackLifeSymbols(v, ["cara"]).total).toBe(0);
  });

  it("reads a planeswalker's row too: Norn's Annex taxes those attacks", () => {
    const v = snap([{ kind: "planeswalker", id: "walker", tax: "{W/P}", tax_phyrexian: 1 }]);
    expect(attackLifeSymbols(v, ["walker", "walker"]).total).toBe(2);
  });

  it("counts a mixed tax by its Phyrexian symbols only", () => {
    const v = snap([{ kind: "player", id: "bob", tax: "{2}{W/P}", tax_phyrexian: 1 }]);
    expect(attackLifeSymbols(v, ["bob", "bob"]).total).toBe(2);
  });

  it("is zero without a view or a field (an older server)", () => {
    expect(attackLifeSymbols(null, ["bob"]).total).toBe(0);
    expect(
      attackLifeSymbols(snap([{ kind: "player", id: "bob", tax: "{2}" }]), ["bob"]).total,
    ).toBe(0);
  });

  it("keeps granted symbols inside the total", () => {
    const v = snap([
      { kind: "player", id: "bob", tax: "{B}", tax_phyrexian: 1, tax_phyrexian_granted: 1 },
    ]);
    expect(attackLifeSymbols(v, ["bob", "bob"])).toEqual({ total: 2, granted: 2 });
  });
});

describe("attackLifeAsks — a printed symbol always asks, as a cast does", () => {
  it("asks for a Phyrexian tax and not for a plain one", () => {
    const v = snap([NORN, { kind: "player", id: "cara", tax: "{2}" }]);
    expect(attackLifeAsks(v, ["bob"], 40)).toBe(true);
    expect(attackLifeAsks(v, ["cara"], 40)).toBe(false);
  });

  it("does not ask when the life total could not buy a symbol (CR 119.4)", () => {
    const v = snap([NORN]);
    expect(attackLifeAsks(v, ["bob"], 1)).toBe(false);
    expect(attackLifeAsks(v, ["bob"], 2)).toBe(true);
    expect(attackLifeAsks(v, ["bob"], undefined)).toBe(false);
  });

  it("does not ask up front when life pays only through a grant", () => {
    const v = snap([
      { kind: "player", id: "bob", tax: "{B}", tax_phyrexian: 1, tax_phyrexian_granted: 1 },
    ]);
    expect(attackLifeAsks(v, ["bob"], 40)).toBe(false);
  });
});

describe("the ceiling and the claim", () => {
  it("is the symbols, or what the life total buys, whichever is smaller", () => {
    expect(attackLifeCeiling(3, 40)).toBe(3);
    expect(attackLifeCeiling(3, 5)).toBe(2);
    expect(attackLifeCeiling(3, 1)).toBe(0);
    expect(clampAttackLife(9, 3, 40)).toBe(3);
    expect(clampAttackLife(3, 3, 5)).toBe(2);
    expect(clampAttackLife(-1, 3, 40)).toBe(0);
  });

  it("describes the split", () => {
    expect(attackLifeSentence(0, 3)).toBe("all paid with mana");
    expect(attackLifeSentence(2, 3)).toBe("2 with life, 1 with mana");
    expect(attackLifeSentence(3, 3)).toBe("all paid with life");
    expect(attackLifeSentence(1, 1)).toBe("paid with life");
  });

  it("builds the sheet's question from the view", () => {
    const v = snap([NORN]);
    expect(attackLifePrompt(v, ["bob", "bob"], "Attack with 2")).toEqual({
      symbols: 2,
      each: "{W/P}",
      attackers: 2,
      verb: "Attack with 2",
    });
  });
});

describe("the declaration carries phyrexian_life only above zero", () => {
  it("adds it to any params", () => {
    expect(withAttackLife({ a: 1 }, 2)).toEqual({ a: 1, phyrexian_life: 2 });
    expect(withAttackLife({ a: 1 }, 0)).toEqual({ a: 1 });
  });

  it("is on attack-all, the staged set and the chosen-attackers confirm", () => {
    const v = snap([NORN]);
    const plan = planAttackAll(v, "me");
    expect(attackAllParams(plan, "bob", { phyrexianLife: 2 })?.phyrexian_life).toBe(2);
    expect(attackAllParams(plan, "bob", { phyrexianLife: 0 })).not.toHaveProperty("phyrexian_life");
    expect(attackAllParams(plan, "bob")).not.toHaveProperty("phyrexian_life");

    const stage = [{ attacker: "bear1", target: "bob" }];
    expect(stagedParams(stage, 1)).toEqual({
      attackers: stage,
      auto_tap: true,
      phyrexian_life: 1,
    });
    expect(stagedParams(stage)).toEqual({ attackers: stage, auto_tap: true });
    expect(stagedParams([], 2)).toBeNull();

    const attacks = [{ attacker: "bear1", target: "bob" }];
    expect(confirmAttacksParams(attacks, 1)).toEqual({
      attackers: attacks,
      auto_tap: true,
      phyrexian_life: 1,
    });
    expect(confirmAttacksParams(attacks)).toEqual({ attackers: attacks, auto_tap: true });
  });
});
