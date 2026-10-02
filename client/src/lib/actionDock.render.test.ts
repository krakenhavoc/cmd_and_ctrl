// @vitest-environment jsdom
//
// actionDock.render.test.ts — ADR 0111 Delivery PR 2 (S56, #1958). The
// action dock owns the screen's bottom-right corner: PhaseDisplay is its
// header, hold / autopass / bluff its toggles row, Pass turn and `next`
// its action bar. These pin what the e2e suite and the tutorial rely on
// (names, classes, disabled-not-hidden, one copy of each), the key rules
// for `next`, and the layout contract the self panel keeps with it.

import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";

import ActionDock from "./components/board/ActionDock.svelte";
import Board from "./components/board/Board.svelte";
import { _resetForTests as resetBluff } from "./bluff";
import { defaultSettings, settings } from "./settings";
import type { CardView, GameView, PlayerView, TurnView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

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
  resetBluff();
});
afterEach(cleanup);

const ME = "me";
const OPP = "opp";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, name: string, n: number): PlayerView =>
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

const turn = (over: Partial<TurnView> = {}): TurnView => ({
  seq: 5,
  number: 3,
  active_seat: 0,
  priority_holder: 0,
  phase: "precombat_main",
  step: "precombat_main",
  ...over,
});

const gameView = (over: Partial<GameView> = {}): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Opp", 1)],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: turn(),
    mulligans_open: false,
    ...over,
  }) as unknown as GameView;

interface Calls {
  pass: number;
  passTurn: number;
  autopass: number;
  sizes: Array<[number, number]>;
}

function mountDock(props: Record<string, unknown> = {}) {
  const calls: Calls = { pass: 0, passTurn: 0, autopass: 0, sizes: [] };
  const r = render(
    ActionDock as never,
    {
      view: gameView(),
      viewerHasPriority: true,
      viewerIsActive: true,
      activePlayerName: "Me",
      autopassEnabled: false,
      onPassPriority: () => calls.pass++,
      onPassTurn: () => calls.passTurn++,
      onToggleAutopass: () => calls.autopass++,
      onSize: (w: number, h: number) => calls.sizes.push([w, h]),
      ...props,
    } as never,
  );
  const q = <T extends HTMLElement = HTMLElement>(sel: string) => r.container.querySelector<T>(sel);
  return { ...r, calls, q };
}

// The accessible name the e2e suite matches: text content without the
// aria-hidden key cap. ("next" is matched with `exact: true`.)
function accessibleName(el: Element): string {
  const copy = el.cloneNode(true) as Element;
  copy.querySelectorAll('[aria-hidden="true"]').forEach((n) => n.remove());
  return (copy.textContent ?? "").trim();
}

function buttonsNamed(root: ParentNode, name: string): HTMLButtonElement[] {
  return [...root.querySelectorAll<HTMLButtonElement>("button")].filter(
    (b) => accessibleName(b) === name,
  );
}

describe("ActionDock", () => {
  it("is one region named actions, with the header, the toggles and the action bar", () => {
    const d = mountDock();
    const dock = d.q("section.action-dock")!;
    expect(dock.getAttribute("aria-label")).toBe("actions");
    // The header: PhaseDisplay, with the e2e suite's two classes.
    expect(dock.querySelector('[aria-label="turn and phase indicator"]')).not.toBeNull();
    expect(dock.querySelectorAll(".turn-no")).toHaveLength(1);
    expect(dock.querySelectorAll(".step-label")).toHaveLength(1);

    // The toggles row: hold, autopass, bluff, in that order.
    const toggles = dock.querySelector('[role="group"][aria-label="priority controls"]')!;
    const names = [...toggles.querySelectorAll("button")].map(
      (b) => b.getAttribute("aria-label") ?? accessibleName(b),
    );
    expect(names).toEqual(["hold", "autopass", "bluff", "bluff options"]);
    const autopass = toggles.querySelector<HTMLButtonElement>("button.action.autopass")!;
    expect(autopass.getAttribute("aria-pressed")).toBe("false");

    // The action bar: Pass turn on the left, next — the primary — on the right.
    const bar = dock.querySelector(".dock-bar")!;
    const bar_ = [...bar.querySelectorAll("button")];
    expect(bar_.map(accessibleName)).toEqual(["Pass turn", "next"]);
    expect(bar_[1].classList.contains("primary")).toBe(true);
    // `next` and the toggles are not inside each other's group.
    expect(toggles.contains(bar_[1])).toBe(false);
  });

  it("draws exactly one next and one Pass turn", () => {
    const d = mountDock();
    expect(buttonsNamed(d.container, "next")).toHaveLength(1);
    expect(buttonsNamed(d.container, "Pass turn")).toHaveLength(1);
  });

  it("keeps next and Pass turn rendered but disabled when they can't be pressed", () => {
    const d = mountDock({ viewerHasPriority: false, viewerIsActive: false });
    const next = buttonsNamed(d.container, "next")[0]!;
    const passTurn = buttonsNamed(d.container, "Pass turn")[0]!;
    expect(next.disabled).toBe(true);
    expect(next.title).toBe("you don't hold priority");
    expect(passTurn.disabled).toBe(true);
    expect(passTurn.title).toBe("Me is the active player");
    // Not gold without priority.
    expect(next.classList.contains("viewer-priority")).toBe(false);

    d.setProps({ viewerHasPriority: true, viewerIsActive: true } as never);
    expect(next.disabled).toBe(false);
    expect(passTurn.disabled).toBe(false);
    expect(next.classList.contains("viewer-priority")).toBe(true);
  });

  it("wires next, Pass turn and autopass to their handlers", () => {
    const d = mountDock();
    click(buttonsNamed(d.container, "next")[0]!);
    click(buttonsNamed(d.container, "Pass turn")[0]!);
    click(d.q("button.action.autopass")!);
    expect(d.calls).toMatchObject({ pass: 1, passTurn: 1, autopass: 1 });
  });

  it("Enter on the focused next presses it, once, and goes no further", () => {
    const d = mountDock();
    const next = buttonsNamed(d.container, "next")[0]!;
    let reachedWindow = 0;
    const onWindow = (e: KeyboardEvent) => {
      if (e.key === "Enter") reachedWindow++;
    };
    window.addEventListener("keydown", onWindow);
    try {
      next.focus();
      expect(document.activeElement).toBe(next);
      const ev = new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true });
      next.dispatchEvent(ev);
      expect(d.calls.pass).toBe(1);
      // preventDefault stops the browser's own Enter-click, so a real
      // keypress passes exactly once; stopPropagation keeps the
      // targeting walk's window-level Enter from also firing.
      expect(ev.defaultPrevented).toBe(true);
      expect(reachedWindow).toBe(0);
    } finally {
      window.removeEventListener("keydown", onWindow);
    }
  });

  it("Escape never passes, and focus does not move to next on a priority frame", () => {
    const d = mountDock({ viewerHasPriority: false });
    const next = buttonsNamed(d.container, "next")[0]!;
    d.setProps({ viewerHasPriority: true } as never);
    expect(document.activeElement).not.toBe(next);
    next.focus();
    next.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    expect(d.calls.pass).toBe(0);
  });

  it("prints the bound keys: a Space cap on next, and aria-keyshortcuts on every button", () => {
    const d = mountDock();
    const next = buttonsNamed(d.container, "next")[0]!;
    expect(next.querySelector("kbd.cap")?.textContent).toBe("Space");
    expect(next.querySelector("kbd.cap")?.getAttribute("aria-hidden")).toBe("true");
    expect(next.getAttribute("aria-keyshortcuts")).toBe("Space");
    expect(next.title).toContain("(Space)");
    expect(buttonsNamed(d.container, "Pass turn")[0]!.getAttribute("aria-keyshortcuts")).toBe("T");
    expect(d.q("button.action.hold")!.getAttribute("aria-keyshortcuts")).toBe("H");
    expect(d.q("button.action.autopass")!.getAttribute("aria-keyshortcuts")).toBe("Shift+P");
    expect(d.q("button.bluff-main")!.getAttribute("aria-keyshortcuts")).toBe("B");

    // Shortcuts off: no cap, no advertised keys.
    settings.update((s) => ({ ...s, shortcuts: { ...s.shortcuts, enabled: false } }));
    flushSync();
    expect(next.querySelector("kbd.cap")).toBeNull();
    expect(next.hasAttribute("aria-keyshortcuts")).toBe(false);
    expect(accessibleName(next)).toBe("next");
  });

  it("says what next will do, from public state, only while you hold priority", () => {
    const d = mountDock();
    expect(d.q(".dock-status")!.textContent?.trim()).toBe("passing moves to Begin Combat");
    d.setProps({ viewerHasPriority: false } as never);
    expect(d.q(".dock-status")!.textContent?.trim()).toBe("");
    // The status row is there either way, so the dock keeps its height.
    expect(d.q(".dock-status")).not.toBeNull();
  });

  it("puts the CR 732 loop notice in the status line and marks autopass paused", () => {
    const d = mountDock({ autopassEnabled: true, loopNotice: "Mirror Engine — loop." });
    const notice = d.q(".dock-status .loop-notice")!;
    expect(notice.getAttribute("role")).toBe("status");
    expect(notice.textContent).toContain("Mirror Engine — loop. Autopass paused.");
    const autopass = d.q("button.action.autopass")!;
    expect(autopass.classList.contains("paused")).toBe(true);
    expect(autopass.textContent?.trim()).toBe("autopass ⏸");
  });

  it("reports its size for --dock-w / --dock-h", () => {
    const d = mountDock();
    expect(d.calls.sizes.length).toBeGreaterThan(0);
  });

  it("folds the phase track behind ▴ (shown by CSS on a phone only)", () => {
    const d = mountDock();
    const toggle = d.q<HTMLButtonElement>("button.track-toggle")!;
    expect(toggle.getAttribute("aria-expanded")).toBe("false");
    expect(toggle.getAttribute("aria-controls")).toBe("dock-phase-track");
    expect(d.q("#dock-phase-track")).not.toBeNull();
    click(toggle);
    expect(toggle.getAttribute("aria-expanded")).toBe("true");
    expect(d.q(".phase-display")!.classList.contains("track-open")).toBe(true);
  });
});

// The corner is the dock's: the self panel keeps it clear (owner
// decision 1). jsdom applies no component CSS, so the layout half reads
// the rules themselves, as stackLane.render.test.ts does.
describe("the layout around the dock", () => {
  const src = (p: string) => readFileSync(join(process.cwd(), p), "utf8");
  const rule = (css: string, sel: string) => css.slice(css.indexOf(`${sel} {`)).split("}")[0];

  function mountBoard(docked: boolean) {
    return render(
      Board as never,
      {
        view: gameView(),
        viewerID: ME,
        isAdmin: false,
        sendAction: () => {},
        combatMode: "idle",
        selectedCombatCardID: null,
        onSelectCombatCard: () => {},
        onDeclareAttack: () => {},
        onDeclareBlock: () => {},
        docked,
      } as never,
    ).container;
  }

  it("gives the viewer's own panel, and only it, the dock's empty cell", () => {
    const c = mountBoard(true);
    const self = c.querySelector('.slot[data-pos="self"] .panel')!;
    expect(self.classList.contains("docked")).toBe(true);
    expect(self.querySelector(".grid-bottom > .dock-spacer")).not.toBeNull();
    expect(c.querySelectorAll(".panel.docked")).toHaveLength(1);
    expect(c.querySelectorAll(".dock-spacer")).toHaveLength(1);
  });

  it("leaves the panel as it was with no dock (spectating, replaying)", () => {
    const c = mountBoard(false);
    expect(c.querySelector(".panel.docked")).toBeNull();
    expect(c.querySelector(".dock-spacer")).toBeNull();
  });

  it("stops the rail above the dock and sizes the cell from --dock-w / --dock-h", () => {
    const css = src("src/lib/components/board/PlayerPanel.svelte");
    const docked = rule(css, "  .panel.docked");
    expect(docked).toMatch(/"creatures rail"\s*"middle\s+rail"\s*"bottom\s+bottom"/);
    const spacer = rule(css, "  .dock-spacer");
    expect(spacer).toContain("var(--dock-w, 0px)");
    expect(spacer).toContain("var(--dock-h, 0px)");
  });

  it("keeps the hover zoom and the log drawer clear of it", () => {
    const zoom = rule(src("src/lib/components/board/HoverZoomOverlay.svelte"), "  .overlay");
    expect(zoom).toContain("max-height: calc(100% - 20px - var(--dock-zoom-clear, 0px))");
    const log = rule(src("src/lib/components/board/GameLogPanel.svelte"), "  .log-panel");
    expect(log).toContain("bottom: var(--dock-log-clear, 0px)");
    const game = src("src/routes/Game.svelte");
    const vars = rule(game, "  section.has-dock");
    expect(vars).toContain("--dock-inset: 15px");
    expect(vars).toContain("--dock-zoom-clear: calc(var(--dock-h, 0px) + var(--dock-inset))");
    expect(vars).toContain("--dock-log-clear:");
  });

  it("sits in the corner at z 55, and becomes a full-width bar on a phone", () => {
    const css = src("src/lib/components/board/ActionDock.svelte");
    const dock = rule(css, "  .action-dock");
    expect(dock).toContain("position: absolute");
    expect(dock).toContain("right: var(--dock-inset, 15px)");
    expect(dock).toContain("bottom: var(--dock-inset, 15px)");
    expect(dock).toContain("z-index: 55");
    expect(dock).toContain("width: clamp(300px, 26vw, 380px)");
    const phone = css.slice(css.indexOf("@media (max-width: 599px)"));
    expect(rule(phone, "    .action-dock")).toContain("width: auto");
    expect(rule(phone, "    .dock-btn")).toContain("min-height: 44px");
    const game = src("src/routes/Game.svelte");
    const gamePhone = game.slice(game.indexOf("section.has-dock .play-area"));
    expect(gamePhone.split("}")[0]).toContain("padding-bottom: calc(var(--dock-h, 0px) + 6px)");
  });
});
