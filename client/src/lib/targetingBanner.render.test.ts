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
