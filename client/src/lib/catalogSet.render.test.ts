// @vitest-environment jsdom
//
// #4790: the catalogue's set picker narrows the grid to one set and states
// how much of that set the game plays. The filter and the summary line have
// unit tests in catalog.test.ts, and so does the picker's search
// (searchSets); this pins that the page wires them: the picker is an
// accessible combobox, typing a name or a set code narrows its list,
// picking a set shows only its cards, and the line under the bar says the
// counts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import Catalog from "../routes/Catalog.svelte";
import { L } from "./labels";
import { cleanup, flushSync, render } from "./test/render.svelte";

const realFetch = globalThis.fetch;
const realMatchMedia = window.matchMedia;

const data = {
  counts: { total: 2, full: 2, caveats: 0, unreviewed: 0, missing: 0 },
  cards: [
    { oracle_id: "a", name: "Web Slinger", completeness: "full", sets: ["spm"] },
    { oracle_id: "b", name: "Old Bear", completeness: "full", sets: ["old"] },
  ],
  sets: [
    {
      code: "spm",
      name: "Marvel's Spider-Man",
      released: "2026-09-26",
      total: 3,
      full: 1,
      caveats: 0,
      unreviewed: 0,
      unsupported: 2,
    },
    {
      code: "hml",
      name: "Homelands",
      released: "1995-10-01",
      total: 2,
      full: 0,
      caveats: 0,
      unreviewed: 0,
      unsupported: 2,
    },
    {
      code: "old",
      name: "Old Set",
      released: "2001-01-01",
      total: 1,
      full: 1,
      caveats: 0,
      unreviewed: 0,
      unsupported: 0,
    },
  ],
};

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

beforeEach(() => {
  window.matchMedia = ((q: string) => ({
    matches: false,
    media: q,
    addEventListener: () => {},
    removeEventListener: () => {},
  })) as never;
  globalThis.fetch = vi.fn(
    async () =>
      new Response(JSON.stringify(data), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
  ) as never;
});
afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  window.matchMedia = realMatchMedia;
});

function picker(container: HTMLElement): HTMLInputElement {
  const input = container.querySelector<HTMLInputElement>(
    `input[role="combobox"][aria-label="${L.filterCatalogueBySet}"]`,
  );
  expect(input, "the picker is drawn").not.toBeNull();
  return input!;
}

function listbox(container: HTMLElement, input: HTMLInputElement): HTMLElement {
  const id = input.getAttribute("aria-controls");
  expect(id, "the combobox names its listbox").toBeTruthy();
  const list = document.getElementById(id!);
  expect(list?.getAttribute("role")).toBe("listbox");
  expect(container.contains(list), "the listbox is the page's own").toBe(true);
  return list!;
}

function optionNames(list: HTMLElement): string[] {
  return [...list.querySelectorAll('[role="option"]')].map(
    (o) => o.querySelector(".name")?.textContent ?? "",
  );
}

function type(input: HTMLInputElement, text: string): void {
  input.value = text;
  input.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

function key(input: HTMLInputElement, k: string): void {
  input.dispatchEvent(new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true }));
  flushSync();
}

describe("the catalogue's set picker", () => {
  it("opens on every set newest first, with All sets, and is closed until asked", async () => {
    const { container } = render(Catalog as never, {} as never);
    await settle();

    const input = picker(container);
    const list = listbox(container, input);
    expect(input.getAttribute("aria-expanded")).toBe("false");
    expect(list.hidden).toBe(true);

    input.focus();
    key(input, "ArrowDown");
    expect(input.getAttribute("aria-expanded")).toBe("true");
    expect(list.hidden).toBe(false);
    expect(optionNames(list)).toEqual(["All sets", "Marvel's Spider-Man", "Homelands", "Old Set"]);
    // The supported count the old option label carried is still there.
    const spm = [...list.querySelectorAll('[role="option"]')][1];
    expect(spm.getAttribute("aria-label")).toBe("Marvel's Spider-Man (1/3), SPM, 2026");
    expect(spm.textContent).toContain("SPM");
    expect(spm.textContent).toContain("2026");
    expect(spm.textContent).toContain("1/3");
    // Nothing picked: All sets is the selected option.
    expect(list.querySelector('[aria-selected="true"]')?.textContent?.trim()).toBe("All sets");

    key(input, "Escape");
    expect(input.getAttribute("aria-expanded")).toBe("false");
  });

  it("narrows the list by set code or name as you type", async () => {
    const { container } = render(Catalog as never, {} as never);
    await settle();
    const input = picker(container);
    const list = listbox(container, input);
    input.focus();

    type(input, "hml");
    expect(optionNames(list)).toEqual(["Homelands"]);
    type(input, "SPM");
    expect(optionNames(list)).toEqual(["Marvel's Spider-Man"]);
    type(input, "set");
    expect(optionNames(list)).toEqual(["Old Set"]);
    type(input, "zzz");
    expect(optionNames(list)).toEqual([]);
    expect(container.textContent).toContain("No set matches “zzz”.");
  });

  it("picks the first match with Enter and filters the grid to it", async () => {
    const { container } = render(Catalog as never, {} as never);
    await settle();
    const input = picker(container);
    const list = listbox(container, input);
    input.focus();

    expect(container.textContent).toContain("Web Slinger");
    expect(container.textContent).toContain("Old Bear");
    expect(container.textContent).not.toContain("not yet supported");

    type(input, "spi");
    const first = list.querySelector('[role="option"]')!;
    expect(input.getAttribute("aria-activedescendant")).toBe(first.id);
    key(input, "Enter");
    await settle();

    expect(input.getAttribute("aria-expanded")).toBe("false");
    expect(input.value).toBe("Marvel's Spider-Man");
    expect(container.textContent).toContain("Web Slinger");
    expect(container.textContent).not.toContain("Old Bear");
    expect(container.textContent).toContain(
      "Marvel's Spider-Man: 1 of 3 supported — 1 complete; 2 not yet supported",
    );
  });

  it("picks with the arrow keys and with a tap, and the clear button goes back to all sets", async () => {
    const { container } = render(Catalog as never, {} as never);
    await settle();
    const input = picker(container);
    const list = listbox(container, input);
    input.focus();

    // All sets, Marvel's Spider-Man, Homelands, Old Set: two down from All sets.
    key(input, "ArrowDown");
    key(input, "ArrowDown");
    key(input, "ArrowDown");
    key(input, "Enter");
    await settle();
    expect(input.value).toBe("Homelands");
    expect(container.textContent).not.toContain("Web Slinger");
    expect(container.textContent).toContain("Homelands: none of its 2 cards supported yet.");

    // A tap: the option keeps the box focused on press, then picks on click.
    input.click();
    flushSync();
    const old = [...list.querySelectorAll<HTMLElement>('[role="option"]')].find(
      (o) => o.querySelector(".name")?.textContent === "Old Set",
    )!;
    const down = new MouseEvent("mousedown", { bubbles: true, cancelable: true });
    old.dispatchEvent(down);
    expect(down.defaultPrevented, "pressing an option does not blur the box").toBe(true);
    old.click();
    await settle();
    expect(input.value).toBe("Old Set");
    expect(container.textContent).toContain("Old Bear");
    expect(container.textContent).not.toContain("Web Slinger");

    const clear = container.querySelector<HTMLButtonElement>(
      `button[aria-label="${L.clearCatalogueSet}"]`,
    );
    expect(clear, "a picked set can be cleared").not.toBeNull();
    clear!.click();
    await settle();
    expect(input.value).toBe("");
    expect(container.textContent).toContain("Web Slinger");
    expect(container.textContent).toContain("Old Bear");
  });

  it("follows the filter when the set chip is removed", async () => {
    const { container } = render(Catalog as never, {} as never);
    await settle();
    const input = picker(container);
    input.focus();
    type(input, "hml");
    key(input, "Enter");
    await settle();
    expect(input.value).toBe("Homelands");

    const chip = [...container.querySelectorAll<HTMLButtonElement>(".achip")].find((b) =>
      b.textContent?.includes("Homelands"),
    );
    expect(chip, "the set shows as a chip").toBeDefined();
    chip!.click();
    await settle();
    expect(input.value).toBe("");
    expect(container.textContent).toContain("Web Slinger");
  });

  it("is not drawn when the server sent no sets", async () => {
    globalThis.fetch = vi.fn(
      async () =>
        new Response(JSON.stringify({ counts: data.counts, cards: data.cards }), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        }),
    ) as never;
    const { container } = render(Catalog as never, {} as never);
    await settle();
    expect(container.querySelector(`[aria-label="${L.filterCatalogueBySet}"]`)).toBeNull();
    expect(container.textContent).toContain("Web Slinger");
  });
});
