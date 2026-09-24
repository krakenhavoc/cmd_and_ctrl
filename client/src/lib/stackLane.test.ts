import { describe, expect, it } from "vitest";

import type { CardView, PlayerView, StackItemView, ZoneView } from "./protocol";
import { seatColor } from "./colors";
import {
  buildStackLane,
  isStackStyle,
  joinPhrases,
  labelEffect,
  oracleEffect,
  stackLaneLive,
  type StackLaneInput,
} from "./stackLane";

// A table of three: the viewer (LUke, seat 0) and two bots.
const ME = "me";
const BOT1 = "bot1";
const BOT2 = "bot2";

const zone = (kind: string, cards: CardView[] = []): ZoneView =>
  ({ kind, count: cards.length, cards }) as ZoneView;

const seat = (id: string, name: string, n: number, graveyard: CardView[] = []): PlayerView =>
  ({ id, name, seat: n, life: 40, graveyard: zone("graveyard", graveyard) }) as PlayerView;

const card = (id: string, name: string, controller: string, extra: Partial<CardView> = {}) =>
  ({
    instance_id: id,
    name,
    owner: controller,
    controller,
    known_by_you: true,
    ...extra,
  }) as CardView;

const spell = (id: string, controller: string, extra: Partial<StackItemView> = {}) =>
  ({
    id,
    kind: "spell",
    controller,
    owner: controller,
    source_card_id: id,
    ...extra,
  }) as StackItemView;

const trigger = (id: string, controller: string, source: string, label: string) =>
  ({
    id,
    kind: "triggered",
    controller,
    owner: controller,
    source_card_id: source,
    label,
  }) as StackItemView;

const bolt = card("bolt", "Lightning Bolt", ME);
const counterspell = card("cs", "Counterspell", BOT2);
const birds = card("birds", "Birds of Paradise", BOT2);
const vivi = card("vivi", "Vivi Ornitier", ME);

const ORACLE: Record<string, string> = {
  "Lightning Bolt": "Lightning Bolt deals 3 damage to any target.",
  Counterspell: "Counter target spell.",
};

function input(extra: Partial<StackLaneInput> = {}): StackLaneInput {
  return {
    stack: zone("stack", [bolt, counterspell]),
    // Wire order is bottom..top: Bolt was cast first, Counterspell on top of it.
    stackItems: [
      spell("bolt", ME, { targets: [{ kind: "card", id: "birds" }] }),
      spell("cs", BOT2, { targets: [{ kind: "card", id: "bolt" }] }),
    ],
    pendingTriggers: [],
    seats: [seat(ME, "LUke", 0), seat(BOT1, "Bot 1", 1), seat(BOT2, "Bot 2", 2)],
    battlefield: zone("battlefield", [birds, vivi]),
    exile: zone("exile"),
    viewerID: ME,
    priorityHolder: ME,
    splitSecondActive: false,
    oracleTextFor: (c) => ORACLE[c.name],
    ...extra,
  };
}

describe("stack lane ordering", () => {
  it("lists the stack top-first, numbered, with only the top marked", () => {
    const m = buildStackLane(input());
    expect(m.stackItems.map((i) => i.name)).toEqual(["Counterspell", "Lightning Bolt"]);
    expect(m.stackItems.map((i) => i.position)).toEqual([1, 2]);
    expect(m.stackItems.map((i) => i.isTop)).toEqual([true, false]);
    expect(m.top?.id).toBe("cs");
    expect(m.live).toBe(true);
  });

  it("puts pending triggers after the stack, never on top, unnumbered", () => {
    const m = buildStackLane(
      input({ pendingTriggers: [trigger("t1", ME, "vivi", "Vivi Ornitier — +1/+1 counter")] }),
    );
    expect(m.items.map((i) => i.kind)).toEqual(["spell", "spell", "pending"]);
    expect(m.pendingTriggers[0]).toMatchObject({
      isTop: false,
      position: null,
      name: "Vivi Ornitier — +1/+1 counter",
    });
    expect(m.top?.id).toBe("cs");
  });

  it("is live on pending triggers alone, and not live when empty", () => {
    const pendingOnly = input({
      stack: zone("stack"),
      stackItems: [],
      pendingTriggers: [trigger("t1", ME, "vivi", "Vivi Ornitier — +1/+1 counter")],
    });
    expect(buildStackLane(pendingOnly).live).toBe(true);
    expect(buildStackLane(pendingOnly).top).toBeNull();
    expect(buildStackLane(input({ stack: zone("stack"), stackItems: [] })).live).toBe(false);
    expect(stackLaneLive([], [])).toBe(false);
    expect(stackLaneLive(undefined, [trigger("t", ME, "vivi", "x")])).toBe(true);
  });
});

describe("stack lane items", () => {
  it("names an ability by its label verbatim and previews its source permanent", () => {
    const m = buildStackLane(
      input({
        stack: zone("stack"),
        stackItems: [trigger("t1", ME, "vivi", "Mulldrifter — draw two cards")],
      }),
    );
    const it0 = m.stackItems[0];
    expect(it0.name).toBe("Mulldrifter — draw two cards");
    expect(it0.kind).toBe("triggered");
    expect(it0.previewCard?.instance_id).toBe("vivi");
    expect(it0.effect).toEqual({ kind: "text", text: "draw two cards" });
    expect(it0.chips.map((c) => c.label)).toEqual(["triggered"]);
  });

  it("carries the caster's seat, name and colour", () => {
    const cs = buildStackLane(input()).top!;
    expect(cs).toMatchObject({ casterSeat: BOT2, casterName: "Bot 2", casterIsViewer: false });
    expect(cs.casterColor).toBe(seatColor(2));
  });

  it("will not preview a card the viewer may not read", () => {
    const hidden = card("bolt", "", ME, { known_by_you: false });
    const m = buildStackLane(input({ stack: zone("stack", [hidden, counterspell]) }));
    expect(m.stackItems[1].previewCard).toBeNull();
  });

  it("keeps the doubled-trigger label", () => {
    const t = {
      ...trigger("t1", ME, "vivi", "Vivi Ornitier — draw a card"),
      doubled_by: "p",
      doubled_by_name: "Panharmonicon",
    };
    const m = buildStackLane(input({ stackItems: [t] }));
    expect(m.top?.doubledLabel).toBe("additional (Panharmonicon)");
    expect(m.top?.chips.map((c) => c.label)).toContain("additional (Panharmonicon)");
  });

  it("builds the chips in the order the docked card always drew them", () => {
    const x = spell("bolt", ME, {
      x_value: 3,
      alt_cost: "overload",
      gift_to: BOT1,
      mode_labels: ["Destroy target artifact."],
      split_second: true,
      hold_priority: true,
    });
    const m = buildStackLane(
      input({
        stack: zone("stack", [
          card("bolt", "Lightning Bolt", ME, { auto: true, unimplemented: true }),
        ]),
        stackItems: [x],
      }),
    );
    expect(m.top?.chips.map((c) => [c.label, c.tone])).toEqual([
      ["auto", "flag"],
      ["manual", "manual"],
      ["X = 3", "plain"],
      ["overload", "flag"],
      ["Gift → Bot 1", "flag"],
      ["Destroy target artifact.", "plain"],
      ["split-second", "flag"],
      ["held priority", "flag"],
    ]);
  });
});

describe("stack lane targets", () => {
  it("resolves a permanent, a player, a stack item and a graveyard card", () => {
    const gy = card("gyc", "Sol Ring", BOT1);
    const item = spell("bolt", ME, {
      targets: [
        { kind: "card", id: "birds" },
        { kind: "player", id: BOT1 },
        { kind: "card", id: "cs" },
        { kind: "card", id: "gyc" },
        { kind: "self" },
        { kind: "none" },
      ],
    });
    const m = buildStackLane(
      input({
        stackItems: [item, spell("cs", BOT2)],
        seats: [seat(ME, "LUke", 0), seat(BOT1, "Bot 1", 1, [gy]), seat(BOT2, "Bot 2", 2)],
      }),
    );
    const bolt = m.stackItems.find((i) => i.id === "bolt")!;
    expect(bolt.targets.map((t) => [t.kind, t.id, t.name, t.ownerSeat])).toEqual([
      ["permanent", "birds", "Birds of Paradise", BOT2],
      ["player", BOT1, "Bot 1", BOT1],
      ["stack", "cs", "Counterspell", BOT2],
      ["card", "gyc", "Sol Ring", BOT1],
    ]);
    expect(bolt.targets.find((t) => t.kind === "stack")?.stackPosition).toBe(1);
    // The overlay's old line keeps self / none placeholders verbatim.
    expect(bolt.targetText).toBe(
      "→ Birds of Paradise / Bot 1 / Counterspell / Sol Ring / self / —",
    );
  });

  it("phrases targets from the viewer's side", () => {
    const m = buildStackLane(
      input({
        stackItems: [
          spell("bolt", ME, {
            targets: [
              { kind: "player", id: ME },
              { kind: "card", id: "vivi" },
              { kind: "card", id: "birds" },
            ],
          }),
        ],
      }),
    );
    expect(m.top?.targets.map((t) => t.phrase)).toEqual([
      "you",
      "your Vivi Ornitier",
      "Bot 2's Birds of Paradise",
    ]);
    expect(m.top?.targets.map((t) => t.isViewers)).toEqual([true, true, false]);
  });

  it("says no 'your' to a spectator", () => {
    const m = buildStackLane(input({ viewerID: null }));
    expect(m.top?.targets[0].phrase).toBe("LUke's Lightning Bolt");
    expect(m.priority.viewerHolds).toBe(false);
  });

  it("records which items target an item", () => {
    const m = buildStackLane(input());
    expect(m.stackItems.find((i) => i.id === "bolt")?.targetedBy).toEqual(["cs"]);
    expect(m.top?.targetedBy).toEqual([]);
  });
});

describe("stack lane summary", () => {
  it("says what a counterspell counters, from the viewer's side", () => {
    expect(buildStackLane(input()).summary).toBe(
      "Counterspell resolves next and counters your Lightning Bolt",
    );
  });

  it("says how much damage goes where", () => {
    const m = buildStackLane(
      input({
        stack: zone("stack", [bolt]),
        stackItems: [spell("bolt", ME, { targets: [{ kind: "card", id: "birds" }] })],
      }),
    );
    expect(m.summary).toBe("Lightning Bolt resolves next — 3 damage to Bot 2's Birds of Paradise");
  });

  it("reads X off the item for an X damage spell", () => {
    const fb = card("fb", "Blaze", ME);
    const m = buildStackLane(
      input({
        stack: zone("stack", [fb]),
        stackItems: [spell("fb", ME, { x_value: 5, targets: [{ kind: "player", id: BOT1 }] })],
        oracleTextFor: () => "Blaze deals X damage to any target.",
      }),
    );
    expect(m.summary).toBe("Blaze resolves next — 5 damage to Bot 1");
  });

  it("names an ability by its source and states its label's effect", () => {
    const m = buildStackLane(
      input({
        stack: zone("stack"),
        stackItems: [trigger("t1", ME, "vivi", "Vivi Ornitier — 1 damage to each opponent")],
      }),
    );
    expect(m.summary).toBe("Vivi Ornitier's trigger resolves next — 1 damage to each opponent");
  });

  it("falls back to '<name> resolves next' when nothing can be said honestly", () => {
    const m = buildStackLane(input({ oracleTextFor: undefined, stackItems: [spell("bolt", ME)] }));
    expect(m.summary).toBe("Lightning Bolt resolves next");
  });

  it("still names the target when the effect is unknown", () => {
    const m = buildStackLane(input({ oracleTextFor: undefined }));
    expect(m.summary).toBe("Counterspell resolves next, targeting your Lightning Bolt");
  });

  it("describes pending triggers when the stack is empty", () => {
    const one = buildStackLane(
      input({
        stack: zone("stack"),
        stackItems: [],
        pendingTriggers: [trigger("t1", ME, "vivi", "Vivi Ornitier — draw a card")],
      }),
    );
    expect(one.summary).toBe("Vivi Ornitier — draw a card is waiting to go on the stack");
    const none = buildStackLane(input({ stack: zone("stack"), stackItems: [] }));
    expect(none.summary).toBe("");
  });
});

describe("oracleEffect reads only what the card says", () => {
  it("reads a bare counter and a bare damage sentence", () => {
    expect(oracleEffect("Counter target spell.", "Counterspell")).toEqual({
      kind: "counter",
      text: "counter target spell",
    });
    expect(oracleEffect("Shock deals 2 damage to any target.", "Shock")).toMatchObject({
      kind: "damage",
      amount: "2",
      recipients: "any target",
    });
    expect(
      oracleEffect(
        "This spell can't be countered.\nCounter target noncreature spell.",
        "Dovin's Veto",
      ),
    ).toMatchObject({ kind: "counter" });
  });

  it("refuses a sentence with a rider it would have to drop", () => {
    expect(
      oracleEffect("Counter target spell unless its controller pays {3}.", "Mana Leak"),
    ).toBeNull();
    expect(
      oracleEffect(
        "Counter target noncreature spell. Its controller loses 2 life.",
        "Countersquall",
      ),
    ).toBeNull();
    expect(
      oracleEffect(
        "Lightning Helix deals 3 damage to any target and you gain 3 life.",
        "Lightning Helix",
      ),
    ).toBeNull();
    expect(oracleEffect("Draw two cards.", "Divination")).toBeNull();
    expect(oracleEffect(undefined, "Anything")).toBeNull();
  });

  it("keeps a single recipient list that says 'and each'", () => {
    expect(
      oracleEffect("Earthquake deals X damage to each creature and each player.", "Earthquake", 2),
    ).toMatchObject({ amount: "2", recipients: "each creature and each player" });
  });
});

describe("helpers", () => {
  it("labelEffect takes the half after the dash", () => {
    expect(labelEffect("Mulldrifter — draw two cards")).toBe("draw two cards");
    expect(labelEffect("{T}: Add {G}.")).toBeNull();
    expect(labelEffect(undefined)).toBeNull();
  });

  it("joinPhrases reads as English", () => {
    expect(joinPhrases([])).toBe("");
    expect(joinPhrases(["a"])).toBe("a");
    expect(joinPhrases(["a", "b"])).toBe("a and b");
    expect(joinPhrases(["a", "b", "c"])).toBe("a, b and c");
  });

  it("isStackStyle accepts exactly the four styles", () => {
    for (const s of ["compact", "fan", "spotlight", "ribbon"]) expect(isStackStyle(s)).toBe(true);
    expect(isStackStyle("carousel")).toBe(false);
    expect(isStackStyle(undefined)).toBe(false);
  });

  it("reports who holds priority", () => {
    expect(buildStackLane(input({ priorityHolder: BOT1, considering: true })).priority).toEqual({
      viewerHolds: false,
      holderSeat: BOT1,
      holderName: "Bot 1",
      considering: true,
    });
    expect(buildStackLane(input()).priority.viewerHolds).toBe(true);
  });
});
