// @vitest-environment jsdom
//
// #1467 — the floating stack lane. The model is pinned in
// stackLane.test.ts; this file pins what only exists once the board
// is rendered: which surface draws the stack for which setting, that
// the lane comes and goes with the stack without the board grid
// changing under it, that its controls reach the same handlers the
// docked card's do, and that the top item is announced.

import { describe, it, expect, afterEach, beforeEach } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";

import Board from "./components/board/Board.svelte";
import { updateSettings } from "./settings";
import type { StackStyle } from "./stackLane";
import type { ActionType, CardView, GameView, PlayerView, StackItemView } from "./protocol";
import { render, click, cleanup } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

const ME = "me";
const OPP = "opp";

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
});

afterEach(() => {
  cleanup();
  updateSettings("display", "stackStyle", "compact");
});

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

const seat = (id: string, name: string, n: number): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
  }) as unknown as PlayerView;

const mulldrifter = {
  instance_id: "mull",
  name: "Mulldrifter",
  owner: ME,
  controller: ME,
  known_by_you: true,
  type_line: "Creature — Elemental",
  power: 2,
  toughness: 2,
} as CardView;

const trigger: StackItemView = {
  id: "trig-1",
  kind: "triggered",
  controller: ME,
  owner: ME,
  source_card_id: "mull",
  label: "Mulldrifter — draw two cards",
};

const gameView = (stackItems: StackItemView[]): GameView =>
  ({
    id: "g1",
    state: "active",
    seats: [seat(ME, "Me", 0), seat(OPP, "Bot 2", 1)],
    battlefield: zone("battlefield", undefined, [mulldrifter]),
    stack: zone("stack", undefined),
    exile: zone("exile", undefined),
    stack_items: stackItems,
    pending_triggers: [],
    pending_choices: [],
    turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    mulligans_open: false,
  }) as unknown as GameView;

function mountBoard(style: StackStyle, stackItems: StackItemView[]) {
  updateSettings("display", "stackStyle", style);
  const sent: { type: ActionType; params?: unknown }[] = [];
  const r = render(
    Board as never,
    {
      view: gameView(stackItems),
      viewerID: ME,
      isAdmin: false,
      sendAction: (type: ActionType, params?: unknown) => sent.push({ type, params }),
      combatMode: "idle",
      selectedCombatCardID: null,
      onSelectCombatCard: () => {},
      onDeclareAttack: () => {},
      onDeclareBlock: () => {},
    } as never,
  );
  const q = (sel: string) => r.container.querySelector<HTMLElement>(sel);
  return { ...r, sent, q };
}

// ADR 0111 §4 / §10 (Delivery PR 4): the strip the docked stack card
// lives in is `region "attention"`, the tutorial's step-9 anchor
// (ADR 0076 §2.4) and an e2e contract (board-layout.spec.ts).
describe("the attention strip", () => {
  it('is region "attention", and the docked stack card is in it', () => {
    const b = mountBoard("compact", [trigger]);
    // By role: BattlefieldRow also wears a `strip` class (its compact rows).
    const strip = b.q(".board > .strip");
    expect(strip?.getAttribute("role")).toBe("region");
    expect(strip?.getAttribute("aria-label")).toBe("attention");
    expect(b.container.querySelectorAll('[role="region"][aria-label="attention"]')).toHaveLength(1);
    expect(strip?.querySelector(".overlay")).not.toBeNull();
  });
});

describe("the stack's display style (#1467)", () => {
  it("compact draws the docked card and no lane", () => {
    const b = mountBoard("compact", [trigger]);
    expect(b.q(".strip .overlay")).not.toBeNull();
    expect(b.q(".stack-lane")).toBeNull();
    expect(b.q("[data-stack-lane-announcer]")).toBeNull();
    expect(b.container.textContent).toContain("Mulldrifter — draw two cards");
  });

  for (const style of ["pile", "fan", "spotlight", "ribbon"] as const) {
    it(`${style} floats a lane while the stack is live, instead of the docked card`, () => {
      const b = mountBoard(style, [trigger]);
      const lane = b.q(".stack-lane");
      expect(lane).not.toBeNull();
      expect(b.q(".lane-host")?.dataset.stackStyle).toBe(style);
      // Never both: the docked card is not rendered while the lane shows.
      expect(b.q(".strip .overlay")).toBeNull();
      // The label e2e specs look for, verbatim.
      expect(lane?.textContent).toContain("Mulldrifter — draw two cards");
      expect(lane?.getAttribute("aria-label")).toBe("stack: 1 on the stack");
    });
  }

  it("comes and goes with the stack, and the board grid does not change", () => {
    const b = mountBoard("spotlight", []);
    const slots = () =>
      Array.from(b.container.querySelectorAll<HTMLElement>(".board > .slot")).map(
        (s) => s.dataset.pos,
      );
    const before = slots();
    expect(b.q(".stack-lane")).toBeNull();

    b.setProps({ view: gameView([trigger]) } as never);
    expect(b.q(".stack-lane")).not.toBeNull();
    expect(slots()).toEqual(before);

    b.setProps({ view: gameView([]) } as never);
    expect(b.q(".stack-lane")).toBeNull();
    expect(slots()).toEqual(before);
  });

  it("sits over the board rather than inside a seat, and passes clicks through its gaps", () => {
    const b = mountBoard("ribbon", [trigger]);
    const host = b.q(".lane-host")!;
    // A child of the board, beside the seats, never wrapping one — so
    // it is layered over the grid and cannot take a grid cell.
    expect(host.parentElement?.classList.contains("board")).toBe(true);
    expect(host.querySelector(".slot")).toBeNull();
    // The seats under it are still in the document and still wired.
    expect(b.q('.board > .slot[data-pos="self"]')).not.toBeNull();
    // Only the lane's own content takes pointer events. jsdom does not
    // apply component CSS, so this reads the rule itself. (vitest runs
    // from client/; under jsdom import.meta.url is not a file URL.)
    const css = readFileSync(
      join(process.cwd(), "src/lib/components/board/StackLaneHost.svelte"),
      "utf8",
    );
    const rule = (sel: string) => css.slice(css.indexOf(`${sel} {`)).split("}")[0];
    expect(rule("  .lane-host")).toContain("pointer-events: none");
    expect(rule("  .stack-lane")).toContain("pointer-events: auto");
  });

  it("wires Counter exactly as the docked card does", () => {
    const b = mountBoard("spotlight", [trigger]);
    click(b.q(".stack-lane .counter-btn")!);
    expect(b.sent).toEqual([{ type: "counter_ability", params: { instance_id: "trig-1" } }]);
  });

  // ADR 0111 §7 (S56 PR 2): the action dock's `next` is the one pass
  // button and its hold the one hold. Neither stack surface draws them.
  it.each(["compact", "fan", "spotlight", "ribbon"] as const)(
    "the %s stack has no Pass and no hold",
    (style) => {
      const b = mountBoard(style, [trigger]);
      const surface = b.q(style === "compact" ? ".strip .overlay" : ".stack-lane")!;
      expect(surface).not.toBeNull();
      expect(surface.querySelector(".pass-btn, .hold-toggle")).toBeNull();
      const names = [...surface.querySelectorAll("button")].map((x) =>
        (x.textContent ?? "").trim().toLowerCase(),
      );
      expect(names.some((n) => n.startsWith("pass") || n.startsWith("hold"))).toBe(false);
      expect(names).toContain("counter");
    },
  );

  it("announces the top item in a polite live region", () => {
    const b = mountBoard("fan", []);
    const live = b.q("[data-stack-lane-announcer]");
    // Mounted before anything is on the stack, so the first item is
    // a change a screen reader hears.
    expect(live?.getAttribute("aria-live")).toBe("polite");
    expect(live?.textContent).toBe("");
    b.setProps({ view: gameView([trigger]) } as never);
    expect(b.q("[data-stack-lane-announcer]")?.textContent).toBe(
      "Mulldrifter's trigger resolves next — draw two cards",
    );
  });
});
