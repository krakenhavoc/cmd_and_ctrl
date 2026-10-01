// legalSpecialActions.test.ts — ADR 0105 sub-PR 4 (#1789), the pure
// half: special actions in the ordinary popover, and the pips on hand
// cards and face-down permanents.
//
//   - specialActionItems builds the rows the popover and the admin menu
//     share: ready rows (kind in the digest) take the accent and sort
//     first; an `available` row the EXACT digest leaves out is greyed
//     with a sentence; no digest, or the capped list, greys nothing new.
//   - readyPips counts the live special-action kinds for the star,
//     including a kind this client has no name for.
//   - handHasAction is what keeps a foretell-able, cyclable or Spirit
//     Guide card out of the hand's dim.

import { describe, it, expect } from "vitest";

import { SPECIAL_NOT_RIGHT_NOW, specialActionItems } from "./contextMenu.logic";
import {
  NO_LEGAL_ACTIONS,
  NO_PIPS,
  SPECIAL_ACTION_GENERIC,
  digestRefusesSpecial,
  handHasAction,
  hasPips,
  legalActionsOf,
  readyPips,
  specialActionName,
  specialPipTitle,
  visibleHighlights,
} from "./legalActions";
import type { CardView, GameView, LegalActionsView, LegalMoveView } from "./protocol";

const ME = "me";

const exact = (sources: LegalActionsView["sources"]) =>
  legalActionsOf({ legal_actions: { pass: true, sources } } as unknown as GameView);

const capped = (moves: LegalMoveView[]) =>
  legalActionsOf({ legal_moves: moves } as unknown as GameView);

const card = (id: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: id, name: id, owner: ME, controller: ME, ...extra }) as CardView;

// A hand card with two special actions: foretell (affordable) and
// plot (timing open, but the seat cannot pay it).
const both = () =>
  card("both", {
    type_line: "Creature — Dwarf",
    special_actions: [
      { kind: "plot", label: "Plot {3}{R}", cost: "{3}{R}", available: true },
      { kind: "foretell", label: "Foretell {2}", cost: "{2}", available: true },
    ],
  });

describe("specialActionItems: the popover's special-action rows", () => {
  const digest = exact({
    both: { kinds: ["special_action"], moves: 1, special_actions: ["foretell"] },
  });

  it("lights the row the digest lists and sorts it first", () => {
    const rows = specialActionItems(both(), ME, digest, digest);
    expect(rows.map((r) => r.id)).toEqual(["special-foretell", "special-plot"]);
    expect(rows[0].ready).toBe(true);
    expect(rows[0].disabled).toBe(false);
  });

  it("greys an available row the exact digest leaves out, with a sentence", () => {
    const rows = specialActionItems(both(), ME, digest, digest);
    const plot = rows.find((r) => r.id === "special-plot")!;
    expect(plot.disabled).toBe(true);
    expect(plot.ready).toBeUndefined();
    expect(plot.hint).toBe(SPECIAL_NOT_RIGHT_NOW);
  });

  it("sends the admin menu's payload, from the actor it is given", () => {
    const row = specialActionItems(both(), ME, digest, digest)[0];
    expect(row.action).toEqual({
      type: "special_action",
      params: { card_id: "both", kind: "foretell", strict: true, auto_tap: true },
      player: ME,
    });
  });

  it("highlights off: no accent and the server's order, but the gate stays", () => {
    const rows = specialActionItems(both(), ME, visibleHighlights(digest, false), digest);
    expect(rows.map((r) => r.id)).toEqual(["special-plot", "special-foretell"]);
    expect(rows.some((r) => r.ready)).toBe(false);
    expect(rows.find((r) => r.id === "special-plot")!.disabled).toBe(true);
  });

  it("no digest greys nothing new, and lights nothing", () => {
    const rows = specialActionItems(both(), ME, NO_LEGAL_ACTIONS, NO_LEGAL_ACTIONS);
    expect(rows.every((r) => !r.disabled && !r.ready)).toBe(true);
  });

  it("the capped list lights the row but greys none", () => {
    const list = capped([
      {
        type: "special_action",
        player: ME,
        kind: "special_action",
        label: "Foretell {2}",
        source: "both",
        params: { card_id: "both", kind: "foretell" },
      },
    ]);
    expect(list.exact).toBe(false);
    const rows = specialActionItems(both(), ME, list, list);
    expect(rows[0].id).toBe("special-foretell");
    expect(rows[0].ready).toBe(true);
    expect(rows.every((r) => !r.disabled)).toBe(true);
  });

  it("an unavailable row keeps the server's 'not right now', and is never lit", () => {
    const c = card("sus", {
      special_actions: [{ kind: "suspend", label: "Suspend 1—{R}", cost: "{R}", available: false }],
    });
    const lit = exact({
      sus: { kinds: ["special_action"], moves: 1, special_actions: ["suspend"] },
    });
    const [row] = specialActionItems(c, ME, lit, lit);
    expect(row.disabled).toBe(true);
    expect(row.hint).toBe("not right now");
    expect(row.ready).toBeUndefined();
  });
});

describe("digestRefusesSpecial", () => {
  const digest = exact({
    c: { kinds: ["special_action"], moves: 1, special_actions: ["foretell"] },
  });

  it("refuses a kind the exact digest leaves out, and not one it lists", () => {
    expect(digestRefusesSpecial(digest, "c", "foretell")).toBe(false);
    expect(digestRefusesSpecial(digest, "c", "plot")).toBe(true);
    expect(digestRefusesSpecial(digest, "other", "foretell")).toBe(true);
  });

  it("no information refuses nothing", () => {
    expect(digestRefusesSpecial(NO_LEGAL_ACTIONS, "c", "plot")).toBe(false);
    expect(digestRefusesSpecial(capped([]), "c", "plot")).toBe(false);
  });
});

describe("readyPips: the star", () => {
  const legal = exact({
    // Only a special action: not castable, not dead.
    foretell: { kinds: ["special_action"], moves: 1, special_actions: ["foretell"] },
    // Only a hand ability: cycling.
    cycler: { kinds: ["activate"], moves: 1, abilities: ["own:0"] },
    // Only a hand mana ability: a Spirit Guide.
    guide: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
    // A face-down permanent that can be turned face up.
    morph: { kinds: ["special_action"], moves: 1, special_actions: ["turn_face_up"] },
    // A kind this client has never heard of.
    room: { kinds: ["special_action"], moves: 2, special_actions: ["open_sesame"] },
  });

  it("a hand card with only a special action: the star", () => {
    expect(readyPips(legal, card("foretell"), "hand")).toEqual({
      abilities: 0,
      mana: false,
      special: 1,
    });
  });

  it("a hand card with only cycling: the bolt", () => {
    expect(readyPips(legal, card("cycler"), "hand")).toEqual({
      abilities: 1,
      mana: false,
      special: 0,
    });
  });

  it("a Spirit Guide in hand: the drop (§4: a source in hand is always marked)", () => {
    const guide = card("guide", {
      type_line: "Creature — Elemental Spirit",
      zone_mana_abilities: [{ index: 0, ref: "own:0" }],
    });
    expect(readyPips(legal, guide, "hand")).toEqual({ abilities: 0, mana: true, special: 0 });
  });

  it("a face-down permanent with turn face up: the star", () => {
    const morph = card("morph", { face_down: true, face_down_kind: "morph" });
    expect(readyPips(legal, morph, "battlefield").special).toBe(1);
  });

  it("an unknown kind still lights the star, under the generic name", () => {
    expect(readyPips(legal, card("room"), "battlefield").special).toBe(1);
    expect(specialActionName("open_sesame")).toBe(SPECIAL_ACTION_GENERIC);
    expect(specialPipTitle(["open_sesame"])).toBe(SPECIAL_ACTION_GENERIC);
  });

  it("highlights off, or no digest: no star", () => {
    expect(readyPips(visibleHighlights(legal, false), card("foretell"), "hand")).toBe(NO_PIPS);
    expect(readyPips(NO_LEGAL_ACTIONS, card("morph"), "battlefield")).toBe(NO_PIPS);
  });

  it("hasPips", () => {
    expect(hasPips(NO_PIPS)).toBe(false);
    expect(hasPips(readyPips(legal, card("foretell"), "hand"))).toBe(true);
    expect(hasPips(readyPips(legal, card("cycler"), "hand"))).toBe(true);
  });
});

describe("specialActionName / specialPipTitle", () => {
  it("names the kinds the client knows", () => {
    expect(specialActionName("foretell")).toBe("foretell");
    expect(specialActionName("turn_face_up")).toBe("turn face up");
    // Not a prototype key either.
    expect(specialActionName("toString")).toBe(SPECIAL_ACTION_GENERIC);
  });
  it("joins and deduplicates", () => {
    expect(specialPipTitle(["foretell", "plot"])).toBe("foretell, plot");
    expect(specialPipTitle(["x", "y"])).toBe(SPECIAL_ACTION_GENERIC);
    expect(specialPipTitle([])).toBe(SPECIAL_ACTION_GENERIC);
  });
});

describe("handHasAction: what keeps a hand card out of the dim", () => {
  const gate = exact({
    foretell: { kinds: ["special_action"], moves: 1, special_actions: ["foretell"] },
    cycler: { kinds: ["activate"], moves: 1, abilities: ["own:0"] },
    guide: { kinds: ["mana"], moves: 1, mana_abilities: ["own:0"] },
    bolt: { kinds: ["cast"], moves: 1, zones: ["hand"] },
  });

  it("a special action, a hand ability, or a hand mana ability", () => {
    expect(handHasAction(gate, "foretell")).toBe(true);
    expect(handHasAction(gate, "cycler")).toBe(true);
    expect(handHasAction(gate, "guide")).toBe(true);
  });

  it("not a cast (that is the cast gate's question), and nothing without a digest", () => {
    expect(handHasAction(gate, "bolt")).toBe(false);
    expect(handHasAction(gate, "nothing")).toBe(false);
    expect(handHasAction(NO_LEGAL_ACTIONS, "foretell")).toBe(false);
  });
});
