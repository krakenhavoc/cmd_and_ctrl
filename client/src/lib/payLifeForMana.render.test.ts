// @vitest-environment jsdom
//
// payLifeForMana.render.test.ts — ADR 0131 §4 (#2531), on the REAL
// Board. Under K'rrik the server counts every {B} in a cost in
// `phyrexian_symbols` and says how many are the grant's in
// `phyrexian_granted`. A cast whose symbols are ALL granted opens the
// life stepper only when the auto-tap preview, asked with nothing
// claimed, says mana is short; a printed Phyrexian symbol still always
// asks; and the hand card's menu keeps a "Pay life for {B}…" row that
// opens the stepper whatever the mana says.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import { get } from "svelte/store";

// `ok` is the answer with nothing claimed; `need` is how many symbols paid
// with life make the mana half affordable.
const preview = vi.hoisted(() => ({ ok: true, need: 0 }));

vi.mock("./sounds", () => ({ play: () => {} }));
vi.mock("./api", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  return {
    ...actual,
    fetchAutoTapPreview: (_gid: string, _id: string, opts?: { phyrexianLife?: number }) =>
      Promise.resolve(
        preview.ok || (opts?.phyrexianLife ?? 0) >= preview.need
          ? { ok: true, cost: "{1}{B}{B}", plan: [] }
          : { ok: false, cost: "{1}{B}{B}", missing: ["B", "B"] },
      ),
  };
});

import Board from "./components/board/Board.svelte";
import { activeDockRequest, _resetForTests as resetDock } from "./dock";
import { payLifeRequest } from "./payLifeForMana";
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
  preview.ok = true;
  preview.need = 0;
});

afterEach(() => {
  cleanup();
  resetDock();
  payLifeRequest.set(null);
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

const mine = (id: string, name: string, cost: string, extras = {}): CardView =>
  ({
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Sorcery",
    mana_cost: cost,
    ...extras,
  }) as CardView;

// {1}{B}{B} under K'rrik: two symbols, both granted.
const spell = () =>
  mine("bs", "Black Spell", "{1}{B}{B}", { phyrexian_symbols: 2, phyrexian_granted: 2 });
// {2}{B/P}: a printed Phyrexian symbol, nothing granted.
const printed = () => mine("gp", "Printed Phyrexian", "{2}{B/P}", { phyrexian_symbols: 1 });
const plain = () => mine("pl", "Plain Spell", "{2}{G}");

const seat = (id: string, idx: number, hand: CardView[], life: number): PlayerView =>
  ({
    id,
    name: id === ME ? "Me" : "Them",
    seat: idx,
    life,
    library: zone("library", id),
    hand: zone("hand", id, hand),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

function gameView(hand: CardView[], life = 40): GameView {
  return {
    id: "g1",
    state: "active",
    seats: [seat(ME, 0, hand, life), seat(THEM, 1, [], 40)],
    battlefield: zone("battlefield", undefined),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    legal_moves: hand.map((c) => ({
      type: "cast_spell",
      player: ME,
      kind: "cast",
      source: c.instance_id,
      label: `Cast ${c.name}`,
      params: { instance_id: c.instance_id, from_zone: "hand" },
    })),
  } as unknown as GameView;
}

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountBoard(hand: CardView[], life = 40) {
  const sent: Sent[] = [];
  const r = render(
    Board as never,
    {
      view: gameView(hand, life),
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
  return { ...r, sent, handCard };
}

function rightClick(el: HTMLElement): void {
  el.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
  flushSync();
}

const payLifeRow = (c: HTMLElement) => c.querySelector<HTMLButtonElement>("[data-pay-life]");
const STEPPER = "Phyrexian mana for Black Spell";

// settle lets the preview promise resolve and the chain react.
async function settle(): Promise<void> {
  await Promise.resolve();
  await Promise.resolve();
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
}

describe("the card menu's row", () => {
  it("is offered on a card with a granted symbol, in its own Payment group", () => {
    const b = mountBoard([spell()]);
    rightClick(b.handCard("Black Spell")!);
    const row = payLifeRow(b.container);
    expect(row, "a granted symbol gives the card a menu").toBeTruthy();
    expect(row!.getAttribute("role")).toBe("menuitem");
    expect(row!.textContent?.trim()).toBe("Pay life for {B}…");
    expect(row!.closest("[role='group']")?.getAttribute("aria-label")).toBe("payment");
  });

  it("is not offered on a printed Phyrexian symbol or a plain card", () => {
    const b = mountBoard([printed(), plain()]);
    rightClick(b.handCard("Printed Phyrexian")!);
    expect(payLifeRow(b.container)).toBeNull();
    closeAbilityPopover();
    closeCardMenu();
    rightClick(b.handCard("Plain Spell")!);
    expect(payLifeRow(b.container)).toBeNull();
  });

  it("opens the stepper even though the mana would pay, and sends nothing yet", async () => {
    preview.ok = true;
    const b = mountBoard([spell()]);
    rightClick(b.handCard("Black Spell")!);
    click(payLifeRow(b.container)!);
    await settle();
    expect(b.sent).toEqual([]);
    expect(get(activeDockRequest)?.label).toBe(STEPPER);
  });

  it("sends the claim the stepper collected as phyrexian_life", async () => {
    const b = mountBoard([spell()]);
    rightClick(b.handCard("Black Spell")!);
    click(payLifeRow(b.container)!);
    await settle();
    const more = document.querySelector<HTMLButtonElement>("[aria-label='one more symbol']");
    expect(more).toBeTruthy();
    click(more!);
    await settle();
    expect(get(activeDockRequest)?.primary?.label).toBe("Cast for 2 life");
    get(activeDockRequest)!.primary!.onPress();
    flushSync();
    expect(b.sent).toHaveLength(1);
    expect(b.sent[0].type).toBe("cast_spell");
    expect((b.sent[0].params as Record<string, unknown>).phyrexian_life).toBe(1);
  });
});

describe("a click-cast whose every symbol is granted", () => {
  it("claims nothing and casts at once when the mana pays", async () => {
    preview.ok = true;
    const b = mountBoard([spell()]);
    click(b.handCard("Black Spell")!);
    await settle();
    expect(get(activeDockRequest)?.label).not.toBe(STEPPER);
    expect(b.sent).toHaveLength(1);
    expect(b.sent[0].type).toBe("cast_spell");
    expect(b.sent[0].params as Record<string, unknown>).not.toHaveProperty("phyrexian_life");
  });

  it("opens the stepper when the mana falls short, starting on a payable count", async () => {
    preview.ok = false;
    preview.need = 1;
    const b = mountBoard([spell()]);
    click(b.handCard("Black Spell")!);
    await settle();
    await settle();
    expect(b.sent).toEqual([]);
    expect(get(activeDockRequest)?.label).toBe(STEPPER);
    // One {B} short: the stepper moved itself to the one symbol that pays.
    expect(get(activeDockRequest)?.primary?.label).toBe("Cast for 2 life");
  });

  it("never asks when CR 119.4 leaves no symbol to buy", async () => {
    preview.ok = false;
    const b = mountBoard([spell()], 1);
    click(b.handCard("Black Spell")!);
    await settle();
    expect(get(activeDockRequest)?.label).not.toBe(STEPPER);
  });
});

describe("a printed Phyrexian symbol", () => {
  it("still always asks, whatever the mana says", async () => {
    preview.ok = true;
    const b = mountBoard([printed()]);
    click(b.handCard("Printed Phyrexian")!);
    await settle();
    expect(b.sent).toEqual([]);
    expect(get(activeDockRequest)?.label).toBe("Phyrexian mana for Printed Phyrexian");
  });
});
