import { describe, it, expect } from "vitest";

import { exileCostBadge, exileStripEntries, manaSymbols, pendingHint } from "./exileStrip";
import type { CardView, CastPriceView, ExilePlayView, GameView } from "./protocol";

// exileStrip.test.ts — #1389. The strip beside the hand shows the
// exiled cards the viewer may cast, and a price tag when the price is
// not the printed cost. Both halves read server answers only
// (`castable_here`, `cast_prices`, `exile_play`); these tests pin which
// cards appear, in which state, when they go, and what the tag says.

const ME = "me";
const THEM = "them";

function snap(cards: CardView[], turn = 5): GameView {
  return {
    id: "g",
    state: "active",
    seats: [],
    battlefield: { kind: "battlefield", count: 0, cards: [] },
    stack: { kind: "stack", count: 0, cards: [] },
    exile: { kind: "exile", count: cards.length, cards },
    turn: {
      seq: turn,
      number: turn,
      active_seat: 0,
      priority_holder: 0,
      phase: "main1",
      step: "main",
    },
  } as unknown as GameView;
}

const grant = (extras: Partial<ExilePlayView> = {}): ExilePlayView => ({ player: ME, ...extras });
const price = (cost: string, extras: Partial<CastPriceView> = {}): CastPriceView => ({
  cost,
  ...extras,
});

function exiled(id: string, extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: id,
    name: id,
    owner: THEM,
    controller: THEM,
    type_line: "Sorcery",
    mana_cost: "{1}{R}",
    ...extras,
  };
}

const ids = (v: GameView, viewer: string | null = ME) =>
  exileStripEntries(v, viewer).map((e) => e.card.instance_id);

describe("exileStripEntries — which cards appear", () => {
  it("is empty with no view, no viewer, or an exile with nothing granted", () => {
    expect(exileStripEntries(null, ME)).toEqual([]);
    expect(exileStripEntries(snap([exiled("a", { exile_play: grant() })]), null)).toEqual([]);
    expect(ids(snap([exiled("plain")]))).toEqual([]);
  });

  it("shows a card the viewer may cast now, lit", () => {
    const v = snap([
      exiled("bolt", { exile_play: grant(), castable_here: true, cast_prices: [price("{1}{R}")] }),
    ]);
    const [e] = exileStripEntries(v, ME);
    expect(e.state).toBe("now");
    expect(e.verb).toBe("cast");
    expect(e.hint).toBeUndefined();
  });

  it("leaves out a card whose grant names somebody else", () => {
    // Exile is public, so the viewer's frame carries the other seat's
    // grant — and none of that seat's per-viewer answers.
    const v = snap([exiled("theirs", { exile_play: grant({ player: THEM }) })]);
    expect(ids(v)).toEqual([]);
  });

  it("keeps a card the viewer holds a live permission over even when the public grant names another seat", () => {
    // #1037: two seats may hold permissions over one card; the
    // per-viewer stamps are the viewer's own answer.
    const v = snap([
      exiled("shared", {
        exile_play: grant({ player: THEM }),
        castable_here: true,
        cast_prices: [price("{1}{R}")],
      }),
    ]);
    expect(ids(v)).toEqual(["shared"]);
  });

  it("shows a live permission whose window is shut as waiting, not hidden", () => {
    // A plotted card outside its owner's main phase: priced, not now.
    const v = snap([
      exiled("plotted", { exile_play: grant({ cast_only: true }), cast_prices: [price("{0}")] }),
    ]);
    const [e] = exileStripEntries(v, ME);
    expect(e.state).toBe("waiting");
    expect(e.hint).toBeUndefined();
  });

  it("dims a warp / plot / foretell grant before its later turn, with a hint", () => {
    const v = snap(
      [
        exiled("warped", { exile_play: grant({ not_before_seq: 6 }) }),
        exiled("far", { exile_play: grant({ not_before_seq: 8 }) }),
      ],
      5,
    );
    const got = exileStripEntries(v, ME);
    expect(got.map((e) => [e.card.instance_id, e.state, e.hint])).toEqual([
      ["warped", "later", "next turn"],
      ["far", "later", "a later turn"],
    ]);
  });

  it("orders castable cards first, then waiting, then later — each in exile order", () => {
    const v = snap([
      exiled("later", { exile_play: grant({ not_before_seq: 9 }) }),
      exiled("waiting", { exile_play: grant(), cast_prices: [price("{2}")] }),
      exiled("now-1", { exile_play: grant(), castable_here: true }),
      exiled("now-2", { exile_play: grant(), castable_here: true }),
    ]);
    expect(ids(v)).toEqual(["now-1", "now-2", "waiting", "later"]);
  });

  it("drops a card when the next frame stops carrying it — cast, left exile, or permission ended", () => {
    const live = exiled("impulse", { exile_play: grant(), castable_here: true });
    expect(ids(snap([live]))).toEqual(["impulse"]);
    // Cast or moved: gone from exile.
    expect(ids(snap([]))).toEqual([]);
    // Still in exile, permission swept at cleanup: no grant, no stamps.
    expect(ids(snap([exiled("impulse")]))).toEqual([]);
  });

  it("strands a land under a cast-only grant and offers a play-granted land as a play", () => {
    const v = snap([
      exiled("ragavan-land", {
        type_line: "Basic Land — Mountain",
        mana_cost: "",
        exile_play: grant({ cast_only: true }),
      }),
      exiled("breeches-land", {
        type_line: "Basic Land — Island",
        mana_cost: "",
        exile_play: grant(),
        castable_here: true,
      }),
    ]);
    const got = exileStripEntries(v, ME);
    expect(got.map((e) => e.card.instance_id)).toEqual(["breeches-land"]);
    expect(got[0].verb).toBe("play");
  });

  it("hands the grant's named face to the cast chain, and none for a faceless grant", () => {
    const v = snap([
      exiled("siege", {
        exile_play: grant({ faces: [1] }),
        castable_here: true,
        faces: [
          { name: "Invasion", type_line: "Battle — Siege" },
          { name: "Back Face", type_line: "Creature — Horror" },
        ],
      }),
      exiled("plain", { exile_play: grant(), castable_here: true }),
    ]);
    const got = exileStripEntries(v, ME);
    expect(got.find((e) => e.card.instance_id === "siege")?.face).toBe(1);
    expect(got.find((e) => e.card.instance_id === "plain")?.face).toBeUndefined();
  });

  it("reads castable_here off a face when the grant leaves the half open", () => {
    // An adventure card impulse-exiled by Ragavan: the server stamps
    // each castable face, and only the Adventure half is castable now.
    const v = snap([
      exiled("giant", {
        layout: "adventure",
        exile_play: grant(),
        faces: [
          { name: "Giant", type_line: "Creature — Giant", mana_cost: "{2}{R}" },
          {
            name: "Stomp",
            type_line: "Instant — Adventure",
            mana_cost: "{1}{R}",
            castable_here: true,
          },
        ],
      }),
    ]);
    expect(exileStripEntries(v, ME)[0].state).toBe("now");
  });

  it("never shows an opponent's face-down foretold card — the redaction leaves nothing to read", () => {
    // What a non-knower receives: a card back, no grant, no stamps.
    const v = snap([
      {
        instance_id: "hidden",
        name: "",
        owner: THEM,
        controller: THEM,
        face_down: true,
        face_down_kind: "foretold",
      },
    ]);
    expect(ids(v)).toEqual([]);
  });

  it("shows the viewer's own foretold card with its face", () => {
    const own = exiled("foretold", {
      owner: ME,
      name: "Saw It Coming",
      face_down: true,
      face_down_kind: "foretold",
      face_visible: true,
      exile_play: grant({ cast_only: true }),
      castable_here: true,
      cast_prices: [price("{1}{U}", { alternative_cost: "foretell", label: "Foretell" })],
    });
    const [e] = exileStripEntries(snap([own]), ME);
    expect(e.card.name).toBe("Saw It Coming");
    expect(e.state).toBe("now");
  });
});

describe("pendingHint", () => {
  it("is silent for a grant with no floor, or once the floor is reached", () => {
    expect(pendingHint(exiled("a", { exile_play: grant() }), 5)).toBeUndefined();
    expect(
      pendingHint(exiled("a", { exile_play: grant({ not_before_seq: 5 }) }), 5),
    ).toBeUndefined();
    expect(
      pendingHint(exiled("a", { exile_play: grant({ not_before_seq: 6 }) }), undefined),
    ).toBeUndefined();
  });
});

describe("exileCostBadge — when the tag shows and what it says", () => {
  it("shows nothing without a price: a pending grant, a land, somebody else's card", () => {
    expect(exileCostBadge(exiled("a", { exile_play: grant() }))).toBeNull();
  });

  it("shows nothing when the price is the printed cost — impulse, warp, adventure", () => {
    expect(
      exileCostBadge(exiled("a", { cast_prices: [price("{1}{R}", { printed: true })] })),
    ).toBeNull();
  });

  it("reads 0 for a free cast — a plotted card", () => {
    const b = exileCostBadge(
      exiled("plotted", { mana_cost: "{3}{R}", cast_prices: [price("{0}")] }),
    );
    expect(b?.symbols).toEqual(["0"]);
    expect(b?.label).toBe("costs free to cast from exile");
    expect(b?.title).toContain("printed cost {3}{R}");
  });

  it("reads 2 for airbend", () => {
    const b = exileCostBadge(
      exiled("airbent", { mana_cost: "{5}{G}", cast_prices: [price("{2}")] }),
    );
    expect(b?.symbols).toEqual(["2"]);
  });

  it("spells out a coloured price symbol by symbol — a foretell cost", () => {
    const b = exileCostBadge(
      exiled("foretold", {
        mana_cost: "{1}{U}{U}",
        cast_prices: [price("{1}{U}", { alternative_cost: "foretell", label: "Foretell" })],
      }),
    );
    expect(b?.symbols).toEqual(["1", "U"]);
    expect(b?.title).toContain("(Foretell)");
  });

  it("shows a discounted or taxed price — the server's number after CR 601.2f", () => {
    // An impulse card under a Thalia: printed {1}{R}, charged {2}{R}.
    const b = exileCostBadge(exiled("taxed", { cast_prices: [price("{2}{R}")] }));
    expect(b?.symbols).toEqual(["2", "R"]);
  });

  it("shows the cheapest and lists the rest on hover", () => {
    const b = exileCostBadge(
      exiled("two-ways", {
        mana_cost: "{4}{R}",
        cast_prices: [price("{R}", { alternative_cost: "evoke", label: "Evoke" }), price("{2}")],
      }),
    );
    expect(b?.symbols).toEqual(["R"]);
    expect(b?.title.split("\n")).toEqual([
      "Casts from exile for {R} (Evoke)",
      "printed cost {4}{R}",
      "or {2}",
    ]);
  });

  it("names the printed cost of the face the grant opens, not the face exile shows", () => {
    const b = exileCostBadge(
      exiled("siege", {
        mana_cost: "{2}{W}",
        exile_play: grant({ faces: [1] }),
        faces: [
          { name: "Invasion", type_line: "Battle — Siege", mana_cost: "{2}{W}" },
          { name: "Back Face", type_line: "Creature — Angel", mana_cost: "" },
        ],
        cast_prices: [price("{0}")],
      }),
    );
    expect(b?.title).not.toContain("printed cost {2}{W}");
  });

  it("carries a life half when a price has one", () => {
    const b = exileCostBadge(exiled("citadel", { cast_prices: [price("{0}", { life: 3 })] }));
    expect(b?.life).toBe(3);
    expect(b?.label).toBe("costs free + 3 life to cast from exile");
  });
});

describe("manaSymbols", () => {
  it("splits brace notation, hybrid symbols intact", () => {
    expect(manaSymbols("{X}{2}{W/U}{G}")).toEqual(["X", "2", "W/U", "G"]);
    expect(manaSymbols("")).toEqual([]);
  });
});
