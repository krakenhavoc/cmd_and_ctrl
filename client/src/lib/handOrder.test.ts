// #1524: rearrange the cards in your own hand. The order is the
// viewer's alone and lives in this browser; these pin the reconcile
// rule, the move, the three one-time sorts and the storage, including
// storage that throws. Hand.svelte only wires them to the markup.

import { describe, it, expect } from "vitest";

import {
  HAND_ORDER_KEY_PREFIX,
  HAND_SORTS,
  applyHandOrder,
  colorRank,
  handOrderKey,
  idsOf,
  loadHandOrder,
  manaValue,
  moveCard,
  sameOrder,
  saveHandOrder,
  sortHand,
  typeRank,
  type OrderStorage,
} from "./handOrder";
import type { CardView } from "./protocol";

function card(id: string, extras: Partial<CardView> = {}): CardView {
  return { instance_id: id, name: id, owner: "me", controller: "me", ...extras };
}

const ids = (cs: CardView[]) => cs.map((c) => c.instance_id);

class MemStorage implements OrderStorage {
  data = new Map<string, string>();
  getItem(k: string): string | null {
    return this.data.get(k) ?? null;
  }
  setItem(k: string, v: string): void {
    this.data.set(k, v);
  }
}

const throwing: OrderStorage = {
  getItem() {
    throw new Error("SecurityError");
  },
  setItem() {
    throw new Error("QuotaExceededError");
  },
};

describe("applyHandOrder — the saved order over the server hand", () => {
  const hand = [card("a"), card("b"), card("c")];

  it("with nothing saved, shows the server order", () => {
    expect(ids(applyHandOrder(hand, null))).toEqual(["a", "b", "c"]);
    expect(ids(applyHandOrder(hand, []))).toEqual(["a", "b", "c"]);
  });

  it("shows the saved order", () => {
    expect(ids(applyHandOrder(hand, ["c", "a", "b"]))).toEqual(["c", "a", "b"]);
  });

  it("appends a new card (a draw) on the right, in server order", () => {
    const drawn = [card("a"), card("x"), card("b"), card("c"), card("y")];
    expect(ids(applyHandOrder(drawn, ["c", "a", "b"]))).toEqual(["c", "a", "b", "x", "y"]);
  });

  it("drops a card that left the hand, keeping the others' relative order", () => {
    expect(ids(applyHandOrder([card("a"), card("c")], ["c", "b", "a"]))).toEqual(["c", "a"]);
  });

  it("handles an empty hand", () => {
    expect(applyHandOrder([], ["a", "b"])).toEqual([]);
    expect(applyHandOrder([], null)).toEqual([]);
  });

  it("an order saved for a different hand is ignored: server order", () => {
    expect(ids(applyHandOrder(hand, ["p", "q", "r"]))).toEqual(["a", "b", "c"]);
  });

  it("ignores a duplicated ID in the saved list", () => {
    expect(ids(applyHandOrder(hand, ["b", "b", "a"]))).toEqual(["b", "a", "c"]);
  });

  it("does not mutate its inputs", () => {
    const saved = ["c", "a"];
    applyHandOrder(hand, saved);
    expect(ids(hand)).toEqual(["a", "b", "c"]);
    expect(saved).toEqual(["c", "a"]);
  });
});

describe("sameOrder / idsOf", () => {
  it("compares ID lists by content and order", () => {
    expect(sameOrder(["a", "b"], ["a", "b"])).toBe(true);
    expect(sameOrder(["a", "b"], ["b", "a"])).toBe(false);
    expect(sameOrder(["a"], ["a", "b"])).toBe(false);
    expect(sameOrder(null, ["a"])).toBe(false);
    expect(sameOrder(null, null)).toBe(true);
  });

  it("idsOf reads instance IDs in order", () => {
    expect(idsOf([card("x"), card("y")])).toEqual(["x", "y"]);
  });
});

describe("moveCard", () => {
  const order = ["a", "b", "c", "d"];

  it("moves right: the card ends at the target index", () => {
    expect(moveCard(order, 0, 2)).toEqual(["b", "c", "a", "d"]);
    expect(moveCard(order, 0, 3)).toEqual(["b", "c", "d", "a"]);
  });

  it("moves left", () => {
    expect(moveCard(order, 3, 0)).toEqual(["d", "a", "b", "c"]);
    expect(moveCard(order, 2, 1)).toEqual(["a", "c", "b", "d"]);
  });

  it("to the same index is a copy, unchanged", () => {
    const out = moveCard(order, 1, 1);
    expect(out).toEqual(order);
    expect(out).not.toBe(order);
  });

  it("clamps the target and ignores an out-of-range source", () => {
    expect(moveCard(order, 0, 99)).toEqual(["b", "c", "d", "a"]);
    expect(moveCard(order, 1, -5)).toEqual(["b", "a", "c", "d"]);
    expect(moveCard(order, 9, 0)).toEqual(order);
    expect(moveCard(order, -1, 0)).toEqual(order);
  });
});

describe("the sort keys", () => {
  it("offers three sorts", () => {
    expect(HAND_SORTS.map((s) => s.key)).toEqual(["manaValue", "type", "color"]);
  });

  it("mana value reads the printed cost; a land is 0 and X is 0", () => {
    expect(manaValue(card("x", { mana_cost: "{2}{U}{U}" }))).toBe(4);
    expect(manaValue(card("x", { mana_cost: "{X}{R}" }))).toBe(1);
    expect(manaValue(card("x", { type_line: "Basic Land — Forest" }))).toBe(0);
  });

  it("type ranks lands, creatures, other permanents, instants/sorceries", () => {
    expect(typeRank(card("x", { type_line: "Basic Land — Island" }))).toBe(0);
    expect(typeRank(card("x", { type_line: "Land Creature — Forest Dryad" }))).toBe(0);
    expect(typeRank(card("x", { type_line: "Artifact Creature — Golem" }))).toBe(1);
    expect(typeRank(card("x", { type_line: "Legendary Enchantment" }))).toBe(2);
    expect(typeRank(card("x", { type_line: "Legendary Planeswalker — Jace" }))).toBe(2);
    expect(typeRank(card("x", { type_line: "Artifact — Equipment" }))).toBe(2);
    expect(typeRank(card("x", { type_line: "Instant" }))).toBe(3);
    expect(typeRank(card("x", { type_line: "Sorcery" }))).toBe(3);
    expect(typeRank(card("x"))).toBe(4);
  });

  it("colour ranks WUBRG, then multicolour, then colourless", () => {
    expect(["W", "U", "B", "R", "G"].map((c) => colorRank(card("x", { colors: [c] })))).toEqual([
      0, 1, 2, 3, 4,
    ]);
    expect(colorRank(card("x", { colors: ["U", "R"] }))).toBe(5);
    expect(colorRank(card("x", { colors: [] }))).toBe(6);
    expect(colorRank(card("x"))).toBe(6);
  });
});

describe("sortHand", () => {
  const bolt = card("bolt", {
    name: "Lightning Bolt",
    type_line: "Instant",
    mana_cost: "{R}",
    colors: ["R"],
  });
  const island = card("island", { name: "Island", type_line: "Basic Land — Island" });
  const bear = card("bear", {
    name: "Grizzly Bears",
    type_line: "Creature — Bear",
    mana_cost: "{1}{G}",
    colors: ["G"],
  });
  const ring = card("ring", { name: "Sol Ring", type_line: "Artifact", mana_cost: "{1}" });
  const charm = card("charm", {
    name: "Izzet Charm",
    type_line: "Instant",
    mana_cost: "{U}{R}",
    colors: ["U", "R"],
  });
  const wrath = card("wrath", {
    name: "Wrath of God",
    type_line: "Sorcery",
    mana_cost: "{2}{W}{W}",
    colors: ["W"],
  });
  const hand = [bolt, island, bear, ring, charm, wrath];

  it("by mana value, then name", () => {
    // 0 Island; 1 Lightning Bolt, Sol Ring; 2 Grizzly Bears, Izzet Charm; 4 Wrath.
    expect(sortHand(hand, "manaValue")).toEqual([
      "island",
      "bolt",
      "ring",
      "bear",
      "charm",
      "wrath",
    ]);
  });

  it("by type: lands first, then creatures, other permanents, spells; then mana value", () => {
    expect(sortHand(hand, "type")).toEqual(["island", "bear", "ring", "bolt", "charm", "wrath"]);
  });

  it("by colour: WUBRG, multicolour, colourless; then mana value", () => {
    // W Wrath; R Bolt; G Bears; UR Charm; colourless Island (0) then Sol Ring (1).
    expect(sortHand(hand, "color")).toEqual(["wrath", "bolt", "bear", "charm", "island", "ring"]);
  });

  it("keeps the current order for full ties (two copies of one card)", () => {
    const a = card("forest-a", { name: "Forest", type_line: "Basic Land — Forest" });
    const b = card("forest-b", { name: "Forest", type_line: "Basic Land — Forest" });
    expect(sortHand([b, a], "manaValue")).toEqual(["forest-b", "forest-a"]);
    expect(sortHand([b, a], "type")).toEqual(["forest-b", "forest-a"]);
    expect(sortHand([a, b], "color")).toEqual(["forest-a", "forest-b"]);
  });

  it("breaks a mana-value tie by name", () => {
    const z = card("z", { name: "Zap", mana_cost: "{2}" });
    const a = card("a", { name: "Abrade", mana_cost: "{1}{R}" });
    expect(sortHand([z, a], "manaValue")).toEqual(["a", "z"]);
  });

  it("an empty hand sorts to nothing", () => {
    expect(sortHand([], "type")).toEqual([]);
  });
});

describe("persistence", () => {
  it("the key is namespaced by game and viewer", () => {
    expect(HAND_ORDER_KEY_PREFIX).toBe("cmdctrl.handOrder.v1");
    expect(handOrderKey("g1", "p1")).toBe("cmdctrl.handOrder.v1:g1:p1");
  });

  it("round-trips an order", () => {
    const s = new MemStorage();
    expect(saveHandOrder("g1", "p1", ["c", "a", "b"], s)).toBe(true);
    expect(loadHandOrder("g1", "p1", s)).toEqual(["c", "a", "b"]);
  });

  it("an order saved for another game or another viewer is not read", () => {
    const s = new MemStorage();
    saveHandOrder("g1", "p1", ["c", "a"], s);
    expect(loadHandOrder("g2", "p1", s)).toBeNull();
    expect(loadHandOrder("g1", "p2", s)).toBeNull();
  });

  it("nothing saved, or no game / viewer, reads as null (server order)", () => {
    const s = new MemStorage();
    expect(loadHandOrder("g1", "p1", s)).toBeNull();
    expect(loadHandOrder(null, "p1", s)).toBeNull();
    expect(loadHandOrder("g1", undefined, s)).toBeNull();
    expect(saveHandOrder(null, "p1", ["a"], s)).toBe(false);
    expect(s.data.size).toBe(0);
  });

  it("an unreadable value reads as null", () => {
    const s = new MemStorage();
    s.setItem(handOrderKey("g1", "p1"), "{not json");
    expect(loadHandOrder("g1", "p1", s)).toBeNull();
    s.setItem(handOrderKey("g1", "p1"), JSON.stringify({ a: 1 }));
    expect(loadHandOrder("g1", "p1", s)).toBeNull();
    s.setItem(handOrderKey("g1", "p1"), JSON.stringify(["a", 2]));
    expect(loadHandOrder("g1", "p1", s)).toBeNull();
  });

  it("storage that throws is swallowed: null on read, false on write", () => {
    expect(loadHandOrder("g1", "p1", throwing)).toBeNull();
    expect(saveHandOrder("g1", "p1", ["a"], throwing)).toBe(false);
  });

  it("no storage at all is the same as nothing saved", () => {
    expect(loadHandOrder("g1", "p1", null)).toBeNull();
    expect(saveHandOrder("g1", "p1", ["a"], null)).toBe(false);
  });

  it.runIf(typeof (globalThis as { localStorage?: unknown }).localStorage === "undefined")(
    "the default storage, absent in this node environment, reads as nothing saved",
    () => {
      expect(loadHandOrder("g1", "p1")).toBeNull();
      expect(saveHandOrder("g1", "p1", ["a"])).toBe(false);
    },
  );
});
