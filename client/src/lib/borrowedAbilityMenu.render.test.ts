// @vitest-environment jsdom
//
// borrowedAbilityMenu.render.test.ts — ADR 0155 owner decisions 2 and 3 in
// the light ability popover: a lent row sits under "From <lender>", a row
// lent twice is listed once, and an uncatalogued lender is a disabled
// "<lender>: not automated" line. A click on a lent row sends the shown
// row's index.

import { afterEach, describe, expect, it } from "vitest";
import ManaAbilityMenu from "./components/board/ManaAbilityMenu.svelte";
import { collapseBorrowed } from "./borrowedRows";
import type { ActivatedAbilityView, CardView, ManaAbilityView } from "./protocol";
import { cleanup, render } from "./test/render.svelte";

afterEach(() => {
  cleanup();
});

const ooze = { id: "ooze", name: "Necrotic Ooze" };
const elf = { id: "elf-a", name: "Llanowar Elves" };

const own: ManaAbilityView = { index: 0, ref: "own:0", label: "Add {C}", produced: "{C}" };
const lentG = (index: number, n: number, from = elf): ManaAbilityView => ({
  index,
  ref: `borrow:elf:0:${n}`,
  granted_by: ooze,
  borrowed_from: from,
  label: "{T}: Add {G}",
  produced: "{G}",
});
const lentPing: ActivatedAbilityView = {
  index: 0,
  ref: "borrow:elf:1:0",
  granted_by: ooze,
  borrowed_from: elf,
  label: "{1}: Necrotic Ooze pings.",
};

const card = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "ooze",
    name: "Necrotic Ooze",
    owner: "me",
    controller: "me",
    ...extra,
  }) as CardView;

describe("ManaAbilityMenu — borrowed rows (ADR 0155)", () => {
  it("lists a lender's rows under 'From <lender>' and its own rows above", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [own, lentG(1, 0)],
      activated: [lentPing],
      tapped: false,
      onActivate: () => {},
    });
    const heading = container.querySelector(".lender .section-label");
    expect(heading?.textContent).toBe("From Llanowar Elves");
    const lent = [...container.querySelectorAll(".lender button.menu-item .label")].map(
      (e) => e.textContent,
    );
    expect(lent).toEqual(["{T}: Add {G}", "{1}: Necrotic Ooze pings."]);
    const top = container.querySelector(".mana-menu > button.menu-item .label");
    expect(top?.textContent).toBe("Add {C}");
  });

  it("shows one row for an ability lent twice, and a click sends the first row's index", () => {
    const sent: number[] = [];
    const rows = collapseBorrowed([
      lentG(3, 0),
      lentG(4, 1, { id: "elf-b", name: "Llanowar Elves" }),
    ]);
    const { container } = render(ManaAbilityMenu, {
      abilities: rows,
      tapped: false,
      onActivate: (i: number) => sent.push(i),
    });
    const buttons = container.querySelectorAll(".lender button.menu-item");
    expect(buttons).toHaveLength(1);
    (buttons[0] as HTMLButtonElement).click();
    expect(sent).toEqual([3]);
  });

  it("names a lender the catalog does not automate in a disabled line", () => {
    const { container } = render(ManaAbilityMenu, {
      abilities: [own],
      tapped: false,
      onActivate: () => {},
      card: card({ unautomated_lenders: [{ id: "bear", name: "Mystery Bear" }] }),
    });
    const line = container.querySelector("button[data-not-automated]") as HTMLButtonElement;
    expect(line.textContent?.trim()).toBe("Mystery Bear: not automated");
    expect(line.disabled).toBe(true);
  });
});
