// ADR 0155 owner decisions 2 and 3 on the client: borrowed rows are grouped
// under "From <lender>", identical rows are shown once (and the first one's
// ref is what a click sends), a self-borrower does not say "from <self>",
// and a lender the catalog does not automate is a disabled line.

import { describe, it, expect } from "vitest";

import {
  borrowedRowKey,
  grantedFromLabel,
  isBorrowedRef,
  isBorrowedRow,
  activatedAbilityRef,
} from "./abilityRef";
import { collapseBorrowed, splitBorrowed, unautomatedLines } from "./borrowedRows";
import { abilityPopoverModel, buildMenuSections, menuAbilityRows } from "./contextMenu.logic";
import type { ActivatedAbilityView, CardView, GameView, ManaAbilityView } from "./protocol";

const KEY = "68954295-54e3-4303-a6bc-fc4547a4e3a3";
const ooze = { id: "ooze", name: "Necrotic Ooze" };
const elfA = { id: "elf-a", name: "Llanowar Elves" };
const elfB = { id: "elf-b", name: "Llanowar Elves" };
const bear = { id: "bear", name: "Mystery Bear" };

const lentMana = (
  index: number,
  n: number,
  from: { id: string; name: string },
): ManaAbilityView => ({
  index,
  ref: `borrow:${KEY}:0:${n}`,
  granted_by: ooze,
  borrowed_from: from,
  tap_cost: true,
  produced: "{G}",
  label: "{T}: Add {G}",
});

const lentPing = (
  index: number,
  from: { id: string; name: string },
  n = 0,
): ActivatedAbilityView => ({
  index,
  ref: `borrow:${KEY}:1:${n}`,
  granted_by: ooze,
  borrowed_from: from,
  label: "{1}: Necrotic Ooze pings.",
});

const card = (extra: Partial<CardView> = {}): CardView =>
  ({
    instance_id: "ooze",
    name: "Necrotic Ooze",
    owner: "me",
    controller: "me",
    type_line: "Creature — Ooze",
    ...extra,
  }) as CardView;

const view = (c: CardView): GameView =>
  ({
    seats: [{ id: "me", name: "Me", life: 40 }],
    battlefield: { kind: "battlefield", count: 1, cards: [c] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  }) as unknown as GameView;

describe("the borrow: ref", () => {
  it("is recognised, and its instance count is stripped", () => {
    expect(isBorrowedRef(`borrow:${KEY}:0:1`)).toBe(true);
    expect(isBorrowedRef("own:0")).toBe(false);
    expect(isBorrowedRef("grant:x:0:0")).toBe(false);
    expect(isBorrowedRef(undefined)).toBe(false);
    expect(borrowedRowKey(`borrow:${KEY}:0:1`)).toBe(`borrow:${KEY}:0`);
    expect(borrowedRowKey("borrow:land:G:2")).toBe("borrow:land:G");
    expect(borrowedRowKey("own:0")).toBe("");
    expect(isBorrowedRow({ ref: `borrow:${KEY}:0:0` })).toBe(true);
    expect(isBorrowedRow({ borrowed_from: elfA })).toBe(true);
    expect(isBorrowedRow({ ref: "own:0" })).toBe(false);
  });

  it("is sent back as the row's ref", () => {
    const c = card({ activated_abilities: [lentPing(0, elfA)] });
    expect(activatedAbilityRef(c, 0)).toEqual({ ref: `borrow:${KEY}:1:0` });
  });
});

describe("who a row is 'from'", () => {
  it("names the lender of a lent row, not the borrower", () => {
    expect(grantedFromLabel(lentMana(0, 0, elfA), "ooze")).toBe("from Llanowar Elves");
  });
  it("says nothing for a lent row whose lender the server did not name", () => {
    expect(grantedFromLabel({ ref: `borrow:${KEY}:0:0`, granted_by: ooze }, "ooze")).toBe("");
  });
  it("says nothing for a row the permanent gave itself", () => {
    expect(grantedFromLabel({ ref: "own:0", granted_by: ooze }, "ooze")).toBe("");
    expect(grantedFromLabel({ ref: "grant:x:0:0", granted_by: ooze })).toBe("from Necrotic Ooze");
  });
});

describe("identical lent rows are shown once", () => {
  it("keeps the first of two Llanowar Elves in two graveyards", () => {
    const rows = [lentMana(0, 0, elfA), lentMana(1, 1, elfB)];
    const shown = collapseBorrowed(rows);
    expect(shown).toHaveLength(1);
    expect(shown[0].index).toBe(0);
    expect(shown[0].ref).toBe(`borrow:${KEY}:0:0`);
  });

  it("collapses the same ability lent by two different cards", () => {
    const other: ManaAbilityView = { ...lentMana(1, 0, bear), ref: "borrow:other-card:0:0" };
    expect(collapseBorrowed([lentMana(0, 0, elfA), other])).toHaveLength(1);
  });

  it("drops a lent row the permanent already has as its own (Experiment Kraj)", () => {
    const own: ActivatedAbilityView = {
      index: 0,
      ref: "own:0",
      label: "{1}: Necrotic Ooze pings.",
    };
    const rows = collapseBorrowed([own, lentPing(1, elfA)]);
    expect(rows).toEqual([own]);
  });

  it("never drops an own or granted row, and returns the list itself when nothing repeats", () => {
    const twin = { index: 1, ref: "own:1", label: "{T}: Add {G}" };
    const own = { index: 0, ref: "own:0", label: "{T}: Add {G}" };
    const rows = [own, twin];
    expect(collapseBorrowed(rows)).toBe(rows);
    const distinct = [lentMana(0, 0, elfA), lentPing(1, elfA)];
    expect(collapseBorrowed(distinct)).toBe(distinct);
  });

  it("is what the popover lists, so the click rule and the menu agree", () => {
    const c = card({
      mana_abilities: [lentMana(0, 0, elfA), lentMana(1, 1, elfB)],
      activated_abilities: [lentPing(0, elfA), lentPing(1, elfB, 1)],
    });
    expect(menuAbilityRows(c, "me")).toHaveLength(1);
    const model = abilityPopoverModel({
      card: c,
      viewerID: "me",
      mana: true,
      activated: true,
      special: false,
    });
    expect(model.mana).toHaveLength(1);
    expect(model.activated).toHaveLength(1);
    // The row shown is the first one, with its own index and ref.
    expect(model.mana[0].a.index).toBe(0);
    expect(model.activated[0].a.ref).toBe(`borrow:${KEY}:1:0`);
  });
});

describe("grouping under 'From <lender>'", () => {
  it("splits own rows from one group per lender, in first-seen order", () => {
    const own: ManaAbilityView = { index: 0, ref: "own:0", label: "Add {C}", produced: "{C}" };
    const treetop: ManaAbilityView = {
      ...lentMana(2, 0, { id: "tree", name: "Treetop Village" }),
      ref: "borrow:tree:0:0",
      label: "{T}: Add {G}{G}",
    };
    const { own: mine, groups } = splitBorrowed([own, lentMana(1, 0, elfA), treetop]);
    expect(mine).toEqual([own]);
    expect(groups.map((g) => g.heading)).toEqual(["From Llanowar Elves", "From Treetop Village"]);
    expect(groups[0].rows).toHaveLength(1);
  });

  it("heads a lender the server did not name as 'another object'", () => {
    const row: ActivatedAbilityView = { index: 0, ref: `borrow:${KEY}:1:0`, label: "x" };
    expect(splitBorrowed([row]).groups[0].heading).toBe("From another object");
  });

  it("puts a lender's rows under a section of their own in the card menu", () => {
    const c = card({
      mana_abilities: [lentMana(0, 0, elfA), lentMana(1, 1, elfB)],
      activated_abilities: [lentPing(0, elfA), lentPing(1, bear, 0)],
    });
    const sections = buildMenuSections(view(c), c, "me", false);
    const from = sections.filter((s) => s.id.startsWith("abilities-from-"));
    expect(from.map((s) => s.label)).toEqual(["From Llanowar Elves"]);
    expect(from[0].items.map((i) => i.label)).toEqual([
      "{T}: Add {G}",
      "{1}: Necrotic Ooze pings.",
    ]);
    // The permanent's own section is not made up of lent rows, and a lent
    // row carries no '(from …)' suffix beside its heading.
    expect(sections.find((s) => s.id === "abilities")).toBeUndefined();
  });
});

describe("a lender the catalog does not automate", () => {
  it("is a disabled '<lender>: not automated' line, once per name", () => {
    const c = card({
      unautomated_lenders: [bear, { id: "bear-2", name: "Mystery Bear" }, { id: "x", name: "Dud" }],
    });
    expect(unautomatedLines(c).map((l) => l.label)).toEqual([
      "Mystery Bear: not automated",
      "Dud: not automated",
    ]);
    const items = buildMenuSections(view(c), c, "me", false).find(
      (s) => s.id === "abilities",
    )?.items;
    expect(items?.map((i) => [i.label, i.disabled])).toEqual([
      ["Mystery Bear: not automated", true],
      ["Dud: not automated", true],
    ]);
    expect(items?.every((i) => !i.activate)).toBe(true);
  });

  it("is nothing for a permanent without one", () => {
    expect(unautomatedLines(card())).toEqual([]);
  });
});
