// @vitest-environment jsdom
//
// #1278: commander ninjutsu (CR 702.49c) is an activated ability that
// functions FROM THE COMMAND ZONE. The server ships its row on
// `zone_abilities`, owner-only, exactly as a hand card's (#1221), so the
// wire needed nothing. What the client needed is to show it where the
// command zone's actions live. #2349: that is the castable strip beside
// your hand, since the command zone tile is gone: the strip hands the
// rows to the commander's Card, whose popover (right-click, as on a
// hand card) offers them with #1227's shortfall greying.

import { describe, it, expect, afterEach } from "vitest";

import ExileStrip from "./components/board/ExileStrip.svelte";
import type { CardView, GameView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const yuriko = (returnCards: string[], withRows = true): CardView =>
  ({
    instance_id: "yuriko",
    name: "Yuriko, the Tiger's Shadow",
    owner: "me",
    controller: "me",
    is_commander: true,
    type_line: "Legendary Creature — Human Ninja",
    zone_abilities: withRows
      ? [
          {
            index: 0,
            label:
              "Commander ninjutsu {U}{B} ({U}{B}, Return an unblocked attacker you control to hand: " +
              "Put this card onto the battlefield from your hand or the command zone tapped and attacking.)",
            mana_cost: "{U}{B}",
            return_label: "an unblocked attacker you control",
            return_options: { cards: returnCards, min: 1, max: 1 },
          },
        ]
      : undefined,
  }) as unknown as CardView;

function snap(card: CardView): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      {
        id: "me",
        name: "Me",
        hand: { kind: "hand", count: 0, cards: [] },
        command: { kind: "command", count: 1, cards: [card] },
      },
    ],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: 0, cards: [] },
    turn: {
      seq: 4,
      number: 4,
      active_seat: 0,
      priority_holder: 0,
      phase: "combat",
      step: "declare_blockers",
    },
  } as unknown as GameView;
}

function mountStrip(card: CardView, withHandler = true) {
  const fired: Array<[string, number]> = [];
  const cast: string[] = [];
  const r = render(
    ExileStrip as never,
    {
      view: snap(card),
      viewerID: "me",
      onCastCard: (c: CardView) => cast.push(c.instance_id),
      ...(withHandler
        ? {
            onActivateAbility: (c: CardView, index: number) => fired.push([c.instance_id, index]),
          }
        : {}),
    } as never,
  );
  return { container: r.container, fired, cast };
}

function openPopover(container: HTMLElement): NodeListOf<HTMLButtonElement> {
  const el = container.querySelector<HTMLElement>(".card[aria-label^='Yuriko']")!;
  const r = el.getBoundingClientRect();
  el.dispatchEvent(
    new MouseEvent("contextmenu", {
      bubbles: true,
      cancelable: true,
      clientX: r.left + r.width / 2,
      clientY: r.top + r.height / 2,
    }),
  );
  flushSync();
  return container.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
}

const ninjutsuRow = (rows: NodeListOf<HTMLButtonElement>) =>
  Array.from(rows).find((r) => r.textContent?.includes("Commander ninjutsu {U}{B}"));

describe("commander ninjutsu in the strip beside the hand", () => {
  it("offers the row in the commander's popover and fires it, without casting", () => {
    const strip = mountStrip(yuriko(["rat"]));
    const row = ninjutsuRow(openPopover(strip.container));
    expect(row).toBeDefined();
    expect(row!.disabled).toBe(false);
    click(row!);
    flushSync();
    expect(strip.fired).toEqual([["yuriko", 0]]);
    expect(strip.cast).toEqual([]);
  });

  it("greys the row while no unblocked attacker can pay", () => {
    const strip = mountStrip(yuriko([]));
    const row = ninjutsuRow(openPopover(strip.container));
    expect(row).toBeDefined();
    expect(row!.disabled).toBe(true);
    click(row!);
    flushSync();
    expect(strip.fired).toEqual([]);
  });

  it("offers nothing on a commander with no command-zone ability", () => {
    const strip = mountStrip(yuriko([], false));
    expect(ninjutsuRow(openPopover(strip.container))).toBeUndefined();
  });

  it("offers nothing without an ability handler", () => {
    const strip = mountStrip(yuriko(["rat"]), false);
    expect(ninjutsuRow(openPopover(strip.container))).toBeUndefined();
  });
});
