// @vitest-environment jsdom
//
// hintLayer.render.test.ts — the first-use hints on screen (ADR 0125
// §3.5, §3.6, §8): the card is `complementary "tip"` beside its anchor,
// takes no focus, is read out once and describes its anchor; Got it and
// Hide tips; `i` focuses the card and Escape dismisses it and puts focus
// back; reduced motion drops the motion; a table hint steps aside for a
// dock request; a settings hint draws inside the dialog; a phone gets
// the strip; a missing anchor logs.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get, writable } from "svelte/store";

import type { Route } from "./router";

const routeStore = vi.hoisted(() => ({
  value: null as null | import("svelte/store").Writable<Route>,
}));
vi.mock("./router", () => {
  routeStore.value = writable<Route>({ name: "lobby" });
  return { route: routeStore.value, navigate: vi.fn() };
});

import HintLayer from "./components/hints/HintLayer.svelte";
import HintSlot from "./components/hints/HintSlot.svelte";
import ShortcutLayer from "./components/ShortcutLayer.svelte";
import { L } from "./labels";
import { POLL_MS } from "./tutorial";
import { defaultSettings, settings, settingsOpen } from "./settings";
import { _resetHintsForTests, replayTips, TIP_BODY_ID, TIP_SLOT_ATTR } from "./hints/runtime";
import PageTipSlot from "./components/hints/PageTipSlot.svelte";
import { _resetTableMomentForTests, publishTableMoment } from "./hints/tableMoment";
import { SITE_WINDOW_MS } from "./hints/queue";
import type { Hint, HintPlace } from "./hints/hint";
import { moment } from "./test/hintContexts";
import { ME } from "./test/tutorialBoards";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

function hint(id: `${HintPlace}.${string}`, extra: Partial<Hint> = {}): Hint {
  return {
    id,
    version: 1,
    place: id.split(".")[0] as HintPlace,
    order: 0,
    anchor: { label: L.actions },
    title: "Your controls",
    body: "Anything you need to answer opens here.",
    ...extra,
  };
}

/** Give an element a viewport rect (jsdom lays nothing out). */
function place(el: Element, left: number, top: number, width: number, height: number): void {
  el.getBoundingClientRect = () =>
    ({
      left,
      top,
      width,
      height,
      right: left + width,
      bottom: top + height,
      x: left,
      y: top,
      toJSON: () => ({}),
    }) as DOMRect;
}

/** An anchor on the page: a labelled region with an area. */
function anchor(label: string = L.actions, parent: Element = document.body): HTMLElement {
  const el = document.createElement("section");
  el.setAttribute("aria-label", label);
  place(el, 400, 100, 200, 40);
  parent.appendChild(el);
  return el;
}

const tip = () => document.querySelector<HTMLElement>(`[aria-label="${L.tip}"]`);
const live = () => document.querySelector(".hint-live")?.textContent ?? "";
const buttonNamed = (name: string) =>
  [...document.querySelectorAll<HTMLButtonElement>("button")].find(
    (b) => (b.textContent ?? "").trim() === name,
  );

function poll(ms = POLL_MS): void {
  vi.advanceTimersByTime(ms);
  flushSync();
}

let logs: string[] = [];

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(1_000_000);
  localStorage.clear();
  settings.set(defaultSettings());
  settingsOpen.set(false);
  routeStore.value!.set({ name: "lobby" });
  logs = [];
  _resetHintsForTests({ log: (m) => logs.push(m) });
  _resetTableMomentForTests();
  Object.defineProperty(window, "innerWidth", { value: 1280, configurable: true });
  Object.defineProperty(window, "innerHeight", { value: 800, configurable: true });
});

afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  vi.useRealTimers();
});

describe("HintLayer on a site page", () => {
  it('shows the card beside its anchor as complementary "tip", without taking focus', () => {
    const before = document.createElement("button");
    before.textContent = "somewhere else";
    document.body.appendChild(before);
    before.focus();
    const target = anchor();
    render(HintLayer as never, { hints: [hint("lobby.controls")] } as never);
    poll();

    const card = tip();
    expect(card).not.toBeNull();
    expect(card!.tagName).toBe("ASIDE");
    expect(card!.textContent).toContain("Your controls");
    expect(card!.textContent).toContain("Anything you need to answer opens here.");
    // Below the anchor, with a ring round it that takes no clicks.
    expect(card!.style.top).toBe("150px");
    expect(document.querySelector(".hint-ring")).not.toBeNull();
    // No focus theft.
    expect(document.activeElement).toBe(before);
    // Read out once, and described on the anchor.
    expect(live()).toBe("Tip: Your controls. Anything you need to answer opens here.");
    expect(target.getAttribute("aria-describedby")).toBe(TIP_BODY_ID);
    expect(document.getElementById(TIP_BODY_ID)?.textContent).toBe(
      "Anything you need to answer opens here.",
    );
  });

  it("Got it marks the hint seen at its version, and the visit shows no second hint", () => {
    const target = anchor();
    render(
      HintLayer as never,
      {
        hints: [hint("lobby.controls", { version: 3 }), hint("lobby.next", { order: 1 })],
      } as never,
    );
    poll();
    click(buttonNamed("Got it")!);
    poll();
    expect(tip()).toBeNull();
    expect(get(settings).help.seen).toEqual({ "lobby.controls": 3 });
    expect(target.hasAttribute("aria-describedby")).toBe(false);
    expect(live()).toBe("");
    poll(10_000);
    expect(tip()).toBeNull();
  });

  it("Hide tips marks it seen and turns tips off", () => {
    anchor();
    render(HintLayer as never, { hints: [hint("lobby.controls")] } as never);
    poll();
    click(buttonNamed("Hide tips")!);
    expect(get(settings).help).toEqual({ seen: { "lobby.controls": 1 }, tipsOff: true });
  });

  it("an action link reads Not now beside it", () => {
    anchor();
    render(
      HintLayer as never,
      {
        hints: [
          hint("lobby.practice", { action: { label: "Start practice", href: "#/practice" } }),
        ],
      } as never,
    );
    poll();
    expect(buttonNamed("Not now")).toBeDefined();
    const link = tip()!.querySelector("a");
    expect(link?.getAttribute("href")).toBe("#/practice");
    expect(link?.textContent).toBe("Start practice");
  });

  it("keeps every label of a three-action tip on one line at 1280 wide", () => {
    anchor();
    render(
      HintLayer as never,
      {
        hints: [
          hint("lobby.practice", { action: { label: "Start practice", href: "#/practice" } }),
        ],
      } as never,
    );
    poll();
    expect(window.innerWidth).toBe(1280);
    // jsdom lays nothing out, so the no-wrap rule is what is checked:
    // every button and the action link are nowrap, and the card is the
    // wider one that fits the three of them.
    const labels = [buttonNamed("Hide tips")!, tip()!.querySelector("a")!, buttonNamed("Not now")!];
    for (const el of labels) {
      expect(getComputedStyle(el).whiteSpace).toBe("nowrap");
    }
    expect(tip()!.classList.contains("with-action")).toBe(true);
    // The action is the primary (orange) button; dismissing is the quiet one.
    expect(tip()!.querySelector("a")!.classList.contains("hint-primary")).toBe(true);
    expect(buttonNamed("Not now")!.classList.contains("hint-action")).toBe(true);
  });

  it("offers nothing with tips off, until the Help menu replays the page's tips", () => {
    anchor();
    settings.update((s) => ({ ...s, help: { seen: { "lobby.controls": 1 }, tipsOff: true } }));
    const hints = [hint("lobby.controls")];
    render(HintLayer as never, { hints } as never);
    poll();
    expect(tip()).toBeNull();
    replayTips("lobby", hints);
    poll();
    expect(tip()).not.toBeNull();
    expect(get(settings).help.seen).toEqual({});
  });

  it("logs a hint whose anchor never comes, and shows nothing", () => {
    render(HintLayer as never, { hints: [hint("lobby.controls")] } as never);
    poll(SITE_WINDOW_MS + POLL_MS);
    expect(tip()).toBeNull();
    expect(logs).toEqual(["hint: lobby.controls has no anchor on the page"]);
  });

  it("with reduced motion the card simply appears: no fade and no pulse", () => {
    anchor();
    render(HintLayer as never, { hints: [hint("lobby.controls")] } as never);
    poll();
    expect(tip()!.classList.contains("animate")).toBe(true);
    expect(document.querySelector(".hint-ring")!.classList.contains("pulse")).toBe(true);

    settings.update((s) => ({ ...s, accessibility: { ...s.accessibility, reduceMotion: true } }));
    poll();
    expect(tip()!.classList.contains("animate")).toBe(false);
    expect(document.querySelector(".hint-ring")!.classList.contains("pulse")).toBe(false);
  });

  it("is a one-line strip on a phone that a tap opens", () => {
    Object.defineProperty(window, "innerWidth", { value: 400, configurable: true });
    anchor();
    render(HintLayer as never, { hints: [hint("lobby.controls")] } as never);
    poll();
    const card = tip()!;
    expect(card.classList.contains("strip")).toBe(true);
    expect(buttonNamed("Got it")).toBeUndefined();
    // The body is still there for the anchor's description.
    expect(document.getElementById(TIP_BODY_ID)).not.toBeNull();
    const head = card.querySelector<HTMLButtonElement>("button[aria-expanded]")!;
    click(head);
    expect(buttonNamed("Got it")).toBeDefined();
    expect(card.textContent).toContain("Anything you need to answer opens here.");
  });
});

// ADR 0156 §3 item 5: a page in the site frame has a tip slot under its
// lede, and its tips are drawn there, in the page's flow, not over it.
describe("HintLayer on a page in the site frame", () => {
  function frame(): HTMLElement {
    const { container } = render(PageTipSlot as never, {} as never);
    return container.querySelector<HTMLElement>(`[${TIP_SLOT_ATTR}]`)!;
  }

  it("draws the page's tip in the slot, with no floating card, and Got it still works", () => {
    const slot = frame();
    const target = anchor();
    render(HintLayer as never, { hints: [hint("lobby.controls")] } as never);
    poll();

    const cards = document.querySelectorAll(`[aria-label="${L.tip}"]`);
    expect(cards).toHaveLength(1);
    const card = cards[0] as HTMLElement;
    expect(slot.contains(card)).toBe(true);
    expect(card.classList.contains("inline")).toBe(true);
    // In the flow: no viewport position of its own.
    expect(card.style.top).toBe("");
    expect(card.style.left).toBe("");
    expect(target.getAttribute("aria-describedby")).toBe(TIP_BODY_ID);
    expect(live()).toBe("Tip: Your controls. Anything you need to answer opens here.");

    click(buttonNamed("Got it")!);
    poll();
    expect(tip()).toBeNull();
    expect(get(settings).help.seen["lobby.controls"]).toBe(1);
  });

  it("leaves a table hint floating even with a slot on the page", () => {
    const slot = frame();
    routeStore.value!.set({ name: "game", gameID: "g" });
    anchor();
    publishTableMoment({ moment: moment(), view: null, viewerID: ME });
    render(HintLayer as never, { hints: [hint("table.dock")] } as never);
    poll();
    const card = tip();
    expect(card).not.toBeNull();
    expect(slot.contains(card)).toBe(false);
    expect(card!.classList.contains("inline")).toBe(false);
  });
});

describe("Go to the tip (i)", () => {
  function press(key: string, target: EventTarget = document.activeElement ?? document.body) {
    target.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true, cancelable: true }));
    flushSync();
  }

  it("moves focus to the tip, and Escape dismisses it and puts focus back", () => {
    const before = document.createElement("button");
    before.textContent = "where I was";
    document.body.appendChild(before);
    anchor();
    render(ShortcutLayer as never, {} as never);
    render(HintLayer as never, { hints: [hint("lobby.controls")] } as never);
    poll();
    before.focus();

    press("i");
    expect(document.activeElement).toBe(tip());

    press("Escape");
    expect(tip()).toBeNull();
    expect(get(settings).help.seen).toEqual({ "lobby.controls": 1 });
    expect(document.activeElement).toBe(before);
  });

  it("does nothing with no tip on screen", () => {
    const before = document.createElement("button");
    document.body.appendChild(before);
    render(ShortcutLayer as never, {} as never);
    before.focus();
    press("i");
    expect(document.activeElement).toBe(before);
    expect(document.querySelector(".sc-toast")).toBeNull();
  });
});

describe("HintLayer at the table", () => {
  it("shows a table hint in a quiet moment and steps aside, unseen, for a dock request", () => {
    routeStore.value!.set({ name: "game", gameID: "g" });
    anchor();
    const table = publishTableMoment({ moment: moment(), view: null, viewerID: ME });
    render(HintLayer as never, { hints: [hint("table.dock")] } as never);
    poll();
    expect(tip()).not.toBeNull();

    table.update({ moment: moment({ dockRequest: true }), view: null, viewerID: ME });
    poll();
    expect(tip()).toBeNull();
    expect(get(settings).help.seen).toEqual({});

    table.update({ moment: moment(), view: null, viewerID: ME });
    poll();
    expect(tip()).not.toBeNull();
  });

  it("shows nothing while the tutorial coach is visible", () => {
    routeStore.value!.set({ name: "game", gameID: "g" });
    anchor();
    publishTableMoment({ moment: moment({ coachVisible: true }), view: null, viewerID: ME });
    render(HintLayer as never, { hints: [hint("table.dock")] } as never);
    poll(1_000);
    expect(tip()).toBeNull();
  });
});

describe("a hint inside the Settings dialog", () => {
  it("is drawn by the dialog's HintSlot, inside the dialog, and only there", () => {
    const dialog = document.createElement("div");
    dialog.setAttribute("role", "dialog");
    place(dialog, 200, 50, 800, 600);
    document.body.appendChild(dialog);
    const nav = document.createElement("nav");
    nav.setAttribute("aria-label", L.adminViews);
    place(nav, 220, 100, 160, 300);
    dialog.appendChild(nav);
    const slot = document.createElement("div");
    dialog.appendChild(slot);

    settingsOpen.set(true);
    render(
      HintLayer as never,
      {
        hints: [hint("settings.display", { anchor: { label: L.adminViews } })],
      } as never,
    );
    // The slot lives in the dialog.
    slot.appendChild(render(HintSlot as never, {} as never).container);
    poll();

    const cards = document.querySelectorAll(`[aria-label="${L.tip}"]`);
    expect(cards).toHaveLength(1);
    expect(dialog.contains(cards[0])).toBe(true);
  });
});
