// #3379 — naming what moved a life total: pairing a popup's deltas
// with the seat's newest `life` entries, the damage fallback, and a
// name only from a card this viewer can see.

import { describe, expect, it } from "vitest";

import { lifeCauses, visibleCardName } from "./lifeCause";
import type { CardView, GameView, LogEvent } from "./protocol";

const card = (id: string, name: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: id, name, owner: "a", controller: "a", ...extra }) as CardView;

const zone = (cards: CardView[] = []) => ({ kind: "z", count: cards.length, cards });

function view(log: LogEvent[], opts: { field?: CardView[]; grave?: CardView[] } = {}): GameView {
  return {
    battlefield: zone(opts.field),
    stack: zone(),
    exile: zone(),
    seats: [
      { id: "a", graveyard: zone(opts.grave), command: zone(), hand: zone() },
      { id: "b", graveyard: zone(), command: zone(), hand: zone() },
    ],
    log,
  } as unknown as GameView;
}

const life = (seq: number, seat: number, amount: number, card_id?: string): LogEvent => ({
  seq,
  kind: "life",
  seat,
  amount,
  card_id,
  text: "",
});

describe("lifeCauses", () => {
  const vista = card("v", "Prismatic Vista");
  const ardent = card("w", "Weftstalker Ardent");

  it("names each delta's source, oldest first", () => {
    const v = view([life(1, 0, -1, "v"), life(2, 1, -1, "w"), life(3, 0, -2, "w")], {
      field: [ardent],
      grave: [vista],
    });
    expect(lifeCauses(v, 0, [-1, -2])).toEqual(["Prismatic Vista", "Weftstalker Ardent"]);
    expect(lifeCauses(v, 0, [-2])).toEqual(["Weftstalker Ardent"]);
    expect(lifeCauses(v, 1, [-1])).toEqual(["Weftstalker Ardent"]);
  });

  it("names nothing when the amounts do not line up", () => {
    const v = view([life(1, 0, -1, "v")], { grave: [vista] });
    expect(lifeCauses(v, 0, [-3])).toEqual([null]);
    expect(lifeCauses(v, 0, [-1, -1])).toEqual([null, null]);
    expect(lifeCauses(v, -1, [-1])).toEqual([null]);
  });

  const damage = (seq: number, target_seat: number, amount: number, card_id: string): LogEvent => ({
    seq,
    kind: "damage",
    seat: 1,
    target_seat,
    amount,
    card_id,
    text: "",
  });

  it("counts damage to the player, which moves life with no life entry", () => {
    const v = view([damage(1, 0, 3, "w"), damage(2, 1, 3, "w")], { field: [ardent] });
    expect(lifeCauses(v, 0, [-3])).toEqual(["Weftstalker Ardent"]);
  });

  it("does not credit an older life entry with newer damage of the same size", () => {
    // Sign in Blood cost 2 life earlier; now Purphoros deals 2.
    const v = view([life(1, 0, -2, "sib"), damage(2, 0, 2, "p"), damage(3, 1, 2, "p")], {
      field: [card("p", "Purphoros, God of the Forge")],
      grave: [card("sib", "Sign in Blood")],
    });
    expect(lifeCauses(v, 0, [-2])).toEqual(["Purphoros, God of the Forge"]);
    expect(lifeCauses(v, 0, [-2, -2])).toEqual(["Sign in Blood", "Purphoros, God of the Forge"]);
  });

  it("ignores damage to permanents", () => {
    const v = view(
      [
        life(1, 0, -1, "v"),
        { seq: 2, kind: "damage", seat: 1, target: "x", amount: 4, card_id: "w", text: "" },
      ],
      { grave: [vista], field: [ardent] },
    );
    expect(lifeCauses(v, 0, [-1])).toEqual(["Prismatic Vista"]);
  });

  it("leaves a source it cannot see unnamed", () => {
    const v = view([life(1, 0, 2, "hidden"), life(2, 0, -1, "fd")], {
      field: [card("fd", "Secret Sphinx", { face_down: true })],
    });
    expect(lifeCauses(v, 0, [2, -1])).toEqual([null, null]);
  });
});

describe("visibleCardName", () => {
  it("finds a card in any zone the frame carries", () => {
    const v = view([], { grave: [card("g", "Grizzly Bears")] });
    expect(visibleCardName(v, "g")).toBe("Grizzly Bears");
    expect(visibleCardName(v, "nope")).toBeNull();
  });
});
