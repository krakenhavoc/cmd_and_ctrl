import { describe, it, expect } from "vitest";

import { kicksOf, stepsFor } from "./targeting";
import type { CardView, LegalTargetsView } from "./protocol";

// #6150: "choose any target, then choose another target for each time
// this spell was kicked" (Comet Storm). The clause ships min / max 1 and
// count_from_kicks; the picker asks for one plus the kicks announced.
describe("count_from_kicks", () => {
  const card = {
    instance_id: "comet",
    name: "Comet Storm",
    optional_costs: [{ index: 0, key: "multikicker", max_times: 20 }],
  } as unknown as CardView;
  const clause: LegalTargetsView = {
    min: 1,
    max: 1,
    count_from_kicks: true,
    players: ["p1", "p2"],
    cards: ["c1"],
  } as LegalTargetsView;

  it("counts kicker and multikicker payments", () => {
    expect(kicksOf(card, [])).toBe(0);
    expect(kicksOf(card, [0, 0])).toBe(2);
  });

  it("asks for one target unkicked", () => {
    const [step] = stepsFor("any", clause, undefined, -1, { optionalCosts: [] });
    expect(step.min).toBe(1);
    expect(step.max).toBe(1);
  });

  it("asks for one plus the kicks", () => {
    const [step] = stepsFor("any", clause, undefined, -1, { optionalCosts: [0, 0], kicks: 2 });
    expect(step.min).toBe(3);
    expect(step.max).toBe(3);
  });
});
