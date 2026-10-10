// @vitest-environment jsdom
//
// #2920: the game-end fanfare. The tone table (win, spectated win, loss,
// draw), the winner's name in their seat colour, the reduced-motion
// path, and the two ways out.

import { afterEach, describe, expect, it, vi } from "vitest";

import GameEndFanfare from "./components/GameEndFanfare.svelte";
import { seatColor } from "./colors";
import { fanfareFor, type Fanfare } from "./gameEndFanfare";
import type { GameView, PlayerView } from "./protocol";
import { cleanup, flushSync, render } from "./test/render.svelte";

afterEach(cleanup);

const seat = (id: string, n: number, name: string): PlayerView =>
  ({ id, seat: n, name, life: 40 }) as unknown as PlayerView;

const SEATS = [seat("a", 0, "Alice"), seat("b", 1, "Bob")];

const ended = (outcome?: GameView["outcome"]): GameView =>
  ({ state: "ended", seats: SEATS, outcome }) as unknown as GameView;

const WIN = { kind: "win", winner: "b", cause: "last_standing" } as GameView["outcome"];

function mount(f: Fanfare, motion: boolean) {
  const onback = vi.fn();
  const ondismiss = vi.fn();
  const r = render(GameEndFanfare, { fanfare: f, motion, onback, ondismiss });
  return { r, onback, ondismiss, q: (s: string) => r.container.querySelector<HTMLElement>(s) };
}

describe("fanfareFor", () => {
  it("is null until the game has ended", () => {
    expect(fanfareFor({ state: "active", seats: SEATS } as unknown as GameView, "a")).toBeNull();
    expect(fanfareFor(null, "a")).toBeNull();
  });

  it("celebrates the viewer's own win", () => {
    const f = fanfareFor(ended(WIN), "b")!;
    expect(f.tone).toBe("own-win");
    expect(f.headline).toBe("Bob");
    expect(f.color).toBe(seatColor(1));
    expect(f.celebrate).toBe(true);
  });

  it("is quiet for the viewer's loss, and still names the winner", () => {
    const f = fanfareFor(ended(WIN), "a")!;
    expect(f.tone).toBe("loss");
    expect(f.headline).toBe("Bob");
    expect(f.celebrate).toBe(false);
  });

  it("says what won an effect win", () => {
    const f = fanfareFor(
      ended({
        kind: "win",
        winner: "b",
        cause: "effect",
        source_name: "Felidar Sovereign",
      } as GameView["outcome"]),
      "a",
    )!;
    expect(f.detail).toContain("Felidar Sovereign");
  });

  it("gives a spectator the winner with confetti", () => {
    const f = fanfareFor(ended(WIN), null)!;
    expect(f.tone).toBe("watch");
    expect(f.celebrate).toBe(true);
  });

  it("is neutral for a draw", () => {
    const f = fanfareFor(ended({ kind: "draw" } as GameView["outcome"]), "a")!;
    expect(f.tone).toBe("draw");
    expect(f.headline).toBe("A draw");
    expect(f.color).toBeNull();
    expect(f.celebrate).toBe(false);
  });
});

describe("GameEndFanfare", () => {
  it("win: names the winner in their seat colour with confetti", () => {
    const { q } = mount(fanfareFor(ended(WIN), "b")!, true);
    expect(q("[role=dialog]")).not.toBeNull();
    expect(q("#fanfare-headline")!.textContent).toBe("Bob");
    expect(q("[data-testid=game-end-fanfare]")!.getAttribute("style")).toContain(seatColor(1));
    expect(q("[data-testid=game-end-fanfare]")!.dataset.tone).toBe("own-win");
    expect(q("[data-testid=fanfare-confetti]")!.children.length).toBeGreaterThan(0);
  });

  it("loss: shows the winner without confetti", () => {
    const { q } = mount(fanfareFor(ended(WIN), "a")!, true);
    expect(q("#fanfare-headline")!.textContent).toBe("Bob");
    expect(q("[data-testid=game-end-fanfare]")!.dataset.tone).toBe("loss");
    expect(q("[data-testid=fanfare-confetti]")).toBeNull();
  });

  it("draw: a neutral headline, no seat colour, no confetti", () => {
    const { q } = mount(fanfareFor(ended({ kind: "draw" } as GameView["outcome"]), "a")!, true);
    expect(q("#fanfare-headline")!.textContent).toBe("A draw");
    expect(q("[data-testid=game-end-fanfare]")!.getAttribute("style") ?? "").not.toContain(
      "--seat",
    );
    expect(q("[data-testid=fanfare-confetti]")).toBeNull();
  });

  it("reduced motion: the same card and words, nothing that moves", () => {
    const { q } = mount(fanfareFor(ended(WIN), "b")!, false);
    expect(q("#fanfare-headline")!.textContent).toBe("Bob");
    expect(q("[data-testid=game-end-fanfare]")!.dataset.motion).toBe("0");
    expect(q("[data-testid=fanfare-confetti]")).toBeNull();
    expect(q(".card")!.classList.contains("pop")).toBe(false);
  });

  it("offers Back to lobby, and Keep looking (button and Escape)", () => {
    const { r, q, onback, ondismiss } = mount(fanfareFor(ended(WIN), "b")!, true);
    const buttons = [...r.container.querySelectorAll("button")];
    buttons.find((b) => b.textContent === "Back to lobby")!.click();
    expect(onback).toHaveBeenCalledTimes(1);
    buttons.find((b) => b.textContent?.startsWith("Keep looking"))!.click();
    expect(ondismiss).toHaveBeenCalledTimes(1);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(ondismiss).toHaveBeenCalledTimes(2);
    expect(q("[role=dialog]")).not.toBeNull();
  });

  // #2934: the scrim sat over the dock and swallowed the click on its
  // Back to lobby button. It is a tint now, so it must not be a modal.
  it("is not a modal, so the dock under it stays reachable", () => {
    const { q } = mount(fanfareFor(ended(WIN), "b")!, false);
    expect(q("[role=dialog]")?.getAttribute("aria-modal")).toBe("false");
  });
});
