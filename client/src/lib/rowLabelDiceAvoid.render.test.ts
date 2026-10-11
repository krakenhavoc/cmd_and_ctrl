// @vitest-environment jsdom
//
// #2995: the opening roll's d20 sat over a row's label ("D20 ures").
// DiceLayer keeps dice off whatever carries data-dice-avoid, so a
// battlefield row's label has to carry it.

import { afterEach, describe, expect, it } from "vitest";

import BattlefieldRow from "./components/board/BattlefieldRow.svelte";
import { cleanup, render } from "./test/render.svelte";

afterEach(cleanup);

describe("a battlefield row's label", () => {
  it("is something the dice stay off", () => {
    const { container } = render(BattlefieldRow as never, {
      label: "creatures",
      cards: [],
      viewerID: null,
    });
    const label = container.querySelector(".row-label");
    expect(label).not.toBeNull();
    expect(label?.hasAttribute("data-dice-avoid")).toBe(true);
  });
});
