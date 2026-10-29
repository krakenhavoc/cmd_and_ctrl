// @vitest-environment jsdom
//
// choiceSheets.render.test.ts — ADR 0111 Delivery PR 6 (S56, #1958).
// Every pending-choice kind that is not inline (PR 5) is a SHEET that
// grows up out of the action dock (owner decision 2), drawn by
// ChoicePromptModal through DockSheet. For each kind group:
//
//   - it opens as a sheet (`.dock-sheet`) inside region "actions", in
//     one non-modal dialog named as its modal was;
//   - its primary and secondary are in the dock's action bar and send
//     the right resolve_choice params;
//   - no backdrop, no centred modal, no aria-modal anywhere;
//   - Enter presses a confirm that commits what was picked, and never a
//     decline ("Find nothing") or a payment ("Pay {4}", Y / N only);
//   - Escape does nothing: a pending choice has no way out.

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import { get } from "svelte/store";
import { _resetForTests as resetDock } from "./dock";
import { boardChoicePick, pickOnBoard } from "./boardChoicePick";
import { L } from "./labels";
import { _resetForTests as resetModals } from "./modalLayers";
import { defaultSettings, settings } from "./settings";
import type { ActionType, CardView, GameView, PendingChoiceView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import {
  barButton,
  barPrimary,
  barSecondaries,
  dockDialog,
  dockRegion,
  nameOf,
  pressKey,
  sheetPanel,
} from "./test/dockView";

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
  resetDock();
  resetModals();
});
afterEach(cleanup);

const card = (id: string, name: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: id, name, owner: "me", controller: "me", ...extra }) as CardView;

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, name: string, n: number, hand: CardView[] = []) => ({
  id,
  name,
  seat: n,
  life: 40,
  library: zone("library", id),
  hand: zone("hand", id, hand),
  graveyard: zone("graveyard", id),
  command: zone("command", id),
  commander_damage: {},
  life_history: [],
  mana_pool: [],
});

function snap(
  choice: Partial<PendingChoiceView>,
  over: { battlefield?: CardView[]; hand?: CardView[] } = {},
): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat("me", "Me", 0, over.hand ?? []), seat("them", "Them", 1)],
    battlefield: zone("battlefield", undefined, over.battlefield ?? []),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: [],
    pending_triggers: [],
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "precombat_main",
      step: "precombat_main",
    },
    mulligans_open: false,
    pending_choices: [{ id: "choice-1", chooser: "me", from_player: "me", options: [], ...choice }],
  } as unknown as GameView;
}

interface Sent {
  type: ActionType;
  params?: Record<string, unknown>;
}

function mount(view: GameView) {
  const sent: Sent[] = [];
  const r = render(
    ChoiceDockHarness as never,
    {
      snap: view,
      viewerID: "me",
      sendAction: (type: ActionType, params?: unknown) =>
        sent.push({ type, params: params as Record<string, unknown> }),
      lastError: null,
    } as never,
  );
  return { ...r, sent };
}

// The prompt is a sheet in the dock, in a dialog of exactly this name,
// and nothing of it is a modal.
function expectSheet(name: string): HTMLElement {
  const dlg = dockDialog();
  expect(dlg, `a dialog in the dock`).not.toBeNull();
  expect(dlg!.getAttribute("aria-label")).toBe(name);
  expect(dlg!.dataset.rank).toBe("choice");
  expect(dlg!.hasAttribute("aria-modal")).toBe(false);
  expect(dockRegion()!.contains(dlg)).toBe(true);
  const sheet = sheetPanel();
  expect(sheet).not.toBeNull();
  expect(dlg!.contains(sheet)).toBe(true);
  expect(document.querySelector(".prompt-backdrop")).toBeNull();
  expect(document.querySelector(".prompt-modal")).toBeNull();
  expect(document.querySelector('[aria-modal="true"]')).toBeNull();
  return sheet!;
}

// Keys with focus nowhere in particular (the page body).
function keyFromBody(key: string): void {
  (document.activeElement as HTMLElement | null)?.blur?.();
  pressKey(key);
}

const inSheet = (sel: string) => [...(sheetPanel()?.querySelectorAll<HTMLElement>(sel) ?? [])];

describe("the scry family is a sheet", () => {
  const cards = [card("a", "Island"), card("b", "Forest")];

  it("scry: Done in the bar, Enter sends the lanes, Escape nothing", () => {
    const { sent } = mount(snap({ kind: "scry", reason: "Preordain — scry 2", options: cards }));
    const sheet = expectSheet("Preordain — scry 2");
    expect(sheet.textContent).toContain("CR 701.22");
    expect(sheet.querySelector(".prompt-count")?.textContent).toContain(
      "2 on top · 0 on the bottom",
    );
    const done = barPrimary()!;
    expect(nameOf(done)).toBe("Done");
    expect(done.getAttribute("aria-keyshortcuts")).toBe("Enter");
    expect(barSecondaries()).toHaveLength(0);

    keyFromBody("Escape");
    expect(sent).toEqual([]);

    // Bottom the Island, then Enter.
    click(inSheet("button").find((b) => nameOf(b) === "put Island on the bottom")!);
    keyFromBody("Enter");
    expect(sent).toEqual([
      {
        type: "resolve_choice",
        params: { choice_id: "choice-1", bottom: ["a"], top_order: ["b"] },
      },
    ]);
  });

  it("surveil answers on the graveyard key", () => {
    const { sent } = mount(snap({ kind: "surveil", options: cards }));
    expectSheet("Surveil");
    click(barPrimary()!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", graveyard: [], top_order: ["a", "b"] });
  });

  it("look_at_top is a pure reorder", () => {
    const { sent } = mount(snap({ kind: "look_at_top", options: cards }));
    expectSheet("Look at the top");
    click(barPrimary()!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", top_order: ["a", "b"] });
  });
});

describe("a card grid is a sheet", () => {
  it("search: Find nothing takes no Enter; Take does; Clear is a plain secondary", () => {
    const opts = [card("f", "Forest"), card("i", "Island")];
    const { sent } = mount(
      snap({
        kind: "search_library",
        reason: "Solemn Simulacrum — a basic land",
        search_max: 1,
        options: opts,
      }),
    );
    const sheet = expectSheet("Solemn Simulacrum — a basic land");
    expect(sheet.querySelectorAll("button.card-pick")).toHaveLength(2);

    // #2964: nothing picked means no primary; Find nothing is a plain secondary.
    expect(barPrimary()).toBeNull();
    const fail = barButton("Find nothing")!;
    expect(barSecondaries()).toContain(fail);
    expect(fail.hasAttribute("aria-keyshortcuts")).toBe(false);
    keyFromBody("Enter");
    expect(sent).toEqual([]);

    const clear = barButton("Clear")!;
    expect(clear.disabled).toBe(true);
    expect(clear.hasAttribute("aria-keyshortcuts")).toBe(false);

    click(inSheet("button.card-pick").find((b) => nameOf(b) === "select Island")!);
    // #3053: declining stays one click away while a card is picked.
    expect(barButton("Find nothing")).not.toBeNull();
    expect(nameOf(barPrimary()!)).toBe("Take");
    expect(barPrimary()!.getAttribute("aria-keyshortcuts")).toBe("Enter");
    expect(sheet.querySelector(".prompt-count")?.textContent).toContain("1 / 1 selected");
    keyFromBody("Escape");
    expect(sent).toEqual([]);
    keyFromBody("Enter");
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: ["i"] } },
    ]);
  });

  it("search: a lone match opens taken; Enter takes it, Find nothing declines (#3053)", () => {
    // The search the engine used to resolve without asking. It asks now
    // (CR 701.23b), and the default is the answer nearly everyone gives.
    const lone = [card("f", "Forest")];
    const a = mount(
      snap({
        kind: "search_library",
        reason: "Rampant Growth — a basic land",
        search_max: 1,
        options: lone,
      }),
    );
    expectSheet("Rampant Growth — a basic land");
    expect(nameOf(barPrimary()!)).toBe("Take");
    expect(barPrimary()!.disabled).toBe(false);
    expect(barPrimary()!.getAttribute("aria-keyshortcuts")).toBe("Enter");
    expect(sheetPanel()!.querySelector(".prompt-count")?.textContent).toContain("1 / 1 selected");
    keyFromBody("Enter");
    expect(a.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: ["f"] } },
    ]);
    cleanup();

    const b = mount(
      snap({
        kind: "search_library",
        reason: "Rampant Growth — a basic land",
        search_max: 1,
        options: lone,
      }),
    );
    const none = barButton("Find nothing")!;
    expect(barSecondaries()).toContain(none);
    expect(none.hasAttribute("aria-keyshortcuts")).toBe(false);
    click(none);
    expect(b.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: [] } },
    ]);
  });

  it("search: a search that must find offers no Find nothing and waits for the pick (#3053)", () => {
    const opts = [card("a", "Bear"), card("b", "Elf")];
    const { sent } = mount(
      snap({
        kind: "search_library",
        reason: "Demonic Tutor — a card",
        search_max: 1,
        search_min: 1,
        options: opts,
      }),
    );
    expectSheet("Demonic Tutor — a card");
    expect(barButton("Find nothing")).toBeNull();
    expect(nameOf(barPrimary()!)).toBe("Take");
    expect(barPrimary()!.disabled).toBe(true);
    click(inSheet("button.card-pick")[1]);
    expect(barPrimary()!.disabled).toBe(false);
    keyFromBody("Enter");
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: ["b"] } },
    ]);
  });

  it("search: a you-may search asks before it looks, with no cards shown (#4134)", () => {
    // Looking through the library is searching (CR 701.19): the question
    // comes first, so declining is never a free look.
    const a = mount(
      snap({
        kind: "search_library",
        reason: "Path to Exile — you may search for a basic land",
        search_max: 1,
        search_optional: true,
        search_undecided: true,
        options: [],
      }),
    );
    expectSheet("Path to Exile — you may search for a basic land");
    expect(inSheet("button.card-pick")).toHaveLength(0);
    expect(barButton("Find nothing")).toBeNull();
    expect(nameOf(barPrimary()!)).toBe("Search");
    const dont = barButton("Don't search")!;
    expect(barSecondaries()).toContain(dont);
    keyFromBody("Enter");
    expect(a.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", apply: true } },
    ]);
    cleanup();

    const b = mount(
      snap({
        kind: "search_library",
        reason: "Path to Exile — you may search for a basic land",
        search_max: 1,
        search_optional: true,
        search_undecided: true,
        options: [],
      }),
    );
    click(barButton("Don't search")!);
    // Not searching is apply:false, never the empty card list: that is a
    // search that finds nothing, and it still shuffles.
    expect(b.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", apply: false } },
    ]);
  });

  it("search: once the search has begun there is no Don't search, only Find nothing (#4134)", () => {
    const opts = [card("f", "Forest"), card("i", "Island")];
    const a = mount(
      snap({
        kind: "search_library",
        reason: "Path to Exile — you may search for a basic land",
        search_max: 1,
        options: opts,
      }),
    );
    expectSheet("Path to Exile — you may search for a basic land");
    expect(inSheet("button.card-pick")).toHaveLength(2);
    expect(barButton("Don't search")).toBeNull();
    click(barButton("Find nothing")!);
    expect(a.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: [] } },
    ]);
  });

  it("search: a prompt with no candidates still asks, then offers only Find nothing (#4140)", () => {
    // Optional: the question first, as for any optional search.
    const a = mount(
      snap({
        kind: "search_library",
        reason: "Path to Exile — you may search for a basic land",
        search_max: 1,
        search_optional: true,
        search_undecided: true,
        options: [],
      }),
    );
    expect(nameOf(barPrimary()!)).toBe("Search");
    click(barButton("Don't search")!);
    expect(a.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", apply: false } },
    ]);
    cleanup();

    // Begun (or mandatory): an empty list and Find nothing.
    const b = mount(
      snap({
        kind: "search_library",
        reason: "Rampant Growth — a basic land",
        search_max: 1,
        options: [],
      }),
    );
    expectSheet("Rampant Growth — a basic land");
    expect(inSheet("button.card-pick")).toHaveLength(0);
    expect(barPrimary()).toBeNull();
    expect(barButton("Don't search")).toBeNull();
    expect(sheetPanel()!.textContent).toContain("Nothing in the library matches");
    click(barButton("Find nothing")!);
    expect(b.sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: [] } },
    ]);
  });

  it("search: a mandatory search has no Don't search (#4134)", () => {
    mount(
      snap({
        kind: "search_library",
        reason: "Rampant Growth — a basic land",
        search_max: 1,
        options: [card("f", "Forest"), card("i", "Island")],
      }),
    );
    expectSheet("Rampant Growth — a basic land");
    expect(barButton("Don't search")).toBeNull();
  });

  it("search: a search the viewer answers for another player names whose it is (#3998)", () => {
    // Opposition Agent: the searcher's prompt arrives addressed to the
    // controller, on the searcher's behalf.
    const opts = [card("a", "Bear")];
    const { sent } = mount(
      snap({
        kind: "search_library",
        reason: "Demonic Tutor — a card",
        on_behalf_of: "them",
        search_max: 1,
        search_min: 1,
        options: opts,
      }),
    );
    expectSheet("Demonic Tutor — a card (choosing for Them)");
    // A lone match opens taken (#3053).
    keyFromBody("Enter");
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: ["a"] } },
    ]);
  });

  it("search: library and graveyard copies of one name are two tiles, the graveyard one marked (#3217)", () => {
    const { sent } = mount(
      snap({
        kind: "search_library",
        reason: "Tower Winder — a card named Command Tower",
        search_max: 1,
        options: [card("lib", "Command Tower"), card("gy", "Command Tower")],
        search_graveyard: ["gy"],
      }),
    );
    expectSheet("Tower Winder — a card named Command Tower");
    const tiles = inSheet("button.card-pick");
    expect(tiles).toHaveLength(2);
    expect(tiles[0].textContent).not.toContain("Graveyard");
    expect(tiles[1].textContent).toContain("Graveyard");
    // "And/or" lets the searcher look only in the library: Find nothing stays.
    expect(barButton("Find nothing")).not.toBeNull();
    click(tiles[1]);
    keyFromBody("Enter");
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: ["gy"] } },
    ]);
  });

  describe("search: which zones were searched (#4516)", () => {
    const andOr = () =>
      snap({
        kind: "search_library",
        reason: "Tower Winder — a card named Command Tower",
        search_max: 1,
        options: [card("lib", "Command Tower"), card("gy", "Command Tower")],
        search_graveyard: ["gy"],
        search_library_optional: true,
      });
    const toggle = () =>
      inSheet("button.remember-toggle").find((b) =>
        b.textContent?.includes(L.searchedMyLibrary),
      ) as HTMLButtonElement | undefined;

    it("a graveyard pick says nothing about the library until the toggle is on", () => {
      const { sent } = mount(andOr());
      const t = toggle();
      expect(t).toBeDefined();
      expect(t!.getAttribute("aria-pressed")).toBe("false");
      click(inSheet("button.card-pick")[1]);
      expect(t!.disabled).toBe(false);
      keyFromBody("Enter");
      expect(sent).toEqual([
        { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: ["gy"] } },
      ]);
    });

    it("the toggle sends library_searched with a graveyard pick", () => {
      const { sent } = mount(andOr());
      click(inSheet("button.card-pick")[1]);
      click(toggle()!);
      expect(toggle()!.getAttribute("aria-pressed")).toBe("true");
      keyFromBody("Enter");
      expect(sent).toEqual([
        {
          type: "resolve_choice",
          params: { choice_id: "choice-1", card_ids: ["gy"], library_searched: true },
        },
      ]);
    });

    it("a library pick is a library search: the toggle reads on and cannot be turned off", () => {
      const { sent } = mount(andOr());
      click(inSheet("button.card-pick")[0]);
      expect(toggle()!.getAttribute("aria-pressed")).toBe("true");
      expect(toggle()!.disabled).toBe(true);
      keyFromBody("Enter");
      expect(sent).toEqual([
        {
          type: "resolve_choice",
          params: { choice_id: "choice-1", card_ids: ["lib"], library_searched: true },
        },
      ]);
    });

    it("Find nothing is a graveyard-only search unless the toggle is on", () => {
      const first = mount(andOr());
      click(barButton("Find nothing")!);
      expect(first.sent).toEqual([
        { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: [] } },
      ]);
      cleanup();

      const second = mount(andOr());
      click(toggle()!);
      click(barButton("Find nothing")!);
      expect(second.sent).toEqual([
        {
          type: "resolve_choice",
          params: { choice_id: "choice-1", card_ids: [], library_searched: true },
        },
      ]);
    });

    it("a search without the flag has no toggle", () => {
      mount(
        snap({
          kind: "search_library",
          reason: "Fetchland — choose a land",
          search_max: 1,
          options: [card("a", "Island"), card("b", "Swamp")],
        }),
      );
      expect(toggle()).toBeUndefined();
    });
  });

  it("search: graveyard, hand and library copies are three tiles, each marked (#3988)", () => {
    const { sent } = mount(
      snap({
        kind: "search_library",
        reason: "Kotose, the Silent Spider — cards named Lightning Bolt",
        search_max: 3,
        options: [
          card("lib", "Lightning Bolt"),
          card("gy", "Lightning Bolt"),
          card("hand", "Lightning Bolt"),
        ],
        search_graveyard: ["gy"],
        search_hand: ["hand"],
        search_all_zones: true,
      }),
    );
    expectSheet("Kotose, the Silent Spider — cards named Lightning Bolt");
    const tiles = inSheet("button.card-pick");
    expect(tiles).toHaveLength(3);
    expect(tiles[0].textContent).not.toContain("Graveyard");
    expect(tiles[0].textContent).not.toContain("Hand");
    expect(tiles[1].textContent).toContain("Graveyard");
    expect(tiles[2].textContent).toContain("Hand");
    expect(sheetPanel()!.textContent).toContain("the library is shuffled either way");
    expect(barButton("Find nothing")).not.toBeNull();
    // Every match fits under the ceiling, so all three open taken (#3053);
    // releasing the library copy leaves the graveyard and hand copies.
    expect(tiles.every((t) => t.getAttribute("aria-pressed") === "true")).toBe(true);
    click(tiles[0]);
    keyFromBody("Enter");
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: ["gy", "hand"] } },
    ]);
  });

  it("search: identical copies are one tile; a pick takes one copy", () => {
    const mountains = [1, 2, 3].map((n) => card(`m${n}`, "Mountain"));
    const { sent } = mount(
      snap({
        kind: "search_library",
        reason: "Fetchland — choose a land",
        search_max: 2,
        options: [...mountains, card("i", "Island")],
      }),
    );
    expectSheet("Fetchland — choose a land");
    const tiles = inSheet("button.card-pick");
    expect(tiles.map(nameOf)).toEqual(["select Mountain", "select Island"]);
    expect(tiles[0].textContent).toContain("×3");
    expect(tiles[0].querySelector(".pick-name")?.textContent).toBe("Mountain");

    click(tiles[0]);
    click(tiles[0]);
    expect(inSheet("button.card-pick")[0].getAttribute("aria-pressed")).toBe("true");
    const sheet = sheetPanel()!;
    expect(sheet.querySelector(".prompt-count")?.textContent).toContain("2 / 2 selected");

    // Full: a further click on the group releases one of its copies.
    click(inSheet("button.card-pick")[0]);
    expect(sheet.querySelector(".prompt-count")?.textContent).toContain("1 / 2 selected");
    click(inSheet("button.card-pick")[0]);
    keyFromBody("Enter");
    expect(sent).toHaveLength(1);
    expect(sent[0].type).toBe("resolve_choice");
    expect([...(sent[0].params?.card_ids as string[])].sort()).toEqual(["m1", "m2"]);
  });

  it("sacrifice: the verb waits for the pick, then Enter sends it", () => {
    const { sent } = mount(
      snap({
        kind: "sacrifice_choice",
        reason: "Grave Pact — sacrifice a creature",
        count: 1,
        options: [card("c1", "Bear"), card("c2", "Elf")],
      }),
    );
    expectSheet("Grave Pact — sacrifice a creature");
    expect(nameOf(barPrimary()!)).toBe("Sacrifice");
    expect(barPrimary()!.disabled).toBe(true);
    expect(barButton("Clear")).toBeNull();
    click(inSheet("button.card-pick")[1]);
    keyFromBody("Enter");
    expect(sent[0].params).toEqual({ choice_id: "choice-1", card_ids: ["c2"] });
  });

  it("the default grid is named as its heading was", () => {
    mount(
      snap({ kind: "discard_choice", reason: "Mind Rot", count: 2, options: [card("x", "X")] }),
    );
    expectSheet("Mind Rot — pick 2 cards");
  });

  it("an entry discard's decline takes no Enter", () => {
    const { sent } = mount(
      snap({
        kind: "entry_discard_from_hand",
        reason: "Mox Diamond — discard a land?",
        choose_min: 0,
        choose_max: 1,
        options: [card("f", "Forest")],
      }),
    );
    expectSheet("Mox Diamond — discard a land?");
    expect(barPrimary()!.hasAttribute("aria-keyshortcuts")).toBe(false);
    keyFromBody("Enter");
    expect(sent).toEqual([]);
  });
});

describe("damage assignment is a sheet", () => {
  it("Deal damage waits for the whole power, then Enter sends the split", () => {
    const { sent } = mount(
      snap(
        {
          kind: "damage_assignment",
          damage_assignment: {
            attacker_card_id: "atk",
            attacker_power: 4,
            blocker_card_ids: ["b1", "b2"],
          },
        } as Partial<PendingChoiceView>,
        {
          battlefield: [card("atk", "Craw Wurm"), card("b1", "Bear"), card("b2", "Elf")],
        },
      ),
    );
    const sheet = expectSheet("Assign combat damage");
    const deal = barPrimary()!;
    expect(nameOf(deal)).toBe("Deal damage");
    expect(deal.disabled).toBe(true);
    const inputs = [...sheet.querySelectorAll<HTMLInputElement>('input[type="number"]')];
    for (const [el, v] of [
      [inputs[0], "2"],
      [inputs[1], "2"],
    ] as const) {
      el.value = v;
      el.dispatchEvent(new Event("input", { bubbles: true }));
    }
    flushSync();
    expect(sheet.querySelector(".prompt-count")?.textContent).toContain("4 / 4 assigned");
    keyFromBody("Escape");
    expect(sent).toEqual([]);
    keyFromBody("Enter");
    expect(sent[0].params).toEqual({
      choice_id: "choice-1",
      assignments: [
        { blocker_id: "b1", amount: 2 },
        { blocker_id: "b2", amount: 2 },
      ],
      trample_to_player: 0,
    });
  });
});

describe("the order kinds are sheets", () => {
  const opts = [
    { id: "t1", label: "Draw a card" },
    { id: "t2", label: "Gain 2 life" },
  ];

  it("trigger_order: Resolve in this order once every trigger is placed", () => {
    const { sent } = mount(snap({ kind: "trigger_order", trigger_options: opts }));
    expectSheet("Order your triggers");
    expect(nameOf(barPrimary()!)).toBe("Resolve in this order");
    expect(barPrimary()!.disabled).toBe(true);
    click(inSheet("button.prompt-opt")[1]);
    click(inSheet("button.prompt-opt")[0]);
    keyFromBody("Enter");
    expect(sent[0].params).toEqual({ choice_id: "choice-1", order: ["t2", "t1"] });
  });

  it("trigger_order: says why it asked, under the heading, when the server gave a reason (#3015)", () => {
    const why = "Ajani's Pridemate watches life totals.";
    const { sent } = mount(
      snap({ kind: "trigger_order", trigger_options: opts, trigger_order_why: why }),
    );
    expectSheet("Order your triggers");
    const line = inSheet(".prompt-why");
    expect(line).toHaveLength(1);
    expect(line[0].textContent?.trim()).toBe(why);
    // It is text only: the prompt is answered exactly as before.
    click(inSheet("button.prompt-opt")[1]);
    click(inSheet("button.prompt-opt")[0]);
    click(barButton("Resolve in this order")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", order: ["t2", "t1"] });
  });

  it("trigger_order: no reason, no line; replacement_order never has one", () => {
    mount(snap({ kind: "trigger_order", trigger_options: opts }));
    expectSheet("Order your triggers");
    expect(inSheet(".prompt-why")).toHaveLength(0);
    cleanup();
    mount(
      snap({ kind: "replacement_order", replacement_options: opts, trigger_order_why: "stray" }),
    );
    expect(inSheet(".prompt-why")).toHaveLength(0);
  });

  it("trigger_order: Auto-order answers in the order a skipped batch stacks", () => {
    const { sent } = mount(snap({ kind: "trigger_order", trigger_options: opts }));
    click(barButton("Auto-order")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", order: ["t2", "t1"] });
  });

  it("replacement_order has no Auto-order", () => {
    mount(snap({ kind: "replacement_order", replacement_options: opts }));
    expect(barButton("Auto-order")).toBeFalsy();
  });

  it("replacement_order: Apply in this order", () => {
    const { sent } = mount(snap({ kind: "replacement_order", replacement_options: opts }));
    expectSheet("Order replacement effects");
    click(inSheet("button.prompt-opt")[0]);
    click(inSheet("button.prompt-opt")[1]);
    click(barButton("Apply in this order")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", order: ["t1", "t2"] });
  });
});

// #3016 (ADR 0018's eighth amendment): "Use this order again".
describe("trigger_order: Use this order again", () => {
  const rows = [
    { id: "p1", label: "Plunderer", row: 1 },
    { id: "n", label: "Noble", row: 2 },
    { id: "p2", label: "Plunderer", row: 1 },
  ];
  const keyedSnap = (over: Partial<PendingChoiceView> = {}) =>
    snap({
      kind: "trigger_order",
      trigger_options: rows,
      auto_answer_key: "trigger_order|noble+plunderer",
      ...over,
    });
  const toggle = () =>
    inSheet("button.remember-toggle").find((b) => b.textContent?.includes(L.useThisOrderAgain));

  it("is absent on a prompt with no key", () => {
    mount(snap({ kind: "trigger_order", trigger_options: rows }));
    expect(toggle()).toBeUndefined();
  });

  it("is off by default, and sends remember with the answer", () => {
    const { sent } = mount(keyedSnap());
    const t = toggle()!;
    expect(t.getAttribute("aria-pressed")).toBe("false");
    click(t);
    expect(toggle()!.getAttribute("aria-pressed")).toBe("true");
    click(inSheet("button.prompt-opt")[1]);
    click(inSheet("button.prompt-opt")[0]);
    click(inSheet("button.prompt-opt")[2]);
    click(barButton("Resolve in this order")!);
    expect(sent[0].params).toEqual({
      choice_id: "choice-1",
      order: ["n", "p1", "p2"],
      remember: true,
    });
  });

  it("starts on when the batch already has a remembered order", () => {
    const { sent } = mount(keyedSnap({ trigger_order_remembered: true }));
    expect(toggle()!.getAttribute("aria-pressed")).toBe("true");
    click(toggle()!);
    click(inSheet("button.prompt-opt")[0]);
    click(inSheet("button.prompt-opt")[2]);
    click(inSheet("button.prompt-opt")[1]);
    click(barButton("Resolve in this order")!);
    expect(sent[0].params).toEqual({
      choice_id: "choice-1",
      order: ["p1", "p2", "n"],
      remember: false,
    });
  });

  it("is disabled, with its hint, on an order that splits a card's triggers", () => {
    const { sent } = mount(keyedSnap({ trigger_order_remembered: true }));
    click(inSheet("button.prompt-opt")[0]);
    click(inSheet("button.prompt-opt")[1]);
    click(inSheet("button.prompt-opt")[2]);
    const t = toggle() as HTMLButtonElement;
    expect(t.disabled).toBe(true);
    expect(t.getAttribute("aria-pressed")).toBe("false");
    expect(sheetPanel()?.textContent).toContain(
      "Only an order that keeps each card's triggers together can be reused.",
    );
    click(barButton("Resolve in this order")!);
    expect(sent[0].params).toEqual({
      choice_id: "choice-1",
      order: ["p1", "n", "p2"],
      remember: false,
    });
  });
});

describe("mode_pick is a sheet", () => {
  it("is named by its source, and Choose takes Enter once a mode is picked", () => {
    const { sent } = mount(
      snap(
        {
          kind: "mode_pick",
          source: "src",
          reason: "choose one",
          mode_options: ["Draw a card", "Gain 3 life"],
          mode_indexes: [0, 1],
          mode_min: 1,
          mode_max: 1,
        },
        { battlefield: [card("src", "Glissa Sunslayer")] },
      ),
    );
    expectSheet("Glissa Sunslayer");
    expect(nameOf(barPrimary()!)).toBe("Choose");
    expect(barPrimary()!.disabled).toBe(true);
    click(inSheet("button.prompt-opt")[1]);
    keyFromBody("Enter");
    expect(sent[0].params).toEqual({ choice_id: "choice-1", modes: [1] });
  });
});

describe("the type and name pickers are sheets", () => {
  it("creature type: the filter takes focus, a click answers, no bar", () => {
    const { sent } = mount(
      snap({ kind: "choose_creature_type", type_options: ["Elf", "Goblin", "Sliver"] }),
    );
    const sheet = expectSheet("Choose a creature type");
    expect(document.activeElement).toBe(sheet.querySelector("input.type-filter"));
    expect(barPrimary()).toBeNull();
    click(inSheet("button.type-pick").find((b) => nameOf(b) === "Goblin")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", creature_type: "Goblin" });
  });

  it("card name: Name “…” waits for text, then Enter sends it", () => {
    const { sent } = mount(
      snap({
        kind: "choose_card_name",
        reason: "Pithing Needle — choose a card name",
        name_options: ["Sol Ring"],
      }),
    );
    const sheet = expectSheet("Pithing Needle — choose a card name");
    expect(barPrimary()!.disabled).toBe(true);
    const box = sheet.querySelector<HTMLInputElement>("input.type-filter")!;
    box.value = "Rhystic Study";
    box.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(nameOf(barPrimary()!)).toBe("Name “Rhystic Study”");
    keyFromBody("Enter");
    expect(sent[0].params).toEqual({ choice_id: "choice-1", card_name: "Rhystic Study" });
  });
});

describe("a long option_pick is a sheet", () => {
  it("lists its options in the sheet, and a click answers", () => {
    const { sent } = mount(
      snap({
        kind: "option_pick",
        reason: "Fact or Fiction — choose a pile",
        pick_options: [
          { label: "Pile 1", cards: [card("c1", "Island")] },
          { label: "Pile 2", cards: [] },
        ],
      }),
    );
    expectSheet("Fact or Fiction — choose a pile");
    expect(barPrimary()).toBeNull();
    keyFromBody("Enter");
    keyFromBody("Escape");
    expect(sent).toEqual([]);
    click(inSheet("button.pick-option")[1]);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", option_index: 1 });
  });
});

describe("pay_unless with picks is a sheet", () => {
  it("pay_cards: Pay waits for the picks, answers Y / N, never Enter", () => {
    const { sent } = mount(
      snap(
        {
          kind: "pay_unless",
          reason: "Echo — discard a card?",
          pay_cost: "Discard a card",
          pay_cards: { action: "discard", count: 1, options: ["h1", "h2"] },
        } as Partial<PendingChoiceView>,
        { hand: [card("h1", "Opt"), card("h2", "Ponder")] },
      ),
    );
    expectSheet("Echo — discard a card?");
    const pay = barPrimary()!;
    expect(nameOf(pay)).toBe("Pay Discard a card");
    expect(pay.getAttribute("aria-keyshortcuts")).toBe("Y");
    expect(pay.disabled).toBe(true);
    const [dont] = barSecondaries();
    expect(nameOf(dont)).toBe("Don't pay");
    expect(dont.getAttribute("aria-keyshortcuts")).toBe("N");

    click(inSheet("button.prompt-opt").find((b) => nameOf(b) === "Ponder")!);
    keyFromBody("Enter");
    keyFromBody("Escape");
    expect(sent).toEqual([]);
    keyFromBody("y");
    expect(sent[0].params).toEqual({ choice_id: "choice-1", apply: true, card_ids: ["h2"] });
  });

  it("a waterbend tap list: the taps ride the Pay answer", () => {
    const { sent } = mount(
      snap(
        {
          kind: "pay_unless",
          reason: "Ward — waterbend {4}?",
          pay_cost: "{4}",
          tap_cost: {
            key: "waterbend",
            label: "Waterbend {4}",
            max: 4,
            options: { cards: ["a1"] },
          },
        } as Partial<PendingChoiceView>,
        { battlefield: [card("a1", "Sol Ring")] },
      ),
    );
    expectSheet("Ward — waterbend {4}?");
    click(inSheet("button.prompt-opt")[0]);
    click(barButton("Pay {4}")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", apply: true, tap_ids: ["a1"] });
  });

  it("Don't pay declines", () => {
    const { sent } = mount(
      snap({
        kind: "pay_unless",
        reason: "Ward — waterbend {4}?",
        pay_cost: "{4}",
        tap_cost: { key: "waterbend", max: 4, options: { cards: [] } },
      } as Partial<PendingChoiceView>),
    );
    click(barButton("Don't pay")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", apply: false });
  });
});

// ADR 0187 §5, §8 (#3575): an either payment is one question with a button
// for each way to pay; the answer names the branch.
describe("an either payment is a sheet with a button for each way to pay", () => {
  const worms = (
    over: Partial<PendingChoiceView> = {},
    sacrificePayable = true,
  ): Partial<PendingChoiceView> =>
    ({
      kind: "pay_unless",
      reason: "Worms of the Earth — sacrifice two lands or take 5 damage?",
      pay_cost: "sacrifice two lands or take 5 damage",
      pay_either: [
        {
          action: "sacrifice",
          count: 2,
          words: "sacrifice two lands",
          payable: sacrificePayable,
          options: sacrificePayable ? ["l1", "l2", "l3"] : undefined,
        },
        { action: "damage", count: 5, words: "take 5 damage", payable: true },
      ],
      ...over,
    }) as Partial<PendingChoiceView>;
  const lands = [card("l1", "Forest"), card("l2", "Island"), card("l3", "Swamp")];

  it("Take 5 damage answers at once with pay_option 1", () => {
    const { sent } = mount(snap(worms(), { battlefield: lands }));
    expectSheet("Worms of the Earth — sacrifice two lands or take 5 damage?");
    click(barButton("Take 5 damage")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", apply: true, pay_option: 1 });
  });

  it("Sacrifice two lands waits for the picks, then names branch 0 and the lands", () => {
    const { sent } = mount(snap(worms(), { battlefield: lands }));
    const sacrifice = barPrimary()!;
    expect(nameOf(sacrifice)).toBe("Sacrifice two lands");
    expect(sacrifice.disabled).toBe(true);
    click(inSheet("button.prompt-opt").find((b) => nameOf(b) === "Island")!);
    expect(barPrimary()!.disabled).toBe(true);
    click(inSheet("button.prompt-opt").find((b) => nameOf(b) === "Swamp")!);
    expect(barPrimary()!.disabled).toBe(false);
    click(barPrimary()!);
    expect(sent[0].params).toEqual({
      choice_id: "choice-1",
      apply: true,
      pay_option: 0,
      card_ids: ["l2", "l3"],
    });
  });

  it("never answers on Y or Enter: yes names no branch", () => {
    const { sent } = mount(snap(worms(), { battlefield: lands }));
    keyFromBody("y");
    keyFromBody("Enter");
    expect(sent).toEqual([]);
  });

  it("Don't pay declines with no pay_option", () => {
    const { sent } = mount(snap(worms(), { battlefield: lands }));
    click(barButton("Don't pay")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", apply: false });
  });

  it("a sacrifice the seat can't make is greyed, and the damage still answers", () => {
    const { sent } = mount(snap(worms({}, false), { battlefield: lands }));
    const sacrifice = barPrimary()!;
    expect(nameOf(sacrifice)).toBe("Sacrifice two lands");
    expect(sacrifice.disabled).toBe(true);
    click(sacrifice);
    expect(sent).toEqual([]);
    click(barButton("Take 5 damage")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", apply: true, pay_option: 1 });
  });
});

describe("a sheet's refusal", () => {
  it("is shown in the dock as Not accepted, and the sheet stays open", () => {
    const r = mount(snap({ kind: "scry", reason: "Opt — scry 1", options: [card("a", "Island")] }));
    click(barPrimary()!);
    expect(r.sent).toHaveLength(1);
    r.setProps({
      lastError: { code: "bad_request", message: "that is not a legal order", at: new Date() },
    } as never);
    const alert = dockDialog()!.querySelector('[role="alert"]');
    expect(alert?.textContent).toContain("Not accepted");
    expect(alert?.textContent).toContain("that is not a legal order");
    expectSheet("Opt — scry 1");
  });
});

// #2880: a choice whose options are permanents on the battlefield shares
// its selection with the board (lib/boardChoicePick.ts). Board.svelte's
// click is pickOnBoard; here it is called directly, as the board would.
describe("a sacrifice picked on the board", () => {
  const nazgul = [card("n1", "Nazgûl"), card("n2", "Nazgûl")];
  const fleshbag = (over: Partial<PendingChoiceView> = {}) =>
    snap(
      {
        kind: "sacrifice_choice",
        reason: "Fleshbag Marauder — sacrifice a creature",
        count: 1,
        options: nazgul,
        ...over,
      },
      { battlefield: [...nazgul, card("f", "Forest")] },
    );

  afterEach(() => boardChoicePick.set(null));

  it("publishes the permanents the board may pick, and clears them with the prompt", () => {
    const r = mount(fleshbag());
    const s = get(boardChoicePick)!;
    expect(s.choiceID).toBe("choice-1");
    expect([...s.eligible].sort()).toEqual(["n1", "n2"]);
    expect(s.selected.size).toBe(0);
    expect([s.min, s.max]).toEqual([1, 1]);
    expect(sheetPanel()!.textContent).toContain("click the highlighted permanents on the board");
    r.setProps({ snap: { ...fleshbag(), pending_choices: [] } } as never);
    flushSync();
    expect(get(boardChoicePick)).toBeNull();
  });

  it("a board click picks in the sheet, enables the dock's Sacrifice, and sends the sheet's answer", () => {
    const { sent } = mount(fleshbag());
    expect(nameOf(barPrimary()!)).toBe(L.sacrifice);
    expect(barPrimary()!.disabled).toBe(true);

    expect(pickOnBoard("n2")).toBe(true);
    flushSync();
    const tiles = inSheet("button.card-pick");
    expect(tiles[1].getAttribute("aria-pressed")).toBe("true");
    expect(tiles[0].getAttribute("aria-pressed")).toBe("false");
    expect(sheetPanel()!.querySelector(".prompt-count")?.textContent).toContain("1 / 1 selected");
    expect(barPrimary()!.disabled).toBe(false);

    click(barPrimary()!);
    expect(sent).toEqual([
      { type: "resolve_choice", params: { choice_id: "choice-1", card_ids: ["n2"] } },
    ]);
  });

  it("a second board click puts it back, and the dock waits again", () => {
    mount(fleshbag());
    pickOnBoard("n1");
    flushSync();
    expect(barPrimary()!.disabled).toBe(false);
    pickOnBoard("n1");
    flushSync();
    expect(barPrimary()!.disabled).toBe(true);
    expect(inSheet("button.card-pick")[0].getAttribute("aria-pressed")).toBe("false");
  });

  it("a pick in the sheet shows on the board", () => {
    mount(fleshbag());
    click(inSheet("button.card-pick")[0]);
    flushSync();
    expect([...get(boardChoicePick)!.selected]).toEqual(["n1"]);
  });

  it("a permanent the choice does not offer is not picked", () => {
    mount(fleshbag());
    expect(pickOnBoard("f")).toBe(true);
    flushSync();
    expect(get(boardChoicePick)!.selected.size).toBe(0);
    expect(barPrimary()!.disabled).toBe(true);
  });

  it("choose N of M: picks up to the ceiling, and the confirm waits for the floor", () => {
    const perms = [card("a", "Bear"), card("b", "Elf"), card("c", "Ogre")];
    const { sent } = mount(
      snap(
        {
          kind: "own_permanents",
          reason: "Choose two of your permanents",
          choose_min: 2,
          choose_max: 2,
          options: perms,
        },
        { battlefield: perms },
      ),
    );
    pickOnBoard("a");
    flushSync();
    expect(barPrimary()!.disabled).toBe(true);
    pickOnBoard("c");
    pickOnBoard("b");
    flushSync();
    expect(sheetPanel()!.querySelector(".prompt-count")?.textContent).toContain("2 / 2 selected");
    click(barPrimary()!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", card_ids: ["a", "c"] });
  });

  it("a grid of cards in a hand publishes nothing", () => {
    mount(
      snap({
        kind: "discard_choice",
        reason: "Mind Rot",
        count: 1,
        options: [card("x", "X")],
      }),
    );
    expect(get(boardChoicePick)).toBeNull();
  });
});
