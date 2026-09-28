// @vitest-environment jsdom
//
// #1563 — DivideDamageModal, the per-target number picker a divided
// step opens once two or more targets are picked. What only the markup
// can say: it opens on the even split, Confirm is refused until every
// target has at least 1 and the shares add up, and confirming sends the
// division the player set.

import { afterEach, describe, expect, it } from "vitest";

import DivideDamageModal from "./components/board/DivideDamageModal.svelte";
import { cleanup, click, render } from "./test/render.svelte";

afterEach(cleanup);

function mount(total: number) {
  const confirmed: Array<Record<string, number>> = [];
  const r = render(
    DivideDamageModal as never,
    {
      sourceName: "Fury",
      targets: [
        { id: "a", name: "Ox" },
        { id: "b", name: "Elk" },
      ],
      total,
      onConfirm: (d: Record<string, number>) => confirmed.push(d),
      onCancel: () => {},
    } as never,
  );
  return { container: r.container, confirmed };
}

const text = (el: Element | null | undefined): string =>
  el?.textContent?.replace(/\s+/g, " ").trim() ?? "";
const shares = (c: HTMLElement): string[] =>
  [...c.querySelectorAll("output.count")].map((o) => text(o));
const button = (c: HTMLElement, label: string): HTMLButtonElement =>
  c.querySelector<HTMLButtonElement>(`button[aria-label="${label}"]`)!;
const confirmButton = (c: HTMLElement): HTMLButtonElement =>
  c.querySelector<HTMLButtonElement>("button.primary")!;

describe("DivideDamageModal (#1563)", () => {
  it("opens on the even split, remainder first", () => {
    const { container } = mount(5);
    expect(text(container.querySelector("h2"))).toContain("Divide 5");
    expect(shares(container)).toEqual(["3", "2"]);
    expect(confirmButton(container).disabled).toBe(false);
  });

  it("gates Confirm on the sum and moves points between targets", () => {
    const { container, confirmed } = mount(4);
    expect(shares(container)).toEqual(["2", "2"]);
    click(button(container, "one less to Elk"));
    expect(shares(container)).toEqual(["2", "1"]);
    expect(confirmButton(container).disabled).toBe(true);
    expect(text(container.querySelector(".status"))).toContain("assign 4");
    // A target never drops below 1.
    expect(button(container, "one less to Elk").disabled).toBe(true);
    click(button(container, "one more to Ox"));
    expect(shares(container)).toEqual(["3", "1"]);
    expect(confirmButton(container).disabled).toBe(false);
    click(confirmButton(container));
    expect(confirmed).toEqual([{ a: 3, b: 1 }]);
  });

  it("will not assign past the amount", () => {
    const { container } = mount(2);
    expect(shares(container)).toEqual(["1", "1"]);
    expect(button(container, "one more to Ox").disabled).toBe(true);
  });
});
