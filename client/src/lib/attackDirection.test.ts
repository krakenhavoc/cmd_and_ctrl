import { describe, it, expect } from "vitest";

import { attackDirectionLines } from "./attackDirection";
import type { AttackDirectionView, PlayerView } from "./protocol";

// attackDirection.test.ts — ADR 0206 (#4211): the dock's line for each
// "may attack only the nearest opponent that way" rule in force.

const seats = [
  { id: "a", name: "Alice", seat: 0 },
  { id: "b", name: "Bob", seat: 1 },
  { id: "c", name: "Carol", seat: 2 },
  { id: "d", name: "Dave", seat: 3 },
] as unknown as PlayerView[];

const left: AttackDirectionView = {
  direction: "left",
  source: "Pramikon, Sky Rampart",
  nearest: { a: "b", b: "c", c: "d", d: "a" },
};
const right: AttackDirectionView = {
  direction: "right",
  source: "Teyo, Geometric Tactician",
  nearest: { a: "d", b: "a", c: "b", d: "c" },
};

describe("attackDirectionLines", () => {
  it("says nothing while no rule is in force", () => {
    expect(attackDirectionLines(undefined, seats, "a")).toEqual([]);
    expect(attackDirectionLines([], seats, "a")).toEqual([]);
  });

  it("names the direction, who the viewer may attack, and the source", () => {
    const [line] = attackDirectionLines([left], seats, "a");
    expect(line.text).toBe("Attacks go left: you may attack only Bob (Pramikon, Sky Rampart)");
    expect(line.direction).toBe("left");
    expect(line.title).toMatch(/nearest opponent to the left \(the next player in turn order\)/);
    expect(line.title).toMatch(/Pramikon, Sky Rampart/);
    expect(line.title).toMatch(/You may attack only Bob and planeswalkers Bob controls/);
  });

  it("reads right as the player before in turn order", () => {
    const [line] = attackDirectionLines([right], seats, "a");
    expect(line.text).toBe(
      "Attacks go right: you may attack only Dave (Teyo, Geometric Tactician)",
    );
    expect(line.title).toMatch(/the player before in turn order/);
  });

  it("tells a spectator the direction and source only", () => {
    const [line] = attackDirectionLines([left], seats, null);
    expect(line.text).toBe("Attacks go left (Pramikon, Sky Rampart)");
    expect(line.title).not.toMatch(/You may attack/);
  });

  it("says so when the viewer has no opponent that way", () => {
    const [line] = attackDirectionLines([{ ...left, nearest: {} }], seats, "a");
    expect(line.text).toBe("Attacks go left: you have no one to attack (Pramikon, Sky Rampart)");
  });

  it("lists every rule, and warns when two disagree", () => {
    const lines = attackDirectionLines([left, right], seats, "a");
    expect(lines.map((l) => l.direction)).toEqual(["left", "right", "conflict"]);
    expect(lines[2].text).toBe("These rules disagree: no one can attack");
  });

  it("does not warn about two rules that agree", () => {
    const lines = attackDirectionLines([left, { ...left, source: "Pramikon (copy)" }], seats, "a");
    expect(lines.map((l) => l.direction)).toEqual(["left", "left"]);
  });

  it("does not warn when two rules name the same opponent (two players left)", () => {
    const two = { a: "b", b: "a" };
    const lines = attackDirectionLines(
      [
        { direction: "left", source: "X", nearest: two },
        { direction: "right", source: "Y", nearest: two },
      ],
      seats,
      "a",
    );
    expect(lines.map((l) => l.direction)).toEqual(["left", "right"]);
  });

  it("leaves out a direction this build does not know", () => {
    expect(attackDirectionLines([{ direction: "up", source: "X" }], seats, "a")).toEqual([]);
  });
});
