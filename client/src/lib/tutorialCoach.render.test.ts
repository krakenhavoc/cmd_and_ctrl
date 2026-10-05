// @vitest-environment jsdom
//
// tutorialCoach.render.test.ts — the tutorial's walking skeleton (ADR 0076
// §2.3, §2.4; #1079): the coach card's six states, the scrim, the anchor
// resolver and the missing-anchor self-advance, and steps 1 and 14 end to
// end.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";

import Board from "./components/board/Board.svelte";
import CoachCard from "./components/tutorial/CoachCard.svelte";
import TutorialCoach from "./components/tutorial/TutorialCoach.svelte";
import { get } from "svelte/store";
import { ANCHOR_GRACE_MS, POLL_MS, type CoachState, type TutorialStep } from "./tutorial";
import {
  COMMANDER,
  HANDOFF,
  ON_THE_STACK,
  OPENING_ROLL,
  TUTORIAL_STEPS,
  WELCOME,
} from "./tutorialSteps";
import { anchorRect, resolveAnchor } from "./tutorialAnchor";
import { L } from "./labels";
import { emit } from "./tutorialBus";
import { defaultSettings, settings } from "./settings";
import { render, click, cleanup, flushSync, type Rendered } from "./test/render.svelte";
import type { GameView, PlayerView } from "./protocol";

vi.mock("./sounds", () => ({ play: () => {} }));

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  settings.set(defaultSettings());
});
afterEach(() => {
  cleanup();
  document.body.innerHTML = "";
  vi.useRealTimers();
});

const buttons = (root: ParentNode) =>
  [...root.querySelectorAll("button")].map((b) => (b.textContent ?? "").trim()).filter(Boolean);
const button = (root: ParentNode, name: string) =>
  [...root.querySelectorAll<HTMLButtonElement>("button")].find(
    (b) => (b.textContent ?? "").trim() === name,
  );

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

describe("CoachCard: the six states", () => {
  const card = (state: CoachState, extra: Record<string, unknown> = {}) =>
    render(
      CoachCard as never,
      {
        state,
        n: 5,
        total: 14,
        title: "Tap a land for mana",
        body: "Click one of the Forests.",
        hint: "A land already on the table.",
        ...extra,
      } as never,
    ).container;

  it("opening: Skip tutorial and Start, no status", () => {
    const c = card("opening", { n: 1 });
    expect(c.textContent).toContain("1 / 14");
    expect(buttons(c)).toEqual(["Skip tutorial", "Start"]);
    expect(c.querySelector(".coach-status")).toBeNull();
  });

  it("action: no primary, only Skip step and what it waits for", () => {
    const c = card("action");
    expect(buttons(c)).toEqual(["Skip step"]);
    expect(c.querySelector(".coach-status")?.textContent).toContain("Waiting for you");
    expect(c.querySelector(".coach-hint")).toBeNull();
  });

  it("hint: the same card with the hint added", () => {
    const c = card("hint");
    expect(buttons(c)).toEqual(["Skip step"]);
    expect(c.querySelector(".coach-hint")?.textContent).toContain("A land already on the table.");
  });

  it("watch: no ask, the bot is thinking", () => {
    const c = card("watch");
    expect(buttons(c)).toEqual(["Skip step"]);
    expect(c.querySelector(".coach-status")?.textContent).toContain("Bot is thinking");
    expect(c.querySelector(".coach-hint")).toBeNull();
  });

  it("recovered: the step stays live and the hint shows at once", () => {
    const c = card("recovered", { title: "Close — that played a land" });
    expect(c.textContent).toContain("Close — that played a land");
    expect(buttons(c)).toEqual(["Skip step"]);
    expect(c.querySelector(".coach-hint")).not.toBeNull();
  });

  it("done: Replay and Finish", () => {
    const c = card("done", { n: 14 });
    expect(c.textContent).toContain("14 / 14");
    expect(buttons(c)).toEqual(["Replay", "Finish"]);
  });

  it("marks progress: past steps full, this one half", () => {
    const c = card("action");
    expect(c.querySelectorAll(".seg")).toHaveLength(14);
    expect(c.querySelectorAll(".seg.past")).toHaveLength(4);
    expect(c.querySelectorAll(".seg.now")).toHaveLength(1);
  });

  it("folds to its one-line header, which still names the step", () => {
    const c = card("action", { minimised: true });
    expect(c.querySelector(".coach-body")).toBeNull();
    expect(c.querySelector(".coach-min-title")?.textContent).toBe("Tap a land for mana");
    expect(
      c
        .querySelector('button[aria-label="restore: Tap a land for mana"]')
        ?.getAttribute("aria-expanded"),
    ).toBe("false");
  });
});

describe("the anchor resolver", () => {
  function table(): void {
    document.body.innerHTML = `
      <section aria-label="Opp board"><div role="list" aria-label="lands" id="opp-lands"></div></section>
      <section aria-label="your board"><div role="list" aria-label="lands" id="my-lands"></div>
        <div aria-label="your hand" id="hand"></div>
        <div data-instance-id="c-1" id="elves"></div></section>`;
  }

  it("finds a label, scoped to its container", () => {
    table();
    // Unscoped, "lands" would be the first opponent's row.
    expect(resolveAnchor({ label: L.lands })?.id).toBe("opp-lands");
    expect(resolveAnchor({ label: L.lands, within: L.yourBoard })?.id).toBe("my-lands");
    expect(resolveAnchor({ label: L.yourHand })?.id).toBe("hand");
  });

  it("matches a dynamic label by its full name, or by its stem (ADR 0125 §2.2)", () => {
    document.body.innerHTML = `
      <section aria-label="Practice Bot board" id="bot-board">
        <div aria-label="Practice Bot command zone, 1 card" id="bot-cz"></div></section>
      <section aria-label="your board">
        <div aria-label="Player command zone, 1 card" id="my-cz"></div>
        <section aria-label="stack: 2 on the stack" id="pile"></section></section>`;
    expect(resolveAnchor({ label: L.seatBoard("Practice Bot") })?.id).toBe("bot-board");
    // Unscoped, the stem finds the first command zone; scoped, the viewer's.
    expect(resolveAnchor({ label: L.commandZone.any })?.id).toBe("bot-cz");
    expect(resolveAnchor({ label: L.commandZone.any, within: L.yourBoard })?.id).toBe("my-cz");
    expect(resolveAnchor({ label: L.stackPile.any })?.id).toBe("pile");
    expect(resolveAnchor({ label: L.discard.any })).toBeNull();
  });

  it("finds a card by instance id", () => {
    table();
    expect(resolveAnchor({ cardID: "c-1" })?.id).toBe("elves");
  });

  it("is null when the label, its container or the card is not there", () => {
    table();
    expect(resolveAnchor({ label: L.attention })).toBeNull();
    expect(resolveAnchor({ label: L.lands, within: L.seatBoard("No such") })).toBeNull();
    expect(resolveAnchor({ cardID: "gone" })).toBeNull();
  });

  it("unions several anchors, and treats an element with no area as missing", () => {
    table();
    place(document.getElementById("hand")!, 100, 600, 400, 120);
    place(document.getElementById("elves")!, 200, 300, 80, 110);
    place(document.getElementById("my-lands")!, 0, 0, 160, 0);
    expect(anchorRect([{ label: L.yourHand }, { cardID: "c-1" }])).toEqual({
      left: 100,
      top: 300,
      width: 400,
      height: 420,
    });
    expect(anchorRect([{ label: L.lands, within: L.yourBoard }])).toBeNull();
    expect(anchorRect([{ label: L.gameActions }, { cardID: "c-1" }])).toEqual({
      left: 200,
      top: 300,
      width: 80,
      height: 110,
    });
  });

  it("finds a seat's portrait, and cuts a card to what its scrolling row shows", () => {
    document.body.innerHTML = `
      <div data-seat-id="bot" id="bot"></div>
      <div role="list" aria-label="creatures" id="row" style="overflow: auto">
        <div data-instance-id="c-2" id="tall"></div></div>`;
    expect(resolveAnchor({ seatID: "bot" })?.id).toBe("bot");
    // A creature card taller than the creature row: the row scrolls,
    // and the hole goes round the part a player can see and click.
    place(document.getElementById("row")!, 0, 300, 1000, 100);
    place(document.getElementById("tall")!, 500, 310, 120, 168);
    expect(anchorRect([{ cardID: "c-2" }])).toEqual({
      left: 500,
      top: 310,
      width: 120,
      height: 90,
    });
    // Scrolled out of sight entirely: missing.
    place(document.getElementById("tall")!, 500, 420, 120, 168);
    expect(anchorRect([{ cardID: "c-2" }])).toBeNull();
  });
});

describe("TutorialCoach", () => {
  const spot: TutorialStep = {
    id: "read-hand",
    n: 2,
    kind: "action",
    title: "Read your hand",
    body: "Hover the fan to lift it.",
    anchor: { label: L.yourHand },
    done: (c) => c.event === "hand-hovered",
  };

  function mount(steps: TutorialStep[], extra: Record<string, unknown> = {}) {
    const sizes: Array<[number, number]> = [];
    const replays: number[] = [];
    const log = vi.fn();
    const r = render(
      TutorialCoach as never,
      {
        view: null,
        viewerID: "me",
        steps,
        log,
        onSize: (w: number, h: number) => sizes.push([w, h]),
        onReplay: () => replays.push(1),
        ...extra,
      } as never,
    );
    return { ...r, sizes, replays, log };
  }

  it("walks steps 1 and 14: Start, then Finish hides the card and frees its cell", () => {
    const m = mount([WELCOME, HANDOFF]);
    const c = m.container;
    expect(c.querySelector('[aria-label="tutorial coach"]')).not.toBeNull();
    expect(c.textContent).toContain("A five-minute practice game");
    // Step 1 leaves the board lit.
    expect(document.querySelector(".tutorial-scrim")).toBeNull();
    click(button(c, "Start")!);
    expect(c.textContent).toContain("That is the whole interface");
    expect(c.textContent).toContain("Press ? for the keymap and , for settings.");
    expect(document.querySelector(".tutorial-scrim")).toBeNull();
    click(button(c, "Finish")!);
    expect(c.querySelector('[aria-label="tutorial coach"]')).toBeNull();
    expect(m.sizes.at(-1)).toEqual([0, 0]);
    expect(m.replays).toHaveLength(0);
  });

  it("Skip tutorial hides the card at once", () => {
    const m = mount([WELCOME, spot, HANDOFF]);
    click(button(m.container, "Skip tutorial")!);
    expect(m.container.querySelector('[aria-label="tutorial coach"]')).toBeNull();
    expect(m.sizes.at(-1)).toEqual([0, 0]);
  });

  it("Replay hides the card and opens a fresh practice table", () => {
    const m = mount([WELCOME, HANDOFF]);
    click(button(m.container, "Start")!);
    click(button(m.container, "Replay")!);
    expect(m.replays).toHaveLength(1);
    expect(m.container.querySelector('[aria-label="tutorial coach"]')).toBeNull();
  });

  it("spotlights a step's anchor with a click-through scrim, and advances on its event", () => {
    document.body.insertAdjacentHTML("beforeend", '<div aria-label="your hand" id="hand"></div>');
    place(document.getElementById("hand")!, 340, 670, 570, 112);
    const m = mount([WELCOME, spot, HANDOFF]);
    click(button(m.container, "Start")!);
    const scrim = document.querySelector<HTMLElement>(".tutorial-scrim");
    expect(scrim).not.toBeNull();
    expect(scrim!.getAttribute("aria-hidden")).toBe("true");
    // The hole is the anchor plus its padding.
    expect(scrim!.style.left).toBe("334px");
    expect(scrim!.style.top).toBe("664px");
    expect(scrim!.style.width).toBe("582px");
    expect(scrim!.style.height).toBe("124px");
    // The board stays playable: a click on the anchor reaches it, not
    // the scrim (the scrim is not an ancestor and swallows nothing).
    const hand = document.getElementById("hand")!;
    const clicked = vi.fn();
    hand.addEventListener("click", clicked);
    hand.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    expect(clicked).toHaveBeenCalledTimes(1);
    expect(scrim!.contains(hand)).toBe(false);
    emit("hand-hovered");
    flushSync();
    expect(m.container.textContent).toContain("That is the whole interface");
    expect(document.querySelector(".tutorial-scrim")).toBeNull();
  });

  it("advances a step whose anchor never appears, and logs it, after the grace period", () => {
    vi.useFakeTimers();
    const m = mount([WELCOME, spot, HANDOFF]);
    click(button(m.container, "Start")!);
    expect(m.container.textContent).toContain("Read your hand");
    expect(document.querySelector(".tutorial-scrim")).toBeNull();
    vi.advanceTimersByTime(ANCHOR_GRACE_MS - POLL_MS);
    flushSync();
    expect(m.container.textContent).toContain("Read your hand");
    vi.advanceTimersByTime(POLL_MS * 2);
    flushSync();
    expect(m.container.textContent).toContain("That is the whole interface");
    expect(m.log).toHaveBeenCalledWith(expect.stringMatching(/read-hand has no anchor/));
  });

  it("waits out an anchor that renders a moment late", () => {
    vi.useFakeTimers();
    const m = mount([WELCOME, spot, HANDOFF]);
    click(button(m.container, "Start")!);
    vi.advanceTimersByTime(ANCHOR_GRACE_MS / 2);
    document.body.insertAdjacentHTML("beforeend", '<div aria-label="your hand" id="hand"></div>');
    place(document.getElementById("hand")!, 10, 10, 100, 100);
    vi.advanceTimersByTime(ANCHOR_GRACE_MS);
    flushSync();
    expect(m.container.textContent).toContain("Read your hand");
    expect(document.querySelector(".tutorial-scrim")).not.toBeNull();
    expect(m.log).not.toHaveBeenCalled();
  });

  it("advances when the anchor disappears mid-step rather than spotlighting nothing", () => {
    vi.useFakeTimers();
    document.body.insertAdjacentHTML("beforeend", '<div aria-label="your hand" id="hand"></div>');
    place(document.getElementById("hand")!, 10, 10, 100, 100);
    const m = mount([WELCOME, spot, HANDOFF]);
    click(button(m.container, "Start")!);
    expect(document.querySelector(".tutorial-scrim")).not.toBeNull();
    document.getElementById("hand")!.remove();
    vi.advanceTimersByTime(ANCHOR_GRACE_MS + POLL_MS);
    flushSync();
    expect(m.container.textContent).toContain("That is the whole interface");
  });

  it("Skip step moves on silently", () => {
    vi.useFakeTimers();
    const m = mount([WELCOME, spot, HANDOFF]);
    click(button(m.container, "Start")!);
    click(button(m.container, "Skip step")!);
    expect(m.container.textContent).toContain("That is the whole interface");
    expect(m.log).not.toHaveBeenCalled();
  });
});

// Sub-PR 4 (#1081): what the coach does for the middle steps.
describe("TutorialCoach: the middle steps", () => {
  const rest: TutorialStep = {
    id: "read-hand",
    n: 2,
    kind: "action",
    title: "Read your hand",
    body: "Rest the pointer on a card.",
    anchor: { label: L.yourHand },
    hover: { ms: 600, event: "hand-hovered" },
  };
  const view = { id: "g", seats: [], turn: { seq: 1 } } as unknown as GameView;

  function mount(steps: TutorialStep[], extra: Record<string, unknown> = {}) {
    const log = vi.fn();
    const r = render(
      TutorialCoach as never,
      { view, viewerID: "me", steps, log, onSize: () => {}, ...extra } as never,
    ) as unknown as Rendered<Record<string, unknown>>;
    click(button(r.container, "Start")!);
    return { ...r, log };
  }
  const handEl = () => {
    document.body.insertAdjacentHTML("beforeend", '<div aria-label="your hand" id="hand"></div>');
    place(document.getElementById("hand")!, 340, 670, 570, 112);
  };
  const on = (m: { container: HTMLElement }, text: string) =>
    m.container.textContent?.includes(text);

  it("completes a hover step once the pointer has rested on the anchor for 600ms", () => {
    vi.useFakeTimers();
    handEl();
    let over = false;
    const m = mount([WELCOME, rest, HANDOFF], { canHover: true, isHovering: () => over });
    // The bus saying the pointer arrived is not a rest: a pointer
    // crossing the hand on its way to the dock fires it too.
    emit("hand-hovered");
    flushSync();
    expect(on(m, "Read your hand")).toBe(true);
    over = true;
    vi.advanceTimersByTime(400);
    over = false;
    vi.advanceTimersByTime(200);
    flushSync();
    expect(on(m, "Read your hand")).toBe(true);
    // Rest again: the clock started over when the pointer left.
    over = true;
    vi.advanceTimersByTime(500);
    flushSync();
    expect(on(m, "Read your hand")).toBe(true);
    vi.advanceTimersByTime(300);
    flushSync();
    expect(on(m, "That is the whole interface")).toBe(true);
    expect(m.log).not.toHaveBeenCalled();
  });

  it("takes the step's own event as the whole gesture on a device with no hover", () => {
    vi.useFakeTimers();
    handEl();
    const m = mount([WELCOME, rest, HANDOFF], { canHover: false, isHovering: () => false });
    emit("pile-hovered");
    flushSync();
    expect(on(m, "Read your hand")).toBe(true);
    emit("hand-hovered");
    flushSync();
    expect(on(m, "That is the whole interface")).toBe(true);
  });

  it("moves a hover step on after a read on a device with no hover, and logs it", () => {
    vi.useFakeTimers();
    handEl();
    const m = mount([WELCOME, rest, HANDOFF], { canHover: false, touchHoverStepMs: 5_000 });
    vi.advanceTimersByTime(4_900);
    flushSync();
    expect(on(m, "Read your hand")).toBe(true);
    vi.advanceTimersByTime(200);
    flushSync();
    expect(on(m, "That is the whole interface")).toBe(true);
    expect(m.log).toHaveBeenCalledWith(
      "tutorial: step read-hand cannot be hovered on this device; advancing",
    );
  });

  it("starts a phone's read only once the opening roll has dealt the hand (ADR 0125 §5.2)", () => {
    vi.useFakeTimers();
    handEl();
    const rolling = {
      ...view,
      opening_roll: { rounds: [{ seats: [0, 1], rolls: [] }], chooser: -1 },
    } as unknown as GameView;
    const m = mount([WELCOME, rest, HANDOFF], {
      view: rolling,
      canHover: false,
      touchHoverStepMs: 5_000,
    });
    // Neither the phone's timer nor a touch on the empty hand moves it.
    vi.advanceTimersByTime(6_000);
    emit("hand-hovered");
    flushSync();
    expect(on(m, "Read your hand")).toBe(true);
    // The deal: the read starts now, and runs its full length.
    m.setProps({ view: { ...view, turn: { seq: 1, step: "upkeep" } } });
    vi.advanceTimersByTime(4_900);
    flushSync();
    expect(on(m, "Read your hand")).toBe(true);
    vi.advanceTimersByTime(200);
    flushSync();
    expect(on(m, "That is the whole interface")).toBe(true);
  });

  it("shows a detour's copy and spotlights what it names, then the step's own", () => {
    vi.useFakeTimers();
    handEl();
    document.body.insertAdjacentHTML(
      "beforeend",
      '<section aria-label="actions" id="dock"></section>',
    );
    place(document.getElementById("dock")!, 920, 620, 330, 150);
    let upkeep = true;
    const land: TutorialStep = {
      id: "play-land",
      n: 3,
      kind: "action",
      title: "Play a land",
      body: "Click a Forest in your hand.",
      hint: "A land waits for your main phase.",
      anchor: { label: L.yourHand },
      done: (c) => c.event === "ability-menu-opened",
      first: () =>
        upkeep
          ? {
              id: "to-main",
              title: "First, your main phase",
              body: (k) => `Press next (${k.nextKey}) until it reads Main.`,
              anchor: { label: L.actions },
            }
          : null,
    };
    const m = mount([WELCOME, land, HANDOFF]);
    expect(on(m, "First, your main phase")).toBe(true);
    // The player's own binding for `next`: Space by default.
    expect(on(m, "Press next (Space) until it reads Main.")).toBe(true);
    expect(document.querySelector<HTMLElement>(".tutorial-scrim")!.style.left).toBe("914px");
    // A detour says what to do already: no hint over it.
    vi.advanceTimersByTime(25_000);
    flushSync();
    expect(m.container.querySelector(".coach-hint")).toBeNull();
    upkeep = false;
    m.setProps({ view: { ...view, turn: { seq: 1, step: "precombat_main" } } });
    vi.advanceTimersByTime(200);
    flushSync();
    expect(on(m, "Play a land")).toBe(true);
    expect(document.querySelector<HTMLElement>(".tutorial-scrim")!.style.left).toBe("334px");
  });

  it("spotlights a card read off the board, and moves on when there is none", () => {
    vi.useFakeTimers();
    document.body.insertAdjacentHTML("beforeend", '<div data-instance-id="elf-7" id="elf"></div>');
    place(document.getElementById("elf")!, 500, 400, 90, 125);
    let pick: string | null = "elf-7";
    const card: TutorialStep = {
      id: "right-click",
      n: 7,
      kind: "action",
      title: "Abilities live on right-click",
      body: "Right-click this one.",
      anchor: (c) => (c.view && pick ? { cardID: pick } : null),
      done: (c) => c.event === "ability-menu-opened",
    };
    const m = mount([WELCOME, card, HANDOFF]);
    expect(document.querySelector<HTMLElement>(".tutorial-scrim")!.style.left).toBe("494px");
    pick = null;
    vi.advanceTimersByTime(ANCHOR_GRACE_MS + POLL_MS * 2);
    flushSync();
    expect(on(m, "That is the whole interface")).toBe(true);
    expect(m.log).toHaveBeenCalledWith(
      "tutorial: step right-click has no anchor on the page; advancing",
    );
  });

  it("draws a status line read off the board", () => {
    const watch: TutorialStep = {
      id: "watch-bot",
      n: 9,
      kind: "watch",
      title: "The bot takes its turn",
      body: "Watch the strip.",
      status: (c) => ((c.view?.turn as { seq: number }).seq === 2 ? "Waiting for you" : undefined),
      done: () => false,
    };
    const m = mount([WELCOME, watch, HANDOFF]);
    expect(m.container.querySelector(".coach-status")?.textContent).toContain("Bot is thinking");
    m.setProps({ view: { ...view, turn: { seq: 2 } } });
    expect(m.container.querySelector(".coach-status")?.textContent).toContain("Waiting for you");
  });

  it("hands the dock's autopass toggle to the step, which the wire never carries", () => {
    const toggle: TutorialStep = {
      id: "watch-bot",
      n: 9,
      kind: "action",
      title: "Let the bot play",
      body: "Turn on autopass.",
      status: (c) => (c.client?.autopass ? "Autopass is on" : undefined),
      done: () => false,
    };
    const m = mount([WELCOME, toggle, HANDOFF], { autopass: false });
    expect(m.container.querySelector(".coach-status")?.textContent).toContain("Waiting for you");
    m.setProps({ autopass: true });
    expect(m.container.querySelector(".coach-status")?.textContent).toContain("Autopass is on");
  });
});

// ADR 0125 §5: the roll, the commander's eventless hover, what a
// completed step teaches, and the count out of fourteen.
describe("TutorialCoach: the refreshed tutorial", () => {
  const view = {
    id: "g",
    seats: [
      { id: "me", name: "Player", seat: 0 },
      { id: "bot", name: "Practice Bot", seat: 1 },
    ],
    turn: { seq: 1 },
  } as unknown as GameView;
  const rolling = (rolls: Array<{ seat: number; result: number }>, chooser?: number) =>
    ({
      ...view,
      opening_roll: {
        rounds: [{ seats: [0, 1], rolls }],
        ...(chooser === undefined ? {} : { chooser }),
      },
    }) as unknown as GameView;

  function mount(steps: TutorialStep[], extra: Record<string, unknown> = {}) {
    const log = vi.fn();
    const r = render(
      TutorialCoach as never,
      { view, viewerID: "me", steps, log, onSize: () => {}, ...extra } as never,
    ) as unknown as Rendered<Record<string, unknown>>;
    return { ...r, log };
  }
  const add = (html: string, id: string, l: number) => {
    document.body.insertAdjacentHTML("beforeend", html);
    place(document.getElementById(id)!, l, 100, 200, 100);
  };
  const seen = () => get(settings).help.seen;

  it("counts out of fourteen on the real script", () => {
    const m = mount(TUTORIAL_STEPS);
    expect(m.container.textContent).toContain("1 / 14");
    expect(m.container.querySelectorAll(".seg")).toHaveLength(14);
  });

  it("walks the roll during the roll: Roll, then the banner, then the chooser's sheet", () => {
    vi.useFakeTimers();
    add('<div aria-label="roll for the first turn" id="roll"></div>', "roll", 900);
    add('<div aria-label="opening roll" id="banner"></div>', "banner", 20);
    add('<div aria-label="choose who takes the first turn" id="sheet"></div>', "sheet", 700);
    const m = mount([WELCOME, OPENING_ROLL, HANDOFF], { view: rolling([]) });
    click(button(m.container, "Start")!);
    flushSync();
    expect(m.container.textContent).toContain("Roll for the first turn");
    expect(document.querySelector<HTMLElement>(".tutorial-scrim")!.style.left).toBe("894px");
    // Rolled; the bot has not: the banner, and who the table waits for.
    m.setProps({ view: rolling([{ seat: 0, result: 15 }]) });
    vi.advanceTimersByTime(POLL_MS);
    flushSync();
    expect(document.querySelector<HTMLElement>(".tutorial-scrim")!.style.left).toBe("14px");
    expect(m.container.querySelector(".coach-status")?.textContent).toContain(
      "Waiting for Practice Bot to roll",
    );
    // Won: the detour onto the sheet.
    m.setProps({
      view: rolling(
        [
          { seat: 0, result: 15 },
          { seat: 1, result: 2 },
        ],
        0,
      ),
    });
    vi.advanceTimersByTime(POLL_MS);
    flushSync();
    expect(m.container.textContent).toContain("You won the roll");
    expect(document.querySelector<HTMLElement>(".tutorial-scrim")!.style.left).toBe("694px");
    // The choice closes the roll: the step is complete, and taught its hint.
    m.setProps({ view: { ...view, turn: { seq: 1, step: "upkeep" } } });
    flushSync();
    expect(m.container.textContent).toContain("That is the whole interface");
    expect(seen()["table.opening-roll"]).toBe(1);
    expect(m.log).not.toHaveBeenCalled();
  });

  it("completes the commander's hover with a mouse, and marks its hint seen", () => {
    vi.useFakeTimers();
    document.body.insertAdjacentHTML(
      "beforeend",
      '<section aria-label="your board"><div aria-label="castable from other zones" id="cz"></div></section>',
    );
    place(document.getElementById("cz")!, 40, 500, 120, 170);
    let over = false;
    const m = mount([WELCOME, COMMANDER, HANDOFF], {
      view: { ...view, seats: [{ id: "me", name: "Player", seat: 0, command: { count: 1 } }] },
      canHover: true,
      isHovering: () => over,
    });
    click(button(m.container, "Start")!);
    expect(m.container.textContent).toContain("Your commander");
    over = true;
    vi.advanceTimersByTime(800);
    flushSync();
    expect(m.container.textContent).toContain("That is the whole interface");
    expect(seen()["table.commander"]).toBe(1);
  });

  it("moves the commander's step on after a read on a touch screen, and teaches nothing", () => {
    vi.useFakeTimers();
    document.body.insertAdjacentHTML(
      "beforeend",
      '<section aria-label="your board"><div aria-label="castable from other zones" id="cz"></div></section>',
    );
    place(document.getElementById("cz")!, 40, 500, 120, 170);
    const m = mount([WELCOME, COMMANDER, HANDOFF], {
      view: { ...view, seats: [{ id: "me", name: "Player", seat: 0, command: { count: 1 } }] },
      canHover: false,
      touchHoverStepMs: 5_000,
    });
    click(button(m.container, "Start")!);
    // No bus event stands for this step: none of the three completes it.
    emit("hand-hovered");
    emit("pile-hovered");
    emit("ability-menu-opened");
    flushSync();
    expect(m.container.textContent).toContain("Your commander");
    vi.advanceTimersByTime(5_100);
    flushSync();
    expect(m.container.textContent).toContain("That is the whole interface");
    expect(m.log).toHaveBeenCalledWith(
      "tutorial: step commander cannot be hovered on this device; advancing",
    );
    expect(seen()["table.commander"]).toBeUndefined();
  });

  it("marks the hand-off's hints on Finish, and nothing on Skip tutorial", () => {
    const m = mount([WELCOME, HANDOFF]);
    click(button(m.container, "Skip tutorial")!);
    expect(seen()).toEqual({});
    cleanup();
    const n = mount([WELCOME, HANDOFF]);
    click(button(n.container, "Start")!);
    click(button(n.container, "Finish")!);
    expect(seen()).toEqual({ "table.more": 1, "table.shortcuts": 1 });
  });

  it("hands a completed step's teaches to onTaught, and a skipped one's to nobody", () => {
    const taught: string[][] = [];
    const step: TutorialStep = {
      ...ON_THE_STACK,
      anchor: undefined,
      done: (c) => c.event === "ability-menu-opened",
      cannot: undefined,
    };
    const m = mount([WELCOME, step, { ...step, id: "again" }, HANDOFF], {
      onTaught: (ids: readonly string[]) => taught.push([...ids]),
    });
    click(button(m.container, "Start")!);
    click(button(m.container, "Skip step")!);
    expect(taught).toEqual([]);
    emit("ability-menu-opened");
    flushSync();
    expect(taught).toEqual([["table.stack"]]);
    expect(seen()).toEqual({});
  });
});

// jsdom applies no component CSS, so the rules that make the scrim a
// suggestion rather than a lock, and the placement, are read from source
// (the way actionDock.render.test.ts reads the dock's).
describe("the coach's CSS contract", () => {
  const src = (p: string) => readFileSync(join(process.cwd(), p), "utf8");
  const rule = (css: string, sel: string) => css.slice(css.indexOf(`${sel} {`)).split("}")[0];

  it("never lets the scrim take a pointer event", () => {
    const scrim = rule(
      src("src/lib/components/tutorial/TutorialScrim.svelte"),
      "  .tutorial-scrim",
    );
    expect(scrim).toContain("pointer-events: none");
    expect(scrim).toContain("box-shadow: 0 0 0 9999px");
    expect(scrim).toContain("position: fixed");
    expect(scrim).toContain("z-index: 56");
  });

  it("docks the card bottom-left at the dock's inset, and on the dock bar on a phone", () => {
    const css = src("src/lib/components/tutorial/TutorialCoach.svelte");
    const slot = rule(css, "  .coach-slot");
    expect(slot).toContain("left: var(--dock-inset, 15px)");
    expect(slot).toContain("bottom: var(--dock-inset, 15px)");
    expect(slot).toContain("width: clamp(280px, 24vw, 340px)");
    expect(slot).toContain("z-index: 57");
    const phone = css.slice(css.indexOf("@media (max-width: 599px)"));
    expect(rule(phone, "    .coach-slot")).toContain("bottom: calc(var(--dock-h, 0px) + 6px)");
  });
});

// The board's half of the placement: the viewer's own panel keeps an
// empty cell the card's size at the left of its bottom row, and ONLY while
// the coach shows. Without it the panel is exactly what it was.
describe("the coach card's cell in the self panel", () => {
  const src = (p: string) => readFileSync(join(process.cwd(), p), "utf8");
  const rule = (css: string, sel: string) => css.slice(css.indexOf(`${sel} {`)).split("}")[0];
  const zone = (kind: string, owner?: string) => ({ kind, owner, count: 0, cards: [] });
  const seat = (id: string, name: string, n: number) =>
    ({
      id,
      name,
      seat: n,
      life: 40,
      library: zone("library", id),
      hand: zone("hand", id),
      graveyard: zone("graveyard", id),
      command: zone("command", id),
      commander_damage: {},
      life_history: [],
      mana_pool: [],
    }) as unknown as PlayerView;
  const gameView = () =>
    ({
      id: "g1",
      state: "active",
      seats: [seat("me", "Me", 0), seat("opp", "Opp", 1)],
      battlefield: zone("battlefield"),
      stack: zone("stack"),
      exile: zone("exile"),
      stack_items: [],
      pending_triggers: [],
      pending_choices: [],
      turn: {
        seq: 5,
        number: 3,
        active_seat: 0,
        priority_holder: 0,
        phase: "precombat_main",
        step: "precombat_main",
      },
      mulligans_open: false,
    }) as unknown as GameView;

  function mountBoard(coached: boolean | undefined) {
    return render(
      Board as never,
      {
        view: gameView(),
        viewerID: "me",
        isAdmin: false,
        sendAction: () => {},
        combatMode: "idle",
        selectedCombatCardID: null,
        onSelectCombatCard: () => {},
        onDeclareAttack: () => {},
        onDeclareBlock: () => {},
        docked: true,
        ...(coached === undefined ? {} : { coached }),
      } as never,
    ).container;
  }

  it("is the first cell of the viewer's own bottom row, before the hand", () => {
    const c = mountBoard(true);
    const row = c.querySelector('.slot[data-pos="self"] .panel .grid-bottom')!;
    expect(row.firstElementChild?.classList.contains("coach-spacer")).toBe(true);
    expect(row.querySelector(".coach-spacer + .hand-zone")).not.toBeNull();
    expect(c.querySelectorAll(".coach-spacer")).toHaveLength(1);
  });

  it("does not exist when no tutorial is running", () => {
    expect(mountBoard(undefined).querySelector(".coach-spacer")).toBeNull();
    cleanup();
    expect(mountBoard(false).querySelector(".coach-spacer")).toBeNull();
  });

  it("takes the card's width from the hand row and no height, and goes on a phone", () => {
    const panel = src("src/lib/components/board/PlayerPanel.svelte");
    const spacer = rule(panel, "  .coach-spacer");
    expect(spacer).toContain("flex: 0 0 var(--coach-w, 0px)");
    // #1081 follow-up: it stretches to the row the hand and the dock
    // already make, so the battlefield rows keep their height. The e2e
    // spec tutorial-layout.spec.ts measures it on a real table.
    expect(spacer).toContain("align-self: stretch");
    expect(spacer).not.toContain("height");
    const phone = panel.slice(panel.indexOf("@media (max-width: 599px)"));
    expect(phone.slice(0, phone.indexOf("display: none"))).toContain(".coach-spacer");
    // The play area grows by the strip only under .has-coach, which
    // Game.svelte sets only while the card shows.
    const game = src("src/routes/Game.svelte");
    expect(rule(game, "    section.has-dock.has-coach .play-area")).toContain(
      "padding-bottom: calc(var(--dock-h, 0px) + var(--coach-h, 0px) + 12px)",
    );
    expect(rule(game, "    section.has-dock .play-area")).not.toContain("coach");
  });
});
