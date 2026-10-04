// @vitest-environment jsdom
//
// castAnyway.render.test.ts — ADR 0118 §2 (#2188), on the REAL Board.
// With strict payment on, every card the viewer could cast offers
// "Cast anyway (don't pay)" in its right-click popover's Sandbox
// section: a hand card (even a plain spell, which had no menu before),
// an exile entry and a commander in the castable-from-other-zones strip,
// and the command-zone panel's commander. Choosing the row sends
// nothing: it opens the dock's "Cast <card> without paying its mana
// cost?" with Cast and Cancel. Cancel and Escape send nothing; Cast runs
// the ordinary cast chain and sends `strict: true, force_cast: true`.
// With admin overrides on, the override menu carries the same row and
// opens the same confirmation.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));
vi.mock("./api", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  return {
    ...actual,
    fetchAutoTapPreview: () => Promise.resolve({ ok: true, cost: "{G}", plan: [] }),
  };
});

import Board from "./components/board/Board.svelte";
import { activeDockRequest, dockKeyAction, _resetForTests as resetDock } from "./dock";
import { clearCastAnyway } from "./castAnyway";
import { resetSettings, updateSettings } from "./settings";
import { closeAbilityPopover } from "./abilityPopover";
import { closeCardMenu } from "./contextMenu";
import type { ActionType, CardView, GameView, PlayerView } from "./protocol";
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
  updateSettings("gameplay", "strictMana", true);
});

afterEach(() => {
  cleanup();
  resetDock();
  clearCastAnyway();
  closeAbilityPopover();
  closeCardMenu();
  resetSettings();
});

const ME = "me";
const THEM = "them";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const mine = (id: string, name: string, typeLine: string, cost: string, extras = {}): CardView =>
  ({
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: typeLine,
    mana_cost: cost,
    ...extras,
  }) as CardView;

const wurm = () => mine("wurm", "Craw Wurm", "Creature — Wurm", "{4}{G}{G}");
const forest = () => mine("forest", "Forest", "Basic Land — Forest", "");
const fireball = () => mine("fireball", "Fireball", "Sorcery", "{X}{R}");
const kenrith = () =>
  mine("kenrith", "Kenrith", "Legendary Creature — Human Noble", "{4}{W}", { is_commander: true });
const impulsed = (): CardView =>
  ({
    instance_id: "bolt",
    name: "Impulse Bolt",
    owner: THEM,
    controller: THEM,
    known_by_you: true,
    type_line: "Instant",
    mana_cost: "{R}",
    exile_play: { player: ME },
    castable_here: true,
    cast_prices: [{ cost: "{R}", printed: true }],
  }) as CardView;

const seat = (id: string, idx: number, hand: CardView[], command: CardView[]): PlayerView =>
  ({
    id,
    name: id === ME ? "Me" : "Them",
    seat: idx,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id, command),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

function gameView(o: { hand?: CardView[]; command?: CardView[]; exile?: CardView[] }): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME, 0, o.hand ?? [], o.command ?? []), seat(THEM, 1, [], [])],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined, o.exile ?? []),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    // The seat owes a decision and can only pass: nothing here is
    // payable, which is exactly when Cast anyway matters.
    legal_moves: [{ type: "pass_priority", player: ME, kind: "pass", label: "Pass" }],
  } as unknown as GameView;
}

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountBoard(o: Parameters<typeof gameView>[0]) {
  const sent: Sent[] = [];
  const r = render(
    Board as never,
    {
      view: gameView(o),
      viewerID: ME,
      isAdmin: false,
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
  const handCard = (name: string) =>
    r.container.querySelector<HTMLElement>(`.hand .card[aria-label='${name}']`);
  const stripCard = (name: string) =>
    r.container.querySelector<HTMLElement>(`.exile-strip .card[aria-label='${name}']`);
  const panelCard = (name: string) =>
    r.container.querySelector<HTMLElement>(`.cmd-zone .card[aria-label='${name}']`);
  return { ...r, sent, handCard, stripCard, panelCard };
}

function rightClick(el: HTMLElement): void {
  el.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
  flushSync();
}

function castAnywayRow(container: HTMLElement): HTMLButtonElement | null {
  return container.querySelector<HTMLButtonElement>("[data-cast-anyway]");
}

const DIALOG = "Cast Craw Wurm without paying its mana cost?";

describe("the row on a hand card", () => {
  it("opens the popover on a plain spell's right-click, in the Sandbox section", () => {
    const b = mountBoard({ hand: [wurm()] });
    rightClick(b.handCard("Craw Wurm")!);
    const row = castAnywayRow(b.container);
    expect(row, "a plain spell's right-click opens the popover").toBeTruthy();
    expect(row!.getAttribute("role")).toBe("menuitem");
    expect(row!.textContent?.trim()).toBe("Cast anyway (don't pay)");
    expect(row!.disabled).toBe(false);
    expect(row!.closest("[role='group']")?.getAttribute("aria-label")).toBe("sandbox");
    // On a hand card it is the section's only row.
    expect(b.container.querySelectorAll("[role='menu'] [role='menuitem']").length).toBe(1);
  });

  it("opens the confirmation and sends nothing", () => {
    const b = mountBoard({ hand: [wurm()] });
    rightClick(b.handCard("Craw Wurm")!);
    click(castAnywayRow(b.container)!);
    flushSync();
    expect(b.sent).toEqual([]);
    const r = get(activeDockRequest);
    expect(r?.label).toBe(DIALOG);
    expect(r?.primary?.label).toBe("Cast");
    expect(r?.secondary?.map((a) => a.label)).toEqual(["Cancel"]);
    // The popover closed behind it.
    expect(b.container.querySelector("[role='menu']")).toBeNull();
  });

  it("Cancel and Escape send nothing", () => {
    const b = mountBoard({ hand: [wurm()] });
    rightClick(b.handCard("Craw Wurm")!);
    click(castAnywayRow(b.container)!);
    flushSync();
    get(activeDockRequest)!.secondary![0].onPress();
    flushSync();
    expect(get(activeDockRequest)).toBeNull();
    expect(b.sent).toEqual([]);

    rightClick(b.handCard("Craw Wurm")!);
    click(castAnywayRow(b.container)!);
    flushSync();
    dockKeyAction(get(activeDockRequest), "Escape")!.onPress();
    flushSync();
    expect(get(activeDockRequest)).toBeNull();
    expect(b.sent).toEqual([]);
  });

  it("Cast sends strict and force_cast through the cast chain", () => {
    const b = mountBoard({ hand: [wurm()] });
    rightClick(b.handCard("Craw Wurm")!);
    click(castAnywayRow(b.container)!);
    flushSync();
    get(activeDockRequest)!.primary!.onPress();
    flushSync();
    expect(b.sent).toEqual([
      { type: "cast_spell", params: { instance_id: "wurm", strict: true, force_cast: true } },
    ]);
    expect(get(activeDockRequest)).toBeNull();
  });

  it("Cast walks the chain: an X spell asks for X before anything is sent", () => {
    const b = mountBoard({ hand: [fireball()] });
    rightClick(b.handCard("Fireball")!);
    click(castAnywayRow(b.container)!);
    flushSync();
    get(activeDockRequest)!.primary!.onPress();
    flushSync();
    expect(b.sent).toEqual([]);
    expect(get(activeDockRequest)?.label).toBe("Choose X for Fireball");
  });

  it("is not offered on a land", () => {
    const b = mountBoard({ hand: [forest()] });
    rightClick(b.handCard("Forest")!);
    expect(castAnywayRow(b.container)).toBeNull();
  });

  it("is not offered with strict payment off: a plain spell has no menu", () => {
    updateSettings("gameplay", "strictMana", false);
    const b = mountBoard({ hand: [wurm()] });
    rightClick(b.handCard("Craw Wurm")!);
    expect(castAnywayRow(b.container)).toBeNull();
    expect(b.container.querySelector("[role='menu']")).toBeNull();
  });
});

describe("the row on the other surfaces", () => {
  it("on a strip exile entry, cast out of exile", () => {
    const b = mountBoard({ exile: [impulsed()] });
    rightClick(b.stripCard("Impulse Bolt")!);
    click(castAnywayRow(b.container)!);
    flushSync();
    expect(b.sent).toEqual([]);
    expect(get(activeDockRequest)?.label).toBe("Cast Impulse Bolt without paying its mana cost?");
    get(activeDockRequest)!.primary!.onPress();
    flushSync();
    expect(b.sent).toEqual([
      {
        type: "cast_spell",
        params: { instance_id: "bolt", from_zone: "exile", strict: true, force_cast: true },
      },
    ]);
  });

  it("on a strip commander and on the command-zone panel, cast out of the command zone", () => {
    for (const surface of ["strip", "panel"] as const) {
      const b = mountBoard({ command: [kenrith()] });
      const el = surface === "strip" ? b.stripCard("Kenrith") : b.panelCard("Kenrith");
      expect(el, surface).toBeTruthy();
      rightClick(el!);
      click(castAnywayRow(b.container)!);
      flushSync();
      expect(get(activeDockRequest)?.label).toBe("Cast Kenrith without paying its mana cost?");
      get(activeDockRequest)!.primary!.onPress();
      flushSync();
      expect(b.sent, surface).toEqual([
        {
          type: "cast_spell",
          params: { instance_id: "kenrith", from_zone: "command", strict: true, force_cast: true },
        },
      ]);
      cleanup();
      resetDock();
    }
  });
});

describe("the admin override menu", () => {
  it("carries the row, and it opens the same confirmation", () => {
    updateSettings("gameplay", "adminOverrides", true);
    const b = mountBoard({ hand: [wurm()] });
    rightClick(b.handCard("Craw Wurm")!);
    const menu = document.querySelector<HTMLElement>("[role='menu'][aria-label='card overrides']");
    expect(menu).toBeTruthy();
    const row = [...menu!.querySelectorAll<HTMLButtonElement>("[role='menuitem']")].find(
      (el) => el.textContent?.trim() === "Cast anyway (don't pay)",
    );
    expect(row).toBeTruthy();
    click(row!);
    flushSync();
    expect(b.sent).toEqual([]);
    expect(get(activeDockRequest)?.label).toBe(DIALOG);
  });
});
