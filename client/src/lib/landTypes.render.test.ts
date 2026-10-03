// @vitest-environment jsdom
//
// ADR 0109 §1, the client half: a land a resolved effect has changed
// wears a badge naming the new type, with the effect and its duration in
// the tooltip. landTypeBadge is unit-tested in landTypes.test.ts; this
// file checks Card.svelte draws it.

import { describe, it, expect, afterEach } from "vitest";

import Card from "./components/board/Card.svelte";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function mount(card: CardView) {
  return render(Card as never, { card } as Record<string, unknown>);
}

const forest = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "forest",
    name: "Forest",
    owner: "them",
    controller: "them",
    type_line: "Basic Land — Island",
    ...extra,
  }) as unknown as CardView;

describe("the land-type badge", () => {
  it("names the type and explains it in the tooltip", () => {
    const { container } = mount(
      forest({
        land_type_effects: [
          { types: ["Island"], until: "until end of turn", source: "Tidal Warrior" },
        ],
      }),
    );
    const badge = container.querySelector(".badge.land-type");
    expect(badge?.textContent?.trim()).toBe("ISLAND");
    expect(badge?.getAttribute("title")).toBe("Island until end of turn — Tidal Warrior");
  });

  it("is absent on a land nothing has changed", () => {
    const { container } = mount(forest());
    expect(container.querySelector(".badge.land-type")).toBeNull();
  });
});
