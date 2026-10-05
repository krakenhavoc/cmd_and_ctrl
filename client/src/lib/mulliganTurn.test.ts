import { describe, expect, it } from "vitest";
import { mulliganDecider, mulliganWaitingText } from "./mulliganTurn";

const seat = (name: string, extra: object = {}) => ({ name, ...extra });

describe("mulligan turn order", () => {
  const seats = [seat("Ann"), seat("Bo", { mulligan_turn: true }), seat("Cy"), seat("Di")];

  it("finds the one seat whose turn it is", () => {
    expect(mulliganDecider(seats)?.name).toBe("Bo");
  });

  it("finds nobody when no seat is marked", () => {
    expect(mulliganDecider([seat("Ann"), seat("Bo")])).toBeNull();
  });

  it("never names a seat that kept or left", () => {
    expect(mulliganDecider([seat("Ann", { mulligan_turn: true, hand_kept: true })])).toBeNull();
    expect(mulliganDecider([seat("Ann", { mulligan_turn: true, eliminated: true })])).toBeNull();
  });

  it("tells a waiting seat who is deciding", () => {
    expect(mulliganWaitingText(seats, false)).toBe("Waiting for Bo to decide");
  });

  it("tells the deciding seat nothing", () => {
    expect(mulliganWaitingText(seats, true)).toBeNull();
  });

  it("says nothing when nobody is deciding", () => {
    expect(mulliganWaitingText([seat("Ann")], false)).toBeNull();
  });
});
