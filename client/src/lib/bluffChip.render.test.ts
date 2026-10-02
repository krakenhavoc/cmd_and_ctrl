// @vitest-environment jsdom
//
// bluffChip.render.test.ts — ADR 0111 §5 (S56 PR 1). The bluff
// control is a split button that is ALWAYS on the table: the main part
// arms and disarms in one click, the ▾ part sets which bluff and how.
// Both write the settings the Settings panel writes.

import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { get } from "svelte/store";
import { flushSync } from "svelte";

import ActionDock from "./components/board/ActionDock.svelte";
import { bluffArmed, _resetForTests as resetBluff, setBluffArmed } from "./bluff";
import { defaultSettings, settings } from "./settings";
import type { GameView, PlayerView, TurnView } from "./protocol";
import { render, cleanup } from "./test/render.svelte";

beforeEach(() => {
  settings.set(defaultSettings());
  resetBluff();
});
afterEach(cleanup);

const seats = [
  { id: "a", name: "Alice", seat: 0 },
  { id: "b", name: "Bob", seat: 1 },
] as unknown as PlayerView[];

const turn: TurnView = {
  seq: 5,
  number: 3,
  active_seat: 0,
  priority_holder: 0,
  phase: "precombat_main",
  step: "precombat_main",
};

// ADR 0111 PR 2: the chip lives in the action dock's toggles row.
function mount() {
  return render(
    ActionDock as never,
    {
      view: { turn, seats, mulligans_open: false } as unknown as GameView,
      viewerHasPriority: true,
      viewerIsActive: true,
      autopassEnabled: false,
      onPassPriority: () => {},
      onPassTurn: () => {},
      onToggleAutopass: () => {},
    } as never,
  ).container;
}

function main(c: HTMLElement): HTMLButtonElement {
  const el = c.querySelector<HTMLButtonElement>("button.bluff-main");
  if (!el) throw new Error("no bluff button");
  return el;
}
function caret(c: HTMLElement): HTMLButtonElement {
  const el = c.querySelector<HTMLButtonElement>("button.bluff-caret");
  if (!el) throw new Error("no bluff caret");
  return el;
}
const gp = () => get(settings).gameplay;
const POP = '[role="group"][aria-label="bluff settings"]';

describe("bluff chip", () => {
  it("renders with default settings, off and enabled", () => {
    const c = mount();
    const b = main(c);
    expect(b.textContent?.trim()).toBe("bluff");
    expect(b.getAttribute("aria-pressed")).toBe("false");
    expect(b.disabled).toBe(false);
    expect(caret(c).getAttribute("aria-label")).toBe("bluff options");
  });

  it("one click with nothing set up turns on represent-a-counterspell and arms", () => {
    const c = mount();
    main(c).click();
    flushSync();
    expect(gp().bluffCounterspell).toBe(true);
    expect(gp().bluffInstant).toBe(false);
    expect(get(bluffArmed)).toBe(true);
    expect(main(c).getAttribute("aria-pressed")).toBe("true");
    expect(main(c).textContent?.trim()).toBe("bluff ✓ counter");
  });

  it("a click while armed disarms and leaves the settings alone", () => {
    const c = mount();
    main(c).click();
    flushSync();
    main(c).click();
    flushSync();
    expect(get(bluffArmed)).toBe(false);
    expect(gp().bluffCounterspell).toBe(true);
    expect(main(c).getAttribute("aria-pressed")).toBe("false");
  });

  it("arming with a kind already chosen changes no setting", () => {
    settings.update((s) => ({ ...s, gameplay: { ...s.gameplay, bluffInstant: true } }));
    const c = mount();
    main(c).click();
    flushSync();
    expect(get(bluffArmed)).toBe(true);
    expect(gp().bluffCounterspell).toBe(false);
    expect(gp().bluffInstant).toBe(true);
  });

  it("the caret opens a popover that writes the same settings", () => {
    const c = mount();
    expect(c.querySelector(POP)).toBeNull();
    caret(c).click();
    flushSync();
    expect(caret(c).getAttribute("aria-expanded")).toBe("true");
    const pop = c.querySelector<HTMLElement>(POP);
    expect(pop).not.toBeNull();

    const boxes = pop!.querySelectorAll<HTMLInputElement>('input[type="checkbox"]');
    expect(boxes).toHaveLength(2);
    boxes[1].click();
    flushSync();
    expect(gp().bluffInstant).toBe(true);
    boxes[0].click();
    flushSync();
    expect(gp().bluffCounterspell).toBe(true);

    pop!.querySelector<HTMLInputElement>('input[type="radio"][value="manual"]')!.click();
    flushSync();
    expect(gp().bluffMode).toBe("manual");
    pop!.querySelector<HTMLInputElement>('input[type="radio"][value="timed"]')!.click();
    flushSync();
    expect(gp().bluffMode).toBe("timed");
  });

  it("closes the popover on Escape", () => {
    const c = mount();
    caret(c).click();
    flushSync();
    c.querySelector(POP)!.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Escape", bubbles: true }),
    );
    flushSync();
    expect(c.querySelector(POP)).toBeNull();
  });

  it("is disabled, with a reason, while smart autopass is off", () => {
    settings.update((s) => ({ ...s, gameplay: { ...s.gameplay, smartAutoPass: false } }));
    const c = mount();
    const b = main(c);
    expect(b.disabled).toBe(true);
    expect(b.title).toContain("needs smart auto-pass");
    b.click();
    flushSync();
    expect(get(bluffArmed)).toBe(false);
    expect(gp().bluffCounterspell).toBe(false);
  });

  it("shows armed state when armed from elsewhere", () => {
    const c = mount();
    setBluffArmed(true);
    flushSync();
    expect(main(c).getAttribute("aria-pressed")).toBe("true");
    expect(main(c).classList.contains("on")).toBe(true);
  });
});
