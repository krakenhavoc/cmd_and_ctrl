// @vitest-environment jsdom
//
// ADR 0119 §3 — the linger, rendered. The rules are pinned as pure
// functions in stackLinger.test.ts; this file pins what only exists once
// the overlay is mounted: a departed card drawn where it was with its
// badge, aria-hidden, the announcer's line, the 200 ms fade with no
// badge, priming on mount and on a prime-key change, and the flight.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./animations", async (importOriginal) => {
  const real = await importOriginal<typeof import("./animations")>();
  return { ...real, flyTo: vi.fn(() => Promise.resolve()), etbPulse: vi.fn() };
});

import StackLinger from "./components/board/StackLinger.svelte";
import Board from "./components/board/Board.svelte";
import { STACK_STYLES } from "./stackLane";
import { flyTo } from "./animations";
import { updateSettings } from "./settings";
import { LINGER_FADE_MS, LINGER_MS } from "./stackLinger";
import type {
  ActionType,
  CardView,
  GameView,
  LogEvent,
  PlayerView,
  StackItemView,
} from "./protocol";
import { render, cleanup } from "./test/render.svelte";
import { flushSync } from "svelte";

const ME = "me";
const OPP = "opp";

const realRect = Element.prototype.getBoundingClientRect;

const rect = (left: number, top: number, width: number, height: number) =>
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

let boardEl: HTMLDivElement;

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}
const g = globalThis as Record<string, unknown>;

beforeEach(() => {
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  vi.useFakeTimers();
  vi.mocked(flyTo).mockClear();
  Element.prototype.getBoundingClientRect = function (this: Element) {
    if (this.classList.contains("board")) return rect(0, 0, 1400, 900);
    if (this.hasAttribute("data-stack-item-id")) return rect(12, 300, 240, 335);
    if (this.hasAttribute("data-pile")) return rect(1300, 800, 50, 70);
    return rect(0, 0, 0, 0);
  };
  boardEl = document.createElement("div");
  boardEl.className = "board";
  document.body.appendChild(boardEl);
});

afterEach(() => {
  cleanup();
  boardEl.remove();
  Element.prototype.getBoundingClientRect = realRect;
  updateSettings("animations", "enabled", true);
  updateSettings("animations", "cardPlay", true);
  updateSettings("display", "stackStyle", "pile");
  vi.useRealTimers();
});

const zone = (kind: string, cards: CardView[] = []) => ({ kind, count: cards.length, cards });

const card = (id: string, name: string, owner: string) =>
  ({
    instance_id: id,
    name,
    owner,
    controller: owner,
    known_by_you: true,
    scryfall_id: `sf-${id}`,
  }) as CardView;

const bolt = card("bolt", "Lightning Bolt", OPP);
const cs = card("cs", "Counterspell", ME);

const spell = (id: string, controller: string): StackItemView =>
  ({ id, kind: "spell", controller, owner: controller, source_card_id: id }) as StackItemView;

const seat = (id: string, name: string, n: number, grave: CardView[] = []): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library"),
    hand: zone("hand"),
    graveyard: zone("graveyard", grave),
    command: zone("command"),
  }) as unknown as PlayerView;

function gameView(opts: {
  items?: StackItemView[];
  stackCards?: CardView[];
  log?: LogEvent[];
  oppGrave?: CardView[];
}): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Bot 2", 1, opts.oppGrave)],
    battlefield: zone("battlefield"),
    stack: zone("stack", opts.stackCards ?? []),
    exile: zone("exile"),
    stack_items: opts.items ?? [],
    pending_triggers: [],
    turn: { number: 3, active_seat: 1, priority_holder: 0, phase: "main1", step: "main" },
    log: opts.log ?? [],
  } as unknown as GameView;
}

const cast: LogEvent = {
  seq: 10,
  kind: "cast",
  seat: 1,
  card_id: "bolt",
  text: "Bot 2 cast Lightning Bolt",
};
const resolved: LogEvent = {
  seq: 11,
  kind: "resolve",
  seat: 1,
  card_id: "bolt",
  stack_item_id: "bolt",
  text: "Lightning Bolt resolved",
};

// Draw a stack item the way every style does: an element carrying its ID.
function drawItem(id: string): HTMLElement {
  const el = document.createElement("div");
  el.dataset.stackItemId = id;
  boardEl.appendChild(el);
  return el;
}

const onStack = () => gameView({ items: [spell("bolt", OPP)], stackCards: [bolt], log: [cast] });

function mount(view = onStack(), beatsPrimeKey = "live:false") {
  return render(StackLinger, { view, viewerID: ME, boardEl, beatsPrimeKey });
}

// Advance the fake clock and flush what the timers changed.
function tick(ms: number): void {
  vi.advanceTimersByTime(ms);
  flushSync();
}

const ghost = (c: HTMLElement, id = "bolt") =>
  c.querySelector<HTMLElement>(`[data-stack-linger="${id}"]`);
const announcer = (c: HTMLElement) =>
  c.querySelector("[data-stack-linger-announcer]")?.textContent?.trim();

describe("the linger", () => {
  it("draws a resolved spell where it was, badged, aria-hidden, and announces it", () => {
    const item = drawItem("bolt");
    const r = mount();
    expect(ghost(r.container)).toBeNull();

    item.remove();
    r.setProps({ view: gameView({ log: [cast, resolved] }) });

    const g = ghost(r.container)!;
    expect(g).not.toBeNull();
    expect(g.dataset.lingerOutcome).toBe("resolved");
    expect(g.style.left).toBe("12px");
    expect(g.style.top).toBe("300px");
    expect(g.querySelector(".badge")?.textContent).toContain("Resolved");
    expect(g.closest("[aria-hidden='true']")).not.toBeNull();
    // Never inside a labelled region: the stack's label stays gone.
    expect(g.closest("[aria-label]")).toBeNull();
    expect(announcer(r.container)).toBe("Lightning Bolt resolved");
  });

  it("names what countered it, and lingers a burst in resolution order", () => {
    drawItem("bolt");
    const csEl = drawItem("cs");
    const both = gameView({
      items: [spell("bolt", OPP), spell("cs", ME)],
      stackCards: [bolt, cs],
      log: [cast],
    });
    const r = mount(both);
    boardEl.innerHTML = "";
    void csEl;
    r.setProps({
      view: gameView({
        log: [
          cast,
          {
            seq: 11,
            kind: "counter",
            seat: 0,
            card_id: "cs",
            target: "bolt",
            stack_item_id: "bolt",
            text: "Counterspell countered Lightning Bolt",
          },
          { seq: 12, kind: "resolve", seat: 0, card_id: "cs", stack_item_id: "cs", text: "" },
        ],
      }),
    });
    // A burst lingers one at a time, in resolution order.
    const first = ghost(r.container, "bolt")!;
    expect(first.dataset.lingerOutcome).toBe("countered");
    expect(first.querySelector(".badge")?.textContent).toContain("by Counterspell");
    expect(ghost(r.container, "cs")).toBeNull();
    expect(announcer(r.container)).toBe(
      "Lightning Bolt was countered by Counterspell. Counterspell resolved",
    );
    tick(LINGER_MS / 2);
    expect(ghost(r.container, "cs")?.dataset.lingerOutcome).toBe("resolved");
  });

  it("titles a fizzled badge with CR 608.2b", () => {
    const item = drawItem("bolt");
    const r = mount();
    item.remove();
    r.setProps({
      view: gameView({
        log: [
          cast,
          { seq: 11, kind: "fizzle", seat: 1, card_id: "bolt", stack_item_id: "bolt", text: "" },
        ],
      }),
    });
    const badge = ghost(r.container)!.querySelector<HTMLElement>(".badge")!;
    expect(badge.textContent).toContain("Fizzled");
    expect(badge.title).toBe("countered on resolution: no legal targets (CR 608.2b)");
  });

  it("fades an item that left with no log entry in 200 ms, with no badge", () => {
    const item = drawItem("bolt");
    const r = mount();
    item.remove();
    r.setProps({ view: gameView({ log: [cast] }) });
    const g = ghost(r.container)!;
    expect(g.dataset.lingerOutcome).toBe("none");
    expect(g.querySelector(".badge")).toBeNull();
    expect(g.classList.contains("fading")).toBe(true);
    expect(announcer(r.container)).toBe("");
    tick(LINGER_FADE_MS);
    expect(ghost(r.container)).toBeNull();
  });

  it("primes on mount: entries already on the wire linger nothing", () => {
    const r = mount(gameView({ log: [cast, resolved] }));
    expect(ghost(r.container)).toBeNull();
    expect(announcer(r.container)).toBe("");
  });

  it("primes again on a reconnect or replay toggle, without lingering", () => {
    const item = drawItem("bolt");
    const r = mount();
    item.remove();
    r.setProps({ view: gameView({ log: [cast, resolved] }), beatsPrimeKey: "offline:false" });
    expect(ghost(r.container)).toBeNull();
  });

  it("shows the badge with animations off, and drops it without motion", () => {
    updateSettings("animations", "enabled", false);
    const item = drawItem("bolt");
    const r = mount();
    item.remove();
    r.setProps({ view: gameView({ log: [cast, resolved], oppGrave: [bolt] }) });
    expect(ghost(r.container)?.dataset.lingerOutcome).toBe("resolved");
    tick(LINGER_MS);
    expect(flyTo).not.toHaveBeenCalled();
    expect(ghost(r.container)).toBeNull();
  });

  it("flies to its owner's graveyard pile after the linger", async () => {
    const pile = document.createElement("button");
    pile.setAttribute("data-pile", "graveyard");
    pile.setAttribute("data-pile-owner", OPP);
    boardEl.appendChild(pile);
    const item = drawItem("bolt");
    const r = mount();
    item.remove();
    r.setProps({ view: gameView({ log: [cast, resolved], oppGrave: [bolt] }) });
    tick(LINGER_MS - 1);
    expect(flyTo).not.toHaveBeenCalled();
    tick(1);
    expect(flyTo).toHaveBeenCalledTimes(1);
    const [el, to] = vi.mocked(flyTo).mock.calls[0];
    expect(el).toBe(ghost(r.container));
    expect(to.scale).toBeCloseTo(70 / 335);
    await Promise.resolve();
    await Promise.resolve();
    flushSync();
    expect(ghost(r.container)).toBeNull();
  });

  it("fades in place when cardPlay is off", () => {
    updateSettings("animations", "cardPlay", false);
    const item = drawItem("bolt");
    const r = mount();
    item.remove();
    r.setProps({ view: gameView({ log: [cast, resolved], oppGrave: [bolt] }) });
    tick(LINGER_MS);
    expect(flyTo).not.toHaveBeenCalled();
    expect(ghost(r.container)?.classList.contains("fading")).toBe(true);
    tick(LINGER_FADE_MS);
    expect(ghost(r.container)).toBeNull();
  });
});

describe("the linger on a real board, in every stack style", () => {
  function mountBoard(view: GameView) {
    return render(
      Board as never,
      {
        view,
        viewerID: ME,
        isAdmin: false,
        sendAction: (_type: ActionType) => {},
        combatMode: "idle",
        selectedCombatCardID: null,
        onSelectCombatCard: () => {},
        onDeclareAttack: () => {},
        onDeclareBlock: () => {},
      } as never,
    );
  }

  for (const style of STACK_STYLES) {
    it(`lingers a resolved spell drawn by the ${style} style`, () => {
      updateSettings("display", "stackStyle", style);
      const b = mountBoard(onStack());
      const board = b.container.querySelector<HTMLElement>(".board")!;
      expect(board.querySelector('[data-stack-item-id="bolt"]')).not.toBeNull();

      b.setProps({ view: gameView({ log: [cast, resolved], oppGrave: [bolt] }) } as never);

      expect(board.querySelector('[data-stack-item-id="bolt"]')).toBeNull();
      const g = board.querySelector<HTMLElement>('[data-stack-linger="bolt"]');
      expect(g?.dataset.lingerOutcome).toBe("resolved");
      // The board-layout contract: no stack label on an empty stack.
      expect(board.querySelector('[aria-label^="stack: "]')).toBeNull();
      expect(b.container.querySelector("[data-stack-linger-announcer]")?.textContent?.trim()).toBe(
        "Lightning Bolt resolved",
      );
    });
  }
});
