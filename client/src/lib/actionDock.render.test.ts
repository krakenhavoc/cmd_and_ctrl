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
import { _resetForTests as resetDock, pushDockRequest } from "./dock";
import { attackRowRequest, blockRequest, combatSelectionRequest } from "./combatDock";
import type { AttackAllPlan } from "./attackAll";
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
  resetDock();
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

    // The toggles row: hold, autopass, bluff and the one Undo, in that order.
    const toggles = dock.querySelector('[role="group"][aria-label="priority controls"]')!;
    const names = [...toggles.querySelectorAll("button")].map(
      (b) => b.getAttribute("aria-label") ?? accessibleName(b),
    );
    expect(names).toEqual(["hold", "autopass", "bluff", "bluff options", "Undo (0 left)"]);
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

// ADR 0111 PR 3: the one Undo, out of the ⋯ menu and the attack row.
describe("the dock's one Undo", () => {
  it("sits last in the toggles row with the count left, and undoes", () => {
    let undone = 0;
    const d = mountDock({ undosLeft: 1, canUndo: true, onUndo: () => undone++ });
    const toggles = d.q('[role="group"][aria-label="priority controls"]')!;
    const undo = toggles.querySelector<HTMLButtonElement>("button.action.undo")!;
    expect(toggles.lastElementChild).toBe(undo);
    expect(undo.getAttribute("aria-label")).toBe("Undo (1 left)");
    expect(undo.querySelector(".undo-count")?.textContent).toBe("1");
    expect(undo.disabled).toBe(false);
    expect(undo.title).toBe("undo your most recent action — 1 left this turn (U)");
    expect(undo.getAttribute("aria-keyshortcuts")).toBe("U");
    click(undo);
    expect(undone).toBe(1);
  });

  it("is disabled, not hidden, when the budget is spent", () => {
    const d = mountDock({ undosLeft: 0, canUndo: false });
    const undo = d.q<HTMLButtonElement>("button.action.undo")!;
    expect(undo.disabled).toBe(true);
    expect(undo.getAttribute("aria-label")).toBe("Undo (0 left)");
    expect(undo.title).toContain("no undos remaining this turn");
  });

  it("reads ∞ on a table with no undo limit", () => {
    const d = mountDock({ undosLeft: -1, canUndo: true });
    const undo = d.q<HTMLButtonElement>("button.action.undo")!;
    expect(undo.getAttribute("aria-label")).toBe("Undo (no limit)");
    expect(undo.querySelector(".undo-count")?.textContent).toBe("∞");
    expect(undo.title).toContain("this table has no undo limit");
  });

  it("is the only Undo in the dock, whatever request is open", () => {
    pushDockRequest(attackRowRequest(attackInput()));
    const d = mountDock({ undosLeft: 1, canUndo: true });
    const undos = [...d.container.querySelectorAll("button")].filter((b) =>
      /undo/i.test(b.getAttribute("aria-label") ?? accessibleName(b)),
    );
    expect(undos).toHaveLength(1);
  });
});

// The attack row's inputs, for a duel: two bears ready, one opponent.
const bear = (id: string): CardView =>
  ({
    instance_id: id,
    name: "Grizzly Bears",
    controller: ME,
    type_line: "Creature — Bear",
  }) as CardView;
function attackPlan(over: Partial<AttackAllPlan> = {}): AttackAllPlan {
  return {
    eligible: [bear("b1"), bear("b2")],
    declared: [],
    blocked: [],
    defenders: [seat(OPP, "Opp", 1)],
    targets: {},
    ...over,
  };
}
const attacks: string[] = [];
const chosen: string[] = [];
function attackInput(over: Partial<Parameters<typeof attackRowRequest>[0]> = {}) {
  return {
    view: gameView({ turn: turn({ step: "declare_attackers", phase: "combat" }) }),
    plan: attackPlan(),
    ready: true,
    blockedHint: "",
    attackAllChord: "a",
    seatColor: () => "#f00",
    onAttackAll: (id: string) => attacks.push(id),
    onChooseAttackers: (id: string) => chosen.push(id),
    ...over,
  };
}

// ADR 0111 PR 3: combat's requests, drawn by the dock.
describe("the dock's requests", () => {
  beforeEach(() => {
    attacks.length = 0;
    chosen.length = 0;
  });

  const dialog = (c: ParentNode, name: string) =>
    c.querySelector<HTMLElement>(`[role="dialog"][aria-label="${name}"]`);

  it("draws the attack row in a non-modal dialog, and keeps next and Pass turn", () => {
    pushDockRequest(
      attackRowRequest(
        attackInput({
          plan: attackPlan({ declared: [bear("b0")] }),
          blockedHint: "1 tapped",
        }),
      ),
    );
    const d = mountDock();
    const dlg = dialog(d.container, "declare attackers")!;
    expect(dlg).not.toBeNull();
    expect(dlg.hasAttribute("aria-modal")).toBe(false);
    // ADR 0111 §10: the attack row is `group "declare attackers"`, with
    // the honest count and the reason the rest are out.
    const group = dlg.querySelector<HTMLElement>('[role="group"][aria-label="declare attackers"]')!;
    expect(group.textContent).toContain("2 ready to attack");
    expect(group.textContent).toContain("1 already declared");
    expect(group.textContent).toContain("can't: 1 tapped");
    // The name the e2e suite matches, unchanged.
    const all = buttonsNamed(group, "Attack Opp with all 2 creatures")[0]!;
    expect(all.getAttribute("aria-keyshortcuts")).toBe("A");
    expect(all.title).toContain("(A)");
    click(all);
    expect(attacks).toEqual([OPP]);
    // A step row does not take the bar: next and Pass turn are still
    // there, once each, and not inside the request.
    expect(buttonsNamed(d.container, "next")).toHaveLength(1);
    expect(buttonsNamed(d.container, "Pass turn")).toHaveLength(1);
    expect(dlg.contains(buttonsNamed(d.container, "next")[0]!)).toBe(false);
  });

  it("puts one Attack all button per opponent at a wider table, with the picker where it applies", () => {
    const view = gameView({
      seats: [seat(ME, "Me", 0), seat(OPP, "Opp", 1), seat("o2", "Two", 2)],
      turn: turn({
        step: "declare_attackers",
        attack_targets: [{ kind: "player", id: "o2", tax: "{2}" }],
      } as Partial<TurnView>),
    });
    pushDockRequest(
      attackRowRequest(
        attackInput({
          view,
          plan: attackPlan({ defenders: [seat(OPP, "Opp", 1), seat("o2", "Two", 2)] }),
        }),
      ),
    );
    const d = mountDock();
    const group = d.q('[role="group"][aria-label="declare attackers"]')!;
    expect(group.textContent).toContain("Attack all →");
    expect(buttonsNamed(group, "Opp")).toHaveLength(1);
    const choose = group.querySelector<HTMLButtonElement>(
      'button[aria-label="Choose attackers against Two"]',
    )!;
    click(choose);
    expect(chosen).toEqual(["o2"]);
  });

  it("answers an attack-tax refusal in the row, with Choose attackers… and dismiss", () => {
    let picked = 0;
    let dismissed = 0;
    pushDockRequest(
      attackRowRequest(
        attackInput({
          refusal: {
            kind: "tax",
            message: "attack tax unpaid",
            reason: "{4}",
            missing: ["{2}"],
            limitRoom: null,
            onChoose: () => picked++,
            onDismiss: () => dismissed++,
          },
        }),
      ),
    );
    const d = mountDock();
    const alert = d.q('[role="dialog"] [role="alert"]')!;
    expect(alert.textContent).toContain(
      "Attacking with all of them costs {4} and you can't pay it",
    );
    expect(alert.textContent).toContain("missing {2}");
    click(buttonsNamed(alert, "Choose attackers…")[0]!);
    click(alert.querySelector('button[aria-label="dismiss"]')!);
    expect([picked, dismissed]).toEqual([1, 1]);
  });

  it("offers Choose up to N… for an attack-limit refusal, and nothing when the limit is used up", () => {
    const refusal = (limitRoom: number | null) => ({
      kind: "limit" as const,
      message: "No more than one creature can attack each combat.",
      limitRoom,
      onChoose: () => {},
      onDismiss: () => {},
    });
    const h = pushDockRequest(attackRowRequest(attackInput({ refusal: refusal(1) })));
    const d = mountDock();
    const alert = () => d.q('[role="dialog"] [role="alert"]')!;
    expect(alert().textContent).toContain("No more than one creature can attack each combat.");
    expect(buttonsNamed(alert(), "Choose up to 1…")).toHaveLength(1);
    h.update(attackRowRequest(attackInput({ refusal: refusal(0) })));
    flushSync();
    expect(alert().textContent).toContain("no more creatures can attack this combat");
    expect(alert().querySelectorAll("button")).toHaveLength(1); // dismiss only
  });

  it("makes No blocks the primary, in place of next and Pass turn, and Done blocking once staged", () => {
    let finished = 0;
    const h = pushDockRequest(blockRequest(0, () => finished++));
    const d = mountDock({ viewerHasPriority: true });
    const dlg = dialog(d.container, "declare blockers")!;
    expect(
      dlg.querySelector('[role="group"][aria-label="declare blockers"]')?.textContent,
    ).toContain("Choose blockers, or declare none");
    const primary = buttonsNamed(dlg, "No blocks")[0]!;
    expect(primary.classList.contains("primary")).toBe(true);
    expect(dlg.querySelector(".dock-bar")?.lastElementChild).toBe(primary);
    // A stronger request takes the bar: next and Pass turn give way.
    expect(buttonsNamed(d.container, "next")).toHaveLength(0);
    expect(buttonsNamed(d.container, "Pass turn")).toHaveLength(0);
    click(primary);
    expect(finished).toBe(1);

    h.update(blockRequest(2, () => finished++));
    flushSync();
    expect(dlg.textContent).toContain("2 blockers declared");
    expect(buttonsNamed(dlg, "Done blocking")).toHaveLength(1);

    // It closes; next and Pass turn come back.
    h.close();
    flushSync();
    expect(dialog(d.container, "declare blockers")).toBeNull();
    expect(buttonsNamed(d.container, "next")).toHaveLength(1);
    expect(buttonsNamed(d.container, "Pass turn")).toHaveLength(1);
  });

  it("focuses the blockers primary when it opens and focus is on the body", () => {
    (document.activeElement as HTMLElement | null)?.blur?.();
    const d = mountDock();
    pushDockRequest(blockRequest(0, () => {}));
    flushSync();
    expect(document.activeElement).toBe(buttonsNamed(d.container, "No blocks")[0]);
  });

  it("does not take focus from a control the player is on", () => {
    const outside = document.createElement("button");
    document.body.appendChild(outside);
    outside.focus();
    const d = mountDock();
    pushDockRequest(blockRequest(0, () => {}));
    flushSync();
    expect(buttonsNamed(d.container, "No blocks")).toHaveLength(1);
    expect(document.activeElement).toBe(outside);
    outside.remove();
  });

  it("draws a combat selection as its hint and Cancel, with no primary", () => {
    let cancelled = 0;
    pushDockRequest(combatSelectionRequest("attacker", "Grizzly Bears", () => cancelled++));
    const d = mountDock();
    const dlg = dialog(d.container, "attacking with Grizzly Bears")!;
    const q = dlg.querySelector(".dock-question")!;
    expect(q.getAttribute("role")).toBe("status");
    expect(q.getAttribute("aria-live")).toBe("polite");
    expect(q.textContent).toContain("Attacking with Grizzly Bears — click an opponent's seat");
    const bar = dlg.querySelector(".dock-bar")!;
    const cancel = buttonsNamed(bar, "Cancel")[0]!;
    expect(cancel.classList.contains("secondary")).toBe(true);
    expect(cancel.getAttribute("aria-keyshortcuts")).toBe("Escape");
    expect(cancel.querySelector("kbd.cap")?.textContent).toBe("Esc");
    expect(bar.querySelector(".primary")).toBeNull();
    expect(buttonsNamed(d.container, "next")).toHaveLength(0);
    click(cancel);
    expect(cancelled).toBe(1);
  });

  it("draws only the strongest request: a selection hides the attack row", () => {
    pushDockRequest(attackRowRequest(attackInput()));
    const sel = pushDockRequest(combatSelectionRequest("attacker", "Grizzly Bears", () => {}));
    const d = mountDock();
    expect(d.container.querySelectorAll('[role="dialog"]')).toHaveLength(1);
    expect(dialog(d.container, "declare attackers")).toBeNull();
    sel.close();
    flushSync();
    expect(dialog(d.container, "declare attackers")).not.toBeNull();
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

  it("scrolls a short creature row inside its area rather than over the lands", () => {
    const css = src("src/lib/components/board/PlayerPanel.svelte");
    const area = rule(css, "  .grid-creatures");
    expect(area).toContain("min-height: 0");
    expect(area).toContain("display: flex");
    expect(area).toContain("flex-direction: column");
    // The row may shrink below its cards; BattlefieldRow's .row is
    // overflow: auto, so it scrolls.
    expect(rule(css, "  .grid-creatures > :global(.row)")).toContain("flex: 0 1 auto");
    expect(rule(src("src/lib/components/board/BattlefieldRow.svelte"), "  .row")).toContain(
      "overflow: auto",
    );
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
