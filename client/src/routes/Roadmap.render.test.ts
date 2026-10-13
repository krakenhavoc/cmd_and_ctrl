// @vitest-environment jsdom
//
// ADR 0156 §7: the roadmap's status sections start collapsed to their
// heading and count, "Up next" stays open, and a search or chip filter
// opens every section that has a match.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import Roadmap from "./Roadmap.svelte";
import { cleanup, flushSync, render } from "../lib/test/render.svelte";

const realFetch = globalThis.fetch;

function item(slug: string, name: string, status: string) {
  return { slug, name, kind: "keyword", status, summary: `${name} summary`, missing: "a gap" };
}

beforeEach(() => {
  const items = [
    item("a", "Flying", "implemented"),
    item("b", "Madness", "partial"),
    item("c", "Banding", "missing"),
  ];
  globalThis.fetch = vi.fn(
    async () =>
      new Response(
        JSON.stringify({
          counts: {
            implemented: 1,
            partial: 1,
            missing: 1,
            cards: { total: 3, full: 1, caveats: 1, unreviewed: 1 },
          },
          next: ["c"],
          items,
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
  ) as never;
});

afterEach(() => {
  cleanup();
  globalThis.fetch = realFetch;
  vi.restoreAllMocks();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

function sections(root: HTMLElement): HTMLDetailsElement[] {
  return [...root.querySelectorAll<HTMLDetailsElement>("details.status-sec")];
}

describe("the roadmap's collapsing sections", () => {
  it("starts every status section closed, with its heading and count showing", async () => {
    const { container } = render(Roadmap as never, {} as never);
    await settle();
    const secs = sections(container);
    expect(secs.map((d) => d.dataset.status)).toEqual(["implemented", "partial", "missing"]);
    for (const d of secs) {
      expect(d.open).toBe(false);
      expect(d.querySelector("summary h2")!.textContent!.replace(/\s+/g, " ")).toMatch(/\w+ \d/);
    }
  });

  it("opens a section when it is toggled", async () => {
    const { container } = render(Roadmap as never, {} as never);
    await settle();
    const d = sections(container)[1]!;
    d.open = true;
    d.dispatchEvent(new Event("toggle"));
    await settle();
    expect(sections(container)[1]!.open).toBe(true);
    expect(sections(container)[0]!.open).toBe(false);
  });

  it("opens the sections that have a match while a search is active", async () => {
    const { container } = render(Roadmap as never, {} as never);
    await settle();
    const input = container.querySelector<HTMLInputElement>("input.search")!;
    input.value = "madness";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await settle();
    const secs = sections(container);
    expect(secs).toHaveLength(1);
    expect(secs[0]!.dataset.status).toBe("partial");
    expect(secs[0]!.open).toBe(true);
  });

  it("opens the matching section when a status chip is pressed", async () => {
    const { container } = render(Roadmap as never, {} as never);
    await settle();
    container.querySelector<HTMLButtonElement>("button.fchip.missing")!.click();
    await settle();
    const secs = sections(container);
    expect(secs.map((d) => d.dataset.status)).toEqual(["missing"]);
    expect(secs[0]!.open).toBe(true);
  });
});
