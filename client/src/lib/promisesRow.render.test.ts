// @vitest-environment jsdom
//
// promisesRow.render.test.ts — #2483. Promise tokens stay off the table
// while nothing is owed either way: the one control is "+ Promise",
// which records a promise you owe this player. Once either count is
// above zero, the counts are named in words, not arrows.

import { describe, it, expect, afterEach } from "vitest";
import PromisesRow from "./components/board/PromisesRow.svelte";
import type { GameView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

afterEach(cleanup);

function mount(promises: Record<string, number>) {
  const sent: unknown[][] = [];
  const c = render(
    PromisesRow as never,
    {
      view: { promises } as unknown as GameView,
      viewerID: "me",
      opponentID: "them",
      sendAction: (...a: unknown[]) => sent.push(a),
    } as never,
  );
  return { el: c.container, sent };
}

describe("PromisesRow (#2483)", () => {
  it("shows only + Promise while nothing is owed, and it records one you owe", () => {
    const { el, sent } = mount({});
    const buttons = [...el.querySelectorAll("button")];
    expect(buttons.map((b) => b.textContent?.trim())).toEqual(["+ Promise"]);
    expect(el.textContent).not.toContain("Owes you");
    buttons[0].click();
    expect(sent).toEqual([["set_promise", { from: "me", to: "them", count: 1 }]]);
  });

  it("names both counts in words once something is owed", () => {
    const { el } = mount({ "them->me": 2, "me->them": 1 });
    expect(el.textContent).toContain("Owes you");
    expect(el.textContent).toContain("You owe");
    expect(el.querySelector(".seg.owed .count")?.textContent).toBe("2");
    expect(el.querySelector(".seg.owe .count")?.textContent).toBe("1");
  });

  it("leaves out Owes you when only you owe", () => {
    const { el } = mount({ "me->them": 1 });
    expect(el.textContent).not.toContain("Owes you");
    expect(el.querySelector('[aria-label="decrement promise"]')).not.toBeNull();
  });
});
