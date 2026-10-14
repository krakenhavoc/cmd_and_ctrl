import { describe, expect, it } from "vitest";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { planAttackAll } from "./attackAll";
import {
  attackDefaultAction,
  attackDefaultStep,
  attackStepKey,
  preStage,
  type AttackDefaultInput,
} from "./attackDefault";

function zone(kind: string, cards: CardView[]): ZoneView {
  return { kind, count: cards.length, cards };
}
function creature(id: string, controller: string, extra: Partial<CardView> = {}): CardView {
  return {
    instance_id: id,
    name: id,
    owner: controller,
    controller,
    type_line: "Creature — Test",
    ...extra,
  };
}
function seat(id: string, extra: Partial<PlayerView> = {}): PlayerView {
  return {
    id,
    name: id,
    seat: 0,
    life: 40,
    library: zone("library", []),
    hand: zone("hand", []),
    graveyard: zone("graveyard", []),
    command: zone("command", []),
    commander_damage: {},
    life_history: [],
    ...extra,
  };
}
function view(seats: PlayerView[], cards: CardView[]): GameView {
  return {
    id: "g1",
    state: "active",
    seats,
    battlefield: zone("battlefield", cards),
    stack: zone("stack", []),
    exile: zone("exile", []),
    turn: {
      seq: 7,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
    },
    mulligans_open: false,
  };
}

const sick = creature("sick", "a", { summoning_sick: true });
const board = [creature("c1", "a"), creature("c2", "a"), sick];
const duel = planAttackAll(view([seat("a"), seat("b")], board), "a");
const pod = planAttackAll(view([seat("a"), seat("b"), seat("c")], board), "a");
const oneLeft = planAttackAll(
  view([seat("a"), seat("b"), seat("c", { eliminated: true })], board),
  "a",
);
const nothing = planAttackAll(view([seat("a"), seat("b")], [sick]), "a");

function input(over: Partial<AttackDefaultInput> = {}): AttackDefaultInput {
  return {
    mode: "none",
    canDeclare: true,
    plan: duel,
    stopped: false,
    holding: false,
    promptOpen: false,
    ...over,
  };
}

describe("attackDefaultAction", () => {
  it("does nothing on Choose each time", () => {
    expect(attackDefaultAction(input({ mode: "ask" }))).toBeNull();
  });

  describe("Attack with everything", () => {
    it("pre-stages the eligible set at the one opponent", () => {
      expect(attackDefaultAction(input({ mode: "all" }))).toEqual({
        kind: "stage",
        stage: [
          { attacker: "c1", target: "b" },
          { attacker: "c2", target: "b" },
        ],
      });
    });
    it("counts attackable opponents, so an eliminated one does not count", () => {
      expect(preStage(oneLeft).map((s) => s.target)).toEqual(["b", "b"]);
    });
    it("stages nothing with two or more opponents", () => {
      expect(attackDefaultAction(input({ mode: "all", plan: pod }))).toBeNull();
      expect(preStage(pod)).toEqual([]);
    });
    it("keeps a creature that must attack", () => {
      const p = planAttackAll(
        view([seat("a"), seat("b")], [creature("must", "a", { must_attack: true })]),
        "a",
      );
      expect(preStage(p)).toEqual([{ attacker: "must", target: "b" }]);
    });
    it("stages nothing when no creature can attack", () => {
      expect(attackDefaultAction(input({ mode: "all", plan: nothing }))).toBeNull();
    });
    it("only stages, so a stop, a hold or a prompt do not change it", () => {
      const r = attackDefaultAction(input({ mode: "all", stopped: true, holding: true }));
      expect(r?.kind).toBe("stage");
    });
  });

  describe("No attacks", () => {
    it("declares no attackers", () => {
      expect(attackDefaultAction(input())).toEqual({ kind: "declare-none" });
    });
    it("declares none with several opponents too", () => {
      expect(attackDefaultAction(input({ plan: pod }))).toEqual({ kind: "declare-none" });
    });
    it("holds off on a stop on declare attackers", () => {
      expect(attackDefaultAction(input({ stopped: true }))).toBeNull();
    });
    it("holds off while holding priority", () => {
      expect(attackDefaultAction(input({ holding: true }))).toBeNull();
    });
    it("holds off with a combat prompt open", () => {
      expect(attackDefaultAction(input({ promptOpen: true }))).toBeNull();
    });
    it("does nothing when the viewer cannot declare", () => {
      expect(attackDefaultAction(input({ canDeclare: false }))).toBeNull();
    });
    it("does nothing with nothing able to attack", () => {
      expect(attackDefaultAction(input({ plan: nothing }))).toBeNull();
    });
    it("does nothing once attackers are declared", () => {
      const p = planAttackAll(
        view([seat("a"), seat("b")], [...board, creature("d", "a", { attacking_target: "b" })]),
        "a",
      );
      expect(attackDefaultAction(input({ plan: p }))).toBeNull();
    });
  });
});

describe("attackStepKey", () => {
  it("is null outside declare attackers", () => {
    expect(attackStepKey({ seq: 7, number: 3, step: "main" })).toBeNull();
    expect(attackStepKey(null)).toBeNull();
  });
  it("names the turn sequence and number", () => {
    expect(attackStepKey({ seq: 7, number: 3, step: "declare_attackers" })).toBe("7:3");
  });
});

describe("attackDefaultStep (once per step)", () => {
  it("acts on the first frame, then never again in the same step", () => {
    const i = input({ mode: "all" });
    const first = attackDefaultStep(null, "7:3", i);
    expect(first.action?.kind).toBe("stage");
    // The player un-stages; the next frame must not re-stage.
    expect(attackDefaultStep(first.latch, "7:3", i).action).toBeNull();
  });
  it("does not resend a refused empty declaration", () => {
    const first = attackDefaultStep(null, "7:3", input());
    expect(first.action).toEqual({ kind: "declare-none" });
    expect(attackDefaultStep(first.latch, "7:3", input()).action).toBeNull();
  });
  it("takes the latch even when a stop means nothing is done", () => {
    const first = attackDefaultStep(null, "7:3", input({ stopped: true }));
    expect(first.action).toBeNull();
    expect(first.latch).toBe("7:3");
    // The stop is lifted mid-step: still no automatic declaration.
    expect(attackDefaultStep(first.latch, "7:3", input()).action).toBeNull();
  });
  it("waits for a frame the viewer can declare on", () => {
    const early = attackDefaultStep(null, "7:3", input({ canDeclare: false }));
    expect(early).toEqual({ latch: null, action: null });
    expect(attackDefaultStep(early.latch, "7:3", input()).action).toEqual({
      kind: "declare-none",
    });
  });
  it("ignores Choose each time without latching", () => {
    expect(attackDefaultStep(null, "7:3", input({ mode: "ask" }))).toEqual({
      latch: null,
      action: null,
    });
  });
  it("clears outside the step, so the next step acts", () => {
    const first = attackDefaultStep(null, "7:3", input());
    const out = attackDefaultStep(first.latch, null, input({ canDeclare: false }));
    expect(out.latch).toBeNull();
    expect(attackDefaultStep(out.latch, "7:3", input()).action).toEqual({ kind: "declare-none" });
    expect(attackDefaultStep(first.latch, "8:4", input()).action).toEqual({
      kind: "declare-none",
    });
  });
});
