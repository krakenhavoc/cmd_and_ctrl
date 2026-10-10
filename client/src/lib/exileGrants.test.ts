import { describe, expect, it } from "vitest";
import { grantHolderLine, grantWindowText, heldPlayableCount } from "./exileGrants";
import type { CardView, ExilePlayView } from "./protocol";

// exileGrants.test.ts — #2559: Memory Vessel and Rocco, Street Chef give
// every player a grant over their own exiled cards, so the table reads
// each other's grants and their windows.

const names: Record<string, string> = { ana: "Ana", bo: "Bo" };
const nameOf = (id: string) => names[id];

function card(grant?: ExilePlayView, id = "c1"): CardView {
  return {
    instance_id: id,
    name: "Forest",
    owner: "ana",
    controller: "ana",
    exile_play: grant,
  } as CardView;
}

describe("grant windows", () => {
  it("names the player the window is counted against", () => {
    expect(
      grantWindowText({ player: "ana", until: "next_turn", until_player: "bo" }, nameOf, "ana"),
    ).toBe("until Bo's next turn");
    expect(
      grantWindowText({ player: "ana", until: "next_turn", until_player: "ana" }, nameOf, "ana"),
    ).toBe("until your next turn");
    expect(
      grantWindowText({ player: "ana", until: "next_end_step", until_player: "bo" }, nameOf, null),
    ).toBe("until Bo's next end step");
    expect(
      grantWindowText(
        { player: "ana", until: "end_of_next_turn", until_player: "ana" },
        nameOf,
        "bo",
      ),
    ).toBe("until the end of Ana's next turn");
  });

  it("falls back to end of turn for a grant with no window", () => {
    expect(grantWindowText({ player: "ana" }, nameOf, "ana")).toBe("until end of turn");
    expect(grantWindowText(null, nameOf, "ana")).toBe("until end of turn");
    expect(grantWindowText({ player: "ana", until: "while_exiled" }, nameOf, "ana")).toBe(
      "while it stays exiled",
    );
  });
});

describe("another player's grant", () => {
  const vessel: ExilePlayView = { player: "ana", until: "next_turn", until_player: "bo" };

  it("labels a card someone else may play", () => {
    expect(grantHolderLine(card(vessel), nameOf, "bo")).toBe(
      "Ana may play it until your next turn",
    );
    expect(grantHolderLine(card({ ...vessel, cast_only: true }), nameOf, "zed")).toBe(
      "Ana may cast it until Bo's next turn",
    );
  });

  it("says nothing for the viewer's own grant or no grant", () => {
    expect(grantHolderLine(card(vessel), nameOf, "ana")).toBeNull();
    expect(grantHolderLine(card(), nameOf, "bo")).toBeNull();
  });

  it("counts the cards a seat holds a grant over", () => {
    const pile = [
      card(vessel, "a"),
      card(vessel, "b"),
      card(undefined, "c"),
      card({ player: "bo" }, "d"),
    ];
    expect(heldPlayableCount(pile, "ana")).toBe(2);
    expect(heldPlayableCount(pile, "bo")).toBe(1);
    expect(heldPlayableCount(undefined, "ana")).toBe(0);
  });
});
