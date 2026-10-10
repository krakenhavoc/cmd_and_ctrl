// @vitest-environment jsdom
//
// damageAssignment.render.test.ts — #2956, ADR 0147, the rendered half:
//   - the damage sheet opens on the server's suggested split, so a
//     trampler's leftover is on the player and not 0;
//   - the blockers carry − / + steppers on the board (DamageStepper),
//     and a tick there moves the sheet's numbers and its count;
//   - with "Auto-assign combat damage" on, a split that covers lethal
//     is sent without a sheet, and the dock says what went where.

import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { get } from "svelte/store";

import ChoiceDockHarness from "./test/ChoiceDockHarness.svelte";
import DamageStepper from "./components/board/DamageStepper.svelte";
import DamageAutoAssign from "./components/board/DamageAutoAssign.svelte";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { autoAssignRecords, boardDamageAssign } from "./damageAssignment";
import { L } from "./labels";
import { defaultSettings, settings } from "./settings";
import type { ActionType, CardView, DamageAssignmentView, GameView } from "./protocol";
import { click, cleanup, flushSync, render } from "./test/render.svelte";
import { barPrimary, dockDialog, dockTestView, nameOf, sheetPanel } from "./test/dockView";

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
  autoAssignRecords.set({});
  boardDamageAssign.set(null);
  resetDock();
  resetModals();
});
afterEach(cleanup);

const frame: DamageAssignmentView = {
  attacker_card_id: "vivi",
  blocker_card_ids: ["solphim", "breaker", "captain"],
  attacker_power: 44,
  allow_trample: true,
  lethal: [4, 3, 2],
  covers_lethal: true,
  suggested: {
    assignments: [
      { blocker_id: "solphim", amount: 4 },
      { blocker_id: "breaker", amount: 3 },
      { blocker_id: "captain", amount: 2 },
    ],
    trample_to_player: 35,
  },
};

const card = (id: string, name: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: id, name, owner: "me", controller: "me", ...extra }) as CardView;

function view(f: DamageAssignmentView = frame): GameView {
  return dockTestView({
    battlefield: {
      kind: "battlefield",
      count: 4,
      cards: [
        card("vivi", "Vivi Ornitier", { attacking_target: "opp" }),
        card("solphim", "Solphim", { owner: "opp", controller: "opp" }),
        card("breaker", "Professional Face-Breaker", { owner: "opp", controller: "opp" }),
        card("captain", "Corsair Captain", { owner: "opp", controller: "opp" }),
      ],
    },
    pending_choices: [
      {
        id: "c1",
        kind: "damage_assignment",
        chooser: "me",
        from_player: "me",
        count: 3,
        damage_assignment: f,
      },
    ],
  } as unknown as Partial<GameView>);
}

interface Sent {
  type: ActionType;
  params?: Record<string, unknown>;
}

function mountSheet(v: GameView) {
  const sent: Sent[] = [];
  const sendAction = (type: ActionType, params?: unknown) =>
    sent.push({ type, params: params as Record<string, unknown> });
  render(ChoiceDockHarness as never, { snap: v, viewerID: "me", sendAction, lastError: null });
  return sent;
}

function numbers(sheet: HTMLElement): number[] {
  return [...sheet.querySelectorAll<HTMLInputElement>('input[type="number"]')].map((el) =>
    Number(el.value),
  );
}

describe("the damage sheet, with auto-assign off", () => {
  beforeEach(() => {
    settings.set({
      ...defaultSettings(),
      gameplay: { ...defaultSettings().gameplay, autoAssignCombatDamage: false },
    });
  });

  it("opens on the suggested split, 35 over to the player", () => {
    const sent = mountSheet(view());
    const sheet = sheetPanel()!;
    expect(sheet).not.toBeNull();
    expect(numbers(sheet)).toEqual([4, 3, 2, 35]);
    expect(sheet.querySelector(".prompt-count")?.textContent).toContain("44 / 44 assigned");
    expect(barPrimary()!.disabled).toBe(false);
    click(barPrimary()!);
    expect(sent[0].params).toEqual({
      choice_id: "c1",
      assignments: [
        { blocker_id: "solphim", amount: 4 },
        { blocker_id: "breaker", amount: 3 },
        { blocker_id: "captain", amount: 2 },
      ],
      trample_to_player: 35,
    });
  });

  it("ticks damage on the blockers themselves, and the sheet follows", () => {
    const sent = mountSheet(view());
    const sheet = sheetPanel()!;
    // The board half: one stepper per blocker, the total on the attacker.
    const captain = render(DamageStepper as never, { cardID: "captain", name: "Corsair Captain" });
    const attacker = render(DamageStepper as never, { cardID: "vivi", name: "Vivi Ornitier" });
    const add = captain.container.querySelector<HTMLButtonElement>(
      `button[aria-label="${L.addDamage("Corsair Captain")}"]`,
    )!;
    const remove = captain.container.querySelector<HTMLButtonElement>(
      `button[aria-label="${L.removeDamage("Corsair Captain")}"]`,
    )!;
    // Lethal already: a tick, and the lethal figure in the title.
    expect(captain.container.textContent).toContain("✓");
    expect(captain.container.querySelector(".amount")?.getAttribute("title")).toBe(
      "lethal damage: 2",
    );
    expect(attacker.container.textContent).toContain("all assigned");
    expect(attacker.container.textContent).toContain("35 over");

    // + on a full split takes the point from the trample.
    click(add);
    flushSync();
    expect(numbers(sheet)).toEqual([4, 3, 3, 34]);
    expect(get(boardDamageAssign)?.shares.amounts.captain).toBe(3);

    // − frees damage, which the sheet and the attacker count as left.
    click(remove);
    click(remove);
    flushSync();
    expect(numbers(sheet)).toEqual([4, 3, 1, 34]);
    expect(sheet.querySelector(".prompt-count")?.textContent).toContain("42 / 44 assigned");
    expect(attacker.container.textContent).toContain("2 of 44 left");
    // Short of lethal with damage going over: no confirm (CR 702.19b).
    expect(barPrimary()!.disabled).toBe(true);
    expect(captain.container.textContent).toContain("/2");

    click(add);
    const over = attacker.container.querySelector<HTMLButtonElement>(
      `button[aria-label="${L.addDamage("the defending player")}"]`,
    )!;
    click(over);
    flushSync();
    expect(numbers(sheet)).toEqual([4, 3, 2, 35]);
    click(barPrimary()!);
    expect(sent[0].params?.trample_to_player).toBe(35);
  });

  it("draws nothing on a card that is not in the assignment", () => {
    mountSheet(view());
    const other = render(DamageStepper as never, { cardID: "elsewhere", name: "Llanowar Elves" });
    expect(other.container.querySelector("button")).toBeNull();
  });
});

describe("auto-assign combat damage (on by default)", () => {
  it("sends the split without a sheet, and says what went where", () => {
    const v = view();
    const sent = mountSheet(v);
    const auto: Sent[] = [];
    render(DamageAutoAssign as never, {
      view: v,
      viewerID: "me",
      live: true,
      sendAction: (type: ActionType, params?: unknown) =>
        auto.push({ type, params: params as Record<string, unknown> }),
      canUndo: true,
      onUndo: () => {},
    });
    flushSync();
    expect(sent).toEqual([]);
    expect(auto).toHaveLength(1);
    expect(auto[0]).toEqual({
      type: "resolve_choice",
      params: {
        choice_id: "c1",
        assignments: [
          { blocker_id: "solphim", amount: 4 },
          { blocker_id: "breaker", amount: 3 },
          { blocker_id: "captain", amount: 2 },
        ],
        trample_to_player: 35,
      },
    });
    expect(sheetPanel()).toBeNull();
    const notice = dockDialog()!;
    expect(notice.getAttribute("aria-label")).toBe(L.damageAutoAssigned);
    expect(notice.textContent).toContain(
      "Vivi Ornitier: 4 to Solphim, 3 to Professional Face-Breaker, 2 to Corsair Captain, 35 to Opp (trample)",
    );
    const alwaysAsk = [...notice.querySelectorAll("button")].find(
      (b) => nameOf(b) === L.alwaysAskCombatDamage,
    )!;
    click(alwaysAsk);
    expect(get(settings).gameplay.autoAssignCombatDamage).toBe(false);
  });

  it("asks, pre-filled, when the damage does not cover every blocker", () => {
    const short: DamageAssignmentView = {
      ...frame,
      attacker_power: 5,
      covers_lethal: false,
      suggested: {
        assignments: [
          { blocker_id: "breaker", amount: 3 },
          { blocker_id: "captain", amount: 2 },
          { blocker_id: "solphim", amount: 0 },
        ],
      },
    };
    const v = view(short);
    mountSheet(v);
    const auto: Sent[] = [];
    render(DamageAutoAssign as never, {
      view: v,
      viewerID: "me",
      live: true,
      sendAction: (type: ActionType, params?: unknown) =>
        auto.push({ type, params: params as Record<string, unknown> }),
      canUndo: true,
      onUndo: () => {},
    });
    flushSync();
    expect(auto).toEqual([]);
    expect(numbers(sheetPanel()!)).toEqual([0, 3, 2, 0]);
  });
});
