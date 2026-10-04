// @vitest-environment jsdom
//
// abilityClick.render.test.ts — ADR 0117 §1 and §2 on a real panel. A
// left-click on a permanent does what the card does: the light ability
// popover when it has a usable non-mana row, its mana when only mana
// rows are usable, and nothing otherwise. The headline test runs the
// click and the popover's greying over the same fixtures and checks
// they agree: the popover a right-click opens is the evidence, and the
// left-click must do what that popover says is possible.

import { describe, it, expect, afterEach, vi } from "vitest";
import { get } from "svelte/store";

vi.mock("./sounds", () => ({ play: () => {} }));

import PlayerPanel from "./components/board/PlayerPanel.svelte";
import { abilityPopover } from "./abilityPopover";
import { closeManaSourcePicker, manaSourcePicker } from "./manaSourcePicker";
import { battlefieldClickIntent } from "./contextMenu.logic";
import type { ActionType, CardView, GameView, ManaAbilityView, PlayerView } from "./protocol";
import { subscribe, type TutorialEvent } from "./tutorialBus";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  closeManaSourcePicker();
  cleanup();
});

const ME = "me";
const BOB = "bob";

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, n: number): PlayerView =>
  ({
    id,
    name: id,
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

// The viewer's precombat main phase, the viewer holding priority, the
// stack empty: every window this file's fixtures care about is open
// unless a fixture says otherwise on its own rows.
const gameView = (battlefield: CardView[]): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, 0), seat(BOB, 1)],
    battlefield: zone("battlefield", undefined, battlefield),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "precombat_main" },
    mulligans_open: false,
  }) as unknown as GameView;

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

const mana = (index: number, extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index,
  ref: `own:m${index}`,
  tap_cost: true,
  ...extra,
});

const viviMana = (power: number, extra: Partial<ManaAbilityView> = {}): ManaAbilityView => ({
  index: 0,
  ref: "own:m0",
  mana_cost: "{0}",
  label:
    "{0}: Add X mana in any combination of {U} and/or {R}, where X is Vivi Ornitier's power. Activate only during your turn and only once each turn.",
  produced: Array.from({ length: power }, () => "{U|R}").join(""),
  color_options: power > 0 ? Array.from({ length: power }, () => ["U", "R"]) : undefined,
  ...extra,
});

const vivi = (power: number, extra: Partial<CardView> = {}, m: Partial<ManaAbilityView> = {}) =>
  permanent("vivi", "Vivi Ornitier", "Legendary Creature — Wizard", {
    power,
    toughness: 3,
    mana_abilities: [viviMana(power, m)],
    ...extra,
  });

const relicDraw = {
  index: 0,
  ref: "own:0",
  label: "{3}, {T}: Draw two cards, then discard a card.",
  tap_cost: true,
  mana_cost: "{3}",
};
const relic = (extra: Partial<CardView> = {}) =>
  permanent("relic", "Relic of Sauron", "Legendary Artifact", {
    mana_abilities: [
      mana(0, {
        label: "{T}: Add two mana in any combination of {U}, {B}, and/or {R}.",
        produced: "{U|B|R}{U|B|R}",
        color_options: [
          ["U", "B", "R"],
          ["U", "B", "R"],
        ],
      }),
    ],
    activated_abilities: [relicDraw],
    ...extra,
  });

const walker = (extra: Partial<CardView> = {}) =>
  permanent("walker", "Teferi, Time Raveler", "Legendary Planeswalker — Teferi", {
    counters: { loyalty: 4 },
    activated_abilities: [
      { index: 0, ref: "own:0", label: "+1: …", loyalty_cost: 1, sorcery_speed: true },
      { index: 1, ref: "own:1", label: "−3: …", loyalty_cost: -3, sorcery_speed: true },
    ],
    ...extra,
  });

// The fixtures of ADR 0117 §1's before-and-after table, on the
// viewer's own panel.
const FIXTURES: Record<string, () => CardView> = {
  "untapped Forest": () =>
    permanent("forest", "Forest", "Basic Land — Forest", {
      mana_abilities: [mana(0, { produced: "{G}", label: "{T}: Add {G}." })],
    }),
  "tapped Forest": () =>
    permanent("forest", "Forest", "Basic Land — Forest", {
      tapped: true,
      mana_abilities: [mana(0, { produced: "{G}", label: "{T}: Add {G}." })],
    }),
  "vanilla creature": () =>
    permanent("bear", "Grizzly Bears", "Creature — Bear", { power: 2, toughness: 2 }),
  "summoning-sick Llanowar Elves": () =>
    permanent("elves", "Llanowar Elves", "Creature — Elf Druid", {
      summoning_sick: true,
      mana_abilities: [mana(0, { produced: "{G}", label: "{T}: Add {G}." })],
    }),
  "Birds of Paradise": () =>
    permanent("birds", "Birds of Paradise", "Creature — Bird", {
      mana_abilities: [
        mana(0, {
          produced: "{W|U|B|R|G}",
          color_options: [["W", "U", "B", "R", "G"]],
        }),
      ],
    }),
  "Mystic Gate": () =>
    permanent("gate", "Mystic Gate", "Land", {
      mana_abilities: [
        mana(0, { produced: "{C}", label: "{T}: Add {C}." }),
        mana(1, {
          mana_cost: "{W/U}",
          produced: "{W|U}{W|U}",
          color_options: [
            ["W", "U"],
            ["W", "U"],
          ],
        }),
      ],
    }),
  "tapped Vivi, power 3": () => vivi(3, { tapped: true }),
  "Vivi, power 1": () => vivi(1),
  "Vivi, power 0 (adds_no_mana)": () => vivi(0, {}, { adds_no_mana: true }),
  "Vivi, already used this turn": () => vivi(3, {}, { condition_unmet: true }),
  "Relic of Sauron": () => relic(),
  "Relic of Sauron, draw row refused": () =>
    relic({ activated_abilities: [{ ...relicDraw, cant_activate: "an effect says no" }] }),
  "tapped Relic of Sauron": () => relic({ tapped: true }),
  "Rogue's Passage": () =>
    permanent("passage", "Rogue's Passage", "Land", {
      mana_abilities: [mana(0, { produced: "{C}", label: "{T}: Add {C}." })],
      activated_abilities: [
        {
          index: 0,
          ref: "own:0",
          label: "{4}, {T}: Target creature can't be blocked this turn.",
          tap_cost: true,
          mana_cost: "{4}",
        },
      ],
    }),
  "Prodigal Sorcerer": () =>
    permanent("prodigal", "Prodigal Sorcerer", "Creature — Human Wizard", {
      activated_abilities: [
        {
          index: 0,
          ref: "own:0",
          label: "{T}: This creature deals 1 damage to any target.",
          tap_cost: true,
          legal_targets: { players: [ME, BOB], min: 1 },
        },
      ],
    }),
  "an Equipment in your main phase": () =>
    permanent("sword", "Bonesplitter", "Artifact — Equipment", {
      activated_abilities: [
        { index: 0, ref: "own:0", label: "Equip {1}", mana_cost: "{1}", sorcery_speed: true },
      ],
    }),
  "an Equipment with its window shut": () =>
    permanent("sword", "Bonesplitter", "Artifact — Equipment", {
      activated_abilities: [
        {
          index: 0,
          ref: "own:0",
          label: "Equip {1}",
          mana_cost: "{1}",
          sorcery_speed: true,
          timing_closed: true,
        },
      ],
    }),
  "an Arrested creature": () =>
    permanent("prodigal", "Prodigal Sorcerer", "Creature — Human Wizard", {
      restrictions: ["cant_activate"],
      activated_abilities: [
        { index: 0, ref: "own:0", label: "{T}: 1 damage to any target.", tap_cost: true },
      ],
    }),
  "a catalogued planeswalker": () => walker(),
  "a planeswalker that activated this turn": () => walker({ loyalty_activated: true }),
  "an uncatalogued planeswalker": () => walker({ activated_abilities: undefined }),
  "an uncatalogued planeswalker that activated this turn": () =>
    walker({ activated_abilities: undefined, loyalty_activated: true }),
  "a face-down creature": () =>
    permanent("morph", "Akroma, Angel of Fury", "Creature", {
      face_down: true,
      face_visible: true,
      power: 2,
      toughness: 2,
      special_actions: [
        {
          kind: "turn_face_up",
          label: "Turn face up {3}{R}{R}{R}",
          cost: "{3}{R}{R}{R}",
          available: true,
        },
      ],
    }),
  "a land under Squirrel Nest": () =>
    permanent("nestland", "Forest", "Basic Land — Forest", {
      mana_abilities: [mana(0, { produced: "{G}", label: "{T}: Add {G}." })],
      activated_abilities: [
        {
          index: 0,
          ref: "grant:squirrel-nest:0:0",
          label: "{T}: Create a 1/1 green Squirrel creature token.",
          tap_cost: true,
          granted_by: { id: "nest", name: "Squirrel Nest" },
        },
      ],
    }),
};

interface Sent {
  type: ActionType;
  params?: unknown;
}

function mountPanel(card: CardView) {
  const sent: Sent[] = [];
  const tapped: string[] = [];
  const activated: [string, number][] = [];
  const view = gameView([card]);
  const r = render(
    PlayerPanel as never,
    {
      seat: view.seats[0],
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
      onTapToggle: (c: CardView) => tapped.push(c.instance_id),
      onPlayCard: () => {},
      onDrawCard: () => {},
      onActivateAbility: (c: CardView, i: number) => activated.push([c.instance_id, i]),
      onManaAbilityCost: () => {},
    } as never,
  );
  const tile = () =>
    r.container.querySelector<HTMLElement>(`.card[data-instance-id="${card.instance_id}"]`)!;
  const rows = () => [
    ...tile().querySelectorAll<HTMLButtonElement>(".mana-menu .menu-item:not([data-raw-tap])"),
  ];
  const rightClick = () => {
    tile().dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
    flushSync();
  };
  return { ...r, view, sent, tapped, activated, tile, rows, rightClick };
}

type Outcome = "popover" | "mana" | "none";

// What the popover says a click can do: a live non-mana row means the
// popover, else a live mana row means mana, else nothing.
function popoverVerdict(rows: HTMLButtonElement[]): Outcome {
  const live = rows.filter((b) => !b.disabled);
  if (live.some((b) => b.dataset.kind !== "mana")) return "popover";
  if (live.some((b) => b.dataset.kind === "mana")) return "mana";
  return "none";
}

// What the left-click did.
function clickOutcome(p: ReturnType<typeof mountPanel>): Outcome {
  if (get(abilityPopover)?.cardID === p.tile().dataset.instanceId) return "popover";
  if (get(manaSourcePicker) || p.sent.some((s) => s.type === "activate_mana_ability")) {
    return "mana";
  }
  return "none";
}

const EXPECTED: Record<string, Outcome> = {
  "untapped Forest": "mana",
  "tapped Forest": "none",
  "vanilla creature": "none",
  "summoning-sick Llanowar Elves": "none",
  "Birds of Paradise": "mana",
  "Mystic Gate": "mana",
  "tapped Vivi, power 3": "mana",
  "Vivi, power 1": "mana",
  "Vivi, power 0 (adds_no_mana)": "none",
  "Vivi, already used this turn": "none",
  "Relic of Sauron": "popover",
  "Relic of Sauron, draw row refused": "mana",
  "tapped Relic of Sauron": "none",
  "Rogue's Passage": "popover",
  "Prodigal Sorcerer": "popover",
  "an Equipment in your main phase": "popover",
  "an Equipment with its window shut": "none",
  "an Arrested creature": "none",
  "a catalogued planeswalker": "popover",
  "a planeswalker that activated this turn": "none",
  "an uncatalogued planeswalker": "popover",
  "an uncatalogued planeswalker that activated this turn": "none",
  "a face-down creature": "popover",
  "a land under Squirrel Nest": "popover",
};

describe("the click rule and the popover agree (ADR 0117 §2)", () => {
  for (const [name, make] of Object.entries(FIXTURES)) {
    it(name, () => {
      const p = mountPanel(make());
      // The popover, opened the way that never routes through the rule.
      p.rightClick();
      const verdict = popoverVerdict(p.rows());
      expect(verdict).toBe(EXPECTED[name]);
      // Close it, then left-click.
      p.rightClick();
      expect(get(abilityPopover)).toBeNull();
      click(p.tile());
      expect(clickOutcome(p)).toBe(verdict);
      // Never a raw tap from a plain click, and never a refused send.
      expect(p.tapped).toEqual([]);
      // The cursor follows the rule.
      expect(p.tile().classList.contains("clickable")).toBe(verdict !== "none");
    });
  }
});

describe("the two bugs ADR 0117 fixes", () => {
  it("a tapped Vivi at power 3 opens the colour picker and never untaps", () => {
    const p = mountPanel(vivi(3, { tapped: true }));
    click(p.tile());
    expect(get(manaSourcePicker)?.cardID).toBe("vivi");
    expect(p.tapped).toEqual([]);
    expect(p.sent).toEqual([]);
  });

  it("a power-0 Vivi sends nothing, so her once-per-turn use is not spent", () => {
    const p = mountPanel(vivi(0, {}, { adds_no_mana: true }));
    click(p.tile());
    expect(p.sent).toEqual([]);
    expect(get(manaSourcePicker)).toBeNull();
    // The popover keeps the row, greyed with the generic reason.
    p.rightClick();
    const [row] = p.rows();
    expect(row.disabled).toBe(true);
    expect(row.title).toBe("adds no mana right now");
  });
});

describe("the left-click popover", () => {
  it("opens Relic of Sauron's popover at the card with the mana row first", () => {
    const p = mountPanel(relic());
    click(p.tile());
    expect(get(abilityPopover)?.cardID).toBe("relic");
    expect(p.rows().map((b) => b.dataset.kind)).toEqual(["mana", "activated"]);
    expect(p.sent).toEqual([]);
    p.rows()[1].click();
    flushSync();
    expect(p.activated).toEqual([["relic", 0]]);
    expect(get(abilityPopover)).toBeNull();
  });

  it("closes when the card is clicked again", () => {
    const p = mountPanel(relic());
    click(p.tile());
    click(p.tile());
    expect(get(abilityPopover)).toBeNull();
    expect(p.sent).toEqual([]);
  });

  it("emits ability-menu-opened, so the tutorial's step 7 completes this way too", () => {
    const seen: TutorialEvent[] = [];
    const off = subscribe((e) => seen.push(e));
    try {
      const p = mountPanel(relic());
      click(p.tile());
      expect(seen).toEqual(["ability-menu-opened"]);
    } finally {
      off();
    }
  });

  it("opens on Enter, and Enter on a row activates that row", () => {
    const p = mountPanel(relic());
    const key = (el: Element) => {
      el.dispatchEvent(
        new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
      );
      flushSync();
    };
    key(p.tile());
    expect(get(abilityPopover)?.cardID).toBe("relic");
    // Enter on a focused row is the row's own: the card's handler must
    // not take it as another click on the card.
    const row = p.rows()[1];
    const ev = new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true });
    row.dispatchEvent(ev);
    expect(ev.defaultPrevented).toBe(false);
    expect(get(abilityPopover)?.cardID).toBe("relic");
  });

  it("lists an uncatalogued planeswalker's manual loyalty rows, greyed once one was activated", () => {
    const p = mountPanel(walker({ activated_abilities: undefined }));
    click(p.tile());
    const loyalty = [...p.tile().querySelectorAll<HTMLButtonElement>("[data-loyalty]")];
    expect(loyalty.map((b) => b.textContent?.trim())).toContain("+1: activate a loyalty ability");
    expect(loyalty.every((b) => !b.disabled)).toBe(true);
    loyalty.find((b) => b.dataset.loyalty === "loyalty-1")!.click();
    flushSync();
    expect(p.sent).toEqual([
      {
        type: "activate_loyalty",
        params: { planeswalker_id: "walker", delta: 1, label: "+1" },
      },
    ]);

    cleanup();
    const q = mountPanel(walker({ activated_abilities: undefined, loyalty_activated: true }));
    q.rightClick();
    const greyed = [...q.tile().querySelectorAll<HTMLButtonElement>("[data-loyalty]")];
    expect(greyed.length).toBeGreaterThan(0);
    expect(greyed.every((b) => b.disabled)).toBe(true);
    expect(greyed[0].title).toBe("Already activated this turn");
  });

  it("an admin's plain click on another seat's land does nothing; the rule says so", () => {
    const land = permanent("forest", "Forest", "Basic Land — Forest", {
      controller: BOB,
      owner: BOB,
      mana_abilities: [mana(0, { produced: "{G}" })],
    });
    expect(battlefieldClickIntent(land, ME, true)).toBe("none");
    expect(battlefieldClickIntent(land, ME, true, { rawTap: true })).toBe("tap");
  });
});
