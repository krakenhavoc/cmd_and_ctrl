// #3831: where a card menu's "Rulings" row points, and which cards get it.

import { describe, expect, it } from "vitest";

import { offersRulings, rulingsURL } from "./rulings";

const GATHERER = "https://gatherer.wizards.com/Pages/Card/Details.aspx?multiverseid=";

describe("offersRulings", () => {
  it("offers a face-up real card the viewer can see", () => {
    expect(offersRulings({ scryfall_id: "abc" })).toBe(true);
  });

  it("does not offer a card the viewer cannot see (no scryfall_id on the wire)", () => {
    expect(offersRulings({})).toBe(false);
    expect(offersRulings({ scryfall_id: "" })).toBe(false);
    expect(offersRulings(null)).toBe(false);
    expect(offersRulings(undefined)).toBe(false);
  });

  it("does not offer a face-down card, even one the viewer may look at", () => {
    expect(offersRulings({ scryfall_id: "abc", face_down: true })).toBe(false);
  });

  it("does not offer a token", () => {
    expect(offersRulings({ scryfall_id: "tok", is_token: true })).toBe(false);
  });
});

describe("rulingsURL", () => {
  it("is Gatherer when the printing has a multiverse id", () => {
    const meta = { set: "lea", collector_number: "232", multiverse_ids: [1234] };
    expect(rulingsURL({ scryfall_id: "abc" }, meta)).toBe(`${GATHERER}1234`);
  });

  it("is the Scryfall rulings section by set and collector number when Gatherer has no id", () => {
    const meta = { set: "ecc", collector_number: "57", multiverse_ids: [] };
    expect(rulingsURL({ scryfall_id: "abc" }, meta)).toBe(
      "https://scryfall.com/card/ecc/57#rulings",
    );
    expect(rulingsURL({ scryfall_id: "abc" }, { set: "ecc", collector_number: "57" })).toBe(
      "https://scryfall.com/card/ecc/57#rulings",
    );
  });

  it("names the Scryfall id until the metadata is here", () => {
    expect(rulingsURL({ scryfall_id: "04002706-2236-4b79-bdea-4f263e43cb9c" }, null)).toBe(
      "https://scryfall.com/card/04002706-2236-4b79-bdea-4f263e43cb9c#rulings",
    );
    expect(rulingsURL({ scryfall_id: "abc" }, {})).toBe("https://scryfall.com/card/abc#rulings");
  });

  it("opens the side of a double-faced card that is up", () => {
    const meta = { set: "isd", collector_number: "51", multiverse_ids: [226749, 226755] };
    expect(rulingsURL({ scryfall_id: "abc" }, meta)).toBe(`${GATHERER}226749`);
    expect(rulingsURL({ scryfall_id: "abc", active_face: 0 }, meta)).toBe(`${GATHERER}226749`);
    expect(rulingsURL({ scryfall_id: "abc", active_face: 1 }, meta)).toBe(`${GATHERER}226755`);
    expect(rulingsURL({ scryfall_id: "abc", active_face: 0 }, meta, 1)).toBe(`${GATHERER}226755`);
  });

  it("falls back to the front's id when a face has none", () => {
    const meta = { set: "x", collector_number: "1", multiverse_ids: [42] };
    expect(rulingsURL({ scryfall_id: "abc", active_face: 1 }, meta)).toBe(`${GATHERER}42`);
  });

  it("links a flavor-name printing by its set and collector number (Kefka Palazzo)", () => {
    // Final Fantasy's Purphoros, God of the Forge, printed as "Kefka
    // Palazzo": it has no Gatherer entry, and Scryfall's page is keyed by
    // the printing, not by the name on the card.
    const meta = { set: "fca", collector_number: "14", multiverse_ids: [] };
    expect(rulingsURL({ scryfall_id: "fbd25e79-d29b-4b51-9653-3932c980b931" }, meta)).toBe(
      "https://scryfall.com/card/fca/14#rulings",
    );
  });

  it("escapes a collector number that is not a plain number", () => {
    const meta = { set: "pf06", collector_number: "12a/★" };
    expect(rulingsURL({ scryfall_id: "abc" }, meta)).toBe(
      "https://scryfall.com/card/pf06/12a%2F%E2%98%85#rulings",
    );
  });

  it("is null for a hidden or face-down card and a token", () => {
    const meta = { set: "ecc", collector_number: "57", multiverse_ids: [1] };
    expect(rulingsURL({}, meta)).toBeNull();
    expect(rulingsURL({ scryfall_id: "abc", face_down: true }, meta)).toBeNull();
    expect(rulingsURL({ scryfall_id: "abc", is_token: true }, meta)).toBeNull();
    expect(rulingsURL(null, meta)).toBeNull();
  });

  it("ignores a malformed multiverse id", () => {
    const meta = { set: "ecc", collector_number: "57", multiverse_ids: [0, -3] };
    expect(rulingsURL({ scryfall_id: "abc" }, meta)).toBe(
      "https://scryfall.com/card/ecc/57#rulings",
    );
  });
});
