// @vitest-environment jsdom
//
// costSheets.render.test.ts — ADR 0111 Delivery PR 6 (S56, #1958). Every
// cost picker Board.svelte mounts during a cast or an activation is a
// sheet that grows up out of the action dock (owner decision 2), no
// longer a centred modal over a blurred board. For each one: it opens
// as a `.dock-sheet` inside region "actions", in a non-modal dialog
// named as the modal was; its confirm is the action bar's primary and
// its Cancel (Back, for the division) the bar's secondary, each sending
// what the modal's footer sent; no backdrop and no `aria-modal` is
// left; and Enter confirms and Escape cancels through the dock's one key
// handler (the pickers' own document listeners are gone).
//
// (The X picker is pinned by dockSheet.render.test.ts.)

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";
import type { Component } from "svelte";

vi.mock("./api", async (orig) => ({
  ...(await orig<typeof import("./api")>()),
  fetchAutoTapPreview: vi.fn(async () => ({ ok: true, cost: "{2}", plan: [] })),
}));

import PhyrexianCostModal from "./components/board/PhyrexianCostModal.svelte";
import SacrificeCostModal from "./components/board/SacrificeCostModal.svelte";
import CrewCostModal from "./components/board/CrewCostModal.svelte";
import CounterCostModal from "./components/board/CounterCostModal.svelte";
import FacePickerModal from "./components/board/FacePickerModal.svelte";
import AlternativeCostModal from "./components/board/AlternativeCostModal.svelte";
import AltCostPaymentModal from "./components/board/AltCostPaymentModal.svelte";
import DiscardCostModal from "./components/board/DiscardCostModal.svelte";
import DivideDamageModal from "./components/board/DivideDamageModal.svelte";
import TapCostModal from "./components/board/TapCostModal.svelte";
import DelveCostModal from "./components/board/DelveCostModal.svelte";
import ModePickerModal from "./components/board/ModePickerModal.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { defaultSettings, settings } from "./settings";
import type { CardView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import {
  barPrimary,
  barSecondaries,
  dockDialog,
  dockRegion,
  nameOf,
  pressKey,
  sheetPanel,
} from "./test/dockView";

beforeEach(() => {
  settings.set(defaultSettings());
  resetDock();
  resetModals();
});
afterEach(cleanup);

const card = (id: string, name: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: id, name, owner: "me", controller: "me", ...extra }) as CardView;

interface Calls {
  confirmed: unknown[][];
  cancelled: number;
}

interface Case {
  title: string;
  component: Component<never>;
  props: (c: Calls) => Record<string, unknown>;
  // The dialog's name: the modal's heading without its source tag.
  dialog: string;
  // The primary's name once `pick` has run.
  primary: string;
  cancel?: string;
  // Clicks in the sheet that make the confirm enabled, if it starts off.
  pick?: (sheet: HTMLElement) => void;
  // What onConfirm receives.
  confirms: unknown[];
}

const ox = card("ox", "Ox", { power: 3, toughness: 3 });
const elk = card("elk", "Elk", { power: 2, toughness: 2 });
const firstOption = (sheet: HTMLElement): void =>
  click(sheet.querySelector<HTMLElement>(".prompt-options .prompt-opt")!);

const CASES: Case[] = [
  {
    title: "the Phyrexian mana stepper",
    component: PhyrexianCostModal as never,
    props: (c) => ({
      gameID: "g1",
      card: card("dm", "Dismember", { mana_cost: "{1}{B/P}{B/P}" }),
      symbols: 2,
      life: 20,
      onConfirm: (n: number) => c.confirmed.push([n]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Phyrexian mana for Dismember",
    primary: "Cast",
    confirms: [0],
  },
  {
    title: "the sacrifice picker",
    component: SacrificeCostModal as never,
    props: (c) => ({
      source: card("altar", "Ashnod's Altar"),
      label: "a creature",
      options: [ox, elk],
      onConfirm: (ids: string[]) => c.confirmed.push([ids]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Ashnod's Altar",
    primary: "Sacrifice",
    pick: firstOption,
    confirms: [["ox"]],
  },
  {
    title: "the crew picker",
    component: CrewCostModal as never,
    props: (c) => ({
      card: card("cart", "Smuggler's Copter"),
      ability: { index: 0, crew_cost: 1 },
      options: [ox],
      onConfirm: (ids: string[]) => c.confirmed.push([ids]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Smuggler's Copter",
    primary: "Crew",
    pick: firstOption,
    confirms: [["ox"]],
  },
  {
    title: "the counter-cost picker",
    component: CounterCostModal as never,
    props: (c) => ({
      card: card("hok", "Heart of Kiran"),
      ability: {
        index: 0,
        counter_cost_kind: "loyalty",
        counter_cost_n: 1,
        counter_cost_options: [
          { card_id: "pw1", kinds: [{ kind: "loyalty", count: 4 }] },
          { card_id: "pw2", kinds: [{ kind: "loyalty", count: 3 }] },
        ],
      },
      board: [card("pw1", "Nissa"), card("pw2", "Jace")],
      onConfirm: (choices: unknown[]) => c.confirmed.push([choices]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Heart of Kiran",
    primary: "Remove",
    pick: firstOption,
    confirms: [[{ cardID: "pw1", kind: "loyalty", count: 4, n: 1 }]],
  },
  {
    title: "the face picker",
    component: FacePickerModal as never,
    props: (c) => ({
      card: card("sg", "Sea Gate Restoration", {
        layout: "modal_dfc",
        faces: [
          { name: "Sea Gate Restoration", type_line: "Sorcery", mana_cost: "{4}{U}{U}{U}" },
          { name: "Sea Gate, Reborn", type_line: "Land", mana_cost: "" },
        ],
      }),
      onConfirm: (face: number, fused: boolean) => c.confirmed.push([face, fused]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Sea Gate Restoration",
    primary: "Cast Sea Gate Restoration",
    confirms: [0, false],
  },
  {
    title: "the alternative-cost picker",
    component: AlternativeCostModal as never,
    props: (c) => ({
      card: card("rift", "Cyclonic Rift", {
        mana_cost: "{1}{U}",
        alternative_costs: [{ key: "overload", label: "Overload {6}{U}", mana_cost: "{6}{U}" }],
      }),
      onConfirm: (key: string | undefined, optional: number[]) => c.confirmed.push([key, optional]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Cyclonic Rift",
    primary: "Cast",
    confirms: [undefined, []],
  },
  {
    title: "the alternative-cost payment picker",
    component: AltCostPaymentModal as never,
    props: (c) => ({
      card: card("fow", "Force of Will"),
      offer: { key: "pitch", label: "Exile a blue card", pay_label: "a blue card" },
      options: [card("brainstorm", "Brainstorm")],
      onConfirm: (ids: string[]) => c.confirmed.push([ids]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Force of Will",
    primary: "Pay",
    pick: firstOption,
    confirms: [["brainstorm"]],
  },
  {
    title: "the alternative-cost discard picker",
    component: AltCostPaymentModal as never,
    props: (c) => ({
      card: card("snag", "Snag"),
      offer: {
        key: "discard",
        label: "Discard a Forest card rather than pay this spell's mana cost",
        pay_label: "a Forest card",
        discards: true,
        pay_options: { cards: ["forest"], min: 1, max: 1 },
      },
      options: [card("forest", "Forest")],
      onConfirm: (ids: string[]) => c.confirmed.push([ids]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Snag",
    primary: "Discard",
    pick: firstOption,
    confirms: [["forest"]],
  },
  {
    title: "the discard-cost picker",
    component: DiscardCostModal as never,
    props: (c) => ({
      card: card("fs", "Fauna Shaman"),
      options: [card("bear", "Grizzly Bears")],
      need: 1,
      label: "Discard a creature card",
      onConfirm: (ids: string[]) => c.confirmed.push([ids]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Fauna Shaman",
    primary: "Discard",
    pick: firstOption,
    confirms: [["bear"]],
  },
  {
    title: "the division picker",
    component: DivideDamageModal as never,
    props: (c) => ({
      sourceName: "Fury",
      targets: [
        { id: "a", name: "Ox" },
        { id: "b", name: "Elk" },
      ],
      total: 4,
      onConfirm: (d: Record<string, number>) => c.confirmed.push([d]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Divide 4 — Fury",
    primary: "Confirm",
    cancel: "Back",
    confirms: [{ a: 2, b: 2 }],
  },
  {
    title: "the convoke / waterbend picker",
    component: TapCostModal as never,
    props: (c) => ({
      card: card("cs", "Chord of Calling"),
      cost: { key: "convoke", label: "Convoke" },
      options: [ox],
      limit: 3,
      onConfirm: (ids: string[]) => c.confirmed.push([ids]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Chord of Calling",
    primary: "Tap nothing",
    confirms: [[]],
  },
  {
    title: "the delve picker",
    component: DelveCostModal as never,
    props: (c) => ({
      gameID: "g1",
      card: card("tc", "Treasure Cruise"),
      options: [card("g1c", "Opt")],
      onConfirm: (ids: string[]) => c.confirmed.push([ids]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Treasure Cruise",
    primary: "Exile nothing",
    confirms: [[]],
  },
  {
    title: "the cast-time mode picker",
    component: ModePickerModal as never,
    props: (c) => ({
      card: card("cc", "Cryptic Command", {
        modes: {
          prompt: "Choose two —",
          min: 2,
          max: 2,
          options: [{ label: "Counter target spell." }, { label: "Draw a card." }],
        },
      }),
      onConfirm: (modes: number[]) => c.confirmed.push([modes]),
      onCancel: () => c.cancelled++,
    }),
    dialog: "Cryptic Command",
    primary: "Cast",
    pick: (sheet) => {
      for (const b of sheet.querySelectorAll<HTMLElement>(".prompt-options .prompt-opt")) click(b);
    },
    confirms: [[0, 1]],
  },
];

function mount(k: Case) {
  const calls: Calls = { confirmed: [], cancelled: 0 };
  render(DockHarness as never, { component: k.component, props: k.props(calls) } as never);
  flushSync();
  return calls;
}

describe.each(CASES)("$title is a dock sheet (ADR 0111 PR 6)", (k) => {
  it("opens in region actions, in a non-modal dialog with the modal's name", () => {
    mount(k);
    const dialog = dockDialog();
    expect(dialog).not.toBeNull();
    expect(dockRegion()!.contains(dialog)).toBe(true);
    expect(dialog!.getAttribute("aria-label")).toBe(k.dialog);
    expect(dialog!.hasAttribute("aria-modal")).toBe(false);
    const sheet = sheetPanel();
    expect(sheet).not.toBeNull();
    expect(dialog!.contains(sheet)).toBe(true);
    expect(document.querySelector(".prompt-backdrop")).toBeNull();
    expect(document.querySelector(".prompt-modal")).toBeNull();
    expect(document.querySelector('[aria-modal="true"]')).toBeNull();
  });

  it("puts its confirm and its Cancel in the dock's bar, and they answer", () => {
    const calls = mount(k);
    k.pick?.(sheetPanel()!);
    const primary = barPrimary()!;
    expect(nameOf(primary)).toBe(k.primary);
    expect(primary.disabled).toBe(false);
    expect(primary.getAttribute("aria-keyshortcuts")).toBe("Enter");
    const cancel = barSecondaries().find((b) => nameOf(b) === (k.cancel ?? "Cancel"))!;
    expect(cancel).toBeDefined();
    expect(cancel.getAttribute("aria-keyshortcuts")).toBe("Escape");
    click(primary);
    expect(calls.confirmed).toEqual([k.confirms]);
    click(cancel);
    expect(calls.cancelled).toBe(1);
  });

  it("answers Enter and Escape through the dock's one handler", () => {
    const calls = mount(k);
    k.pick?.(sheetPanel()!);
    (document.activeElement as HTMLElement | null)?.blur();
    pressKey("Enter");
    expect(calls.confirmed).toEqual([k.confirms]);
    pressKey("Escape");
    expect(calls.cancelled).toBe(1);
  });
});

describe("a cost sheet's keys", () => {
  it("leave a disabled confirm alone on Enter", () => {
    const calls = mount(CASES.find((k) => k.title === "the sacrifice picker")!);
    expect(barPrimary()!.disabled).toBe(true);
    pressKey("Enter");
    expect(calls.confirmed).toEqual([]);
  });

  it("the division's Escape goes Back to the picks, and nothing else", () => {
    const calls = mount(CASES.find((k) => k.title === "the division picker")!);
    pressKey("Escape");
    expect(calls.cancelled).toBe(1);
    expect(calls.confirmed).toEqual([]);
  });
});

// ADR 0135 §2: Foil's "an Island card and another card" is a discard
// with a set rule. The confirm stays off until the picks fill both
// parts one-to-one: two non-Islands do not, an Island and another card
// do.
describe("an alternative-cost discard with a set rule", () => {
  const foil = (c: Calls) => ({
    card: card("foil", "Foil"),
    offer: {
      key: "discard",
      label: "Discard an Island card and another card rather than pay this spell's mana cost",
      pay_label: "an Island card and another card",
      discards: true,
      pay_options: {
        cards: ["isl", "mtn", "spell"],
        min: 2,
        max: 2,
        each_of: [
          { label: "an Island card", cards: ["isl"] },
          { label: "another card", cards: ["isl", "mtn", "spell"] },
        ],
      },
    },
    options: [card("isl", "Island"), card("mtn", "Mountain"), card("spell", "Opt")],
    onConfirm: (ids: string[]) => c.confirmed.push([ids]),
    onCancel: () => c.cancelled++,
  });
  const option = (name: string): HTMLElement =>
    [...sheetPanel()!.querySelectorAll<HTMLElement>(".prompt-options .prompt-opt")].find((b) =>
      (b.textContent ?? "").includes(name),
    )!;

  it("holds confirm on two non-Island cards and opens it on an Island and another card", () => {
    const calls: Calls = { confirmed: [], cancelled: 0 };
    render(
      DockHarness as never,
      { component: AltCostPaymentModal as never, props: foil(calls) } as never,
    );
    flushSync();
    expect(sheetPanel()!.textContent).toContain(
      "One card for each part: an Island card, another card.",
    );
    click(option("Mountain"));
    click(option("Opt"));
    expect(barPrimary()!.disabled).toBe(true);
    click(option("Opt"));
    click(option("Island"));
    expect(barPrimary()!.disabled).toBe(false);
    click(barPrimary()!);
    expect(calls.confirmed).toEqual([[["mtn", "isl"]]]);
  });
});
