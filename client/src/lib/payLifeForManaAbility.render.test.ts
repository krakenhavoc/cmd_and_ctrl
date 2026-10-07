// @vitest-environment jsdom
//
// payLifeForManaAbility.render.test.ts — ADR 0131 §2 (#2531), PR 2: a
// mana ability whose own mana component has symbols the viewer could pay
// 2 life each for (a filter land's {B} under K'rrik) gets a "Pay life for
// {B}…" row per count in its popover, and choosing it sends the count as
// `phyrexian_life`. Auto-tap never pays life, so the row is how a player
// chooses it.

import { afterEach, describe, expect, it, vi } from "vitest";
import ManaAbilityMenu from "./components/board/ManaAbilityMenu.svelte";
import { L } from "./labels";
import type { ManaAbilityView } from "./protocol";
import { cleanup, render } from "./test/render.svelte";

afterEach(() => {
  cleanup();
});

function filter(extra: Partial<ManaAbilityView> = {}): ManaAbilityView {
  return {
    index: 1,
    label: "{B}, {T}: Add {B}{B}",
    produced: "{B}{B}",
    tap_cost: true,
    mana_cost: "{B}",
    phyrexian_symbols: 1,
    phyrexian_granted: 1,
    ...extra,
  };
}

function lifeRows(container: HTMLElement): HTMLButtonElement[] {
  return [...container.querySelectorAll<HTMLButtonElement>('button[data-kind="mana-pay-life"]')];
}

describe("ManaAbilityMenu — Pay life for {B}… on a mana ability", () => {
  it("lists the row, named by the label contract, and sends the claim", () => {
    const onActivate = vi.fn();
    const { container } = render(ManaAbilityMenu, {
      abilities: [filter()],
      tapped: false,
      onActivate,
      payerLife: 20,
    });
    const rows = lifeRows(container);
    expect(rows).toHaveLength(1);
    expect(rows[0].textContent).toContain(L.payLifeForMana);
    expect(rows[0].textContent).toContain("2");
    rows[0].click();
    expect(onActivate).toHaveBeenCalledWith(1, 1);
  });

  it("the ordinary row still activates with no claim", () => {
    const onActivate = vi.fn();
    const { container } = render(ManaAbilityMenu, {
      abilities: [filter()],
      tapped: false,
      onActivate,
      payerLife: 20,
    });
    const first = container.querySelector<HTMLButtonElement>('button[data-kind="mana"]')!;
    first.click();
    expect(onActivate).toHaveBeenCalledWith(1, undefined);
  });

  it("offers a row per count, bounded by the life total (CR 119.4)", () => {
    const two = filter({ mana_cost: "{B}{B}", phyrexian_symbols: 2, phyrexian_granted: 2 });
    const rich = render(ManaAbilityMenu, {
      abilities: [two],
      tapped: false,
      onActivate: () => {},
      payerLife: 20,
    });
    expect(lifeRows(rich.container)).toHaveLength(2);
    cleanup();
    const poor = render(ManaAbilityMenu, {
      abilities: [two],
      tapped: false,
      onActivate: () => {},
      payerLife: 3,
    });
    expect(lifeRows(poor.container)).toHaveLength(1);
  });

  it("offers none without a symbol to pay, or without life to pay it", () => {
    for (const [a, life] of [
      [filter({ phyrexian_symbols: 0, phyrexian_granted: 0 }), 20],
      [filter({ phyrexian_symbols: undefined, phyrexian_granted: undefined }), 20],
      [filter(), 1],
      [filter(), undefined],
    ] as const) {
      const { container } = render(ManaAbilityMenu, {
        abilities: [a],
        tapped: false,
        onActivate: () => {},
        payerLife: life,
      });
      expect(lifeRows(container)).toHaveLength(0);
      cleanup();
    }
  });

  it("greys with the ability when its source is tapped", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [filter()],
      tapped: true,
      onActivate: () => {},
      payerLife: 20,
    });
    const rows = lifeRows(container);
    expect(rows).toHaveLength(1);
    expect(rows[0].disabled).toBe(true);
  });
});
