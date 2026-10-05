import { describe, expect, it } from "vitest";
import { damageBadge } from "./damageBadge";

describe("damageBadge", () => {
  it("is null for a permanent with no damage", () => {
    expect(damageBadge({})).toBeNull();
    expect(damageBadge({ damage_marked: 0, toughness: 4 })).toBeNull();
  });

  it("is the plain count on a creature that isn't indestructible", () => {
    expect(damageBadge({ damage_marked: 2, toughness: 4 })).toEqual({
      text: "2",
      title: "2 damage marked",
      summary: "2 damage",
      lethal: false,
      survives: false,
    });
  });

  // #2257: the reported board. A 5/4 Solphim with an indestructible
  // counter, four damage from a deathtouch blocker.
  it("says why lethal damage didn't destroy an indestructible creature", () => {
    const badge = damageBadge({
      damage_marked: 4,
      toughness: 4,
      abilities: ["indestructible"],
      counters: { indestructible: 1 },
    });
    expect(badge).toEqual({
      text: "4",
      title:
        "4 damage marked, enough to kill a creature with toughness 4, but it has indestructible " +
        "(from an indestructible counter). Lethal damage and deathtouch don't destroy it. " +
        "The damage wears off at end of turn.",
      summary: "4 damage: lethal, but indestructible",
      lethal: true,
      survives: true,
    });
  });

  it("covers deathtouch damage below toughness on an indestructible creature", () => {
    const badge = damageBadge({ damage_marked: 1, toughness: 4, abilities: ["indestructible"] });
    expect(badge?.survives).toBe(true);
    expect(badge?.lethal).toBe(false);
    expect(badge?.title).toBe(
      "1 damage marked. It has indestructible, so lethal damage and deathtouch don't destroy it.",
    );
    expect(badge?.summary).toBe("1 damage: indestructible");
  });

  it("names the counter only when the keyword comes from one", () => {
    const granted = damageBadge({ damage_marked: 5, toughness: 5, abilities: ["indestructible"] });
    expect(granted?.title).not.toContain("counter");
    const counter = damageBadge({
      damage_marked: 5,
      toughness: 5,
      abilities: ["indestructible"],
      counters: { indestructible: 2 },
    });
    expect(counter?.title).toContain("(from an indestructible counter)");
  });

  it("does not call damage lethal on a creature with no toughness on the wire", () => {
    const badge = damageBadge({ damage_marked: 3, abilities: ["indestructible"] });
    expect(badge?.lethal).toBe(false);
  });
});
