import { describe, expect, it } from "vitest";
import {
  applyCastChoices,
  castBlightOffer,
  castTeamworkOffer,
  optionalCostPayOptions,
} from "./targeting";
import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";
import type { CardView } from "./protocol";

// #1703: teamwork (CR 702.194a) and blight (CR 701.68a) as optional
// costs — the offer's payers, the claimed offer the cast chain asks
// about, and the two wire lists.

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "HULK SMASH!", owner: "me", controller: "me", ...extras };
}

const teamworkCard = card({
  optional_costs: [
    {
      index: 0,
      key: "teamwork",
      label: "Teamwork 4",
      teamwork: 4,
      teamwork_options: { cards: ["a", "b"], min: 1, max: 0 },
    },
  ],
});

const blightCard = card({
  name: "Pyrrhic Strike",
  optional_costs: [
    {
      index: 0,
      key: "blight",
      label: "Blight 2",
      blight: 2,
      blight_options: { cards: ["bear"], min: 1, max: 1 },
    },
  ],
});

describe("optionalCostPayOptions — teamwork and blight", () => {
  it("lists the creatures that could pay", () => {
    expect(optionalCostPayOptions(teamworkCard.optional_costs![0])).toEqual(["a", "b"]);
    expect(optionalCostPayOptions(blightCard.optional_costs![0])).toEqual(["bear"]);
  });

  it("present-and-empty is an offer that cannot be taken", () => {
    expect(
      optionalCostPayOptions({
        index: 0,
        key: "teamwork",
        teamwork: 4,
        teamwork_options: { min: 1, max: 0 },
      }),
    ).toEqual([]);
    expect(
      optionalCostPayOptions({
        index: 0,
        key: "blight",
        blight: 1,
        blight_options: { min: 1, max: 1 },
      }),
    ).toEqual([]);
  });
});

describe("castTeamworkOffer / castBlightOffer", () => {
  it("is undefined until the offer is claimed", () => {
    expect(castTeamworkOffer(teamworkCard, {})).toBeUndefined();
    expect(castBlightOffer(blightCard, { optionalCosts: [] })).toBeUndefined();
  });

  it("returns the number and the payers once claimed", () => {
    expect(castTeamworkOffer(teamworkCard, { optionalCosts: [0] })).toMatchObject({
      n: 4,
      options: ["a", "b"],
    });
    expect(castBlightOffer(blightCard, { optionalCosts: [0] })).toMatchObject({
      n: 2,
      options: ["bear"],
    });
  });

  it("does not confuse the two", () => {
    expect(castBlightOffer(teamworkCard, { optionalCosts: [0] })).toBeUndefined();
    expect(castTeamworkOffer(blightCard, { optionalCosts: [0] })).toBeUndefined();
  });
});

describe("the wire", () => {
  it("cast_spell carries teamwork_ids and blight_ids only when paid", () => {
    const params: Record<string, unknown> = { instance_id: "spell" };
    applyCastChoices(params, { optionalCosts: [0], teamworkIDs: ["a", "b"], blightIDs: ["bear"] });
    expect(params).toEqual({
      instance_id: "spell",
      optional_costs: [0],
      teamwork_ids: ["a", "b"],
      blight_ids: ["bear"],
    });
    const none: Record<string, unknown> = { instance_id: "spell" };
    applyCastChoices(none, { teamworkIDs: [], blightIDs: [] });
    expect(none).toEqual({ instance_id: "spell" });
  });

  it("the auto-tap preview excludes them from the plan", () => {
    expect(castPreviewParams({ teamworkIDs: ["a"], blightIDs: ["bear"] })).toEqual({
      teamworkIDs: ["a"],
      blightIDs: ["bear"],
    });
    expect(castPreviewParamsFromPayload({ teamwork_ids: ["a"], blight_ids: ["bear"] })).toEqual({
      teamworkIDs: ["a"],
      blightIDs: ["bear"],
    });
  });
});
