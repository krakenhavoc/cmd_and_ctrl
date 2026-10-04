// castStrip.test.ts — #2202. The pure half of the castable-from-other-
// zones strip: which commanders join the exile cards beside the hand,
// in which order, lit or greyed by the server's move list, and what
// the price tag says. Every input is a server answer (the viewer's own
// command zone, `cast_prices`, `legal_moves`, `commander_casts`).

import { describe, it, expect } from "vitest";

import {
  castStripBadge,
  castStripEntries,
  castStripLegality,
  commanderCostBadge,
  commanderStripEntries,
  commanderTax,
} from "./castStrip";
import type { CardView, CastPriceView, GameView, LegalMoveView } from "./protocol";

const ME = "me";
const THEM = "them";

const commander = (id: string, name: string, prices?: CastPriceView[]): CardView =>
  ({
    instance_id: id,
    name,
    owner: ME,
    controller: ME,
    is_commander: true,
    type_line: "Legendary Creature — Elf Druid",
    mana_cost: "{2}{G}",
    ...(prices ? { cast_prices: prices } : {}),
  }) as CardView;

const zone = (kind: string, cards: CardView[]) => ({ kind, count: cards.length, cards });

function snap(opts: {
  mine?: CardView[];
  theirs?: CardView[];
  exile?: CardView[];
  casts?: Record<string, number>;
  moves?: LegalMoveView[];
}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      {
        id: ME,
        name: "Me",
        hand: zone("hand", []),
        command: zone("command", opts.mine ?? []),
        commander_casts: opts.casts,
      },
      {
        id: THEM,
        name: "Them",
        hand: zone("hand", []),
        command: zone("command", opts.theirs ?? []),
      },
    ],
    battlefield: zone("battlefield", []),
    stack: zone("stack", []),
    exile: zone("exile", opts.exile ?? []),
    turn: { seq: 4, number: 4, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
    legal_moves: opts.moves,
  } as unknown as GameView;
}

const castMove = (source: string, zoneName = "command"): LegalMoveView =>
  ({
    type: "cast_spell",
    player: ME,
    kind: "cast",
    label: "Cast",
    source,
    zone: zoneName,
  }) as LegalMoveView;

const impulse: CardView = {
  instance_id: "bolt",
  name: "Impulse Bolt",
  owner: THEM,
  controller: THEM,
  type_line: "Instant",
  mana_cost: "{R}",
  exile_play: { player: ME },
  castable_here: true,
  cast_prices: [{ cost: "{R}", printed: true }],
} as CardView;

describe("commanderStripEntries", () => {
  it("puts the viewer's commander in the strip, cast from the command zone", () => {
    const e = commanderStripEntries(snap({ mine: [commander("k", "Kenrith")] }), ME);
    expect(e).toHaveLength(1);
    expect(e[0]).toMatchObject({ zone: "command", verb: "cast", casts: 0 });
    expect(e[0].card.instance_id).toBe("k");
  });

  it("is empty when the commander is not in the command zone", () => {
    expect(commanderStripEntries(snap({ mine: [] }), ME)).toEqual([]);
  });

  it("holds both partners, in command zone order", () => {
    const e = commanderStripEntries(
      snap({ mine: [commander("a", "Thrasios"), commander("b", "Tymna")] }),
      ME,
    );
    expect(e.map((x) => x.card.name)).toEqual(["Thrasios", "Tymna"]);
  });

  it("never holds an opponent's commander", () => {
    const view = snap({ theirs: [commander("o", "Their Commander")] });
    expect(commanderStripEntries(view, ME)).toEqual([]);
    expect(castStripEntries(view, ME)).toEqual([]);
  });

  it("holds nothing for a spectator", () => {
    expect(commanderStripEntries(snap({ mine: [commander("k", "Kenrith")] }), null)).toEqual([]);
  });

  it("carries the server's cast count for the hover", () => {
    const e = commanderStripEntries(
      snap({ mine: [commander("k", "Kenrith")], casts: { k: 2 } }),
      ME,
    );
    expect(e[0].casts).toBe(2);
  });
});

describe("castStripEntries", () => {
  it("puts the commanders first, then the exile cards", () => {
    const e = castStripEntries(snap({ mine: [commander("k", "Kenrith")], exile: [impulse] }), ME);
    expect(e.map((x) => [x.card.instance_id, x.zone])).toEqual([
      ["k", "command"],
      ["bolt", "exile"],
    ]);
  });
});

describe("castStripLegality", () => {
  it("lights a commander the move list can cast from the command zone", () => {
    const view = snap({ mine: [commander("k", "Kenrith")], moves: [castMove("k")] });
    const [e] = castStripEntries(view, ME);
    expect(castStripLegality(e, view, ME).legal).toBe(true);
  });

  it("greys a commander the move list leaves out", () => {
    const view = snap({ mine: [commander("k", "Kenrith")], moves: [castMove("other")] });
    const [e] = castStripEntries(view, ME);
    expect(castStripLegality(e, view, ME).legal).toBe(false);
  });

  it("greys a commander when the viewer does not have priority", () => {
    const view = snap({ mine: [commander("k", "Kenrith")] });
    (view.turn as { priority_holder: number }).priority_holder = 1;
    const [e] = castStripEntries(view, ME);
    expect(castStripLegality(e, view, ME)).toMatchObject({ legal: false });
  });
});

describe("the commander's price tag", () => {
  it("has no tag before the first cast: the price is the printed cost", () => {
    const card = commander("k", "Kenrith", [{ cost: "{2}{G}", printed: true }]);
    expect(commanderCostBadge(card, 0)).toBeNull();
  });

  it("tags tax 2 with the total and lists printed cost plus tax on hover", () => {
    const card = commander("k", "Kenrith", [{ cost: "{4}{G}" }]);
    const b = commanderCostBadge(card, 1)!;
    expect(b.symbols).toEqual(["4", "G"]);
    expect(b.title).toContain("Casts from the command zone for {4}{G}");
    expect(b.title).toContain("printed cost {2}{G}");
    expect(b.title).toContain("commander tax +{2} (cast 1 time before)");
    expect(b.label).toBe("costs {4}{G} to cast from the command zone");
  });

  it("tags tax 4 with the total", () => {
    const card = commander("k", "Kenrith", [{ cost: "{6}{G}" }]);
    const b = commanderCostBadge(card, 2)!;
    expect(b.symbols).toEqual(["6", "G"]);
    expect(b.title).toContain("commander tax +{4} (cast 2 times before)");
  });

  it("shows the server's total, not printed + tax, under a cost modifier", () => {
    // {2}{G} + {4} tax − a Medallion's {1}: the server says {5}{G}.
    const card = commander("k", "Kenrith", [{ cost: "{5}{G}" }]);
    expect(commanderCostBadge(card, 2)!.symbols).toEqual(["5", "G"]);
  });

  it("has no tag on a frame that carries no price", () => {
    expect(commanderCostBadge(commander("k", "Kenrith"), 3)).toBeNull();
  });

  it("castStripBadge reads the entry's zone and cast count", () => {
    const view = snap({
      mine: [commander("k", "Kenrith", [{ cost: "{4}{G}" }])],
      casts: { k: 1 },
      exile: [impulse],
    });
    const [cmd, ex] = castStripEntries(view, ME);
    expect(castStripBadge(cmd)?.title).toContain("commander tax +{2}");
    // The impulse card pays its printed cost: no tag, as before.
    expect(castStripBadge(ex)).toBeNull();
  });

  it("commanderTax is {2} per earlier cast (CR 903.8)", () => {
    expect([0, 1, 2, undefined].map(commanderTax)).toEqual([0, 2, 4, 0]);
  });
});
