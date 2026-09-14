import { describe, expect, it } from "vitest";
import type { CardView, GameView, PlayerView, ZoneView } from "./protocol";
import {
  attackAllLabel,
  attackAllParams,
  attackBlocker,
  blockedSummary,
  planAttackAll,
} from "./attackAll";

// Builders mirror contextMenu.test.ts: only the fields the logic reads
// are populated, so unrelated protocol churn doesn't ripple in.

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function creature(
  instance_id: string,
  controller: string,
  extra: Partial<CardView> = {},
): CardView {
  return {
    instance_id,
    name: instance_id,
    owner: controller,
    controller,
    type_line: "Creature — Test",
    ...extra,
  };
}

function seat(id: string, name: string, extra: Partial<PlayerView> = {}): PlayerView {
  return {
    id,
    name,
    seat: 0,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
    ...extra,
  };
}

function view(seats: PlayerView[], battlefield: CardView[] = []): GameView {
  return {
    id: "g1",
    state: "active",
    seats,
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      number: 1,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
    },
    mulligans_open: false,
  };
}

describe("attackBlocker", () => {
  it("clears a plain untapped creature", () => {
    expect(attackBlocker(creature("c1", "a"))).toBeNull();
  });

  it("blocks tapped, summoning-sick, and defender creatures", () => {
    expect(attackBlocker(creature("c1", "a", { tapped: true }))).toBe("tapped");
    expect(attackBlocker(creature("c2", "a", { summoning_sick: true }))).toBe("summoning-sick");
    expect(attackBlocker(creature("c3", "a", { abilities: ["defender"] }))).toBe("defender");
  });

  it("reports the most permanent reason first", () => {
    const wall = creature("c1", "a", {
      tapped: true,
      summoning_sick: true,
      abilities: ["defender", "reach"],
    });
    expect(attackBlocker(wall)).toBe("defender");
  });

  // Vigilance is not an eligibility question at all — it only changes
  // whether the declaration taps. It must never keep a creature out of
  // the plan, which is the trap while #317 repairs printed keywords.
  it("ignores vigilance", () => {
    expect(attackBlocker(creature("c1", "a", { abilities: ["vigilance"] }))).toBeNull();
  });

  // S24: a pacified creature is one the server's bulk verb silently
  // skips, so counting it would make the button lie about how many
  // creatures are swinging.
  it("blocks a creature under a can't-attack restriction", () => {
    expect(attackBlocker(creature("c1", "a", { restrictions: ["cant_attack"] }))).toBe(
      "restricted",
    );
  });

  // The restriction outranks every other reason — it is the one the
  // player has to answer with a removal spell rather than with time.
  it("reports the restriction ahead of defender and tapped", () => {
    const pacified = creature("c1", "a", {
      tapped: true,
      summoning_sick: true,
      abilities: ["defender"],
      restrictions: ["cant_attack"],
    });
    expect(attackBlocker(pacified)).toBe("restricted");
  });

  // "Can't block" and "can't be blocked" say nothing about whether
  // this creature may ATTACK. A Whispersilk Cloak carrier is the whole
  // point of the attack-with-all button, not an exclusion from it.
  it("ignores the block-side restrictions", () => {
    expect(attackBlocker(creature("c1", "a", { restrictions: ["cant_block"] }))).toBeNull();
    expect(attackBlocker(creature("c2", "a", { restrictions: ["cant_be_blocked"] }))).toBeNull();
  });
});

describe("planAttackAll", () => {
  const alice = seat("a", "Alice");
  const bob = seat("b", "Bob", { seat: 1 });
  const carol = seat("c", "Carol", { seat: 2, eliminated: true });

  it("buckets the viewer's creatures by eligibility", () => {
    const v = view(
      [alice, bob],
      [
        creature("ok1", "a"),
        creature("ok2", "a", { abilities: ["vigilance"] }),
        creature("tapped", "a", { tapped: true }),
        creature("sick", "a", { summoning_sick: true }),
        creature("wall", "a", { abilities: ["defender"] }),
        creature("already", "a", { attacking_target: "b", tapped: true }),
      ],
    );
    const plan = planAttackAll(v, "a");
    expect(plan.eligible.map((c) => c.instance_id)).toEqual(["ok1", "ok2"]);
    expect(plan.declared.map((c) => c.instance_id)).toEqual(["already"]);
    expect(plan.blocked.map((b) => b.reason).sort()).toEqual([
      "defender",
      "summoning-sick",
      "tapped",
    ]);
  });

  it("ignores non-creatures and other players' permanents", () => {
    const v = view(
      [alice, bob],
      [
        creature("mine", "a"),
        creature("theirs", "b"),
        { instance_id: "land", name: "Forest", owner: "a", controller: "a", type_line: "Land" },
        { instance_id: "rock", name: "Sol Ring", owner: "a", controller: "a" },
      ],
    );
    const plan = planAttackAll(v, "a");
    expect(plan.eligible.map((c) => c.instance_id)).toEqual(["mine"]);
    expect(plan.blocked).toHaveLength(0);
  });

  it("lists only seated, living opponents as defenders", () => {
    const plan = planAttackAll(view([alice, bob, carol]), "a");
    expect(plan.defenders.map((s) => s.id)).toEqual(["b"]);
  });

  it("returns an empty plan without a view or a viewer", () => {
    expect(planAttackAll(null, "a").eligible).toHaveLength(0);
    expect(planAttackAll(view([alice, bob]), null).defenders).toHaveLength(0);
  });
});

describe("attackAllParams", () => {
  const alice = seat("a", "Alice");
  const bob = seat("b", "Bob", { seat: 1 });
  const carol = seat("c", "Carol", { seat: 2 });

  it("points every eligible creature at one named seat", () => {
    const v = view(
      [alice, bob, carol],
      [creature("x", "a"), creature("y", "a"), creature("z", "a", { tapped: true })],
    );
    const plan = planAttackAll(v, "a");
    expect(attackAllParams(plan, "c")).toEqual({
      attackers: [
        { attacker: "x", target: "c" },
        { attacker: "y", target: "c" },
      ],
    });
  });

  it("refuses to build an empty batch", () => {
    const v = view([alice, bob], [creature("x", "a", { tapped: true })]);
    expect(attackAllParams(planAttackAll(v, "a"), "b")).toBeNull();
  });

  it("refuses a seat that isn't an attackable opponent", () => {
    const v = view([alice, bob], [creature("x", "a")]);
    const plan = planAttackAll(v, "a");
    expect(attackAllParams(plan, "a")).toBeNull();
    expect(attackAllParams(plan, "nobody")).toBeNull();
  });
});

describe("blockedSummary", () => {
  it("is empty when nothing is blocked", () => {
    expect(blockedSummary([])).toBe("");
  });

  it("counts by reason, most common first", () => {
    const v = view(
      [seat("a", "Alice"), seat("b", "Bob", { seat: 1 })],
      [
        creature("t1", "a", { tapped: true }),
        creature("t2", "a", { tapped: true }),
        creature("w", "a", { abilities: ["defender"] }),
      ],
    );
    expect(blockedSummary(planAttackAll(v, "a").blocked)).toBe("2 tapped, 1 defender");
  });
});

describe("attackAllLabel", () => {
  const alice = seat("a", "Alice");
  const bob = seat("b", "Bob", { seat: 1, display_name: "bobby" });

  it("always names the defending seat and the honest count", () => {
    const v = view([alice, bob], [creature("x", "a"), creature("y", "a")]);
    expect(attackAllLabel(planAttackAll(v, "a"), bob)).toBe("Attack bobby with all 2 creatures");
  });

  it("singularises one creature", () => {
    const v = view([alice, bob], [creature("x", "a")]);
    expect(attackAllLabel(planAttackAll(v, "a"), bob)).toBe("Attack bobby with all 1 creature");
  });
});
