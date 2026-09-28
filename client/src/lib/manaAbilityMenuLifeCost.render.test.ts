// @vitest-environment jsdom
//
// manaAbilityMenuLifeCost.render.test.ts — #1695. The left-click mana
// popover (ManaAbilityMenu.svelte) kept its own copy of `abilityBlocked`
// and never checked `life_cost` against the paying player's life, so a
// "Pay N life" ability (Mana Confluence, Greed) stayed clickable there
// even after #1690 fixed the same bug in the right-click context menu.
// The popover now delegates to the shared `abilityBlocked` /
// `notEnoughLife` in contextMenu.logic.ts via a new `payerLife` prop,
// threaded down from PlayerPanel through BattlefieldRow and Card. CR
// 119.4: paying life equal to your total is legal, so the boundary
// case (cost === life) must stay enabled.

import { afterEach, describe, expect, it } from "vitest";
import ManaAbilityMenu from "./components/board/ManaAbilityMenu.svelte";
import { NOT_ENOUGH_LIFE } from "./contextMenu.logic";
import type { ActivatedAbilityView, ManaAbilityView } from "./protocol";
import { cleanup, render } from "./test/render.svelte";

afterEach(() => {
  cleanup();
});

function manaAbility(extra: Partial<ManaAbilityView> = {}): ManaAbilityView {
  return { index: 0, label: "Add {C}", produced: "{C}", ...extra };
}

function activatedAbility(extra: Partial<ActivatedAbilityView> = {}): ActivatedAbilityView {
  return { index: 0, label: "{T}: draw a card", ...extra };
}

function firstButton(container: HTMLElement): HTMLButtonElement {
  const btn = container.querySelector("button.menu-item");
  if (!btn) throw new Error("expected a menu-item button");
  return btn as HTMLButtonElement;
}

describe("ManaAbilityMenu — life-cost row (#1695)", () => {
  it("greys a mana ability whose life cost is more than the payer's life", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [manaAbility({ life_cost: 2 })],
      tapped: false,
      onActivate: () => {},
      payerLife: 1,
    });
    const btn = firstButton(container);
    expect(btn.disabled).toBe(true);
    expect(btn.title).toBe(NOT_ENOUGH_LIFE);
  });

  it("offers the same mana ability when the payer has exactly enough life (CR 119.4)", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [manaAbility({ life_cost: 2 })],
      tapped: false,
      onActivate: () => {},
      payerLife: 2,
    });
    const btn = firstButton(container);
    expect(btn.disabled).toBe(false);
  });

  it("leaves a mana ability with no life cost unaffected at 1 life", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [manaAbility()],
      tapped: false,
      onActivate: () => {},
      payerLife: 1,
    });
    const btn = firstButton(container);
    expect(btn.disabled).toBe(false);
  });

  it("leaves every row enabled when payerLife is absent — the server's own check is the gate", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [manaAbility({ life_cost: 99 })],
      tapped: false,
      onActivate: () => {},
    });
    const btn = firstButton(container);
    expect(btn.disabled).toBe(false);
  });

  it("greys an activated ability with an unaffordable life cost the same way", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [],
      tapped: false,
      onActivate: () => {},
      activated: [activatedAbility({ life_cost: 3 })],
      onActivateAbility: () => {},
      payerLife: 2,
    });
    const btn = firstButton(container);
    expect(btn.disabled).toBe(true);
    expect(btn.title).toBe(NOT_ENOUGH_LIFE);
  });

  it("keeps the tap reason for a tapped source over an unpayable life cost", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [manaAbility({ tap_cost: true, life_cost: 5 })],
      tapped: true,
      onActivate: () => {},
      payerLife: 1,
    });
    const btn = firstButton(container);
    expect(btn.disabled).toBe(true);
    expect(btn.title).toBe("already tapped");
  });
});
