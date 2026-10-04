// targetingDock.test.ts — ADR 0111 Delivery PR 4. What the dock's
// targeting and insufficient-mana requests say and offer. The sentences
// were TargetingBanner's (#1211, #1659), which is gone: the dock's
// question line carries them now.
//
// #1211: the sentence is the only thing that tells a player WHICH half
// of the stack they may click: an ability row and a spell row are drawn
// identically in the stack overlay, and the legal set (which decides
// legality) is invisible until a click is refused. So the three stack
// modes each get their own line, and this is the test that they reach
// it rather than falling through to the "a target" default, which is
// what a mode the switch does not know produces, silently.

import { describe, it, expect, afterEach, vi } from "vitest";
import { get } from "svelte/store";

import { begin, beginChoice, cancel, targeting, togglePick } from "./targeting";
import { insufficientManaRequest, targetingRequest } from "./targetingDock";
import { dockKeyAction, takesBar } from "./dock";
import type { CardView, PendingChoiceView } from "./protocol";

afterEach(() => {
  targeting.set(null);
});

const card = (name: string, over: Partial<CardView> = {}): CardView => ({
  instance_id: name.toLowerCase(),
  name,
  owner: "me",
  controller: "me",
  ...over,
});

const handlers = () => ({ onDone: vi.fn(), onCancel: vi.fn() });

function current() {
  return targetingRequest(get(targeting)!, null, handlers());
}
const lineOf = (r: ReturnType<typeof current>) =>
  [r.question, r.detail].filter(Boolean).join(" · ");

describe("the dock's targeting prompt names the stack half it is asking about (#1211)", () => {
  for (const [mode, sentence] of [
    ["stack_spell", "a spell on the stack"],
    ["stack_ability", "an ability on the stack"],
    ["stack_item", "a spell or ability on the stack"],
  ] as const) {
    it(mode, () => {
      begin(card("Stifle"), mode);
      const line = lineOf(current());
      expect(line).toContain(sentence);
      if (mode !== "stack_spell") expect(line).not.toContain("Click a target to target");
      cancel();
    });
  }
});

// #1659: the amount being divided is named while the player is still
// picking, not only once DivideDamageModal opens.
describe("#1659 — the dock's targeting prompt shows the divided amount", () => {
  it("a fixed divide names the amount and the target ceiling", () => {
    begin(
      card("Arc Lightning", { legal_targets: { min: 1, max: 3, divide: { total: 3 } } }),
      "any",
    );
    expect(lineOf(current())).toContain("divide 3 damage among up to 3 targets");
  });

  it("an X-based divide shows the announced X once it is known", () => {
    begin(
      card("Rolling Thunder", { legal_targets: { min: 0, max: 0, divide: { from_x: true } } }),
      "any",
      { xValue: 5 },
    );
    expect(lineOf(current())).toContain("divide 5 damage among any number of targets");
  });

  it("an X-based divide falls back to 'X' rather than inventing 0", () => {
    begin(
      card("Rolling Thunder", { legal_targets: { min: 0, max: 0, divide: { from_x: true } } }),
      "any",
    );
    const line = lineOf(current());
    expect(line).toContain("divide X damage among any number of targets");
    expect(line).not.toContain("divide 0 damage");
  });

  it("a clause with no divide shows nothing about dividing", () => {
    begin(card("Lightning Bolt"), "any");
    expect(lineOf(current())).not.toContain("divide");
  });
});

describe("the targeting request (ADR 0111 §6)", () => {
  it("is the dialog the e2e suite waits on, and a flow that takes the bar", () => {
    begin(card("Lightning Bolt"), "any");
    const r = current();
    expect(r.label).toBe("Select target for Lightning Bolt");
    expect(r.rank).toBe("flow");
    expect(takesBar(r)).toBe(true);
    expect(r.question).toBe("Click a player or creature to target Lightning Bolt");
  });

  it("a single pick has Cancel and no Done: the click on the board commits", () => {
    begin(card("Lightning Bolt"), "any");
    const r = current();
    expect(r.primary).toBeNull();
    expect(r.secondary?.map((a) => a.label)).toEqual(["Cancel"]);
    expect(dockKeyAction(r, "Escape")?.label).toBe("Cancel");
    expect(dockKeyAction(r, "Enter")).toBeNull();
  });

  it("a pick list has Done, disabled until enough are picked, and Enter presses it then", () => {
    begin(card("Arc Lightning", { legal_targets: { min: 1, max: 3 } }), "any");
    let r = current();
    expect(r.primary?.label).toBe("Done");
    expect(r.primary?.disabled).toBe(true);
    expect(r.detail).toContain("0/3 picked");
    expect(dockKeyAction(r, "Enter")).toBeNull();

    targeting.set(togglePick(get(targeting)!, { kind: "player", id: "opp" }));
    const h = handlers();
    r = targetingRequest(get(targeting)!, null, h);
    expect(r.primary?.disabled).toBe(false);
    expect(r.detail).toContain("1/3 picked");
    dockKeyAction(r, "Enter")!.onPress();
    expect(h.onDone).toHaveBeenCalledTimes(1);
    dockKeyAction(r, "Escape")!.onPress();
    expect(h.onCancel).toHaveBeenCalledTimes(1);
  });

  it("a trigger's target is a choice with no Cancel: the trigger needs one", () => {
    const choice = {
      id: "ch",
      kind: "pick_target",
      chooser: "me",
      from_player: "me",
      count: 0,
      reason: "destroy target artifact",
      pick_target: { cards: ["x"], min: 1, max: 1 },
    } as PendingChoiceView;
    beginChoice(choice, card("Reclamation Sage"));
    const r = current();
    expect(r.rank).toBe("choice");
    expect(r.label).toBe("Select target for Reclamation Sage");
    expect(r.question).toContain("Reclamation Sage triggered — click");
    expect(r.secondary).toEqual([]);
    expect(dockKeyAction(r, "Escape")).toBeNull();
  });
});

describe("the insufficient-mana request (S15, ADR 0111 PR 4)", () => {
  it("offers Auto-tap & cast as the primary, Cast anyway and Cancel beside it", () => {
    const h = { onAutoTap: vi.fn(), onCastAnyway: vi.fn(), onCancel: vi.fn() };
    const r = insufficientManaRequest(["{G}", "{1}"], "Llanowar Elves", h);
    expect(r.rank).toBe("flow");
    expect(r.label).toBe("insufficient mana");
    expect(r.question).toBe("Insufficient mana for Llanowar Elves");
    expect(r.detail).toBe("missing {G} {1}");
    expect(r.primary?.label).toBe("Auto-tap & cast");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Cancel", "Cast anyway"]);

    dockKeyAction(r, "Enter")!.onPress();
    expect(h.onAutoTap).toHaveBeenCalledTimes(1);
    dockKeyAction(r, "Escape")!.onPress();
    expect(h.onCancel).toHaveBeenCalledTimes(1);
    // Cast anyway is a click only: no key reaches the override.
    expect(h.onCastAnyway).not.toHaveBeenCalled();
    expect(r.secondary?.find((a) => a.label === "Cast anyway")?.title).toBe(
      "cast it without paying its mana cost; the game log says so",
    );
  });

  // ADR 0118 §1 (amends ADR 0111 PR 4): a clicked cast is auto-tapped
  // under strict, so its refusal means the planner found no plan, and
  // the preview would find none either.
  it("drops Auto-tap & cast after a refused cast that was already auto-tapped", () => {
    const h = { onAutoTap: vi.fn(), onCastAnyway: vi.fn(), onCancel: vi.fn() };
    const r = insufficientManaRequest(["{4}"], "Craw Wurm", h, { autoTapped: true });
    expect(r.label).toBe("insufficient mana");
    expect(r.question).toBe("Insufficient mana for Craw Wurm");
    expect(r.detail).toBe("missing {4}");
    expect(r.primary?.label).toBe("Cancel");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Cast anyway"]);
    expect([r.primary, ...(r.secondary ?? [])].some((a) => a?.label === "Auto-tap & cast")).toBe(
      false,
    );

    // Escape cancels; Enter reaches nothing, and Cast anyway has no key.
    expect(dockKeyAction(r, "Enter")).toBeNull();
    dockKeyAction(r, "Escape")!.onPress();
    expect(h.onCancel).toHaveBeenCalledTimes(1);
    expect(r.secondary?.[0].keyShortcuts).toBeUndefined();
    expect(r.secondary?.[0].title).toBe(
      "cast it without paying its mana cost; the game log says so",
    );
    r.secondary![0].onPress();
    expect(h.onCastAnyway).toHaveBeenCalledTimes(1);
    expect(h.onAutoTap).not.toHaveBeenCalled();
  });

  it("keeps all three buttons when the refused cast was not auto-tapped", () => {
    const h = { onAutoTap: vi.fn(), onCastAnyway: vi.fn(), onCancel: vi.fn() };
    const r = insufficientManaRequest(["{G}"], "Llanowar Elves", h, { autoTapped: false });
    expect(r.primary?.label).toBe("Auto-tap & cast");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Cancel", "Cast anyway"]);
  });
});
