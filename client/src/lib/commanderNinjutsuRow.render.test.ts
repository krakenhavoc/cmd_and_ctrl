// @vitest-environment jsdom
//
// #1278: commander ninjutsu (CR 702.49c) is an activated ability that
// functions FROM THE COMMAND ZONE. The server ships its row on
// `zone_abilities`, owner-only, exactly as a hand card's (#1221), so the
// wire needed nothing. What the client needed is to show it where the
// command zone's actions live: the CommandZone tile hands the rows to
// its Card (right-click opens the shared popover, with #1227's
// shortfall greying) and shows an "ability" hint beside "cast".

import { describe, it, expect, afterEach } from "vitest";

import CommandZone from "./components/board/CommandZone.svelte";
import type { CardView, ZoneView } from "./protocol";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

afterEach(cleanup);

const yuriko = (returnCards: string[], withRows = true): CardView =>
  ({
    instance_id: "yuriko",
    name: "Yuriko, the Tiger's Shadow",
    owner: "me",
    controller: "me",
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

function mountZone(
  card: CardView,
  isSelf = true,
): { container: HTMLElement; fired: Array<[string, number]>; sent: string[] } {
  const fired: Array<[string, number]> = [];
  const sent: string[] = [];
  const zone = { count: 1, cards: [card] } as unknown as ZoneView;
  const view = render(
    CommandZone as never,
    {
      seat: { id: "me", name: "Me" },
      zone,
      isSelf,
      sendAction: (type: string) => sent.push(type),
      onActivateAbility: (c: CardView, index: number) => fired.push([c.instance_id, index]),
    } as never,
  );
  return { container: view.container, fired, sent };
}

function abilityHint(container: HTMLElement): HTMLButtonElement | null {
  return container.querySelector<HTMLButtonElement>(".ability-hint");
}

describe("commander ninjutsu in the command zone", () => {
  it("shows an ability hint that opens the row and fires it, without casting", () => {
    const tile = mountZone(yuriko(["rat"]));
    const hint = abilityHint(tile.container);
    expect(hint).not.toBeNull();

    click(hint!);
    flushSync();
    const rows = tile.container.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].disabled).toBe(false);
    expect(rows[0].textContent).toContain("Commander ninjutsu {U}{B}");

    click(rows[0]);
    flushSync();
    expect(tile.fired).toEqual([["yuriko", 0]]);
    expect(tile.sent).toEqual([]);
  });

  it("greys the row while no unblocked attacker can pay", () => {
    const tile = mountZone(yuriko([]));
    click(abilityHint(tile.container)!);
    flushSync();
    const rows = tile.container.querySelectorAll<HTMLButtonElement>('[role="menuitem"]');
    expect(rows.length).toBe(1);
    expect(rows[0].disabled).toBe(true);
    click(rows[0]);
    flushSync();
    expect(tile.fired).toEqual([]);
  });

  it("offers nothing on a commander with no command-zone ability", () => {
    const tile = mountZone(yuriko([], false));
    expect(abilityHint(tile.container)).toBeNull();
  });

  it("offers nothing on another seat's command zone", () => {
    const tile = mountZone(yuriko(["rat"]), false);
    expect(abilityHint(tile.container)).toBeNull();
  });
});
