// legalActions.test.ts — ADR 0105 §9, the client half of the
// legal-action contract (#1789).
//
// The fixture is server/internal/legal/testdata/legal_actions_agreement.json.
// Each scenario declares BY HAND which cards are ready and what the
// digest must say about them; server/internal/legal/
// legal_actions_agreement_test.go asserts the Go digest matches that
// declaration, and this file asserts legalActions.ts's lookups match
// the same declaration against the real filtered frame. Neither side
// is the other's oracle.
//
// Regenerate the fixture after any enumerator or digest change:
//
//   go test ./internal/legal/ -run TestLegalActionsAgreement -update

import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

import { describe, it, expect } from "vitest";

import {
  NO_COMBAT_RINGS,
  NO_LEGAL_ACTIONS,
  NO_PIPS,
  attackTargetListed,
  attackTargetOpen,
  combatRings,
  digestRefusesRow,
  highlightsLive,
  legalActionsOf,
  notableManaRefs,
  pipCount,
  readyFirst,
  readyPips,
  visibleHighlights,
  type LegalActions,
  type ReadyZone,
} from "./legalActions";
import { autopassDecision, type AutopassGates } from "./autopassDecision";
import { hasPassMove } from "./timing";
import { planAttackAll } from "./attackAll";
import type { CardView, GameView, LegalMoveView } from "./protocol";

interface Expectation {
  instance_id: string;
  name: string;
  ready: boolean;
  kinds?: string[];
  abilities?: string[];
  mana_abilities?: string[];
  zones?: string[];
  faces?: number[];
  attack_targets?: string[];
  blocks?: string[];
  why: string;
}

interface Scenario {
  name: string;
  note: string;
  viewer: string;
  view: GameView;
  pass: boolean;
  expect: Expectation[];
}

const fixtureURL = new URL(
  "../../../server/internal/legal/testdata/legal_actions_agreement.json",
  import.meta.url,
);
const scenarios: Scenario[] = JSON.parse(readFileSync(fileURLToPath(fixtureURL), "utf8"));

const ZONES: ReadyZone[] = ["hand", "graveyard", "exile", "library", "command", "battlefield"];

const sorted = <T>(xs: readonly T[] | undefined): T[] => [...(xs ?? [])].sort();

// checkAgainst asserts one lookup against one scenario's declarations.
function checkAgainst(la: LegalActions, sc: Scenario, label: string): void {
  for (const e of sc.expect) {
    const id = e.instance_id;
    const msg = `${label} ${sc.name} / ${e.name}: ${e.why}`;
    expect(la.isReady(id), msg).toBe(e.ready);
    expect(sorted(la.kinds(id)), msg).toEqual(sorted(e.kinds));
    expect(sorted(la.readyAbilityRefs(id)), msg).toEqual(sorted(e.abilities));
    expect(sorted(la.readyManaRefs(id)), msg).toEqual(sorted(e.mana_abilities));
    expect(sorted(la.attackTargets(id)), msg).toEqual(sorted(e.attack_targets));
    expect(sorted(la.blockableAttackers(id)), msg).toEqual(sorted(e.blocks));
    expect(la.canAttack(id), msg).toBe((e.kinds ?? []).includes("attack"));
    const castish = (e.kinds ?? []).some((k) => k === "cast" || k === "land");
    for (const z of ZONES) {
      expect(la.castableFrom(id, z), `${msg} (from ${z})`).toBe(
        castish && (e.zones ?? []).includes(z),
      );
    }
  }
}

describe("legalActions.ts agrees with the server digest", () => {
  it("the fixture is present and non-trivial", () => {
    expect(scenarios.length).toBeGreaterThan(0);
    for (const sc of scenarios) {
      expect(sc.expect.length, `${sc.name} declares nothing`).toBeGreaterThan(0);
    }
  });

  for (const sc of scenarios) {
    describe(sc.name, () => {
      it("every declared card reads as declared", () => {
        checkAgainst(legalActionsOf(sc.view), sc, "digest");
      });

      it("pass is the declared pass, and hasPassMove agrees", () => {
        const la = legalActionsOf(sc.view);
        if (!sc.view.legal_actions && !sc.view.legal_moves) {
          // No decision owed: no information, which is not "no".
          expect(la.known).toBe(false);
          expect(la.pass).toBeUndefined();
          expect(hasPassMove(sc.view)).toBeUndefined();
          expect(sc.pass).toBe(false);
          return;
        }
        expect(la.known).toBe(true);
        expect(la.pass).toBe(sc.pass);
        expect(hasPassMove(sc.view)).toBe(sc.pass);
      });

      // The other direction, as an invariant: every source the digest
      // names is ready, and nothing it does not name is.
      it("every digested source is ready, and only those", () => {
        const la = legalActionsOf(sc.view);
        const named = new Set(Object.keys(sc.view.legal_actions?.sources ?? {}));
        const everyCard = [
          ...sc.view.seats.flatMap((s) => [
            ...(s.hand?.cards ?? []),
            ...(s.graveyard?.cards ?? []),
            ...(s.library?.cards ?? []),
            ...(s.command?.cards ?? []),
          ]),
          ...(sc.view.battlefield?.cards ?? []),
          ...(sc.view.exile?.cards ?? []),
        ];
        for (const c of everyCard) {
          expect(la.isReady(c.instance_id), c.name || c.instance_id).toBe(named.has(c.instance_id));
        }
      });

      // ADR 0105 sub-PR 2: an older server sends the list and no
      // digest. Every fixture board is under the wire cap, so folding
      // the list must give the same answers the digest does.
      it("falls back to legal_moves when the digest is absent", () => {
        const older = { ...sc.view, legal_actions: undefined } as GameView;
        const la = legalActionsOf(older);
        checkAgainst(la, sc, "fallback");
        if (sc.view.legal_moves) {
          expect(la.pass).toBe(sc.pass);
          expect(hasPassMove(older)).toBe(sc.pass);
        }
      });
    });
  }

  it("counts ready cards per zone on the main-phase board", () => {
    const sc = scenarios.find((s) => s.name === "main_phase_ready");
    expect(sc).toBeDefined();
    const la = legalActionsOf(sc!.view);
    const seat = sc!.view.seats.find((s) => s.id === sc!.viewer)!;
    const sources = sc!.view.legal_actions?.sources ?? {};
    const want = (cards: CardView[] | undefined) =>
      (cards ?? []).filter((c) => c.instance_id in sources).length;
    expect(la.readyCount("hand", seat.id)).toBe(want(seat.hand.cards));
    expect(la.readyCount("hand", seat.id)).toBeGreaterThanOrEqual(2); // Forest, Bolt
    expect(la.readyCount("command", seat.id)).toBe(want(seat.command.cards));
    expect(la.readyCount("command", seat.id)).toBe(1);
    expect(la.readyCount("graveyard", seat.id)).toBe(want(seat.graveyard.cards));
    // A per-seat zone needs a seat.
    expect(la.readyCount("hand")).toBe(0);
    // The opponent's hand holds nothing the viewer can act on.
    const other = sc!.view.seats.find((s) => s.id !== sc!.viewer)!;
    expect(la.readyCount("hand", other.id)).toBe(0);
  });
});

// ---- the absent-list rule ---------------------------------------------

const ME = "me";

function frame(over: Partial<GameView> = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      {
        id: ME,
        name: "Me",
        hand: { kind: "hand", count: 1, cards: [{ instance_id: "bolt", name: "Bolt" }] },
        graveyard: { kind: "graveyard", count: 0, cards: [] },
        library: { kind: "library", count: 0, cards: [] },
        command: { kind: "command", count: 0, cards: [] },
      },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { seq: 1, number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    ...over,
  } as unknown as GameView;
}

const move = (m: Partial<LegalMoveView> & Pick<LegalMoveView, "kind">): LegalMoveView => ({
  type: "cast_spell",
  player: ME,
  label: m.kind,
  ...m,
});

describe("absent digest and absent list: no information", () => {
  it("highlights nothing, counts nothing, and does not say no to passing", () => {
    for (const la of [legalActionsOf(frame()), legalActionsOf(null), NO_LEGAL_ACTIONS]) {
      expect(la.known).toBe(false);
      expect(la.pass).toBeUndefined();
      expect(la.isReady("bolt")).toBe(false);
      expect(la.castableFrom("bolt", "hand")).toBe(false);
      expect(la.readyAbilityRefs("bolt")).toEqual([]);
      expect(la.readyCount("hand", ME)).toBe(0);
    }
  });

  it("an empty list is information: nothing is ready", () => {
    const la = legalActionsOf(frame({ legal_moves: [] }));
    expect(la.known).toBe(true);
    expect(la.pass).toBe(false);
    expect(la.isReady("bolt")).toBe(false);
  });

  it("the digest wins over the list when both are present", () => {
    const la = legalActionsOf(
      frame({
        legal_moves: [move({ kind: "cast", source: "bolt", params: { from_zone: "hand" } })],
        legal_actions: { sources: {} },
      }),
    );
    expect(la.isReady("bolt")).toBe(false);
    expect(la.pass).toBe(false);
  });
});

// ---- the fallback fold -------------------------------------------------

describe("the legal_moves fallback", () => {
  it("reads the same params the server digest reads", () => {
    const la = legalActionsOf(
      frame({
        legal_moves: [
          move({
            kind: "pass",
            type: "pass_priority",
            source: "00000000-0000-0000-0000-000000000000",
          }),
          move({ kind: "cast", source: "bolt", params: { from_zone: "hand" } }),
          move({ kind: "cast", source: "adv", params: { from_zone: "hand", face: 1 } }),
          move({ kind: "activate", source: "bomb", params: { ref: "own:0" } }),
          move({ kind: "activate", source: "bomb", params: { ref: "own:0" } }),
          move({ kind: "mana", source: "vivi", params: { ref: "own:0" } }),
          move({ kind: "special_action", source: "fore", params: { kind: "foretell" } }),
          move({ kind: "choice", source: "bolt" }),
        ],
      }),
    );
    expect(la.pass).toBe(true);
    expect(la.castableFrom("bolt", "hand")).toBe(true);
    expect(la.castableFrom("bolt", "graveyard")).toBe(false);
    expect(la.readyAbilityRefs("bomb")).toEqual(["own:0"]);
    expect(la.readyManaRefs("vivi")).toEqual(["own:0"]);
    expect(la.readySpecialActions("fore")).toEqual(["foretell"]);
    expect(la.kinds("bolt")).toEqual(["cast"]); // a choice move is not digested
    expect(la.isReady("00000000-0000-0000-0000-000000000000")).toBe(false);
  });

  it("gives every creature in a grouped block declaration its pair", () => {
    const la = legalActionsOf(
      frame({
        legal_moves: [
          move({
            kind: "block",
            type: "declare_blockers",
            source: "a",
            params: {
              blocks: [
                { blocker: "a", attacker: "menace" },
                { blocker: "b", attacker: "menace" },
              ],
            },
          }),
        ],
      }),
    );
    expect(la.blockableAttackers("a")).toEqual(["menace"]);
    expect(la.blockableAttackers("b")).toEqual(["menace"]);
    expect(la.kinds("b")).toEqual(["block"]);
  });

  it("a cast whose zone the wire does not name counts for any zone", () => {
    const la = legalActionsOf(frame({ legal_moves: [move({ kind: "cast", source: "bolt" })] }));
    expect(la.castableFrom("bolt", "hand")).toBe(true);
    expect(la.castableFrom("bolt", "exile")).toBe(true);
  });
});

// ---- the toggle and autopass suppression (ADR 0105 §3, §6) -----------

function gates(over: Partial<AutopassGates> = {}): AutopassGates {
  return {
    viewerHasPriority: true,
    tableBusy: false,
    hasPendingChoice: false,
    owesBlockDecision: false,
    owesAttackRequirement: false,
    loopSuspended: false,
    step: "upkeep",
    autopassToggle: false,
    viewerIsActive: false,
    autopassPersistThroughTurns: false,
    manualStop: false,
    autoPassPriority: true,
    stackEmpty: true,
    holdPriority: false,
    autoPassOwnStack: true,
    ownsEveryStackItem: false,
    stepStop: false,
    smartAutoPass: true,
    alwaysStopOpponentStack: false,
    hasResponse: false,
    hasPlay: false,
    combatWindow: false,
    oppEndWindow: false,
    bluffCounter: false,
    bluffInstant: false,
    bluffManual: false,
    ...over,
  };
}

describe("when highlights show", () => {
  const live = legalActionsOf(
    frame({ legal_moves: [move({ kind: "cast", source: "bolt", params: { from_zone: "hand" } })] }),
  );

  it("the setting off shows nothing, whatever autopass says", () => {
    for (const v of ["hold", "pass", "clear-toggle", { kind: "bluff", manual: false }] as const) {
      expect(highlightsLive(false, v)).toBe(false);
    }
    expect(visibleHighlights(live, false).isReady("bolt")).toBe(false);
  });

  it("a frame smart autopass is about to pass shows nothing", () => {
    const verdict = autopassDecision(gates());
    expect(verdict).toBe("pass");
    expect(highlightsLive(true, verdict)).toBe(false);
    expect(visibleHighlights(live, highlightsLive(true, verdict)).isReady("bolt")).toBe(false);
  });

  it("a window that stays with the player shows what is ready", () => {
    // A ticked step with something to play holds.
    const hold = autopassDecision(gates({ stepStop: true, hasPlay: true }));
    expect(hold).toBe("hold");
    expect(highlightsLive(true, hold)).toBe(true);
    expect(visibleHighlights(live, highlightsLive(true, hold)).isReady("bolt")).toBe(true);
    // A bluff and the safety belt both leave the cursor with the player.
    expect(highlightsLive(true, { kind: "bluff", manual: false })).toBe(true);
    expect(highlightsLive(true, "clear-toggle")).toBe(true);
    // No verdict yet (no frame): nothing to suppress.
    expect(highlightsLive(true, null)).toBe(true);
  });

  it("suppressing the highlights never touches the pass answer the gates read", () => {
    const f = frame({ legal_moves: [move({ kind: "pass", type: "pass_priority" })] });
    expect(visibleHighlights(legalActionsOf(f), false).pass).toBeUndefined();
    expect(hasPassMove(f)).toBe(true);
  });
});

// ---- the mana-noise rule (ADR 0105 §4) --------------------------------

describe("notableManaRefs", () => {
  const mountain: CardView = {
    instance_id: "m",
    name: "Mountain",
    owner: ME,
    controller: ME,
    type_line: "Basic Land — Mountain",
  };
  const vivi: CardView = {
    instance_id: "v",
    name: "Vivi Ornitier",
    owner: ME,
    controller: ME,
    type_line: "Legendary Creature — Wizard",
    mana_abilities: [{ index: 0, ref: "own:0", label: "{0}: Add X mana" }],
  } as unknown as CardView;

  it("a land's ordinary {T} mana ability gets no pip", () => {
    expect(notableManaRefs(mountain, ["land:R"], "battlefield")).toEqual([]);
    const tapLand = {
      ...mountain,
      mana_abilities: [{ index: 0, ref: "own:0", tap_cost: true }],
    } as unknown as CardView;
    expect(notableManaRefs(tapLand, ["own:0"], "battlefield")).toEqual([]);
  });

  it("a land's mana ability without {T} does (a sacrifice, a once-a-turn)", () => {
    const sacLand = {
      ...mountain,
      mana_abilities: [
        { index: 0, ref: "own:0", tap_cost: true },
        { index: 1, ref: "own:1", sacrifice_cost: true },
      ],
    } as unknown as CardView;
    expect(notableManaRefs(sacLand, ["own:0", "own:1"], "battlefield")).toEqual(["own:1"]);
  });

  it("a nonland source does, {T} or not (#1621's Vivi)", () => {
    expect(notableManaRefs(vivi, ["own:0"], "battlefield")).toEqual(["own:0"]);
    const rock = {
      ...vivi,
      type_line: "Artifact",
      mana_abilities: [{ index: 0, ref: "own:0", tap_cost: true }],
    } as unknown as CardView;
    expect(notableManaRefs(rock, ["own:0"], "battlefield")).toEqual(["own:0"]);
  });

  it("a source in hand does (a Spirit Guide)", () => {
    expect(notableManaRefs(mountain, ["own:0"], "hand")).toEqual(["own:0"]);
  });

  it("nothing ready, nothing marked", () => {
    expect(notableManaRefs(vivi, [], "battlefield")).toEqual([]);
  });
});

// ---- pips, menu order and the popover gate (ADR 0105 sub-PR 3) -------

describe("readyPips: bolt and drop, from the digest, after §4", () => {
  const card = (id: string, type_line: string, mana: object[] = []): CardView =>
    ({
      instance_id: id,
      name: id,
      owner: ME,
      controller: ME,
      type_line,
      mana_abilities: mana,
    }) as unknown as CardView;
  const forest = card("forest", "Basic Land — Forest", [
    { index: 0, ref: "own:0", tap_cost: true },
  ]);
  const vivi = card("vivi", "Legendary Creature — Wizard", [{ index: 0, ref: "own:0" }]);
  const treasure = card("treasure", "Token Artifact — Treasure", [
    { index: 0, ref: "own:0", tap_cost: true, sacrifice_cost: true },
  ]);
  const solRing = card("sol", "Artifact", [{ index: 0, ref: "own:0", tap_cost: true }]);
  const guide = {
    ...card("guide", "Creature — Elemental Spirit"),
    mana_abilities: undefined,
    zone_mana_abilities: [{ index: 0, ref: "own:0" }],
  } as unknown as CardView;
  const legal = legalActionsOf(
    frame({
      legal_actions: {
        pass: true,
        sources: {
          forest: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
          vivi: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
          treasure: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
          sol: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
          guide: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
          equip: { kinds: ["activate"], moves: 3, abilities: ["own:0", "own:1"] },
          one: {
            kinds: ["activate", "mana"],
            moves: 2,
            abilities: ["own:2"],
            mana_abilities: ["own:0"],
          },
        },
      },
    }),
  );

  it("a basic land's {T} mana ability: no pip", () => {
    expect(readyPips(legal, forest, "battlefield")).toBe(NO_PIPS);
  });
  it("Vivi's free mana ability (no {T}): drop pip, no bolt", () => {
    expect(readyPips(legal, vivi, "battlefield")).toEqual({
      abilities: 0,
      mana: true,
      special: 0,
    });
  });
  it("a Treasure's sacrifice: drop pip", () => {
    expect(readyPips(legal, treasure, "battlefield").mana).toBe(true);
  });
  it("a nonland {T} rock: drop pip", () => {
    expect(readyPips(legal, solRing, "battlefield").mana).toBe(true);
  });
  it("a Spirit Guide in hand: drop pip", () => {
    expect(readyPips(legal, guide, "hand").mana).toBe(true);
  });
  it("counts the live activated-ability rows for the bolt", () => {
    expect(readyPips(legal, card("equip", "Artifact — Equipment"), "battlefield")).toEqual({
      abilities: 2,
      mana: false,
      special: 0,
    });
    const one = card("one", "Artifact", [{ index: 0, ref: "own:0" }]);
    expect(readyPips(legal, one, "battlefield")).toEqual({
      abilities: 1,
      mana: true,
      special: 0,
    });
  });
  it("no digest, highlights off, or a card not in it: no pip", () => {
    expect(readyPips(NO_LEGAL_ACTIONS, vivi, "battlefield")).toBe(NO_PIPS);
    expect(readyPips(visibleHighlights(legal, false), vivi, "battlefield")).toBe(NO_PIPS);
    expect(readyPips(legal, card("theirs", "Artifact"), "battlefield")).toBe(NO_PIPS);
  });
});

describe("pipCount", () => {
  it("prints nothing for one and the count from two up", () => {
    expect(pipCount(0)).toBe("");
    expect(pipCount(1)).toBe("");
    expect(pipCount(2)).toBe("2");
    expect(pipCount(5)).toBe("5");
  });
});

describe("readyFirst", () => {
  it("moves ready rows up and keeps each half's order", () => {
    const rows = ["a", "B", "c", "D", "e"];
    expect(readyFirst(rows, (r) => r === r.toUpperCase())).toEqual(["B", "D", "a", "c", "e"]);
  });
  it("leaves a list with no ready row as it was", () => {
    expect(readyFirst([3, 1, 2], () => false)).toEqual([3, 1, 2]);
  });
});

describe("digestRefusesRow: the popover's gate", () => {
  const digest = legalActionsOf(
    frame({
      legal_actions: {
        pass: true,
        sources: { shikari: { kinds: ["activate"], moves: 1, abilities: ["own:0"] } },
      },
    }),
  );

  it("the exact digest refuses a row it leaves out, and not one it lists", () => {
    expect(digest.exact).toBe(true);
    expect(digestRefusesRow(digest, "shikari", "own:0")).toBe(false);
    expect(digestRefusesRow(digest, "shikari", "own:1")).toBe(true);
    // A card with no entry has no live row at all.
    expect(digestRefusesRow(digest, "other", "own:0")).toBe(true);
  });

  it("no information refuses nothing: no digest, the capped list, a row with no ref", () => {
    expect(digestRefusesRow(NO_LEGAL_ACTIONS, "shikari", "own:1")).toBe(false);
    expect(digestRefusesRow(legalActionsOf(frame()), "shikari", "own:1")).toBe(false);
    const capped = legalActionsOf(
      frame({
        legal_moves: [
          move({
            kind: "activate",
            type: "activate_ability",
            source: "shikari",
            params: { ref: "own:0" },
          }),
        ],
      }),
    );
    expect(capped.known).toBe(true);
    expect(capped.exact).toBe(false);
    expect(capped.readyAbilityRefs("shikari")).toEqual(["own:0"]);
    expect(digestRefusesRow(capped, "shikari", "own:1")).toBe(false);
    expect(digestRefusesRow(digest, "shikari", undefined)).toBe(false);
  });
});

// ---- ADR 0105 sub-PR 5: combat --------------------------------------

describe("combat against the server's fixture", () => {
  const scenario = (name: string): Scenario => {
    const sc = scenarios.find((x) => x.name === name);
    if (!sc) throw new Error(`fixture has no ${name} scenario`);
    return sc;
  };

  it("declare attackers: the creature with a target is a candidate, and its target lights", () => {
    const sc = scenario("declare_attackers");
    const la = legalActionsOf(sc.view);
    const giant = sc.expect.find((e) => e.name === "Hill Giant")!;
    const defender = giant.attack_targets![0];
    const mine = sc.view.battlefield.cards.filter((c) => c.controller === sc.viewer);

    const idle = combatRings(la, "attack", null, mine);
    expect([...idle.candidates]).toEqual([giant.instance_id]);
    expect(idle.targets.size).toBe(0);

    const picked = combatRings(la, "attack", giant.instance_id, mine);
    expect([...picked.targets]).toEqual([defender]);

    // attack-with-all counts what the server would declare.
    const plan = planAttackAll(sc.view, sc.viewer, la);
    expect(plan.eligible.map((c) => c.instance_id)).toEqual([giant.instance_id]);
    expect(plan.blocked).toEqual([]);
  });

  it("declare blockers: both creatures are candidates, and the attacker lights for either", () => {
    const sc = scenario("declare_blockers");
    const la = legalActionsOf(sc.view);
    const mine = sc.view.battlefield.cards.filter((c) => c.controller === sc.viewer);
    const blockers = sc.expect.filter((e) => (e.blocks ?? []).length > 0);
    expect(blockers).toHaveLength(2);

    const idle = combatRings(la, "block", null, mine);
    expect(sorted([...idle.candidates])).toEqual(sorted(blockers.map((b) => b.instance_id)));
    for (const b of blockers) {
      expect([...combatRings(la, "block", b.instance_id, mine).targets]).toEqual(b.blocks);
    }
  });
});

describe("combatRings", () => {
  const digest = legalActionsOf(
    frame({
      legal_actions: {
        pass: true,
        sources: {
          bear: { kinds: ["attack"], moves: 2, attack_targets: ["opp", "walker"] },
          elf: { kinds: ["attack"], moves: 1, attack_targets: ["opp"] },
          wall: { kinds: ["block"], moves: 1, blocks: ["giant"] },
          vet: { kinds: ["block"], moves: 1, blocks: ["giant", "ogre"] },
        },
      },
    }),
  );
  const card = (id: string, extra: Partial<CardView> = {}): CardView =>
    ({ instance_id: id, name: id, controller: ME, owner: ME, ...extra }) as CardView;

  it("attack: candidates are the creatures with a target, never one already declared", () => {
    const r = combatRings(digest, "attack", null, [
      card("bear"),
      card("elf", { attacking_target: "opp" }),
      card("sick"),
    ]);
    expect([...r.candidates]).toEqual(["bear"]);
    expect(r.targets.size).toBe(0);
  });

  it("attack: a selected attacker lights exactly its own targets", () => {
    expect([...combatRings(digest, "attack", "bear", []).targets]).toEqual(["opp", "walker"]);
    expect([...combatRings(digest, "attack", "elf", []).targets]).toEqual(["opp"]);
    // One the digest has nothing for lights nothing.
    expect(combatRings(digest, "attack", "sick", []).targets.size).toBe(0);
  });

  it("block: candidates are the creatures with an attacker to block, and the selection lights them", () => {
    const r = combatRings(digest, "block", "vet", [
      card("wall"),
      card("vet"),
      card("bear"),
      card("old", { blocking_target: "giant" }),
    ]);
    expect([...r.candidates]).toEqual(["wall", "vet"]);
    expect([...r.targets]).toEqual(["giant", "ogre"]);
  });

  it("idle, no information, or nothing to light: no rings", () => {
    expect(combatRings(digest, "idle", "bear", [card("bear")])).toBe(NO_COMBAT_RINGS);
    expect(combatRings(NO_LEGAL_ACTIONS, "attack", "bear", [card("bear")])).toBe(NO_COMBAT_RINGS);
    expect(combatRings(legalActionsOf(frame()), "block", "vet", [card("vet")])).toBe(
      NO_COMBAT_RINGS,
    );
    // Highlights off is the lookup that knows nothing.
    expect(combatRings(visibleHighlights(digest, false), "attack", "bear", [card("bear")])).toBe(
      NO_COMBAT_RINGS,
    );
  });

  it("reads the legal_moves fallback the same way", () => {
    const la = legalActionsOf(
      frame({
        legal_moves: [
          move({
            kind: "attack",
            type: "declare_attacker",
            source: "bear",
            params: { attacker: "bear", target: "opp" },
          }),
        ],
      }),
    );
    const r = combatRings(la, "attack", "bear", [card("bear"), card("elf")]);
    expect([...r.candidates]).toEqual(["bear"]);
    expect([...r.targets]).toEqual(["opp"]);
  });
});

describe("attackTargetOpen and attackTargetListed: the defender gates", () => {
  const gate = legalActionsOf(
    frame({
      legal_actions: {
        pass: true,
        sources: { bear: { kinds: ["attack"], moves: 1, attack_targets: ["opp"] } },
      },
    }),
  );
  const bear = { instance_id: "bear", name: "Bear" } as CardView;

  it("opens a defender the selected attacker may attack, and shuts one it may not", () => {
    expect(attackTargetOpen(gate, bear, "bear", "opp")).toBe(true);
    expect(attackTargetOpen(gate, bear, "bear", "other")).toBe(false);
    // A summoning-sick creature (no entry) may attack no one.
    expect(attackTargetOpen(gate, undefined, "sick", "opp")).toBe(false);
  });

  it("no list, or a re-point of a declared attacker, keeps every defender open", () => {
    expect(attackTargetOpen(NO_LEGAL_ACTIONS, bear, "bear", "other")).toBe(true);
    expect(attackTargetOpen(legalActionsOf(frame()), bear, "bear", "other")).toBe(true);
    const declared = { ...bear, attacking_target: "opp" };
    expect(attackTargetOpen(gate, declared, "bear", "other")).toBe(true);
  });

  it("a permanent defender is clickable only when the list names it", () => {
    expect(attackTargetListed(gate, "bear", "opp")).toBe(true);
    expect(attackTargetListed(gate, "bear", "walker")).toBe(false);
    expect(attackTargetListed(gate, null, "opp")).toBe(false);
    expect(attackTargetListed(NO_LEGAL_ACTIONS, "bear", "opp")).toBe(false);
  });
});
