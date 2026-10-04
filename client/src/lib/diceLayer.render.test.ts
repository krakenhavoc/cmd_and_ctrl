// @vitest-environment jsdom
//
// ADR 0121 §7, the rendered half: the dice layer draws a card's roll at
// the roller's seat, tumbles it, settles it on the server's number,
// reads it out once, and lets the strip's text cue show only then. The
// schedule itself is pinned in dice.test.ts.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";

import DiceLayer from "./components/board/DiceLayer.svelte";
import RevealBanner from "./components/board/RevealBanner.svelte";
import { DICE_FADE_MS, DICE_HOLD_MS, DICE_TUMBLE_MS } from "./dice";
import { DiceQueue } from "./diceQueue.svelte";
import type { GameView, LogEvent } from "./protocol";
import { resetSettings, updateSettings } from "./settings";
import { dockTestView } from "./test/dockView";
import { cleanup, render } from "./test/render.svelte";

function roll(seq: number, seat: number, results: number[], sides = 20): LogEvent {
  const dice = results.length === 1 ? `a d${sides}` : `${results.length}d${sides}`;
  return {
    seq,
    kind: "roll",
    seat,
    card_id: "ogre",
    text: `${seat === 0 ? "Me" : "Opp"} rolled ${dice} for Hoarding Ogre: ${results.join(", ")}`,
    sides,
    results,
  };
}

function flip(seq: number, seat: number, faces: string[], call?: "heads" | "tails"): LogEvent {
  return { seq, kind: "flip", seat, text: `Opp called heads: ${faces.join(", ")}`, faces, call };
}

function viewWith(log: LogEvent[]): GameView {
  return dockTestView({ log } as Partial<GameView>);
}

function boardWithAvatar(rect: { left: number; top: number; width: number; height: number }) {
  const board = document.createElement("div");
  board.getBoundingClientRect = () => new DOMRect(0, 0, 1000, 600);
  const avatar = document.createElement("div");
  avatar.setAttribute("data-seat-id", "opp");
  avatar.getBoundingClientRect = () => new DOMRect(rect.left, rect.top, rect.width, rect.height);
  board.appendChild(avatar);
  document.body.appendChild(board);
  return board;
}

let queue: DiceQueue;

beforeEach(() => {
  vi.useFakeTimers({
    toFake: [
      "Date",
      "setTimeout",
      "clearTimeout",
      "setInterval",
      "clearInterval",
      "requestAnimationFrame",
      "cancelAnimationFrame",
    ],
  });
  vi.setSystemTime(1_000_000);
  resetSettings();
  updateSettings("animations", "enabled", true);
  updateSettings("animations", "dice", true);
  updateSettings("animations", "speed", 1);
  updateSettings("accessibility", "reduceMotion", false);
  queue = new DiceQueue();
});

afterEach(() => {
  queue.dispose();
  cleanup();
  vi.useRealTimers();
});

function advance(ms: number) {
  vi.advanceTimersByTime(ms);
  flushSync();
}

function groups(root: ParentNode = document): HTMLElement[] {
  return [...root.querySelectorAll<HTMLElement>(".dice-layer .group")];
}

function announcer(root: ParentNode = document): HTMLElement {
  return root.querySelector<HTMLElement>('.dice-announcer[role="status"][aria-live="polite"]')!;
}

function bareBoard(): HTMLElement {
  const board = document.createElement("div");
  board.getBoundingClientRect = () => new DOMRect(0, 0, 1000, 600);
  document.body.appendChild(board);
  return board;
}

function mountLayer(log: LogEvent[], board: HTMLElement = bareBoard()) {
  return render(DiceLayer as never, { view: viewWith(log), boardEl: board, queue });
}

describe("the dice layer", () => {
  it("never animates what was already in the log when it mounted", () => {
    const r = mountLayer([roll(1, 1, [14])]);
    expect(groups(r.container)).toHaveLength(0);
    advance(100);
    expect(announcer(r.container).textContent).toBe("");
  });

  it("is hidden from the accessibility tree, with an always-mounted announcer beside it", () => {
    const r = mountLayer([]);
    expect(r.container.querySelector(".dice-layer")?.getAttribute("aria-hidden")).toBe("true");
    expect(announcer(r.container)).not.toBeNull();
  });

  it("tumbles a new roll, settles it on the server's number, reads it once, then fades", () => {
    const board = boardWithAvatar({ left: 700, top: 20, width: 40, height: 40 });
    const r = mountLayer([], board);
    r.setProps({ view: viewWith([roll(5, 1, [14])]) });

    let [g] = groups(r.container);
    expect(g.dataset.dicePhase).toBe("tumble");
    expect(g.classList.contains("moving")).toBe(true);
    // Mid-tumble the face is never the result, and nothing is read out.
    expect(g.querySelector(".die")?.getAttribute("data-face")).not.toBe("14");
    expect(announcer(r.container).textContent).toBe("");

    advance(DICE_TUMBLE_MS);
    [g] = groups(r.container);
    expect(g.dataset.dicePhase).toBe("hold");
    expect(g.querySelector(".die")?.getAttribute("data-face")).toBe("14");
    expect(g.querySelector(".num")?.textContent).toBe("14");
    expect(g.querySelector(".caption")?.textContent?.trim()).toBe("d20");
    expect(announcer(r.container).textContent).toContain("Opp rolled a d20 for Hoarding Ogre: 14");

    advance(DICE_HOLD_MS);
    expect(groups(r.container)[0].dataset.dicePhase).toBe("fade");
    advance(DICE_FADE_MS);
    expect(groups(r.container)).toHaveLength(0);
    // Read once, not again on the next frame.
    r.setProps({
      view: viewWith([roll(5, 1, [14]), { seq: 6, kind: "step", seat: -1, text: "x" }]),
    });
    expect(announcer(r.container).querySelectorAll("p")).toHaveLength(1);
  });

  it("sits beside the roller's avatar, toward the centre of the table", () => {
    const board = boardWithAvatar({ left: 700, top: 20, width: 40, height: 40 });
    const r = mountLayer([], board);
    r.setProps({ view: viewWith([roll(5, 1, [3])]) });
    const [g] = groups(r.container);
    expect(g.classList.contains("below")).toBe(true);
    expect(parseFloat(g.style.top)).toBe(20 + 40 + 10);
    expect(g.dataset.diceSeat).toBe("1");
  });

  it("falls back to the edge of the strip when the avatar is not on screen", () => {
    const r = mountLayer([]);
    r.setProps({ view: viewWith([roll(5, 1, [3])]) });
    expect(groups(r.container)[0].classList.contains("strip")).toBe(true);
  });

  it("with motion off, shows the result settled at once and reads it at once", () => {
    updateSettings("animations", "dice", false);
    const r = mountLayer([]);
    r.setProps({ view: viewWith([roll(5, 0, [4], 6)]) });
    const [g] = groups(r.container);
    expect(g.dataset.dicePhase).toBe("hold");
    expect(g.classList.contains("moving")).toBe(false);
    expect(g.querySelector(".die")?.getAttribute("data-face")).toBe("4");
    expect(g.querySelectorAll(".pip")).toHaveLength(4);
    expect(announcer(r.container).textContent).toContain("rolled a d6");
    advance(DICE_HOLD_MS + DICE_FADE_MS);
    expect(groups(r.container)).toHaveLength(0);
  });

  it("is still with the master switch off, even with the dice toggle on", () => {
    updateSettings("animations", "enabled", false);
    const r = mountLayer([]);
    r.setProps({ view: viewWith([roll(5, 0, [9])]) });
    expect(groups(r.container)[0].dataset.dicePhase).toBe("hold");
  });

  it("draws six dice of a bigger batch and a +N chip", () => {
    const r = mountLayer([]);
    r.setProps({ view: viewWith([roll(5, 0, [1, 2, 3, 4, 5, 6, 1, 2], 6)]) });
    const [g] = groups(r.container);
    expect(g.querySelectorAll(".die")).toHaveLength(6);
    expect(g.querySelector(".more")?.textContent).toBe("+2");
    expect(g.querySelector(".caption")?.textContent?.trim()).toBe("8d6");
  });

  it("spins a called coin and says won or lost once it lands", () => {
    const r = mountLayer([]);
    r.setProps({ view: viewWith([flip(5, 1, ["tails"], "heads")]) });
    let [g] = groups(r.container);
    expect(g.querySelector(".call")?.textContent).toBe("called heads");
    expect(g.querySelector(".verdict")).toBeNull();
    advance(DICE_TUMBLE_MS);
    [g] = groups(r.container);
    expect(g.querySelector(".verdict")?.textContent).toBe("lost");
    expect(g.querySelector(".caption")?.textContent?.trim()).toBe("tails");
  });

  it("plays two seats' rolls at once", () => {
    const r = mountLayer([]);
    r.setProps({ view: viewWith([roll(5, 0, [3]), roll(6, 1, [17])]) });
    expect(groups(r.container).map((g) => g.dataset.diceSeat)).toEqual(["0", "1"]);
  });
});

describe("the strip cue waits for the die", () => {
  it("shows a roll's line only once its die has settled", () => {
    const board = mountLayer([]);
    const strip = render(RevealBanner as never, { snap: viewWith([]), dice: queue });
    const next = viewWith([roll(5, 1, [14])]);
    board.setProps({ view: next });
    strip.setProps({ snap: next });
    expect(strip.container.querySelector(".random-line")).toBeNull();
    advance(DICE_TUMBLE_MS - 1);
    expect(strip.container.querySelector(".random-line")).toBeNull();
    advance(1);
    const line = strip.container.querySelector(".random-line");
    expect(line?.textContent).toContain("Hoarding Ogre: 14");
    // The dice layer's announcer reads it; the line is not a second live region.
    expect(line?.getAttribute("aria-live")).toBeNull();
  });

  it("shows it at once with motion off", () => {
    updateSettings("animations", "dice", false);
    const board = mountLayer([]);
    const strip = render(RevealBanner as never, { snap: viewWith([]), dice: queue });
    const next = viewWith([roll(5, 1, [14])]);
    board.setProps({ view: next });
    strip.setProps({ snap: next });
    expect(strip.container.querySelector(".random-line")).not.toBeNull();
  });

  it("shows it at once, as a live line, with no dice queue at all", () => {
    const strip = render(RevealBanner as never, { snap: viewWith([]), dice: null });
    strip.setProps({ snap: viewWith([roll(5, 1, [14])]) });
    const line = strip.container.querySelector(".random-line");
    expect(line).not.toBeNull();
    expect(line?.getAttribute("aria-live")).toBe("polite");
  });
});

// ADR 0121 §5: a die rolled at the table from the ⋯ menu.
function tableRoll(seq: number, rollID: number, result = 14): LogEvent {
  return {
    seq,
    kind: "table_roll",
    seat: 1,
    text: `Opp rolled a d20 at the table: ${result}`,
    sides: 20,
    results: [result],
    roll_id: rollID,
  };
}

describe("a table roll", () => {
  it("tumbles at the roller's seat and is read out once, even when an undo moves its line", () => {
    const board = boardWithAvatar({ left: 700, top: 20, width: 40, height: 40 });
    const r = mountLayer([], board);
    r.setProps({ view: viewWith([roll(4, 0, [3], 6), tableRoll(5, 1)]) });
    const keys = groups(r.container).map((g) => g.dataset.diceKey);
    expect(keys).toContain("table:1");
    const g = groups(r.container).find((x) => x.dataset.diceKey === "table:1")!;
    expect(g.dataset.diceSeat).toBe("1");
    expect(g.dataset.dicePhase).toBe("tumble");

    // Another seat undoes their roll from before it: the server writes
    // the table roll's line again, at seq 4. Same roll, not a new one.
    r.setProps({ view: viewWith([tableRoll(4, 1)]) });
    expect(groups(r.container).map((x) => x.dataset.diceKey)).toEqual(["table:1"]);
    advance(DICE_TUMBLE_MS);
    expect(announcer(r.container).textContent).toContain("Opp rolled a d20 at the table: 14");
    advance(DICE_HOLD_MS + DICE_FADE_MS);
    expect(groups(r.container)).toHaveLength(0);
    // A second undo moves it again, after it was read: nothing plays,
    // nothing is read again.
    r.setProps({
      view: viewWith([tableRoll(3, 1), { seq: 4, kind: "step", seat: -1, text: "x" }]),
    });
    advance(1);
    expect(groups(r.container)).toHaveLength(0);
    expect(
      [...announcer(r.container).querySelectorAll("p")].filter((p) =>
        p.textContent?.includes("at the table"),
      ),
    ).toHaveLength(1);
  });
});
