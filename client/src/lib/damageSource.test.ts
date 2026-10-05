import { describe, expect, it } from "vitest";
import { permanentWhoseCaption } from "./damageSource";
import type { CardView, GameView } from "./protocol";

const land = (id: string, controller: string) =>
  ({ instance_id: id, name: "Forest", controller }) as unknown as CardView;

describe("permanentWhoseCaption (#1960)", () => {
  const mine = land("a", "me");
  const theirs = land("b", "opp");
  const snap = {
    battlefield: { cards: [mine, theirs] },
    seats: [
      { id: "me", name: "Me" },
      { id: "opp", name: "Sly" },
    ],
  } as unknown as GameView;

  it("says Yours for the viewer's own permanent", () => {
    expect(permanentWhoseCaption(snap, mine, "me")).toBe("Yours");
  });
  it("names an opponent's seat", () => {
    expect(permanentWhoseCaption(snap, theirs, "me")).toBe("Sly's");
  });
  it("is empty off the battlefield", () => {
    expect(permanentWhoseCaption(snap, land("z", "opp"), "me")).toBe("");
  });
});
