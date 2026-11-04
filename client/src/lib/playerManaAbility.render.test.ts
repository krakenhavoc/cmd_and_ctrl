// @vitest-environment jsdom
//
// playerManaAbility.render.test.ts — #4885. A seat with a player mana
// ability (Channel's "pay 1 life: add {C}" until end of turn) carries
// `mana_abilities`; its own panel shows one button per ability, and each
// press sends `activate_mana_ability { player_ability: <source> }` for the
// seat, with no card_id. The button is greyed while the life can't be
// paid, and nothing renders without one.

import { describe, it, expect, afterEach } from "vitest";

import PlayerManaAbilities from "./components/board/PlayerManaAbilities.svelte";
import type { ActionType, PlayerManaAbilityView } from "./protocol";
import { L } from "./labels";
import { render, click, cleanup } from "./test/render.svelte";

afterEach(cleanup);

const ME = "me";

const channel = (usable = true): PlayerManaAbilityView => ({
  source: "channel-1",
  label: "Channel",
  life_cost: 1,
  produced: "{C}",
  usable,
});

function mount(abilities: PlayerManaAbilityView[] | undefined) {
  const sent: { type: string; params: unknown; player?: string }[] = [];
  const { container } = render(
    PlayerManaAbilities as never,
    {
      seatID: ME,
      abilities,
      sendAction: (type: ActionType, params?: unknown, player?: string) =>
        sent.push({ type, params, player }),
    } as never,
  );
  return { container, sent };
}

const button = (c: HTMLElement) =>
  c.querySelector<HTMLButtonElement>(
    `button[aria-label='${L.playerManaAbility("Channel", 1, "{C}")}']`,
  );

describe("PlayerManaAbilities", () => {
  it("renders nothing when the seat has no player mana ability", () => {
    expect(button(mount(undefined).container)).toBeNull();
    expect(button(mount([]).container)).toBeNull();
  });

  it("names the ability and its price", () => {
    const btn = button(mount([channel()]).container);
    expect(btn).not.toBeNull();
    expect(btn!.getAttribute("aria-label")).toBe("Channel: pay 1 life, add {C}");
    expect(btn!.textContent).toContain("Channel");
    expect(btn!.textContent).toContain("1 life");
  });

  it("sends one activation per press, naming the source and no card", () => {
    const { container, sent } = mount([channel()]);
    const btn = button(container)!;
    click(btn);
    click(btn);
    expect(sent).toEqual([
      { type: "activate_mana_ability", params: { player_ability: "channel-1" }, player: ME },
      { type: "activate_mana_ability", params: { player_ability: "channel-1" }, player: ME },
    ]);
  });

  it("is greyed and sends nothing while the life can't be paid", () => {
    const { container, sent } = mount([channel(false)]);
    const btn = button(container)!;
    expect(btn.disabled).toBe(true);
    click(btn);
    expect(sent).toEqual([]);
  });
});
