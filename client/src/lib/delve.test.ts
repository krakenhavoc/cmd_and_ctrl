import { describe, expect, it } from "vitest";
import type { AutoTapPreview } from "./api";
import { castPreviewParams, castPreviewParamsFromPayload } from "./castPreview";
import { chooseDelveForMe, delveLimit, delveOptionIDs, hasDelveChoice } from "./delve";
import type { CardView } from "./protocol";
import { applyCastChoices } from "./targeting";

// delve.test.ts — ADR 0100: the delve picker's pure half, and the
// delve_ids field on the cast payload and the preview query.

function cruise(options: string[], max = 7): CardView {
  return {
    instance_id: "cruise",
    name: "Treasure Cruise",
    delve: { options: { cards: options }, max },
  } as unknown as CardView;
}

describe("delve picker", () => {
  it("reads the server's options in order", () => {
    expect(delveOptionIDs(cruise(["a", "b", "c"]))).toEqual(["a", "b", "c"]);
    expect(hasDelveChoice(cruise(["a"]))).toBe(true);
    expect(hasDelveChoice(cruise([]))).toBe(false);
    expect(hasDelveChoice({ instance_id: "x" } as unknown as CardView)).toBe(false);
  });

  it("caps by the preview's budget, and by the cards there are", () => {
    const card = cruise(["a", "b", "c"], 7);
    const preview = { ok: true, cost: "{7}{U}", delve_budget: 2 } as AutoTapPreview;
    expect(delveLimit(preview, card)).toBe(2);
    expect(delveLimit(null, card)).toBe(3);
  });

  it("reads an absent budget on an answered preview as zero", () => {
    const preview = { ok: true, cost: "{0}" } as AutoTapPreview;
    expect(delveLimit(preview, cruise(["a", "b"], 7))).toBe(0);
  });

  it("chooses for me from the front of the server's order", () => {
    expect(chooseDelveForMe(["land", "spent", "flashback"], 2)).toEqual(["land", "spent"]);
    expect(chooseDelveForMe(["a"], 0)).toEqual([]);
  });
});

describe("delve on the wire", () => {
  it("rides cast_spell as delve_ids, and is omitted when empty", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { delveIDs: ["a", "b"] });
    expect(params.delve_ids).toEqual(["a", "b"]);
    const none: Record<string, unknown> = {};
    applyCastChoices(none, { delveIDs: [] });
    expect(none.delve_ids).toBeUndefined();
  });

  it("reaches the preview from the choices and from a stashed payload", () => {
    expect(castPreviewParams({ delveIDs: ["a"] }).delveIDs).toEqual(["a"]);
    expect(castPreviewParamsFromPayload({ delve_ids: ["a", 3, ""] }).delveIDs).toEqual(["a"]);
  });
});
