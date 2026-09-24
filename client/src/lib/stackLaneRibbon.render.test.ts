// @vitest-environment jsdom
//
// StackLaneRibbon (#1467, PR 2 of 3) — the style-C lane body: one
// horizontal, numbered row in resolution order. stackLane.test.ts
// pins the model; this pins the presentation contract the issue
// asked for: the top item is ringed and marked "1 · NEXT", items
// read left to right in resolution order, names read off the model
// verbatim, Counter reaches `controls.counter`, and every item
// carries `data-stack-item-id` (CombatArrows' stack-target arrows
// key off it).

import { describe, it, expect, afterEach } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";

import StackLaneRibbon from "./components/board/StackLaneRibbon.svelte";
import { buildStackLane, type StackLaneControls, type StackLaneItem } from "./stackLane";
import type { CardView, PlayerView, StackItemView, ZoneView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";
const BOT2 = "bot2";

const zone = (kind: string, cards: CardView[] = []): ZoneView =>
  ({ kind, count: cards.length, cards }) as ZoneView;

const seat = (id: string, name: string, n: number): PlayerView =>
  ({ id, name, seat: n, life: 40, graveyard: zone("graveyard") }) as PlayerView;

const card = (id: string, name: string, controller: string): CardView =>
  ({
    instance_id: id,
    name,
    owner: controller,
    controller,
    known_by_you: true,
  }) as CardView;

const bolt = card("bolt", "Lightning Bolt", ME);
const counterspell = card("cs", "Counterspell", BOT2);
const birds = card("birds", "Birds of Paradise", BOT2);
const vivi = card("vivi", "Vivi Ornitier", ME);

// Cast order: Lightning Bolt first (bottom of stack, resolves last),
// then Vivi's trigger, then Counterspell last (top, resolves next) —
// countering the Bolt.
const boltSpell: StackItemView = {
  id: "bolt",
  kind: "spell",
  controller: ME,
  owner: ME,
  source_card_id: "bolt",
  targets: [{ kind: "card", id: "birds" }],
};

const viviTrigger: StackItemView = {
  id: "trig-vivi",
  kind: "triggered",
  controller: ME,
  owner: ME,
  source_card_id: "vivi",
  label: "Vivi Ornitier — 1 damage to each opponent",
};

const counterspellSpell: StackItemView = {
  id: "cs",
  kind: "spell",
  controller: BOT2,
  owner: BOT2,
  source_card_id: "cs",
  targets: [{ kind: "card", id: "bolt" }],
};

const pendingTrigger: StackItemView = {
  id: "trig-pending",
  kind: "triggered",
  controller: BOT2,
  owner: BOT2,
  source_card_id: "cs",
  label: "Rite of Flame — waiting to go on the stack",
};

function buildModel() {
  return buildStackLane({
    stack: zone("stack", [bolt, counterspell]),
    stackItems: [boltSpell, viviTrigger, counterspellSpell],
    pendingTriggers: [pendingTrigger],
    seats: [seat(ME, "Me", 0), seat(BOT2, "Bot 2", 1)],
    battlefield: zone("battlefield", [birds, vivi]),
    exile: zone("exile"),
    viewerID: ME,
    priorityHolder: ME,
    splitSecondActive: false,
  });
}

function mount(targetableIDs: Set<string> = new Set()) {
  const model = buildModel();
  const counterCalls: StackLaneItem[] = [];
  const targetCalls: StackLaneItem[] = [];
  const controls: StackLaneControls = {
    counter: (item) => counterCalls.push(item),
    pass: undefined,
    targetable: (item) => targetableIDs.has(item.id),
    target: (item) => targetCalls.push(item),
    hoverEnter: () => {},
    hoverLeave: () => {},
  };
  const r = render(
    StackLaneRibbon as never,
    {
      model,
      controls,
      styleName: "ribbon",
    } as never,
  );
  return {
    ...r,
    model,
    counterCalls,
    targetCalls,
    q: (sel: string) => r.container.querySelector<HTMLElement>(sel),
    qa: (sel: string) => Array.from(r.container.querySelectorAll<HTMLElement>(sel)),
  };
}

describe("StackLaneRibbon (#1467)", () => {
  it("rings the top item and marks it '1 · NEXT'", () => {
    const b = mount();
    const top = b.q('[data-stack-item-id="cs"]')!;
    expect(top.classList.contains("top")).toBe(true);
    expect(top.textContent).toContain("1 · NEXT");
  });

  it("lists items left to right in resolution order", () => {
    const b = mount();
    const ids = b.qa("[data-stack-item-id]").map((el) => el.dataset.stackItemId);
    expect(ids).toEqual(["cs", "trig-vivi", "bolt", "trig-pending"]);
  });

  it("keeps item names verbatim", () => {
    const b = mount();
    expect(b.container.textContent).toContain("Counterspell");
    expect(b.container.textContent).toContain("Vivi Ornitier — 1 damage to each opponent");
    expect(b.container.textContent).toContain("Lightning Bolt");
    expect(b.container.textContent).toContain("Rite of Flame — waiting to go on the stack");
  });

  it("marks a pending trigger as waiting, with no resolution number", () => {
    const b = mount();
    const pending = b.q('[data-stack-item-id="trig-pending"]')!;
    expect(pending.classList.contains("waiting")).toBe(true);
    expect(pending.textContent).toContain("waiting");
  });

  it("flags the item another item on the stack targets", () => {
    const b = mount();
    const boltItem = b.q('[data-stack-item-id="bolt"]')!;
    expect(boltItem.classList.contains("targeted")).toBe(true);
    expect(boltItem.textContent).toContain("targeted by Counterspell");
  });

  it("joins items with a chevron connector", () => {
    const b = mount();
    expect(b.qa(".chevron").length).toBe(3);
  });

  it("scrolls horizontally rather than wrapping", () => {
    const b = mount();
    expect(b.q(".row")).not.toBeNull();
    // jsdom does not apply component CSS, so this reads the rule
    // itself — the same technique stackLane.render.test.ts uses.
    const css = readFileSync(
      join(process.cwd(), "src/lib/components/board/StackLaneRibbon.svelte"),
      "utf8",
    );
    const rule = (sel: string) => css.slice(css.indexOf(`${sel} {`)).split("}")[0];
    expect(rule("  .row")).toContain("overflow-x: auto");
    expect(rule("  .row")).not.toContain("flex-wrap: wrap");
  });

  it("Counter calls controls.counter with the item", () => {
    const b = mount();
    const row = b.q('[data-stack-item-id="cs"]')!.closest(".r-line")!;
    const btn = row.querySelector<HTMLButtonElement>(".counter-btn")!;
    click(btn);
    expect(b.counterCalls.map((i) => i.id)).toEqual(["cs"]);
  });

  it("exposes a keyboard-reachable Counter button with an aria-label", () => {
    const b = mount();
    const row = b.q('[data-stack-item-id="cs"]')!.closest(".r-line")!;
    const btn = row.querySelector<HTMLButtonElement>(".counter-btn")!;
    expect(btn.tagName).toBe("BUTTON");
    expect(btn.getAttribute("aria-label")).toBe("Counter Counterspell");
  });

  it("data-stack-item-id is present on every item", () => {
    const b = mount();
    expect(b.qa("[data-stack-item-id]").length).toBe(4);
  });

  it("routes a click to controls.target only when the item is targetable", () => {
    const b = mount(new Set(["trig-vivi"]));
    const row = b.q('[data-stack-item-id="trig-vivi"]')!;
    expect(row.getAttribute("role")).toBe("button");
    click(row);
    expect(b.targetCalls.map((i) => i.id)).toEqual(["trig-vivi"]);

    const top = b.q('[data-stack-item-id="cs"]')!;
    expect(top.getAttribute("role")).toBeNull();
  });
});
