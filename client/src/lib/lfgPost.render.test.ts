// @vitest-environment jsdom
//
// ADR 0161: the #lfg section of the lobby's table settings.

import { afterEach, describe, expect, it, vi } from "vitest";

import type { GameMeta } from "./api";
import LfgPost from "./components/LfgPost.svelte";
import { LFG_GUEST_HINT } from "./lfgPost";
import { cleanup, render } from "./test/render.svelte";

function table(over: Partial<GameMeta> = {}): GameMeta {
  return {
    id: "g1",
    name: "Friday pod",
    created_at: "2026-10-13T20:00:00Z",
    state: "lobby",
    players: [
      {
        player_id: "p1",
        name: "Alice",
        seat: 0,
        deck_uploaded: false,
        is_host: true,
        discord_id: "501",
      },
    ],
    ...over,
  } as GameMeta;
}

function mount(game: GameMeta, over: Record<string, unknown> = {}) {
  const ontoggle = vi.fn();
  const { container } = render(
    LfgPost as never,
    { game, canManage: true, ontoggle, ...over } as never,
  );
  return { container, ontoggle };
}

const box = (c: HTMLElement) => c.querySelector<HTMLInputElement>("input[type=checkbox]");
const text = (c: HTMLElement, id: string) =>
  c.querySelector(`[data-testid=${id}]`)?.textContent?.trim() ?? null;

describe("LfgPost", () => {
  afterEach(() => cleanup());

  it("gives the host the box, unticked, and posts on a tick", () => {
    const { container, ontoggle } = mount(table());
    const b = box(container)!;
    expect(b.checked).toBe(false);
    expect(b.disabled).toBe(false);
    expect(container.textContent).toContain("Post to the CMD & CTRL Discord (#lfg)");
    b.click();
    expect(ontoggle).toHaveBeenCalledWith(true);
  });

  it("shows the post's status, and withdraws on an untick", () => {
    const { container, ontoggle } = mount(table({ lfg: true, lfg_status: "posted" }));
    expect(text(container, "lfg-status")).toMatch(/^Posted in #lfg/);
    const b = box(container)!;
    expect(b.checked).toBe(true);
    b.click();
    expect(ontoggle).toHaveBeenCalledWith(false);
  });

  it("says when the post is unavailable", () => {
    const { container } = mount(table({ lfg: true, lfg_status: "unavailable" }));
    expect(text(container, "lfg-status")).toMatch(/unavailable/);
  });

  it("disables the box for a guest host, with the hint", () => {
    const guestHosted = table({
      players: [{ player_id: "p1", name: "Gus", seat: 0, deck_uploaded: false, is_host: true }],
    });
    const { container, ontoggle } = mount(guestHosted);
    expect(box(container)!.disabled).toBe(true);
    expect(text(container, "lfg-hint")).toBe(LFG_GUEST_HINT);
    box(container)!.click();
    expect(ontoggle).not.toHaveBeenCalled();
  });

  it("gives someone who cannot manage the table no box", () => {
    const { container } = mount(table(), { canManage: false });
    expect(box(container)).toBeNull();
    expect(text(container, "lfg-status")).toMatch(/not posted/);
  });
});
