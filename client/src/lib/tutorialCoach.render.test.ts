// @vitest-environment jsdom
//
// tutorialCoach.render.test.ts — the tutorial's walking skeleton (ADR 0076
// §2.3, §2.4; #1079): the coach card's six states, the scrim, the anchor
// resolver and the missing-anchor self-advance, and steps 1 and 11 end to
// end.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";

import Board from "./components/board/Board.svelte";
import CoachCard from "./components/tutorial/CoachCard.svelte";
import TutorialCoach from "./components/tutorial/TutorialCoach.svelte";
import { ANCHOR_GRACE_MS, POLL_MS, type CoachState, type TutorialStep } from "./tutorial";
import { HANDOFF, WELCOME } from "./tutorialSteps";
import { anchorRect, resolveAnchor } from "./tutorialAnchor";
import { emit } from "./tutorialBus";
import { defaultSettings, settings } from "./settings";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
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
        total: 11,
        title: "Tap a land for mana",
        body: "Click one of the Forests.",
        hint: "A land already on the table.",
        ...extra,
      } as never,
    ).container;

  it("opening: Skip tutorial and Start, no status", () => {
    const c = card("opening", { n: 1 });
    expect(c.textContent).toContain("1 / 11");
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
    const c = card("done", { n: 11 });
    expect(c.textContent).toContain("11 / 11");
    expect(buttons(c)).toEqual(["Replay", "Finish"]);
  });

  it("marks progress: past steps full, this one half", () => {
    const c = card("action");
    expect(c.querySelectorAll(".seg")).toHaveLength(11);
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
    expect(resolveAnchor({ label: "lands" })?.id).toBe("opp-lands");
    expect(resolveAnchor({ label: "lands", within: "your board" })?.id).toBe("my-lands");
    expect(resolveAnchor({ label: "your hand" })?.id).toBe("hand");
  });

  it("finds a card by instance id", () => {
    table();
    expect(resolveAnchor({ cardID: "c-1" })?.id).toBe("elves");
  });

  it("is null when the label, its container or the card is not there", () => {
    table();
    expect(resolveAnchor({ label: "attention" })).toBeNull();
    expect(resolveAnchor({ label: "lands", within: "no such board" })).toBeNull();
    expect(resolveAnchor({ cardID: "gone" })).toBeNull();
  });

  it("unions several anchors, and treats an element with no area as missing", () => {
    table();
    place(document.getElementById("hand")!, 100, 600, 400, 120);
    place(document.getElementById("elves")!, 200, 300, 80, 110);
    place(document.getElementById("my-lands")!, 0, 0, 160, 0);
    expect(anchorRect([{ label: "your hand" }, { cardID: "c-1" }])).toEqual({
      left: 100,
      top: 300,
      width: 400,
      height: 420,
    });
    expect(anchorRect([{ label: "lands", within: "your board" }])).toBeNull();
    expect(anchorRect([{ label: "nowhere" }, { cardID: "c-1" }])).toEqual({
      left: 200,
      top: 300,
      width: 80,
      height: 110,
    });
  });
});

describe("TutorialCoach", () => {
  const spot: TutorialStep = {
    id: "read-hand",
    n: 2,
    kind: "action",
    title: "Read your hand",
    body: "Hover the fan to lift it.",
    anchor: { label: "your hand" },
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

  it("walks steps 1 and 11: Start, then Finish hides the card and frees its cell", () => {
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

  it("is sized from --coach-w / --coach-h, and goes on a phone", () => {
    const panel = src("src/lib/components/board/PlayerPanel.svelte");
    const spacer = rule(panel, "  .coach-spacer");
    expect(spacer).toContain("flex: 0 0 var(--coach-w, 0px)");
    expect(spacer).toContain("height: var(--coach-h, 0px)");
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
