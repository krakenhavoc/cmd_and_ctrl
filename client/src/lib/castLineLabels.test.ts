import { describe, expect, it } from "vitest";
import type { CardView, GameView, LegalMoveView } from "./protocol";
import { cardQualifier, distinctCastLabels, targetCardIDs } from "./castLineLabels";

const card = (id: string, controller: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: id, name: "Grizzly Bears", owner: controller, controller, ...extra }) as CardView;

const view = (cards: CardView[]): GameView =>
  ({
    seats: [
      { id: "me", name: "Me" },
      { id: "bob", name: "Bob" },
    ],
    battlefield: { kind: "battlefield", count: cards.length, cards },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
  }) as unknown as GameView;

const move = (label: string, targets: { kind: string; id: string }[]): LegalMoveView =>
  ({ type: "cast_spell", player: "me", kind: "cast", label, params: { targets } }) as LegalMoveView;

const label = (m: LegalMoveView) => m.label;

describe("castLineLabels (#4709)", () => {
  it("reads card targets only, in order", () => {
    expect(
      targetCardIDs(
        move("x", [
          { kind: "player", id: "bob" },
          { kind: "card", id: "a" },
          { kind: "card", id: "b" },
        ]),
      ),
    ).toEqual(["a", "b"]);
    expect(targetCardIDs({ label: "x" } as LegalMoveView)).toEqual([]);
  });

  it("names a token and someone else's card, and nothing for your own", () => {
    const v = view([]);
    expect(cardQualifier(card("a", "me"), v, "me")).toBe("");
    expect(cardQualifier(card("a", "me", { is_token: true }), v, "me")).toBe("token");
    expect(cardQualifier(card("a", "bob", { is_token: true }), v, "me")).toBe("token, Bob's");
  });

  it("leaves lines that already differ alone", () => {
    const v = view([card("a", "me"), card("b", "bob")]);
    const ms = [
      move("targeting Grizzly Bears", [{ kind: "card", id: "a" }]),
      move("targeting Bob", []),
    ];
    expect(distinctCastLabels(ms, label, v, "me")).toEqual([
      "targeting Grizzly Bears",
      "targeting Bob",
    ]);
  });

  it("qualifies alike lines, and numbers what is still alike", () => {
    const v = view([card("a", "me"), card("b", "me"), card("t", "me", { is_token: true })]);
    const ms = ["a", "b", "t"].map((id) => move("targeting Grizzly Bears", [{ kind: "card", id }]));
    expect(distinctCastLabels(ms, label, v, "me")).toEqual([
      "targeting Grizzly Bears #1",
      "targeting Grizzly Bears #2",
      "targeting Grizzly Bears (token)",
    ]);
  });

  it("marks each target of a two-target line", () => {
    const v = view([card("a", "me"), card("t", "bob", { is_token: true })]);
    const ms = [
      move("targeting Grizzly Bears and Grizzly Bears", [
        { kind: "card", id: "a" },
        { kind: "card", id: "t" },
      ]),
      move("targeting Grizzly Bears and Grizzly Bears", [
        { kind: "card", id: "t" },
        { kind: "card", id: "a" },
      ]),
    ];
    expect(distinctCastLabels(ms, label, v, "me")).toEqual([
      "targeting Grizzly Bears and Grizzly Bears (yours; token, Bob's)",
      "targeting Grizzly Bears and Grizzly Bears (token, Bob's; yours)",
    ]);
  });
});
