// @vitest-environment jsdom
//
// #1438 — click a mana source to tap it FOR mana. manaSource.test.ts
// pins the decisions as pure functions; this file checks that real
// clicks on a real board reach them: that a Swamp sends
// activate_mana_ability, that a painland opens the anchored picker
// instead, that the raw tap is still one Alt-click (or one menu row)
// away, and that the picker and the mana_pick prompt draw symbols and
// answer the keyboard.

import { describe, it, expect, afterEach, vi } from "vitest";
import { get } from "svelte/store";

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import ManaSourcePicker from "./components/board/ManaSourcePicker.svelte";
import ManaSymbolPicker from "./components/board/ManaSymbolPicker.svelte";
import ManaPoolPips from "./components/board/ManaPoolPips.svelte";
import ChoicePromptModal from "./components/board/ChoicePromptModal.svelte";
import { closeManaSourcePicker, manaSourcePicker } from "./manaSourcePicker";
import type { ManaPickOption } from "./manaSource";
import type { ActionType, CardView, GameView, ManaAbilityView, PlayerView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  closeManaSourcePicker();
  cleanup();
});

const ME = "me";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const ability = (index: number, extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index,
  tap_cost: true,
  ...extra,
});

const permanent = (id: string, name: string, typeLine: string, extra: Partial<CardView> = {}) =>
  ({
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: typeLine,
    ...extra,
  }) as CardView;

const swamp = (extra: Partial<CardView> = {}) =>
  permanent("swamp", "Swamp", "Basic Land — Swamp", {
    mana_abilities: [ability(0, { produced: "{B}", label: "Add {B}" })],
    ...extra,
  });

const forge = () =>
  permanent("forge", "Battlefield Forge", "Land", {
    mana_abilities: [
      ability(0, { produced: "{C}", label: "Add {C}" }),
      ability(1, {
        produced: "{R|W}",
        label: "Add {R} or {W}. This land deals 1 damage to you.",
      }),
    ],
  });

const bear = () =>
  permanent("bear", "Grizzly Bears", "Creature — Bear", { power: 2, toughness: 2 });

const seat = (): PlayerView =>
  ({
    id: ME,
    name: "Me",
    seat: 0,
    life: 40,
    library: zone("library", ME),
    hand: zone("hand", ME),
    graveyard: zone("graveyard", ME),
    command: zone("command", ME),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const gameView = (battlefield: CardView[], extra: Partial<GameView> = {}): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat()],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
    ...extra,
  }) as unknown as GameView;

interface Sent {
  type: ActionType;
  params?: unknown;
  player?: string;
}

function mountPanel(cards: CardView[]) {
  const sent: Sent[] = [];
  const tapped: string[] = [];
  const view = gameView(cards);
  const r = render(
    PlayerPanel as never,
    {
      seat: view.seats[0],
      isSelf: true,
      isActive: true,
      hasPriority: true,
      viewerID: ME,
      isAdmin: false,
      sendAction: (type: ActionType, params?: unknown, player?: string) =>
        sent.push({ type, params, player }),
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: cards,
      exile: zone("exile", undefined),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: (c: CardView) => tapped.push(c.instance_id),
      onPlayCard: () => {},
      onDrawCard: () => {},
      onManaAbilityCost: () => {},
    } as never,
  );
  const tile = (id: string) =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${id}"]`)!;
  return { ...r, sent, tapped, tile };
}

describe("left-click on your own permanents (#1438)", () => {
  it("taps a Swamp FOR mana — one click, no prompt", () => {
    const p = mountPanel([swamp()]);
    click(p.tile("swamp"));
    expect(p.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "swamp", ability_index: 0 },
        player: ME,
      },
    ]);
    expect(p.tapped).toEqual([]);
    expect(get(manaSourcePicker)).toBeNull();
  });

  it("opens the anchored picker for a painland and sends nothing yet", () => {
    const p = mountPanel([forge()]);
    click(p.tile("forge"));
    expect(p.sent).toEqual([]);
    expect(p.tapped).toEqual([]);
    expect(get(manaSourcePicker)?.cardID).toBe("forge");
  });

  it("closes the picker when the same card is clicked again", () => {
    const p = mountPanel([forge()]);
    click(p.tile("forge"));
    click(p.tile("forge"));
    expect(get(manaSourcePicker)).toBeNull();
    expect(p.sent).toEqual([]);
  });

  it("keeps click-to-tap for a permanent with no mana ability", () => {
    const p = mountPanel([bear()]);
    click(p.tile("bear"));
    expect(p.tapped).toEqual(["bear"]);
    expect(p.sent).toEqual([]);
  });

  it("raw-taps a mana source on Alt-click", () => {
    const p = mountPanel([swamp()]);
    click(p.tile("swamp"), { altKey: true });
    expect(p.tapped).toEqual(["swamp"]);
    expect(p.sent).toEqual([]);
  });

  it("untaps a tapped mana source with one click", () => {
    const p = mountPanel([swamp({ tapped: true })]);
    click(p.tile("swamp"));
    expect(p.tapped).toEqual(["swamp"]);
    expect(p.sent).toEqual([]);
  });

  it("offers 'Tap (no mana)' on right-click", () => {
    const p = mountPanel([swamp()]);
    p.tile("swamp").dispatchEvent(
      new MouseEvent("contextmenu", { bubbles: true, cancelable: true }),
    );
    flushSync();
    const row = p.container.querySelector<HTMLElement>("[data-raw-tap]");
    expect(row?.textContent).toContain("Tap (no mana)");
    click(row!);
    expect(p.tapped).toEqual(["swamp"]);
    expect(p.sent).toEqual([]);
  });
});

describe("ManaSourcePicker — the anchored picker", () => {
  function mountPicker(cards: CardView[] = [forge()]) {
    const picked: number[] = [];
    const onClose = vi.fn();
    const r = render(
      ManaSourcePicker as never,
      {
        view: gameView(cards),
        open: { cardID: "forge", anchor: { left: 100, top: 500, right: 170, bottom: 600 } },
        onPick: (_c: CardView, i: number) => picked.push(i),
        onClose,
      } as never,
    );
    return { ...r, picked, onClose };
  }

  const key = (k: string) => {
    window.dispatchEvent(new KeyboardEvent("keydown", { key: k, bubbles: true, cancelable: true }));
    flushSync();
  };

  it("draws one mana-symbol option per ability, with tooltip and rider", () => {
    const { container } = mountPicker();
    const buttons = container.querySelectorAll<HTMLButtonElement>(".mana-option");
    expect(buttons.length).toBe(2);
    const symbolsOf = (b: Element) =>
      Array.from(b.querySelectorAll("svg[data-symbol]")).map((s) => s.getAttribute("data-symbol"));
    expect(symbolsOf(buttons[0])).toEqual(["C"]);
    expect(symbolsOf(buttons[1])).toEqual(["R", "W"]);
    expect(buttons[0].title).toBe("Add {C}");
    expect(buttons[1].title).toContain("deals 1 damage to you");
    expect(buttons[1].textContent).toContain("deals 1 damage to you");
  });

  it("picks with a click", () => {
    const p = mountPicker();
    click(p.container.querySelectorAll(".mana-option")[1]);
    expect(p.picked).toEqual([1]);
    expect(p.onClose).toHaveBeenCalled();
  });

  it("picks with the number keys", () => {
    const p = mountPicker();
    key("1");
    expect(p.picked).toEqual([0]);
  });

  it("cancels on Escape and picks nothing", () => {
    const p = mountPicker();
    key("Escape");
    expect(p.onClose).toHaveBeenCalled();
    expect(p.picked).toEqual([]);
  });

  it("closes on a click outside", () => {
    const p = mountPicker();
    click(document.body);
    expect(p.onClose).toHaveBeenCalled();
    expect(p.picked).toEqual([]);
  });

  it("closes when the source leaves the battlefield", () => {
    const p = mountPicker();
    p.setProps({ view: gameView([]) } as never);
    expect(p.onClose).toHaveBeenCalled();
  });
});

describe("ManaSymbolPicker", () => {
  const opts: ManaPickOption[] = [
    { key: "a", symbols: ["G"], caption: "Green", title: "Add {G}" },
    { key: "b", symbols: ["C"], caption: "Colorless", title: "Add {C}", disabled: "no" },
  ];

  it("will not pick a disabled option, by click or by key", () => {
    const picked: string[] = [];
    const { container } = render(
      ManaSymbolPicker as never,
      { options: opts, onPick: (o: ManaPickOption) => picked.push(o.key) } as never,
    );
    const b = container.querySelectorAll<HTMLButtonElement>(".mana-option");
    expect(b[1].disabled).toBe(true);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "2" }));
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "3" }));
    flushSync();
    expect(picked).toEqual([]);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "1" }));
    flushSync();
    expect(picked).toEqual(["a"]);
  });

  it("gives every option at least a 44 px target", () => {
    // jsdom has no layout, so read the declared floor off the stylesheet
    // Svelte injected rather than a measured box.
    render(ManaSymbolPicker as never, { options: opts, onPick: () => {} } as never);
    const css = Array.from(document.querySelectorAll("style"))
      .map((s) => s.textContent ?? "")
      .join("\n");
    const rule = /\.mana-option\.svelte-[\w-]+\s*\{([^}]*)\}/.exec(css)?.[1] ?? "";
    expect(Number(/min-width:\s*(\d+)px/.exec(rule)?.[1])).toBeGreaterThanOrEqual(44);
    expect(Number(/min-height:\s*(\d+)px/.exec(rule)?.[1])).toBeGreaterThanOrEqual(44);
  });
});

describe("the mana_pick prompt uses the same picker", () => {
  const snap = (colors: string[]): GameView =>
    gameView([], {
      pending_choices: [
        {
          id: "pick-1",
          kind: "mana_pick",
          chooser: ME,
          from_player: ME,
          count: 1,
          reason: "Birds of Paradise — pick a color",
          color_options: colors,
        },
      ],
    } as unknown as Partial<GameView>);

  function mountPrompt(colors: string[]) {
    const sent: Sent[] = [];
    const r = render(
      ChoicePromptModal as never,
      {
        snap: snap(colors),
        viewerID: ME,
        sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
        lastError: null,
      } as never,
    );
    return { ...r, sent };
  }

  it("draws the server's colours as symbols, in the server's order", () => {
    const { container } = mountPrompt(["G", "U", "W"]);
    const symbols = Array.from(container.querySelectorAll(".mana-option svg[data-symbol]")).map(
      (s) => s.getAttribute("data-symbol"),
    );
    expect(symbols).toEqual(["G", "U", "W"]);
    expect(container.querySelector(".mana-option")?.getAttribute("title")).toBe("Add Green mana");
  });

  it("answers with a click", () => {
    const p = mountPrompt(["G", "U", "W"]);
    click(p.container.querySelectorAll(".mana-option")[2]);
    expect(p.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "pick-1", color: "W" } },
    ]);
  });

  it("answers with a number key", () => {
    const p = mountPrompt(["G", "U", "W"]);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "2" }));
    flushSync();
    expect(p.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "pick-1", color: "U" } },
    ]);
  });

  it("does not cancel on Escape — the source is already tapped", () => {
    const p = mountPrompt(["G", "U"]);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(p.sent).toEqual([]);
    expect(p.container.querySelector(".mana-option")).not.toBeNull();
  });
});

describe("ManaPoolPips", () => {
  it("draws the pool with the same symbols", () => {
    const { container } = render(ManaPoolPips as never, { pool: ["G", "G", "C"] } as never);
    const pips = container.querySelectorAll(".pip");
    expect(pips.length).toBe(2);
    expect(pips[0].getAttribute("aria-label")).toBe("2 green mana");
    expect(pips[0].querySelector("svg")?.getAttribute("data-symbol")).toBe("G");
    expect(pips[1].querySelector("svg")?.getAttribute("data-symbol")).toBe("C");
  });
});
