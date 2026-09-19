import { describe, it, expect } from "vitest";

import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";
import { applyCastChoices } from "./targeting";
import type { CastChoices } from "./targeting";

// castPreview.test.ts — #696. The auto-tap preview used to be told
// only which card, so it priced the printed cost from the hand and
// disabled "Auto-tap & cast" on every cast whose real price was
// something else. These two builders are what stops that happening
// again, and the fact that matters most is the LAST test here: the
// two of them and applyCastChoices must name the same cast, because
// one builds the preview's question and the other builds the payload
// the confirm button sends.

describe("castPreviewParams — the choices announced so far", () => {
  it("carries the source zone, the alternative cost, the optional costs, the taps and the face", () => {
    const choices: CastChoices = {
      fromZone: "graveyard",
      altCost: "flashback",
      optionalCosts: [0, 0],
      tapIDs: ["bird", "elf"],
      face: 1,
      // Not part of the price question the endpoint takes as cast
      // params — x rides its own `x` param, the rest are paid rather
      // than priced.
      xValue: 3,
      discardIDs: ["a"],
      altCostIDs: ["b"],
    };
    expect(castPreviewParams(choices)).toEqual({
      fromZone: "graveyard",
      alternativeCost: "flashback",
      optionalCosts: [0, 0],
      tapIDs: ["bird", "elf"],
      face: 1,
    });
  });

  it("sends nothing for a plain hand cast, because every omission is the server's default", () => {
    expect(castPreviewParams({})).toEqual({});
    expect(castPreviewParams(null)).toEqual({});
    // Face 0 is the front face and the server's default; an empty
    // list and an absent one are the same announcement.
    expect(castPreviewParams({ face: 0, optionalCosts: [], tapIDs: [] })).toEqual({});
  });

  it("copies the lists rather than aliasing the live choices", () => {
    const choices: CastChoices = { optionalCosts: [0], tapIDs: ["bird"] };
    const out = castPreviewParams(choices);
    out.optionalCosts?.push(1);
    out.tapIDs?.push("elf");
    expect(choices.optionalCosts).toEqual([0]);
    expect(choices.tapIDs).toEqual(["bird"]);
  });
});

describe("castPreviewParamsFromPayload — the stashed cast being retried", () => {
  it("reads the wire payload the auto-tap retry will replay", () => {
    expect(
      castPreviewParamsFromPayload({
        instance_id: "bolt",
        from_zone: "exile",
        alternative_cost: "flashback",
        optional_costs: [1],
        tap_ids: ["bird"],
        face: 1,
        strict: true,
      }),
    ).toEqual({
      fromZone: "exile",
      alternativeCost: "flashback",
      optionalCosts: [1],
      tapIDs: ["bird"],
      face: 1,
    });
  });

  it("drops a field of the wrong shape rather than sending a query the server would refuse", () => {
    expect(
      castPreviewParamsFromPayload({
        from_zone: 7,
        alternative_cost: "",
        optional_costs: "kicker",
        tap_ids: [42],
        face: "back",
      }),
    ).toEqual({});
    expect(castPreviewParamsFromPayload(undefined)).toEqual({});
  });
});

describe("the preview and the cast name the same cast", () => {
  it("prices what applyCastChoices sends", () => {
    const choices: CastChoices = {
      fromZone: "graveyard",
      altCost: "flashback",
      optionalCosts: [0],
      tapIDs: ["bird"],
      face: 1,
    };
    const payload: Record<string, unknown> = { instance_id: "looting" };
    applyCastChoices(payload, choices);
    // The round trip: the choices the pickers hold, and the payload
    // the confirm button dispatched and Game.svelte stashed, must
    // produce the same preview query — otherwise the modal that says
    // "affordable" and the cast that follows it are about different
    // casts.
    expect(castPreviewParamsFromPayload(payload)).toEqual(castPreviewParams(choices));
  });
});
