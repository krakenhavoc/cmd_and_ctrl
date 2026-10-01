// legalActionsA11y.test.ts — ADR 0105 §7, sub-PR 6 (#1789): the words
// the ready highlights say. A ready card's accessible name gains what
// it is ready for; a ready row or button gains "available"; each pip is
// named for what it opens onto; the combat pip follows the rings; and
// the phase display's live region says "N actions available" once per
// decision, counting what the board draws.

import { describe, it, expect } from "vitest";

import {
  DROP_PIP_LABEL,
  NO_COMBAT_RINGS,
  NO_LEGAL_ACTIONS,
  QUIET_ANNOUNCER,
  READY_PHRASE_GENERIC,
  actionableCount,
  announceArrival,
  boltPipLabel,
  combatPipFor,
  combatRings,
  legalActionsOf,
  readyAnnouncement,
  readyCardLabel,
  readyPhrases,
  specialActionPhrase,
  starPipLabel,
  visibleHighlights,
  withAvailable,
  type ReadyAnnouncer,
} from "./legalActions";
import type { CardView, GameView, LegalActionsView } from "./protocol";

const ME = "me";
const OPP = "opp";

const card = (id: string, typeLine: string, extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: id,
    name: id,
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: typeLine,
    ...extra,
  }) as CardView;

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const frame = (
  digest: LegalActionsView | undefined,
  parts: { hand?: CardView[]; battlefield?: CardView[]; graveyard?: CardView[] } = {},
): GameView =>
  ({
    id: "g",
    state: "active",
    seats: [
      {
        id: ME,
        name: "Me",
        seat: 0,
        hand: zone("hand", ME, parts.hand),
        library: zone("library", ME),
        graveyard: zone("graveyard", ME, parts.graveyard),
        command: zone("command", ME),
      },
      {
        id: OPP,
        name: "Opp",
        seat: 1,
        hand: zone("hand", OPP),
        library: zone("library", OPP),
        graveyard: zone("graveyard", OPP),
        command: zone("command", OPP),
      },
    ],
    battlefield: zone("battlefield", undefined, parts.battlefield),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    turn: { seq: 1, number: 2, active_seat: 0, priority_holder: 0, step: "precombat_main" },
    legal_actions: digest,
  }) as unknown as GameView;

const bolt = card("bolt", "Instant");
const forestOut = (f: CardView): CardView => ({ ...f, instance_id: "forest-bf" });
const wurm = card("wurm", "Creature — Wurm");
const forest = card("forest", "Basic Land — Forest", {
  mana_abilities: [{ index: 0, ref: "own:0", tap_cost: true, produced: "{G}" }],
});
const mdfc = card("mdfc", "Sorcery // Land");
const looting = card("looting", "Sorcery");
const engine = card("engine", "Artifact", {
  activated_abilities: [{ index: 0, ref: "own:0", label: "{1}: Scry 1" }],
});
const vivi = card("vivi", "Legendary Creature — Wizard", {
  mana_abilities: [{ index: 0, ref: "own:0", label: "Add X" }],
});
const guide = card("guide", "Creature — Elemental Spirit", {
  zone_mana_abilities: [{ index: 0, ref: "own:0", label: "Exile: Add {R}" }],
});
const foreteller = card("foreteller", "Instant", {
  special_actions: [{ kind: "foretell", label: "Foretell", available: true }],
} as Partial<CardView>);
const bear = card("bear", "Creature — Bear");
const declared = card("declared", "Creature — Bear", { attacking_target: OPP, tapped: true });
const wall = card("wall", "Creature — Wall");
const blocking = card("blocking", "Creature — Wall", { blocking_target: "giant" });

const digest: LegalActionsView = {
  pass: true,
  sources: {
    bolt: { kinds: ["cast"], moves: 1, zones: ["hand"] },
    forest: { kinds: ["land"], moves: 1, zones: ["hand"] },
    "forest-bf": { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
    mdfc: { kinds: ["cast", "land"], moves: 2, zones: ["hand"] },
    looting: { kinds: ["cast"], moves: 1, zones: ["graveyard"] },
    engine: { kinds: ["activate"], moves: 1, abilities: ["own:0"] },
    vivi: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
    guide: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
    foreteller: { kinds: ["special_action"], moves: 1, special_actions: ["foretell"] },
  },
};
const legal = legalActionsOf(frame(digest));

describe("readyPhrases: what a ready card's name gains", () => {
  it("a castable spell is castable; an uncastable one says nothing", () => {
    expect(readyPhrases(legal, bolt, "hand")).toEqual(["castable"]);
    expect(readyPhrases(legal, wurm, "hand")).toEqual([]);
  });

  it("a land in hand is a playable land; on the battlefield its {T} mana says nothing (§4)", () => {
    expect(readyPhrases(legal, forest, "hand")).toEqual(["playable land"]);
    expect(legal.readyManaRefs("forest-bf")).toEqual(["own:0"]);
    expect(readyPhrases(legal, forestOut(forest), "battlefield")).toEqual([]);
  });

  it("a card with a castable face and a land face says both", () => {
    expect(readyPhrases(legal, mdfc, "hand")).toEqual(["castable", "playable land"]);
  });

  it("castable is per zone: a flashback card is castable from the graveyard only", () => {
    expect(readyPhrases(legal, looting, "graveyard")).toEqual(["castable"]);
    expect(readyPhrases(legal, looting, "hand")).toEqual([]);
  });

  it("a live activated ability, and a mana ability worth a drop pip", () => {
    expect(readyPhrases(legal, engine, "battlefield")).toEqual(["has an ability you can activate"]);
    expect(readyPhrases(legal, vivi, "battlefield")).toEqual(["has a mana ability you can use"]);
    expect(readyPhrases(legal, guide, "hand")).toEqual(["has a mana ability you can use"]);
  });

  it("a special action is named by kind, and an unknown kind generically", () => {
    expect(readyPhrases(legal, foreteller, "hand")).toEqual(["can be foretold"]);
    expect(specialActionPhrase("turn_face_up")).toBe("can be turned face up");
    expect(specialActionPhrase("unlock")).toBe("has a door you can unlock");
    expect(specialActionPhrase("something_new")).toBe("has a special action");
  });

  it("combat: a candidate can attack or can block; a declared creature says neither", () => {
    const combat = legalActionsOf(
      frame({
        pass: false,
        sources: {
          bear: { kinds: ["attack"], moves: 1, attack_targets: [OPP] },
          declared: { kinds: ["attack"], moves: 1, attack_targets: [OPP] },
          wall: { kinds: ["block"], moves: 1, blocks: ["giant"] },
          blocking: { kinds: ["block"], moves: 1, blocks: ["giant"] },
        },
      }),
    );
    expect(readyPhrases(combat, bear, "battlefield")).toEqual(["can attack"]);
    expect(readyPhrases(combat, declared, "battlefield")).toEqual([]);
    expect(readyPhrases(combat, wall, "battlefield")).toEqual(["can block"]);
    expect(readyPhrases(combat, blocking, "battlefield")).toEqual([]);
  });

  it("a combat target says what may be done to it", () => {
    const giant = card("giant", "Creature — Giant", { controller: OPP, attacking_target: ME });
    const walker = card("walker", "Planeswalker — Jace", { controller: OPP });
    expect(readyPhrases(NO_LEGAL_ACTIONS, giant, "battlefield", true)).toEqual(["can be blocked"]);
    expect(readyPhrases(NO_LEGAL_ACTIONS, walker, "battlefield", true)).toEqual([
      "can be attacked",
    ]);
  });

  it("highlights off, or no digest: nothing", () => {
    expect(readyPhrases(visibleHighlights(legal, false), bolt, "hand")).toEqual([]);
    expect(readyPhrases(NO_LEGAL_ACTIONS, bolt, "hand")).toEqual([]);
  });
});

describe("labels", () => {
  it("a ready card's name gains its phrases; a plain card keeps its name", () => {
    expect(readyCardLabel("Lightning Bolt", true, ["castable"])).toBe("Lightning Bolt, castable");
    expect(readyCardLabel("Forest", true, ["playable land"])).toBe("Forest, playable land");
    expect(readyCardLabel("Craw Wurm", false, ["castable"])).toBe("Craw Wurm");
  });

  it("a ring nothing explains is still never silent", () => {
    expect(readyCardLabel("Thing", true, [])).toBe(`Thing, ${READY_PHRASE_GENERIC}`);
  });

  it("a ready row or button gains 'available'", () => {
    expect(withAvailable("cast commander", true)).toBe("cast commander, available");
    expect(withAvailable("cast commander", false)).toBe("cast commander");
  });

  it("each pip is named for what it opens onto", () => {
    expect(boltPipLabel(1)).toBe("Activate an ability — open actions");
    expect(boltPipLabel(3)).toBe("3 abilities you can activate — open actions");
    expect(DROP_PIP_LABEL).toBe("Mana ability available — open actions");
    expect(starPipLabel(["foretell"])).toBe("Foretell available — open actions");
    expect(starPipLabel(["turn_face_up"])).toBe("Turn face up available — open actions");
    expect(starPipLabel(["mystery"])).toBe("Special action available — open actions");
  });
});

describe("combatPipFor", () => {
  const combat = legalActionsOf(
    frame({
      pass: false,
      sources: {
        bear: { kinds: ["attack"], moves: 1, attack_targets: [OPP] },
        wall: { kinds: ["block"], moves: 1, blocks: ["giant"] },
      },
    }),
  );

  it("a sword on an attack candidate, a shield on a block candidate", () => {
    const atk = combatRings(combat, "attack", null, [bear, wall]);
    expect(combatPipFor(combat, atk, ["bear"])).toBe("attack");
    const blk = combatRings(combat, "block", null, [bear, wall]);
    expect(combatPipFor(combat, blk, ["wall"])).toBe("block");
  });

  it("a group shows the pip when any member is a candidate", () => {
    const atk = combatRings(combat, "attack", null, [bear]);
    expect(combatPipFor(combat, atk, ["other", "bear"])).toBe("attack");
  });

  it("no candidate, no pip", () => {
    expect(combatPipFor(combat, NO_COMBAT_RINGS, ["bear"])).toBeNull();
  });
});

describe("actionableCount: the N of 'N actions available'", () => {
  const view = frame(digest, {
    hand: [bolt, wurm, forest, guide, foreteller],
    battlefield: [engine, vivi, forestOut(forest)],
    graveyard: [looting],
  });
  const lookup = legalActionsOf(view);

  it("counts the cards the board highlights, one per card", () => {
    // bolt, forest (land drop), guide, foreteller in hand; engine and
    // vivi on the battlefield; looting in the graveyard. Not the wurm,
    // and not a battlefield land's {T} mana.
    expect(actionableCount(lookup, view, ME)).toBe(7);
  });

  it("is 0 with highlights off, autopass passing, or no digest", () => {
    expect(actionableCount(visibleHighlights(lookup, false), view, ME)).toBe(0);
    expect(actionableCount(NO_LEGAL_ACTIONS, view, ME)).toBe(0);
    expect(actionableCount(lookup, null, ME)).toBe(0);
  });

  it("counts only the seat's own cards", () => {
    expect(actionableCount(lookup, view, OPP)).toBe(0);
  });
});

describe("announceArrival: once per decision", () => {
  const feed = (counts: number[]): string[] => {
    let a: ReadyAnnouncer = QUIET_ANNOUNCER;
    return counts.map((n) => {
      a = announceArrival(a, n);
      return a.text;
    });
  };

  it("says the line when the decision arrives, and not again on later frames of it", () => {
    expect(feed([0, 3, 3, 2, 4])).toEqual([
      "",
      "3 actions available",
      "3 actions available",
      "3 actions available",
      "3 actions available",
    ]);
  });

  it("empties when the decision ends, so the next arrival is news again", () => {
    expect(feed([3, 0, 3])).toEqual(["3 actions available", "", "3 actions available"]);
  });

  it("singular for one", () => {
    expect(readyAnnouncement(1)).toBe("1 action available");
    expect(feed([1])).toEqual(["1 action available"]);
  });

  it("a quiet frame keeps the same object, so nothing re-renders", () => {
    expect(announceArrival(QUIET_ANNOUNCER, 0)).toBe(QUIET_ANNOUNCER);
    const live = announceArrival(QUIET_ANNOUNCER, 2);
    expect(announceArrival(live, 5)).toBe(live);
  });
});
