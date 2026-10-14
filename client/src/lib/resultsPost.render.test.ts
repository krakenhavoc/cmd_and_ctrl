// @vitest-environment jsdom
//
// ADR 0174: the results section of the lobby's table settings.

import { afterEach, describe, expect, it, vi } from "vitest";

import type { GameMeta } from "./api";
import ResultsPost from "./components/ResultsPost.svelte";
import { cleanup, render } from "./test/render.svelte";

function table(over: Partial<GameMeta> = {}): GameMeta {
  return {
    id: "g1",
    name: "Friday pod",
    created_at: "2026-10-14T20:00:00Z",
    state: "lobby",
    players: [{ player_id: "p1", name: "Gus", seat: 0, deck_uploaded: false, is_host: true }],
    ...over,
  } as GameMeta;
}

function mount(game: GameMeta, over: Record<string, unknown> = {}) {
  const ontoggle = vi.fn();
  const { container } = render(
    ResultsPost as never,
    { game, canManage: true, ontoggle, ...over } as never,
  );
  return { container, ontoggle };
}

const box = (c: HTMLElement) => c.querySelector<HTMLInputElement>("input[type=checkbox]");
const text = (c: HTMLElement, id: string) =>
  c.querySelector(`[data-testid=${id}]`)?.textContent?.trim() ?? null;

describe("ResultsPost", () => {
  afterEach(() => cleanup());

  it("gives the host the box, unticked, and opts in on a tick, even for a guest host", () => {
    const { container, ontoggle } = mount(table());
    const b = box(container)!;
    expect(b.checked).toBe(false);
    expect(b.disabled).toBe(false);
    expect(container.textContent).toContain("Post the result to the CMD & CTRL Discord");
    expect(text(container, "results-status")).toMatch(/will not be posted/);
    b.click();
    expect(ontoggle).toHaveBeenCalledWith(true);
  });

  it("shows the choice, and opts out on an untick", () => {
    const { container, ontoggle } = mount(table({ results: true }));
    expect(text(container, "results-status")).toMatch(/posted to Discord when the game ends/);
    const b = box(container)!;
    expect(b.checked).toBe(true);
    b.click();
    expect(ontoggle).toHaveBeenCalledWith(false);
  });

  it("disables the box for an ended game, with the hint", () => {
    const { container, ontoggle } = mount(table({ state: "ended" }));
    expect(box(container)!.disabled).toBe(true);
    expect(text(container, "results-hint")).toMatch(/ended/);
    box(container)!.click();
    expect(ontoggle).not.toHaveBeenCalled();
  });

  it("tells everyone at the table, with no box for who cannot manage it", () => {
    const { container } = mount(table({ results: true }), { canManage: false });
    expect(box(container)).toBeNull();
    expect(text(container, "results-status")).toMatch(/posted to Discord/);
  });

  it("shows an error from the route", () => {
    const { container } = mount(table(), { error: "nope" });
    expect(container.querySelector("[role=alert]")?.textContent).toBe("nope");
  });
});
