import { describe, expect, it } from "vitest";
import type { CardView, GameView, LegalActionsView, PlayerView, ZoneView } from "./protocol";
import { NO_LEGAL_ACTIONS, legalActionsOf } from "./legalActions";
import {
  attackAllLabel,
  attackAllParams,
  attackAllTaxLabel,
  attackBlocker,
  attackRefusal,
  attackTaxLabelForCount,
  attackTaxOn,
  blockedSummary,
  eligibleAt,
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
      seq: 1,
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

// ADR 0105 sub-PR 5: eligibility is the server's. The row fields only
// name the reason, and decide only on a frame with no list at all.
describe("planAttackAll reads eligibility from the legal-action digest", () => {
  const alice = seat("a", "Alice");
  const bob = seat("b", "Bob", { seat: 1 });

  const withDigest = (cards: CardView[], sources: LegalActionsView["sources"]): GameView => ({
    ...view([alice, bob], cards),
    legal_actions: { pass: true, sources },
  });

  const board = () => [
    // The server lists it: eligible.
    creature("ok", "a"),
    // Nothing on the row says why, but the server leaves it out: an
    // attack tax it can't pay, a count limit, a goad. The old
    // re-derivation would have swept it in.
    creature("taxed", "a"),
    // Sick by the row fields, and absent from the list: the reason is
    // still the row's sentence.
    creature("sick", "a", { summoning_sick: true }),
    // A "can't attack" effect the row fields carry.
    creature("pacified", "a", { restrictions: ["cant_attack"] }),
    // Defender on the row, but an effect lets it attack (the server
    // lists it): the server wins.
    creature("wall", "a", { abilities: ["defender"] }),
    creature("already", "a", { attacking_target: "b", tapped: true }),
  ];
  const sources: LegalActionsView["sources"] = {
    ok: { kinds: ["attack"], moves: 1, attack_targets: ["b"] },
    wall: { kinds: ["attack"], moves: 1, attack_targets: ["b"] },
  };

  it("eligible is exactly the creatures the server lists", () => {
    const v = withDigest(board(), sources);
    const plan = planAttackAll(v, "a", legalActionsOf(v));
    expect(plan.eligible.map((c) => c.instance_id)).toEqual(["ok", "wall"]);
    expect(plan.declared.map((c) => c.instance_id)).toEqual(["already"]);
    expect(Object.fromEntries(plan.blocked.map((b) => [b.card.instance_id, b.reason]))).toEqual({
      taxed: "unavailable",
      sick: "summoning-sick",
      pacified: "restricted",
    });
    expect(blockedSummary(plan.blocked)).toBe(
      "1 can't attack, 1 summoning sick, 1 can't attack right now",
    );
  });

  it("a creature the old re-derivation allowed is withheld when the server says no", () => {
    const v = withDigest([creature("unlisted", "a")], {});
    // The row fields see an untapped, unrestricted creature,
    expect(attackBlocker(v.battlefield.cards[0])).toBeNull();
    // and the server's answer wins.
    expect(attackRefusal(v.battlefield.cards[0], legalActionsOf(v))).toBe("unavailable");
    expect(planAttackAll(v, "a", legalActionsOf(v)).eligible).toEqual([]);
  });

  it("reads the capped legal_moves list the same way (an older server)", () => {
    const v: GameView = {
      ...view([alice, bob], [creature("ok", "a"), creature("taxed", "a")]),
      legal_moves: [
        {
          type: "declare_attacker",
          player: "a",
          kind: "attack",
          label: "Attack Bob with ok",
          source: "ok",
          params: { attacker: "ok", target: "b" },
        },
      ],
    };
    const plan = planAttackAll(v, "a", legalActionsOf(v));
    expect(plan.eligible.map((c) => c.instance_id)).toEqual(["ok"]);
    expect(plan.blocked.map((b) => b.reason)).toEqual(["unavailable"]);
  });

  it("no list is no information: the row fields decide, and nothing is newly withheld", () => {
    const v = view([alice, bob], board());
    const plans = [
      planAttackAll(v, "a"),
      planAttackAll(v, "a", legalActionsOf(v)),
      planAttackAll(v, "a", NO_LEGAL_ACTIONS),
    ];
    for (const plan of plans) {
      expect(plan.eligible.map((c) => c.instance_id)).toEqual(["ok", "taxed"]);
      expect(plan.blocked.map((b) => b.reason).sort()).toEqual([
        "defender",
        "restricted",
        "summoning-sick",
      ]);
    }
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
      // ADR 0080 (#1063): the payload always carries permission to
      // tap for the CR 508.1a attack tax. Inert without one.
      auto_tap: true,
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

  // #1162: the attack-tax picker's subset and lock-a-land options.
  // Both are additive over the plain "attack with everything" shape
  // above — omitting either keeps that behaviour byte for byte, which
  // the tests in this describe block already pin.
  describe("opts.only — the picker's chosen subset", () => {
    it("declares exactly the named subset rather than every eligible creature", () => {
      const v = view(
        [alice, bob, carol],
        [creature("x", "a"), creature("y", "a"), creature("z", "a")],
      );
      const plan = planAttackAll(v, "a");
      expect(attackAllParams(plan, "c", { only: ["y"] })).toEqual({
        attackers: [{ attacker: "y", target: "c" }],
        auto_tap: true,
      });
    });

    it("keeps plan.eligible's own order rather than the subset list's", () => {
      const v = view(
        [alice, bob, carol],
        [creature("x", "a"), creature("y", "a"), creature("z", "a")],
      );
      const plan = planAttackAll(v, "a");
      expect(attackAllParams(plan, "c", { only: ["z", "x"] })?.attackers).toEqual([
        { attacker: "x", target: "c" },
        { attacker: "z", target: "c" },
      ]);
    });

    it("drops an ID the plan no longer considers eligible rather than sending it", () => {
      const v = view(
        [alice, bob, carol],
        [creature("x", "a"), creature("y", "a", { tapped: true })],
      );
      const plan = planAttackAll(v, "a");
      // "y" is tapped (blocked, not eligible) and "ghost" names nothing
      // on the board at all — a stale picker selection after a
      // snapshot changed the board underneath it.
      expect(attackAllParams(plan, "c", { only: ["x", "y", "ghost"] })).toEqual({
        attackers: [{ attacker: "x", target: "c" }],
        auto_tap: true,
      });
    });

    it("refuses an empty subset the same way it refuses an empty plan", () => {
      const v = view([alice, bob, carol], [creature("x", "a")]);
      const plan = planAttackAll(v, "a");
      expect(attackAllParams(plan, "c", { only: [] })).toBeNull();
      expect(attackAllParams(plan, "c", { only: ["nobody-here"] })).toBeNull();
    });

    it("still refuses a seat that isn't an attackable opponent, subset or not", () => {
      const v = view([alice, bob], [creature("x", "a")]);
      const plan = planAttackAll(v, "a");
      expect(attackAllParams(plan, "a", { only: ["x"] })).toBeNull();
    });
  });

  describe("opts.lockedSources — the lock-a-land toggle", () => {
    it("carries locked_sources through to the payload", () => {
      const v = view([alice, bob], [creature("x", "a")]);
      const plan = planAttackAll(v, "a");
      expect(attackAllParams(plan, "b", { lockedSources: ["forest-1"] })).toEqual({
        attackers: [{ attacker: "x", target: "b" }],
        auto_tap: true,
        locked_sources: ["forest-1"],
      });
    });

    it("omits the field entirely for an empty or absent lock list", () => {
      const v = view([alice, bob], [creature("x", "a")]);
      const plan = planAttackAll(v, "a");
      expect(attackAllParams(plan, "b", { lockedSources: [] })).toEqual({
        attackers: [{ attacker: "x", target: "b" }],
        auto_tap: true,
      });
      expect(attackAllParams(plan, "b")).toEqual({
        attackers: [{ attacker: "x", target: "b" }],
        auto_tap: true,
      });
    });

    it("copies the list rather than aliasing the caller's array", () => {
      const v = view([alice, bob], [creature("x", "a")]);
      const plan = planAttackAll(v, "a");
      const locked = ["forest-1"];
      const params = attackAllParams(plan, "b", { lockedSources: locked });
      locked.push("island-1");
      expect(params?.locked_sources).toEqual(["forest-1"]);
    });

    it("composes with a chosen subset — the picker's actual shape", () => {
      const v = view(
        [alice, bob, carol],
        [creature("x", "a"), creature("y", "a"), creature("z", "a")],
      );
      const plan = planAttackAll(v, "a");
      expect(attackAllParams(plan, "c", { only: ["x", "z"], lockedSources: ["forest-1"] })).toEqual(
        {
          attackers: [
            { attacker: "x", target: "c" },
            { attacker: "z", target: "c" },
          ],
          auto_tap: true,
          locked_sources: ["forest-1"],
        },
      );
    });
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

// --- the CR 508.1a attack tax (ADR 0080, #1063) --------------------
//
// The price is the SERVER's, read off turn.attack_targets[].tax.
// Nothing here derives it, and the total for a wide swing is the
// per-creature string repeated rather than arithmetic on it — which
// is exactly what the engine charges and concatenates.

// taxedView is `view` with attack_targets stamped, the way the server
// stamps them during declare_attackers.
function taxedView(
  seats: PlayerView[],
  battlefield: CardView[],
  taxes: Record<string, string>,
): GameView {
  const v = view(seats, battlefield);
  v.turn.attack_targets = seats.map((s) => ({
    kind: "player" as const,
    id: s.id,
    ...(taxes[s.id] ? { tax: taxes[s.id] } : {}),
  }));
  return v;
}

describe("attackTaxOn", () => {
  const alice = seat("a", "Alice");
  const bob = seat("b", "Bob", { seat: 1 });

  it("reads the server's price for the named seat", () => {
    const v = taxedView([alice, bob], [], { b: "{2}" });
    expect(attackTaxOn(v, "b")).toBe("{2}");
  });

  it("is empty for a seat that charges nothing", () => {
    const v = taxedView([alice, bob], [], { b: "{2}" });
    expect(attackTaxOn(v, "a")).toBe("");
  });

  it("is empty when the server sent no attack targets at all", () => {
    // An older server, or a step that is not declare_attackers.
    expect(attackTaxOn(view([alice, bob]), "b")).toBe("");
    expect(attackTaxOn(null, "b")).toBe("");
  });
});

// #1162: the running-total primitive the subset picker reuses as the
// player checks attackers on and off. attackAllTaxLabel below is now
// this function called with plan.eligible.length, so the two describe
// blocks pin the same arithmetic from two different call shapes.
describe("attackTaxLabelForCount", () => {
  it("is empty for a free attack or a non-positive count", () => {
    expect(attackTaxLabelForCount("", 3)).toBe("");
    expect(attackTaxLabelForCount("{2}", 0)).toBe("");
    expect(attackTaxLabelForCount("{2}", -1)).toBe("");
  });

  it("names the flat price for one attacker", () => {
    expect(attackTaxLabelForCount("{2}", 1)).toBe("costs {2}");
  });

  it("repeats the price once per attacker for a wider count", () => {
    expect(attackTaxLabelForCount("{2}", 3)).toBe("costs {2} each, {2}{2}{2} for all 3");
  });
});

describe("attackAllTaxLabel", () => {
  const alice = seat("a", "Alice");
  const bob = seat("b", "Bob", { seat: 1 });

  it("is empty when attacking the seat is free", () => {
    const v = taxedView([alice, bob], [creature("x", "a")], {});
    expect(attackAllTaxLabel(v, planAttackAll(v, "a"), "b")).toBe("");
  });

  it("names the per-creature price for a single attacker", () => {
    const v = taxedView([alice, bob], [creature("x", "a")], { b: "{2}" });
    expect(attackAllTaxLabel(v, planAttackAll(v, "a"), "b")).toBe("costs {2}");
  });

  it("repeats the price once per attacker rather than adding it up", () => {
    // Three attackers into Propaganda is "{2}{2}{2}" on the wire, six
    // generic to the parser. The client never does that arithmetic.
    const v = taxedView(
      [alice, bob],
      [creature("x", "a"), creature("y", "a"), creature("z", "a")],
      {
        b: "{2}",
      },
    );
    expect(attackAllTaxLabel(v, planAttackAll(v, "a"), "b")).toBe(
      "costs {2} each, {2}{2}{2} for all 3",
    );
  });

  it("carries a stacked price through unchanged", () => {
    const v = taxedView([alice, bob], [creature("x", "a"), creature("y", "a")], {
      b: "{2}{2}",
    });
    expect(attackAllTaxLabel(v, planAttackAll(v, "a"), "b")).toBe(
      "costs {2}{2} each, {2}{2}{2}{2} for all 2",
    );
  });
});

// ADR 0106 §2 (#1794): a creature that "can't attack its owner" is
// eligible, and the server lists every opponent for it but its owner.
// "Attack with all" is per seat, so it counts and sends the creature at
// every seat but that one.
describe("attack with all, per seat", () => {
  const alice = seat("a", "Alice");
  const bob = seat("b", "Bob", { seat: 1 });
  const carol = seat("c", "Carol", { seat: 2 });
  const v: GameView = {
    ...view([alice, bob, carol], [creature("bear", "a"), creature("xantcha", "a", { owner: "b" })]),
    legal_actions: {
      pass: true,
      sources: {
        bear: { kinds: ["attack"], moves: 2, attack_targets: ["b", "c"] },
        xantcha: { kinds: ["attack"], moves: 1, attack_targets: ["c"] },
      },
    },
  };
  const plan = planAttackAll(v, "a", legalActionsOf(v));

  it("leaves the creature out at its owner's seat only", () => {
    expect(plan.eligible.map((c) => c.instance_id)).toEqual(["bear", "xantcha"]);
    expect(eligibleAt(plan, "b").map((c) => c.instance_id)).toEqual(["bear"]);
    expect(eligibleAt(plan, "c").map((c) => c.instance_id)).toEqual(["bear", "xantcha"]);
  });

  it("sends and counts what that seat may be attacked by", () => {
    expect(attackAllParams(plan, "b")?.attackers).toEqual([{ attacker: "bear", target: "b" }]);
    expect(attackAllParams(plan, "c")?.attackers.map((a) => a.attacker)).toEqual([
      "bear",
      "xantcha",
    ]);
    expect(attackAllLabel(plan, bob)).toBe("Attack Bob with all 1 creature");
    expect(attackAllLabel(plan, carol)).toBe("Attack Carol with all 2 creatures");
  });

  it("a frame with no list counts every eligible creature at every seat, as before", () => {
    const bare = planAttackAll(view([alice, bob, carol], v.battlefield.cards), "a");
    expect(eligibleAt(bare, "b").map((c) => c.instance_id)).toEqual(["bear", "xantcha"]);
  });
});
