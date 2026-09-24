// @vitest-environment jsdom
//
// #1533 — AttackDeclarationModal (the #1162 attack-tax picker) opened
// for a CR 508.1c count limit. What only the markup can say: the
// reason is on screen, the selection opens at the cap, rows past the
// cap cannot be checked, and confirming sends a declaration the server
// accepts. The cap is the server's `attack_targets[].attack_limit`; the
// pure rules behind it are in attackLimit.test.ts.

import { afterEach, describe, expect, it } from "vitest";

import AttackDeclarationModal from "./components/board/AttackDeclarationModal.svelte";
import type { AttackTargetView, CardView, GameView, PlayerView, ZoneView } from "./protocol";
import { cleanup, click, render } from "./test/render.svelte";

afterEach(cleanup);

function zone(kind: string, owner: string | undefined, cards: CardView[]): ZoneView {
  return { kind, owner, count: cards.length, cards };
}

function seat(id: string, name: string, n: number): PlayerView {
  return {
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id, []),
    hand: zone("hand", id, []),
    graveyard: zone("graveyard", id, []),
    command: zone("command", id, []),
    commander_damage: {},
    life_history: [],
  };
}

const bears = ["Bear A", "Bear B", "Bear C"];

function snap(rows: AttackTargetView[], extra: CardView[] = []): GameView {
  const bf: CardView[] = bears.map((name, i) => ({
    instance_id: `bear${i + 1}`,
    name,
    owner: "me",
    controller: "me",
    type_line: "Creature — Bear",
  }));
  return {
    id: "g1",
    state: "active",
    seats: [seat("me", "Me", 0), seat("bob", "Bob", 1)],
    battlefield: zone("battlefield", undefined, [...bf, ...extra]),
    stack: zone("stack", undefined, []),
    exile: zone("exile", undefined, []),
    turn: {
      seq: 1,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_attackers",
      attack_targets: rows,
    },
    mulligans_open: false,
  };
}

const ARBITER = "No more than one creature can attack each combat (Silent Arbiter).";

function mount(view: GameView, limitReason: string | null = null) {
  const confirmed: Array<{ attackers: string[]; locked: string[] }> = [];
  const r = render(
    AttackDeclarationModal as never,
    {
      view,
      viewerID: "me",
      defenderSeatID: "bob",
      limitReason,
      onConfirm: (attackers: string[], locked: string[]) => confirmed.push({ attackers, locked }),
      onCancel: () => {},
    } as never,
  );
  return { container: r.container, confirmed };
}

const text = (el: Element | null | undefined): string =>
  el?.textContent?.replace(/\s+/g, " ").trim() ?? "";

const rows = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>('[role="checkbox"]'),
];
const checked = (c: HTMLElement): string[] =>
  rows(c)
    .filter((b) => b.getAttribute("aria-checked") === "true")
    .map((b) => text(b));
const attackButton = (c: HTMLElement): HTMLButtonElement =>
  [...c.querySelectorAll<HTMLButtonElement>("button.primary")].find((b) =>
    text(b).startsWith("Attack with"),
  )!;

describe("AttackDeclarationModal under an attack limit (#1533)", () => {
  it("shows the server's reason and the cap", () => {
    const { container } = mount(snap([{ kind: "player", id: "bob", attack_limit: 1 }]), ARBITER);
    const note = container.querySelector('[role="note"]');
    expect(text(note)).toContain(ARBITER);
    expect(text(note)).toContain("Choose up to 1.");
    expect(text(container.querySelector(".prompt-count"))).toBe("1 / 1 allowed");
  });

  it("opens with the first N checked and the rest disabled", () => {
    const { container } = mount(snap([{ kind: "player", id: "bob", attack_limit: 1 }]), ARBITER);
    expect(checked(container)).toEqual(["Bear A"]);
    const [a, b, c] = rows(container);
    expect(a.disabled).toBe(false);
    expect(b.disabled).toBe(true);
    expect(c.disabled).toBe(true);
    expect(text(attackButton(container))).toBe("Attack with 1");
    expect(attackButton(container).disabled).toBe(false);
  });

  it("lets the player swap which creature attacks, and sends that one", () => {
    const { container, confirmed } = mount(
      snap([{ kind: "player", id: "bob", attack_limit: 1 }]),
      ARBITER,
    );
    const [a, , c] = rows(container);
    click(c); // disabled: nothing changes
    expect(checked(container)).toEqual(["Bear A"]);
    click(a); // uncheck frees the slot
    expect(checked(container)).toEqual([]);
    expect(rows(container).every((r) => !r.disabled)).toBe(true);
    expect(attackButton(container).disabled).toBe(true);
    click(rows(container)[2]);
    expect(checked(container)).toEqual(["Bear C"]);
    click(attackButton(container));
    expect(confirmed).toEqual([{ attackers: ["bear3"], locked: [] }]);
  });

  it("'First N' refills to the cap, not to every creature", () => {
    const { container } = mount(snap([{ kind: "player", id: "bob", attack_limit: 2 }]));
    const first = [...container.querySelectorAll("button.ghost")].find(
      (b) => text(b) === "First 2",
    );
    expect(first).toBeTruthy();
    click(rows(container)[0]);
    click(rows(container)[1]);
    expect(checked(container)).toEqual([]);
    click(first!);
    expect(checked(container)).toEqual(["Bear A", "Bear B"]);
  });

  it("explains the cap from the published number when opened without a refusal", () => {
    const { container } = mount(snap([{ kind: "player", id: "bob", attack_limit: 2 }]));
    expect(text(container.querySelector('[role="note"]'))).toContain(
      "Only 2 more creatures can attack Bob this combat.",
    );
  });

  it("offers no lock-a-land row for a limit alone — there is nothing to pay", () => {
    const land: CardView = {
      instance_id: "forest",
      name: "Forest",
      owner: "me",
      controller: "me",
      type_line: "Basic Land — Forest",
      mana_abilities: [{ index: 0, label: "Add {G}", produced: "{G}", cost: { tap: true } }],
    } as unknown as CardView;
    const { container } = mount(snap([{ kind: "player", id: "bob", attack_limit: 1 }], [land]));
    expect(container.querySelector(".lock-btn")).toBeNull();
    // The same land IS offered once there is a tax to pay, so the
    // absence above is the gate and not a fixture the modal ignores.
    cleanup();
    const taxed = mount(snap([{ kind: "player", id: "bob", attack_limit: 1, tax: "{1}" }], [land]));
    expect(taxed.container.querySelector(".lock-btn")).not.toBeNull();
  });

  it("shows tax and limit together when both apply", () => {
    const { container } = mount(
      snap([{ kind: "player", id: "bob", attack_limit: 2, tax: "{2}" }]),
      null,
    );
    expect(text(container.querySelector('[role="note"]'))).toContain("Only 2 more creatures");
    expect(text(container.querySelector(".prompt-hint"))).toContain("costs {2} per creature");
    expect(checked(container)).toEqual(["Bear A", "Bear B"]);
  });
});

describe("AttackDeclarationModal with no limit (#1162, unchanged)", () => {
  it("opens with every creature checked, no note and nothing disabled", () => {
    const { container } = mount(snap([{ kind: "player", id: "bob", tax: "{2}" }]));
    expect(checked(container)).toEqual(bears);
    expect(container.querySelector('[role="note"]')).toBeNull();
    expect(rows(container).some((r) => r.disabled)).toBe(false);
    expect(text(container.querySelector(".prompt-count"))).toBe("3 / 3 attacking");
  });
});
