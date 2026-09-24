import { describe, it, expect } from "vitest";

import { castableFromZone } from "./zoneBrowser.logic";
import { castableFaceIndex } from "./faces";
import { applyCastChoices } from "./targeting";
import type { CardView } from "./protocol";

// castFromZone.test.ts — S29. Two small surfaces, both of which turn
// into a wrong button or a wrong payload if they drift:
//
//   - castableFromZone decides whether the zone browser offers a
//     cast at all. Since #1055 `castable_here` is the VIEWER's own
//     answer, so this gate is the bit and the zone and nothing else;
//     the ownership check it used to carry existed only because the
//     bit was public and meant the pile owner's answer, and it is the
//     server that answers "whose" now.
//   - applyCastChoices is the single place that knows the wire
//     names, and `from_zone` is the one that decides which pile the
//     server reaches into.

function inYard(extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: "looting",
    name: "Faithless Looting",
    owner: "me",
    controller: "me",
    type_line: "Sorcery",
    ...extras,
  };
}

describe("castableFromZone — who gets the graveyard cast button", () => {
  it("offers the cast when the server marked the card for THIS frame", () => {
    expect(castableFromZone(inYard({ castable_here: true }), "graveyard")).toBe(true);
  });

  // #1055. A permission names an OBJECT, so a card in an opponent's
  // graveyard is castable by the seat that holds one — Wrexial's "cast
  // target instant or sorcery card from that player's graveyard". That
  // used to need a second gate here, "or `exile_play` names me",
  // because the bit was public and meant the pile owner. The server
  // stamps it per seat now, so the same one-field read answers both
  // shapes: the owner's own flashback, and a holder's grant over
  // somebody else's pile.
  it("does not care whose pile it is — the bit already does", () => {
    const granted = inYard({ owner: "them", castable_here: true, exile_play: { player: "me" } });
    expect(castableFromZone(granted, "graveyard")).toBe(true);
  });

  // The bystander's frame: the card, the public grant naming somebody
  // else, and no bit. `exile_play` must not put a button back.
  it("withholds it from a card the server did not mark", () => {
    expect(castableFromZone(inYard(), "graveyard")).toBe(false);
    expect(castableFromZone(inYard({ castable_here: false }), "graveyard")).toBe(false);
    expect(castableFromZone(inYard({ exile_play: { player: "them" } }), "graveyard")).toBe(false);
  });

  // #1185. `castableFromZone` never reads `cant_cast` itself — the
  // docs/protocol.md contract is that `castable_here` already folds it
  // in ("no cant_cast && at least one claimable price"), so a graveyard
  // card blocked by its own printed clause never gets a true bit to
  // read here in the first place. This pins that contract at the one
  // boundary #1185 audited, rather than leaving it implicit.
  it("a cant_cast clause never reaches this gate as a true bit", () => {
    expect(
      castableFromZone(inYard({ cant_cast: "Cast this spell only during combat" }), "graveyard"),
    ).toBe(false);
    expect(
      castableFromZone(
        inYard({ cant_cast: "Cast this spell only during combat", castable_here: false }),
        "graveyard",
      ),
    ).toBe(false);
  });

  it("is graveyard-only — exile keeps its own grant-keyed button", () => {
    const card = inYard({ castable_here: true });
    expect(castableFromZone(card, "exile")).toBe(false);
    expect(castableFromZone(card, "command")).toBe(false);
    expect(castableFromZone(card, "stack")).toBe(false);
  });

  // #1173. `castable_here` is stamped on `faces[i]` for a permission
  // printed on the BACK face (#1171) — the front half of such a card
  // is not a cast surface here, and reading the card's own top-level
  // block alone (always face 0's answer for a front-up pile, CR
  // 712.8a) missed it. Latent today — no catalog card prints a
  // back-face graveyard permission yet — but the union is the
  // documented rule (docs/protocol.md, "castable_here… for every
  // castable FACE") and the fixture below is exactly the shape #1171
  // ships when one does.
  it("shows the button when only the BACK face's castable_here is true", () => {
    const gravecrawlerShaped = inYard({
      name: "Refraction Elemental",
      layout: "modal_dfc",
      castable_here: false,
      faces: [
        { name: "Refraction Elemental", type_line: "Creature — Elemental", castable_here: false },
        { name: "Refraction's Echo", type_line: "Sorcery", castable_here: true },
      ],
    });
    expect(castableFromZone(gravecrawlerShaped, "graveyard")).toBe(true);
    // The same union `castableFromZone` reads (`castableFaces`,
    // faces.ts) is what the face picker's own default reads — so the
    // button and the picker it opens agree on which half this is.
    // The picker opens on face 1 rather than defaulting to the front,
    // which the server would refuse this cast from.
    expect(castableFaceIndex(gravecrawlerShaped, (f) => f.castable_here === true)).toBe(1);
  });

  it("still withholds the button when neither face's castable_here is true", () => {
    const neither = inYard({
      layout: "modal_dfc",
      faces: [
        { name: "Front", type_line: "Sorcery" },
        { name: "Back", type_line: "Land" },
      ],
    });
    expect(castableFromZone(neither, "graveyard")).toBe(false);
  });
});

describe("applyCastChoices — from_zone on the wire", () => {
  it("omits from_zone for a hand cast", () => {
    const params: Record<string, unknown> = { instance_id: "x" };
    applyCastChoices(params, {});
    expect(params).toEqual({ instance_id: "x" });
  });

  it("sends the zone alongside the rest of the announce-time choices", () => {
    const params: Record<string, unknown> = { instance_id: "x" };
    applyCastChoices(params, { fromZone: "graveyard", altCost: "flashback" });
    expect(params).toEqual({
      instance_id: "x",
      from_zone: "graveyard",
      alternative_cost: "flashback",
    });
  });
});
