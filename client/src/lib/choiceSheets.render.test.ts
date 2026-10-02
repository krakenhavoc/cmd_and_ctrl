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
//     decline ("Fail to find") or a payment ("Pay {4}", Y / N only);
//   - Escape does nothing: a pending choice has no way out.

import { describe, it, expect, afterEach, beforeEach } from "vitest";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
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
  it("search: Fail to find takes no Enter; Take does; Clear is a plain secondary", () => {
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

    const primary = barPrimary()!;
    expect(nameOf(primary)).toBe("Fail to find");
    expect(primary.hasAttribute("aria-keyshortcuts")).toBe(false);
    keyFromBody("Enter");
    expect(sent).toEqual([]);

    const clear = barButton("Clear")!;
    expect(clear.disabled).toBe(true);
    expect(clear.hasAttribute("aria-keyshortcuts")).toBe(false);

    click(inSheet("button.card-pick").find((b) => nameOf(b) === "select Island")!);
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

  it("replacement_order: Apply in this order", () => {
    const { sent } = mount(snap({ kind: "replacement_order", replacement_options: opts }));
    expectSheet("Order replacement effects");
    click(inSheet("button.prompt-opt")[0]);
    click(inSheet("button.prompt-opt")[1]);
    click(barButton("Apply in this order")!);
    expect(sent[0].params).toEqual({ choice_id: "choice-1", order: ["t1", "t2"] });
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
