import { describe, expect, it } from "vitest";
import type { CardView, PlayerView } from "./protocol";
import { takenFromByCard } from "./takenFrom";

function card(id: string, owner: string, controller: string): CardView {
  return { instance_id: id, owner, controller } as unknown as CardView;
}

const seats = [
  { id: "a", name: "Alice" },
  { id: "b", name: "Bob" },
] as unknown as PlayerView[];

describe("takenFromByCard (ADR 0104)", () => {
  it("names the owner of a permanent someone else controls", () => {
    const out = takenFromByCard([card("x", "a", "b"), card("y", "b", "b")], seats);
    expect(out).toEqual({ x: "Alice" });
  });

  it("falls back to a neutral name for an owner the view does not seat", () => {
    expect(takenFromByCard([card("x", "gone", "b")], seats)).toEqual({ x: "another player" });
  });

  it("says nothing for a card with no owner or controller on the wire", () => {
    expect(takenFromByCard([card("x", "", "b"), card("y", "a", "")], seats)).toEqual({});
  });

  it("tolerates an empty board", () => {
    expect(takenFromByCard(undefined, undefined)).toEqual({});
  });
});
