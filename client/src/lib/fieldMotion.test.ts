// fieldMotion — permanents arriving and leaving, as pure functions:
// priming and rewinds, where a departed permanent went (the next frame,
// then the log), tokens, combat deaths left to the strikes, control
// changes, arrivals on boards the viewer does not control, the landing
// selectors and the tapped copy's box.

import { describe, expect, it } from "vitest";

import {
  ARRIVAL_MS,
  ArrivalClock,
  combatCards,
  departureCaption,
  departureCaptions,
  emptyFieldTracker,
  fieldDestination,
  fieldLandingSelectors,
  trackField,
  uprightBox,
  type FieldCard,
  type FieldTracker,
} from "./fieldMotion";
import type { CardView, GameView, LogEvent } from "./protocol";

const ME = "me";
const OPP = "opp";

const card = (id: string, extra: Partial<CardView> = {}): CardView =>
  ({ instance_id: id, name: id, owner: OPP, controller: OPP, ...extra }) as CardView;

const zone = (cards: CardView[] = []) => ({ kind: "z", count: cards.length, cards });

function view(opts: {
  field?: CardView[];
  graveyard?: Record<string, CardView[]>;
  hand?: Record<string, CardView[]>;
  command?: Record<string, CardView[]>;
  exile?: CardView[];
  phased?: CardView[];
  log?: LogEvent[];
  eliminated?: string[];
}): GameView {
  return {
    battlefield: zone(opts.field),
    exile: zone(opts.exile),
    phased_out: opts.phased ? zone(opts.phased) : undefined,
    seats: [ME, OPP].map((id) => ({
      id,
      eliminated: opts.eliminated?.includes(id) === true,
      graveyard: zone(opts.graveyard?.[id]),
      hand: zone(opts.hand?.[id]),
      command: zone(opts.command?.[id]),
    })),
    log: opts.log ?? [],
  } as unknown as GameView;
}

const entry = (seq: number, extra: Partial<LogEvent> = {}): LogEvent => ({
  seq,
  kind: "step",
  seat: 0,
  text: "",
  ...extra,
});

const drawnCard = (id: string, extra: Partial<FieldCard> = {}): FieldCard => ({
  id,
  owner: OPP,
  controller: OPP,
  isToken: false,
  tapped: false,
  image: `/img/${id}`,
  rect: { left: 10, top: 20, width: 60, height: 84 },
  ...extra,
});

function drawnOf(...cards: FieldCard[]): Map<string, FieldCard> {
  return new Map(cards.map((c) => [c.id, c]));
}

// A tracker primed on `field`, with the log seen up to seq 1.
function primedOn(field: CardView[]): FieldTracker {
  const res = trackField(emptyFieldTracker(), view({ field, log: [entry(1)] }), ME, new Map());
  expect(res.quiet).toBe(true);
  return res.tracker;
}

describe("trackField", () => {
  it("primes on the first frame: nothing departs or arrives", () => {
    const res = trackField(
      emptyFieldTracker(),
      view({ field: [card("a")], log: [entry(1)] }),
      ME,
      new Map(),
    );
    expect(res).toMatchObject({ quiet: true, departures: [], arrivals: [] });
    expect(res.tracker.field.get("a")).toBe(OPP);
  });

  it("re-primes on request, and is quiet on an undo", () => {
    const t = primedOn([card("a")]);
    const reprimed = trackField(t, view({ log: [entry(2)] }), ME, drawnOf(drawnCard("a")), {
      reprime: true,
    });
    expect(reprimed.quiet).toBe(true);
    // Undo: the frame's highest seq is below the watermark.
    const after = trackField(t, view({ field: [card("a")], log: [entry(5)] }), ME, new Map());
    const undone = trackField(
      after.tracker,
      view({ log: [entry(3)] }),
      ME,
      drawnOf(drawnCard("a")),
    );
    expect(undone.quiet).toBe(true);
    expect(undone.departures).toEqual([]);
  });

  it("flies a destroyed creature to its owner's graveyard", () => {
    const t = primedOn([card("bear")]);
    const res = trackField(
      t,
      view({ graveyard: { [OPP]: [card("bear")] }, log: [entry(2, { kind: "zone" })] }),
      ME,
      drawnOf(drawnCard("bear")),
    );
    expect(res.departures).toHaveLength(1);
    expect(res.departures[0].dest).toEqual({ zone: "graveyard", seat: OPP, cardVisible: false });
    expect(res.departures[0].card.image).toBe("/img/bear");
  });

  it("finds an exiled card in the exile strip, by its owner", () => {
    const t = primedOn([card("x", { controller: ME })]);
    const res = trackField(
      t,
      view({ exile: [card("x")], log: [entry(2)] }),
      ME,
      drawnOf(drawnCard("x", { controller: ME })),
    );
    expect(res.departures[0].dest).toEqual({ zone: "exile", seat: OPP, cardVisible: true });
  });

  it("reads a hidden destination from the log's zone entry", () => {
    const t = primedOn([card("b")]);
    const res = trackField(
      t,
      view({
        log: [entry(2, { kind: "zone", card_id: "b", old_zone: "battlefield", new_zone: "hand" })],
      }),
      ME,
      drawnOf(drawnCard("b")),
    );
    expect(res.departures[0].dest).toEqual({ zone: "hand", seat: OPP, cardVisible: false });
  });

  it("fades a token, which ceases to exist (CR 111.7)", () => {
    const t = primedOn([card("tok", { is_token: true })]);
    const res = trackField(
      t,
      view({ log: [entry(2)] }),
      ME,
      drawnOf(drawnCard("tok", { isToken: true })),
    );
    expect(res.departures).toHaveLength(1);
    expect(res.departures[0].dest).toBeNull();
    expect(departureCaption(res.departures[0])).toBe("gone");
  });

  it("leaves a creature that took combat damage to the strikes", () => {
    const t = primedOn([card("blocker"), card("attacker", { owner: ME, controller: ME })]);
    const res = trackField(
      t,
      view({
        graveyard: { [OPP]: [card("blocker")], [ME]: [card("attacker")] },
        log: [entry(2, { kind: "damage", combat: true, card_id: "attacker", target: "blocker" })],
      }),
      ME,
      drawnOf(drawnCard("blocker"), drawnCard("attacker", { owner: ME, controller: ME })),
    );
    expect(res.departures).toEqual([]);
  });

  it("skips a permanent the board never drew", () => {
    const t = primedOn([card("hidden")]);
    const res = trackField(t, view({ log: [entry(2)] }), ME, new Map());
    expect(res.departures).toEqual([]);
  });

  it("flies a stolen permanent to its new slot and rings it there", () => {
    const t = primedOn([card("m", { controller: ME, owner: ME })]);
    const res = trackField(
      t,
      view({ field: [card("m", { controller: OPP, owner: ME })], log: [entry(2)] }),
      ME,
      drawnOf(drawnCard("m", { controller: ME, owner: ME })),
    );
    expect(res.departures[0].dest).toEqual({ zone: "battlefield", seat: OPP, cardVisible: true });
    expect(departureCaption(res.departures[0])).toBe("changed control");
    expect(res.arrivals).toEqual(["m"]);
  });

  it("rings arrivals on other boards only, and every arrival for a spectator", () => {
    const t = primedOn([]);
    const v = view({
      field: [card("mine", { controller: ME, owner: ME }), card("theirs")],
      log: [entry(2)],
    });
    expect(trackField(t, v, ME, new Map()).arrivals).toEqual(["theirs"]);
    expect(trackField(t, v, null, new Map()).arrivals.sort()).toEqual(["mine", "theirs"]);
  });

  it("counts a phased-out permanent as still there", () => {
    const t = primedOn([card("p")]);
    const res = trackField(
      t,
      view({ phased: [card("p")], log: [entry(2)] }),
      ME,
      drawnOf(drawnCard("p")),
    );
    expect(res.departures).toEqual([]);
    expect(res.arrivals).toEqual([]);
  });
});

describe("fieldDestination", () => {
  it("prefers what the frame shows over the log", () => {
    const v = view({ hand: { [ME]: [card("c")] } });
    const fresh = [
      entry(2, { kind: "zone", card_id: "c", old_zone: "battlefield", new_zone: "library" }),
    ];
    expect(fieldDestination(v, { id: "c", owner: ME }, fresh)).toEqual({
      zone: "hand",
      seat: ME,
      cardVisible: true,
    });
  });

  it("takes the newest log entry, and only one that left the battlefield", () => {
    const fresh = [
      entry(2, { kind: "zone", card_id: "c", old_zone: "battlefield", new_zone: "hand" }),
      entry(3, { kind: "zone", card_id: "c", old_zone: "hand", new_zone: "library" }),
      entry(4, { kind: "zone", card_id: "c", old_zone: "battlefield", new_zone: "library" }),
    ];
    expect(fieldDestination(view({}), { id: "c", owner: OPP }, fresh)?.zone).toBe("library");
    expect(fieldDestination(view({}), { id: "c", owner: OPP }, fresh.slice(0, 2))?.zone).toBe(
      "hand",
    );
  });

  it("is null when neither the frame nor the log says", () => {
    expect(fieldDestination(view({}), { id: "c", owner: OPP }, [])).toBeNull();
    const odd = [
      entry(2, { kind: "zone", card_id: "c", old_zone: "battlefield", new_zone: "stack" }),
    ];
    expect(fieldDestination(view({}), { id: "c", owner: OPP }, odd)).toBeNull();
  });
});

describe("combatCards", () => {
  it("names both ends of combat damage and ignores other damage", () => {
    const got = combatCards([
      entry(1, { kind: "damage", combat: true, card_id: "a", target: "b" }),
      entry(2, { kind: "damage", card_id: "bolt", target: "c" }),
    ]);
    expect([...got].sort()).toEqual(["a", "b"]);
  });
});

describe("fieldLandingSelectors", () => {
  it("lands on the card when it is drawn, else on the pile or the seat", () => {
    expect(fieldLandingSelectors({ zone: "exile", seat: OPP, cardVisible: true }, "x")).toEqual([
      '[data-instance-id="x"]',
      '[data-pile="exile"][data-pile-owner="opp"]',
    ]);
    expect(fieldLandingSelectors({ zone: "library", seat: OPP, cardVisible: false }, "x")).toEqual([
      '[data-pile="library"][data-pile-owner="opp"]',
    ]);
    expect(fieldLandingSelectors({ zone: "hand", seat: OPP, cardVisible: false }, "x")).toEqual([
      '[data-seat-id="opp"]',
    ]);
    expect(
      fieldLandingSelectors({ zone: "battlefield", seat: OPP, cardVisible: true }, "x"),
    ).toEqual(['[data-instance-id="x"]']);
  });
});

describe("departure captions", () => {
  it("says a card left the game with its eliminated owner (CR 800.4a)", () => {
    const t = primedOn([card("a"), card("tok", { is_token: true })]);
    const res = trackField(
      t,
      view({ log: [entry(2)], eliminated: [OPP] }),
      ME,
      drawnOf(drawnCard("a"), drawnCard("tok", { isToken: true })),
    );
    expect(res.departures.map(departureCaption)).toEqual(["left the game", "left the game"]);
  });

  it("gives a board wipe one counted caption per board and destination", () => {
    const dep = (id: string, controller: string, zone: "exile" | "graveyard") => ({
      card: drawnCard(id, { controller }),
      dest: { zone, seat: controller, cardVisible: false },
    });
    expect(
      departureCaptions([
        dep("a", OPP, "exile"),
        dep("b", OPP, "exile"),
        dep("c", ME, "exile"),
        dep("d", OPP, "exile"),
        dep("e", OPP, "graveyard"),
      ]),
    ).toEqual(["exiled ×3", null, "exiled", null, "to graveyard"]);
  });
});

describe("uprightBox", () => {
  it("keeps an untapped tile's box", () => {
    const rect = { left: 0, top: 0, width: 60, height: 84 };
    expect(uprightBox({ rect, tapped: false })).toBe(rect);
  });

  it("turns a tapped tile's rotated box upright about its centre", () => {
    const rect = { left: 100, top: 50, width: 84, height: 60 };
    expect(uprightBox({ rect, tapped: true })).toEqual({
      left: 112,
      top: 38,
      width: 60,
      height: 84,
    });
  });
});

describe("ArrivalClock", () => {
  it("keeps a ring for ARRIVAL_MS and restarts it on a second arrival", () => {
    const c = new ArrivalClock();
    c.ring(["a"], 0);
    expect(c.live(ARRIVAL_MS - 1)).toEqual(["a"]);
    c.ring(["a"], 1000);
    expect(c.next()).toBe(1000 + ARRIVAL_MS);
    expect(c.live(ARRIVAL_MS + 1)).toEqual(["a"]);
    expect(c.live(1000 + ARRIVAL_MS)).toEqual([]);
    expect(c.next()).toBeNull();
  });
});
