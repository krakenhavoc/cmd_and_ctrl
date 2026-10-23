// @vitest-environment jsdom
//
// ignoreEffects.render.test.ts — #1986, CR 116.2d. A seat a Damping Engine
// binds carries `ignorable_effects`; its own panel shows a button that
// opens the sacrifice picker over the permanents the server listed, and a
// confirm sends `special_action { kind: "ignore_effect", card_id,
// sacrifice_id }` for the seat. Nothing renders without an effect.

import { describe, it, expect, afterEach } from "vitest";

import IgnoreEffects from "./components/board/IgnoreEffects.svelte";
import type { ActionType, CardView, IgnorableEffectView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";
import DockHarness from "./test/DockHarness.svelte";
import { barPrimary, barSecondaries, dockTestView } from "./test/dockView";

afterEach(cleanup);

const ME = "me";

function permanent(id: string, name: string, typeLine: string): CardView {
  return {
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    type_line: typeLine,
  } as CardView;
}

const engine = permanent("engine", "Damping Engine", "Artifact");
const token = permanent("token", "Spirit", "Token Creature — Spirit");
const bear = permanent("bear", "Grizzly Bears", "Creature — Bear");

const row: IgnorableEffectView = {
  source_id: "engine",
  source_name: "Damping Engine",
  label: "Sacrifice a permanent: ignore Damping Engine's effect until end of turn",
  sacrifice_ids: ["token", "bear", "engine"],
};

function mount(effects: IgnorableEffectView[] | undefined) {
  const sent: { type: string; params: unknown; player?: string }[] = [];
  const base = dockTestView();
  const view = {
    ...base,
    battlefield: { ...base.battlefield, count: 3, cards: [engine, bear, token] },
  };
  const { container } = render(
    DockHarness as never,
    {
      component: IgnoreEffects,
      props: {
        seatID: ME,
        effects,
        view,
        sendAction: (type: ActionType, params?: unknown, player?: string) =>
          sent.push({ type, params, player }),
      },
    } as never,
  );
  return { container, sent };
}

const button = (c: HTMLElement) =>
  c.querySelector<HTMLButtonElement>(
    "button[aria-label='Ignore Damping Engine (sacrifice a permanent)']",
  );
const rows = (c: HTMLElement): HTMLButtonElement[] => [
  ...c.querySelectorAll<HTMLButtonElement>(".dock-sheet .prompt-options .prompt-opt"),
];

describe("IgnoreEffects", () => {
  it("renders nothing when no effect bears on the seat", () => {
    expect(button(mount(undefined).container)).toBeNull();
    expect(button(mount([]).container)).toBeNull();
  });

  it("offers the listed permanents and sends the pick as the cost", () => {
    const { container, sent } = mount([row]);
    const btn = button(container);
    expect(btn).not.toBeNull();
    expect(btn!.title).toBe(row.label);
    expect(rows(container)).toHaveLength(0);

    click(btn!);
    // The picker lists the server's permanents, in the server's order.
    const picks = rows(container);
    expect(picks.map((r) => r.textContent ?? "")).toEqual([
      expect.stringContaining("Spirit"),
      expect.stringContaining("Grizzly Bears"),
      expect.stringContaining("Damping Engine"),
    ]);
    // The player chooses: not the first.
    click(picks[1]);
    click(barPrimary(container)!);
    expect(sent).toEqual([
      {
        type: "special_action",
        params: { card_id: "engine", kind: "ignore_effect", sacrifice_id: "bear" },
        player: ME,
      },
    ]);
  });

  it("sends nothing when the picker is cancelled", () => {
    const { container, sent } = mount([row]);
    click(button(container)!);
    expect(rows(container)).toHaveLength(3);
    const cancel = barSecondaries(container).find((b) => /cancel/i.test(b.textContent ?? ""));
    expect(cancel).toBeDefined();
    click(cancel!);
    expect(rows(container)).toHaveLength(0);
    expect(sent).toEqual([]);
  });
});
