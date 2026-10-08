// @vitest-environment jsdom
//
// #2630: the lobby's create row. A blank name must be creatable (the
// server names the table), and the dice fills the box.

import { describe, it, expect, afterEach, vi } from "vitest";
import { flushSync } from "svelte";

import NewTableRow from "./components/NewTableRow.svelte";
import { L } from "./labels";
import { click, cleanup, render } from "./test/render.svelte";

afterEach(cleanup);

function mount(props: Record<string, unknown>) {
  return render(NewTableRow as never, { name: "", ...props });
}

const input = (c: HTMLElement) => c.querySelector('input[type="text"]') as HTMLInputElement;
const dice = (c: HTMLElement) =>
  c.querySelector(`[aria-label="${L.suggestTableName}"]`) as HTMLButtonElement;
const create = (c: HTMLElement) => c.querySelector('button[type="submit"]') as HTMLButtonElement;

describe("the create-a-table row", () => {
  it("lets an empty box create, and says a blank name is fine", () => {
    const { container } = mount({});
    expect(create(container).disabled).toBe(false);
    expect(input(container).placeholder).toContain("blank");
  });

  it("still disables create while a create is in flight", () => {
    const { container } = mount({ busy: true });
    expect(create(container).disabled).toBe(true);
  });

  it("fills the box from the dice, and rerolls", async () => {
    const suggest = vi
      .fn()
      .mockResolvedValueOnce("Six Untapped Islands")
      .mockResolvedValueOnce("Goblins Behaving Badly");
    const r = mount({ suggest });
    click(dice(r.container));
    await vi.waitFor(() => expect(input(r.container).value).toBe("Six Untapped Islands"));
    click(dice(r.container));
    await vi.waitFor(() => expect(input(r.container).value).toBe("Goblins Behaving Badly"));
    expect(suggest).toHaveBeenCalledTimes(2);
  });

  it("keeps what was typed when a suggestion fails", async () => {
    const suggest = vi.fn().mockRejectedValue(new Error("429"));
    const r = mount({ suggest, name: "Friday" });
    click(dice(r.container));
    await vi.waitFor(() => expect(suggest).toHaveBeenCalled());
    await vi.waitFor(() => {
      flushSync();
      expect(dice(r.container).disabled).toBe(false);
    });
    expect(input(r.container).value).toBe("Friday");
  });
});
