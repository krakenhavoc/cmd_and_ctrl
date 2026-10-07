// @vitest-environment jsdom
//
// payLifeForManaAbilityPanel.render.test.ts — ADR 0131 §2 (#2531), PR 2,
// on a real PlayerPanel: choosing "Pay life for {B}…" on a filter land's
// popover sends `activate_mana_ability` carrying `phyrexian_life`, with
// the row's ref and no colours (the server asks each picking slot as its
// usual mana_pick). The plain row still sends no claim.

import { describe, it, expect, afterEach } from "vitest";

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import { closeAbilityPopover } from "./abilityPopover";
import { closeManaSourcePicker } from "./manaSourcePicker";
import type { ActionType, CardView, GameView, PlayerView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  closeManaSourcePicker();
  closeAbilityPopover();
  cleanup();
});

const ME = "me";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

// "{B}, {T}: Add {B}{B}" with a {B} the viewer's K'rrik lets them pay for.
const filter = (life?: number): { card: CardView; life: number } => ({
  life: life ?? 40,
  card: {
    instance_id: "heath",
    name: "Black Filter",
    owner: ME,
    controller: ME,
    known_by_you: true,
    type_line: "Land",
    mana_abilities: [
      {
        index: 0,
        tap_cost: true,
        produced: "{B}{B}",
        label: "{B}, {T}: Add {B}{B}",
        mana_cost: "{B}",
        ref: "own:0",
        phyrexian_symbols: 1,
        phyrexian_granted: 1,
      },
    ],
  } as CardView,
});

function mountPanel(card: CardView, life: number) {
  const sent: { type: ActionType; params?: unknown }[] = [];
  const seat = {
    id: ME,
    name: "Me",
    seat: 0,
    life,
    library: zone("library", ME),
    hand: zone("hand", ME),
    graveyard: zone("graveyard", ME),
    command: zone("command", ME),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  } as unknown as PlayerView;
  const view = {
    id: "g1",
    state: "active",
    seats: [seat],
    battlefield: zone("battlefield", undefined, [card]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  } as unknown as GameView;
  const r = render(
    PlayerPanel as never,
    {
      seat,
      isSelf: true,
      isActive: true,
      hasPriority: true,
      viewerID: ME,
      isAdmin: false,
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      isMonarch: false,
      isInitiative: false,
      view,
      controlledCards: [card],
      exile: zone("exile", undefined),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
      onTapToggle: () => {},
      onPlayCard: () => {},
      onDrawCard: () => {},
      onManaAbilityCost: () => {},
    } as never,
  );
  const tile = r.container.querySelector<HTMLElement>(
    `.card[data-instance-id="${card.instance_id}"]`,
  )!;
  tile.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
  flushSync();
  return { ...r, sent };
}

describe("PlayerPanel — Pay life for {B}… on a mana ability", () => {
  it("sends the claim with the activation", () => {
    const { card, life } = filter();
    const p = mountPanel(card, life);
    const row = p.container.querySelector<HTMLElement>('button[data-kind="mana-pay-life"]');
    expect(row).not.toBeNull();
    click(row!);
    expect(p.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "heath", ability_index: 0, ref: "own:0", phyrexian_life: 1 },
      },
    ]);
  });

  it("the ordinary row sends no claim", () => {
    const { card, life } = filter();
    const p = mountPanel(card, life);
    click(p.container.querySelector<HTMLElement>('button[data-kind="mana"]')!);
    expect(p.sent).toEqual([
      {
        type: "activate_mana_ability",
        params: { card_id: "heath", ability_index: 0, ref: "own:0" },
      },
    ]);
  });

  it("offers no life row at 1 life (CR 119.4)", () => {
    const { card } = filter();
    const p = mountPanel(card, 1);
    expect(p.container.querySelector('button[data-kind="mana-pay-life"]')).toBeNull();
  });
});
