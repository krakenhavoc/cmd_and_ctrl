// @vitest-environment jsdom
//
// targetingBanner.render.test.ts — #1211. The banner's sentence is the
// only thing that tells a player WHICH half of the stack they may
// click: an ability row and a spell row are drawn identically in the
// stack overlay, and the legal set (which decides legality) is
// invisible until a click is refused.
//
// So the three stack modes each get their own line, and this is the
// test that they reach the DOM rather than falling through to the
// "a target" default — which is what a mode the switch does not know
// produces, silently.

import { describe, it, expect, afterEach } from "vitest";

import TargetingBanner from "./components/board/TargetingBanner.svelte";
import { begin, cancel, targeting } from "./targeting";
import type { CardView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(() => {
  targeting.set(null);
  cleanup();
});

const card = (name: string): CardView => ({
  instance_id: name.toLowerCase(),
  name,
  owner: "me",
  controller: "me",
});

describe("the targeting banner names the stack half it is asking about", () => {
  it("a spell, an ability and either", () => {
    for (const [mode, sentence] of [
      ["stack_spell", "a spell on the stack"],
      ["stack_ability", "an ability on the stack"],
      ["stack_item", "a spell or ability on the stack"],
    ] as const) {
      begin(card("Stifle"), mode);
      const { container } = render(TargetingBanner, {});
      expect(container.textContent).toContain(sentence);
      // The default arm would say this instead, and it is what an
      // unknown mode silently produces.
      if (mode !== "stack_spell") {
        expect(container.textContent).not.toContain("Click a target to target");
      }
      cancel();
      cleanup();
    }
  });
});

// #1659: the banner used to say nothing about the amount being
// divided until the picks closed and DivideDamageModal opened. Now it
// names the amount, and the target count, while the player is still
// picking.
describe("#1659 — the targeting banner shows the divided amount", () => {
  it("a fixed divide names the amount and the target ceiling", () => {
    const arcLightning: CardView = {
      ...card("Arc Lightning"),
      legal_targets: { min: 1, max: 3, divide: { total: 3 } },
    };
    begin(arcLightning, "any");

    const { container } = render(TargetingBanner, {});

    expect(container.textContent).toContain("divide 3 damage among up to 3 targets");
  });

  it("an X-based divide shows the announced X once it is known", () => {
    const rollingThunder: CardView = {
      ...card("Rolling Thunder"),
      legal_targets: { min: 0, max: 0, divide: { from_x: true } },
    };
    begin(rollingThunder, "any", { xValue: 5 });

    const { container } = render(TargetingBanner, {});

    expect(container.textContent).toContain("divide 5 damage among any number of targets");
  });

  it("an X-based divide falls back to 'X' rather than inventing 0", () => {
    const rollingThunder: CardView = {
      ...card("Rolling Thunder"),
      legal_targets: { min: 0, max: 0, divide: { from_x: true } },
    };
    // No xValue in the choices — the walk doesn't know X yet.
    begin(rollingThunder, "any");

    const { container } = render(TargetingBanner, {});

    expect(container.textContent).toContain("divide X damage among any number of targets");
    expect(container.textContent).not.toContain("divide 0 damage");
  });

  it("a clause with no divide shows nothing about dividing", () => {
    begin(card("Lightning Bolt"), "any");

    const { container } = render(TargetingBanner, {});

    expect(container.textContent).not.toContain("divide");
  });
});
