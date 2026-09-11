import { describe, it, expect } from "vitest";

import {
  canCastFromHand,
  canActivateAbility,
  canActivateLoyalty,
  canActivateSorcerySpeedAbility,
  canPassPriority,
  hasPriority,
  isActivePlayer,
  isMainPhase,
  stackEmpty,
} from "./timing";
import type { CardView, GameView, PlayerView, TurnView, ZoneView, StackItemView } from "./protocol";

// emptyZone returns a zone with no cards. The kind / owner fields
// don't matter for the timing predicates — they only read counts.
function emptyZone(kind: string, owner = ""): ZoneView {
  return { kind, owner, count: 0, cards: [] };
}

function turn(opts: Partial<TurnView> = {}): TurnView {
  return {
    number: 1,
    active_seat: 0,
    priority_holder: 0,
    phase: "precombat_main",
    step: "precombat_main",
    ...opts,
  };
}

function seat(id: string, idx: number, name = ""): PlayerView {
  return {
    id,
    name: name || `seat ${idx}`,
    seat: idx,
    life: 40,
    library: emptyZone("library", id),
    hand: emptyZone("hand", id),
    graveyard: emptyZone("graveyard", id),
    command: emptyZone("command", id),
    commander_damage: {},
    life_history: [],
  };
}

interface SnapOpts {
  step?: string;
  activeSeat?: number;
  priorityHolder?: number;
  splitSecond?: boolean;
  stackItems?: StackItemView[];
  battlefield?: CardView[];
}

function snap(o: SnapOpts = {}): GameView {
  const seats = [seat("p0", 0), seat("p1", 1)];
  return {
    id: "g",
    state: "active",
    seats,
    battlefield: { ...emptyZone("battlefield"), cards: o.battlefield ?? [] },
    stack: emptyZone("stack"),
    exile: emptyZone("exile"),
    turn: turn({
      step: o.step ?? "precombat_main",
      active_seat: o.activeSeat ?? 0,
      priority_holder: o.priorityHolder ?? 0,
    }),
    mulligans_open: false,
    stack_items: o.stackItems ?? [],
    split_second_active: o.splitSecond ?? false,
  };
}

function card(name: string, type: string, extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "c-" + name,
    name,
    owner: "p0",
    controller: "p0",
    type_line: type,
    ...extras,
  };
}

describe("priority + step predicates", () => {
  it("hasPriority is false during Untap / Cleanup (-1 sentinel)", () => {
    const s = snap({ step: "untap", priorityHolder: -1 });
    expect(hasPriority(s, "p0")).toBe(false);
  });

  it("hasPriority is true when priority_holder seat matches viewer", () => {
    const s = snap();
    expect(hasPriority(s, "p0")).toBe(true);
    expect(hasPriority(s, "p1")).toBe(false);
  });

  it("isActivePlayer reads active_seat", () => {
    const s = snap({ activeSeat: 1 });
    expect(isActivePlayer(s, "p1")).toBe(true);
    expect(isActivePlayer(s, "p0")).toBe(false);
  });

  it("isMainPhase covers both pre/post combat mains", () => {
    expect(isMainPhase(snap({ step: "precombat_main" }))).toBe(true);
    expect(isMainPhase(snap({ step: "postcombat_main" }))).toBe(true);
    expect(isMainPhase(snap({ step: "draw" }))).toBe(false);
    expect(isMainPhase(snap({ step: "begin_combat" }))).toBe(false);
  });

  it("stackEmpty reads both stack_items + stack.cards", () => {
    expect(stackEmpty(snap())).toBe(true);
    expect(
      stackEmpty(
        snap({
          stackItems: [
            {
              id: "x",
              kind: "spell",
              controller: "p0",
              owner: "p0",
              source_card_id: "x",
            },
          ],
        }),
      ),
    ).toBe(false);
  });
});

describe("canCastFromHand", () => {
  it("instant on opponent's turn during opponent's main phase = legal (priority is what matters)", () => {
    // Active = seat 1 but priority holder is seat 0 (e.g. seat 1 cast a spell, priority passed back to seat 0).
    const s = snap({ activeSeat: 1, priorityHolder: 0 });
    const c = card("Bolt", "Instant");
    expect(canCastFromHand(c, s, "p0").legal).toBe(true);
  });

  // S20: a targeted spell with an empty legal set can't be cast.
  it("targeted spell with no legal target = illegal", () => {
    const s = snap({ activeSeat: 0, priorityHolder: 0 });
    const blade = card("Doom Blade", "Instant", { legal_targets: { cards: [] } });
    const got = canCastFromHand(blade, s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBe("No legal target");
  });

  // S20 sub-PR 4: a modal spell is castable while enough options
  // are — untargeted ones always, targeted ones with a legal target.
  it("modal spell: castable when an untargeted option exists, not when every option lacks a target", () => {
    const s = snap({ activeSeat: 0, priorityHolder: 0 });
    const charm = card("Rakdos Charm", "Instant", {
      modes: {
        prompt: "Choose one",
        min: 1,
        max: 1,
        options: [
          {
            label: "Destroy target artifact.",
            target_mode: "permanent",
            legal_targets: { cards: [] },
          },
          { label: "Each creature deals 1 damage to its controller." },
        ],
      },
    });
    expect(canCastFromHand(charm, s, "p0").legal).toBe(true);
    const stuck = card("Charm", "Instant", {
      modes: {
        prompt: "Choose one",
        min: 1,
        max: 1,
        options: [
          { label: "A", target_mode: "permanent", legal_targets: { cards: [] } },
          { label: "B", target_mode: "player", legal_targets: { players: [] } },
        ],
      },
    });
    const got = canCastFromHand(stuck, s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBe("No castable mode");
  });

  it("targeted spell with a legal target = legal; free-form card untouched", () => {
    const s = snap({ activeSeat: 0, priorityHolder: 0 });
    const blade = card("Doom Blade", "Instant", { legal_targets: { cards: ["c-Bear"] } });
    expect(canCastFromHand(blade, s, "p0").legal).toBe(true);
    const freeForm = card("Homebrew", "Instant", { target_mode: "creature" });
    expect(canCastFromHand(freeForm, s, "p0").legal).toBe(true);
  });

  it("sorcery on opponent's turn = illegal (not your turn)", () => {
    const s = snap({ activeSeat: 1, priorityHolder: 0 });
    const c = card("Wrath", "Sorcery");
    const got = canCastFromHand(c, s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBe("Not your turn");
  });

  it("sorcery during own upkeep = illegal (only at sorcery speed)", () => {
    const s = snap({ step: "upkeep" });
    const c = card("Wrath", "Sorcery");
    expect(canCastFromHand(c, s, "p0").reason).toBe("Only at sorcery speed");
  });

  it("sorcery with stack non-empty = illegal", () => {
    const s = snap({
      stackItems: [
        {
          id: "x",
          kind: "spell",
          controller: "p1",
          owner: "p1",
          source_card_id: "x",
        },
      ],
    });
    const c = card("Wrath", "Sorcery");
    expect(canCastFromHand(c, s, "p0").reason).toBe("Stack isn't empty");
  });

  it("split-second blocks even instants", () => {
    const s = snap({ splitSecond: true });
    const c = card("Bolt", "Instant");
    expect(canCastFromHand(c, s, "p0").reason).toBe("Split second on the stack");
  });

  it("land on own main phase, stack empty = legal", () => {
    const s = snap();
    const c = card("Forest", "Basic Land — Forest");
    expect(canCastFromHand(c, s, "p0").legal).toBe(true);
  });

  it("land outside main phase = illegal", () => {
    const s = snap({ step: "draw" });
    const c = card("Forest", "Basic Land — Forest");
    expect(canCastFromHand(c, s, "p0").reason).toBe("Lands only on your main phase");
  });

  it("spectator (no viewerID) cannot cast", () => {
    const s = snap();
    const c = card("Bolt", "Instant");
    expect(canCastFromHand(c, s, null).reason).toBe("Spectator can't cast");
  });

  it("viewer without priority cannot cast", () => {
    const s = snap({ priorityHolder: 1 });
    const c = card("Bolt", "Instant");
    expect(canCastFromHand(c, s, "p0").reason).toBe("Not your priority");
  });
});

describe("canActivateAbility", () => {
  it("legal when viewer holds priority", () => {
    const c = card("Goblin Bombardment", "Enchantment");
    expect(canActivateAbility(c, snap(), "p0").legal).toBe(true);
  });

  it("blocked by split-second", () => {
    const c = card("Goblin Bombardment", "Enchantment");
    expect(canActivateAbility(c, snap({ splitSecond: true }), "p0").reason).toBe(
      "Split second on the stack",
    );
  });
});

// S24: equip is the catalog's first "activate only as a sorcery"
// ability, and this is the window it answers to.
describe("canActivateSorcerySpeedAbility", () => {
  it("legal on your own main phase with an empty stack", () => {
    expect(canActivateSorcerySpeedAbility(snap(), "p0").legal).toBe(true);
  });

  it("rejects outside a main phase", () => {
    expect(canActivateSorcerySpeedAbility(snap({ step: "upkeep" }), "p0").reason).toBe(
      "Sorcery-speed only",
    );
  });

  it("rejects on an opponent's turn", () => {
    expect(
      canActivateSorcerySpeedAbility(snap({ activeSeat: 1, priorityHolder: 0 }), "p0").reason,
    ).toBe("Not your turn");
  });

  it("rejects a spectator", () => {
    expect(canActivateSorcerySpeedAbility(snap(), null).reason).toBe("Spectator can't activate");
  });
});

describe("canActivateLoyalty", () => {
  const pwID = "pw-1";
  function snapWithPW(o: SnapOpts = {}): GameView {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    return snap({ ...o, battlefield: [c] });
  }

  it("legal on own main phase, stack empty, not yet activated", () => {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(canActivateLoyalty(c, snapWithPW(), "p0", false).legal).toBe(true);
  });

  it("rejects on opponent's turn", () => {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(
      canActivateLoyalty(c, snapWithPW({ activeSeat: 1, priorityHolder: 0 }), "p0", false).reason,
    ).toBe("Not your turn");
  });

  it("rejects when already activated this turn", () => {
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(canActivateLoyalty(c, snapWithPW(), "p0", true).reason).toBe(
      "Already activated this turn",
    );
  });

  it("rejects when source not on battlefield", () => {
    // snap() with no battlefield card.
    const c = card("Jace", "Legendary Planeswalker — Jace", { instance_id: pwID });
    expect(canActivateLoyalty(c, snap(), "p0", false).reason).toBe(
      "Planeswalker not on the battlefield",
    );
  });
});

describe("canPassPriority", () => {
  it("rejects on Untap / Cleanup (no-priority steps)", () => {
    expect(canPassPriority(snap({ step: "untap", priorityHolder: -1 }), "p0").reason).toBe(
      "No one holds priority this step",
    );
    expect(canPassPriority(snap({ step: "cleanup", priorityHolder: -1 }), "p0").reason).toBe(
      "No one holds priority this step",
    );
  });

  it("rejects when viewer doesn't hold priority", () => {
    expect(canPassPriority(snap({ priorityHolder: 1 }), "p0").reason).toBe("Not your priority");
  });

  it("legal when viewer holds priority on a granting step", () => {
    expect(canPassPriority(snap(), "p0").legal).toBe(true);
  });
});

// ADR 0034 — a modal DFC in hand is legal to play when EITHER face
// is legal right now, because the player has not chosen yet: the
// face picker opens AFTER this gate, not before it. This is the one
// client file whose BEHAVIOUR the face model changes; everywhere
// else keeps working untouched because the wire now hands it one
// clean type line per face instead of a "Sorcery // Land"
// concatenation.
describe("canCastFromHand over multiple faces", () => {
  function mdfc(frontType: string, backType: string): CardView {
    return card("Modal", frontType, {
      layout: "modal_dfc",
      faces: [
        { name: "Front", type_line: frontType },
        { name: "Back", type_line: backType },
      ],
    });
  }

  it("a sorcery-front / land-back MDFC is illegal outside a main phase", () => {
    const s = snap({ step: "declare_attackers" });
    expect(canCastFromHand(mdfc("Sorcery", "Land"), s, "p0").legal).toBe(false);
  });

  it("an INSTANT-front / land-back MDFC stays castable in combat", () => {
    const s = snap({ step: "declare_attackers" });
    expect(canCastFromHand(mdfc("Instant", "Land"), s, "p0").legal).toBe(true);
  });

  it("a creature-front / land-back MDFC is legal in a main phase", () => {
    const s = snap({ step: "precombat_main" });
    expect(canCastFromHand(mdfc("Creature — Elephant", "Land"), s, "p0").legal).toBe(true);
  });

  it("reports a face's denial rather than a blank refusal", () => {
    const s = snap({ step: "draw" });
    const got = canCastFromHand(mdfc("Sorcery", "Land"), s, "p0");
    expect(got.legal).toBe(false);
    expect(got.reason).toBeTruthy();
  });

  it("a transform card is gated on its FRONT face alone — CR 712.4", () => {
    const s = snap({ step: "declare_attackers" });
    const jace = card("Jace", "Legendary Creature — Human Wizard", {
      layout: "transform",
      faces: [
        { name: "Jace, Vryn's Prodigy", type_line: "Legendary Creature — Human Wizard" },
        // An instant back face must NOT make the front castable at
        // instant speed: the back cannot be cast at all.
        { name: "Back", type_line: "Instant" },
      ],
    });
    expect(canCastFromHand(jace, s, "p0").legal).toBe(false);
  });

  it("single-faced cards take exactly the path they always did", () => {
    const s = snap({ step: "declare_attackers" });
    expect(canCastFromHand(card("Bolt", "Instant"), s, "p0").legal).toBe(true);
    expect(canCastFromHand(card("Wrath", "Sorcery"), s, "p0").legal).toBe(false);
  });
});
