// @vitest-environment jsdom
//
// StackLaneSpotlight (#1467, PR 2 of 3) — the style-B lane body: the
// top stack item shown large, with the rest of the stack and any
// pending triggers queued beside it. stackLane.test.ts pins the
// model; this pins the presentation contract the issue asked for:
// the top item is visibly marked next, the queue is in resolution
// order, names read off the model verbatim (e2e specs look for
// them), Counter reaches `controls.counter`, and every item carries
// `data-stack-item-id` (CombatArrows' stack-target arrows key off it).

import { describe, it, expect, afterEach } from "vitest";

import StackLaneSpotlight from "./components/board/StackLaneSpotlight.svelte";
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
    StackLaneSpotlight as never,
    {
      model,
      controls,
      styleName: "spotlight",
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

describe("StackLaneSpotlight (#1467)", () => {
  it("marks the top item as resolving next", () => {
    const b = mount();
    const art = b.q('[data-stack-item-id="cs"]')!;
    expect(art).not.toBeNull();
    expect(art.closest(".hero")!.textContent).toContain("Resolves next");
  });

  it("orders the rest of the stack, then pending triggers, after the top item", () => {
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

  it("shows the top item's target chip with its queue position", () => {
    const b = mount();
    const hero = b.q('[data-stack-item-id="cs"]')!.closest(".hero")!;
    expect(hero.textContent).toContain("queue #3");
  });

  it("flags the item another item on the stack targets", () => {
    const b = mount();
    const boltRow = b.q('[data-stack-item-id="bolt"]')!;
    expect(boltRow.textContent).toContain("targeted by Counterspell");
  });

  it("Counter calls controls.counter with the item", () => {
    const b = mount();
    const btn = b.q(".hero .counter-btn") as HTMLButtonElement;
    click(btn);
    expect(b.counterCalls.map((i) => i.id)).toEqual(["cs"]);
  });

  it("Counter on a queued item calls controls.counter with that item", () => {
    const b = mount();
    const row = b.q('[data-stack-item-id="trig-vivi"]')!.closest(".q-line")!;
    const btn = row.querySelector<HTMLButtonElement>(".counter-btn")!;
    click(btn);
    expect(b.counterCalls.map((i) => i.id)).toEqual(["trig-vivi"]);
  });

  it("exposes a keyboard-reachable Counter button with an aria-label", () => {
    const b = mount();
    const btn = b.q(".hero .counter-btn") as HTMLButtonElement;
    expect(btn.tagName).toBe("BUTTON");
    expect(btn.getAttribute("aria-label")).toBe("Counter Counterspell");
  });

  it("data-stack-item-id is present on every item, hero included", () => {
    const b = mount();
    expect(b.qa("[data-stack-item-id]").length).toBe(4);
  });

  it("routes a click to controls.target only when the item is targetable", () => {
    const b = mount(new Set(["trig-vivi"]));
    const row = b.q('[data-stack-item-id="trig-vivi"]')!;
    expect(row.getAttribute("role")).toBe("button");
    click(row);
    expect(b.targetCalls.map((i) => i.id)).toEqual(["trig-vivi"]);

    const hero = b.q('[data-stack-item-id="cs"]')!;
    expect(hero.getAttribute("role")).toBeNull();
  });
});
