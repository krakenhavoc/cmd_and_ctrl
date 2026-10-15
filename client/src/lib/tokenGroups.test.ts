// #1724 — identical tokens group on the board. The pure rules: what
// the group key is, how a group splits into its untapped and tapped
// cards, and what the member list's "select N" and "all" reach for.
// The markup is in tokenGroups.render.test.ts.

import { describe, expect, it } from "vitest";

import type { CardView } from "./protocol";
import {
  bulkCandidates,
  groupTokens,
  isToken,
  liveSelection,
  ptCounterDelta,
  rowBadges,
  rowEntries,
  cardGroupKey,
  selectFirstN,
  tokenGroupKey,
} from "./tokenGroups";

let n = 0;
function soldier(extra: Partial<CardView> = {}): CardView {
  n++;
  return {
    instance_id: `s${n}`,
    name: "Soldier",
    owner: "me",
    controller: "me",
    type_line: "Token Creature — Soldier",
    colors: ["W"],
    power: 1,
    toughness: 1,
    battle_x: n,
    ...extra,
  };
}

// #3046: identical nontoken permanents pile, but only when nothing the
// player can see tells them apart.
let b = 0;
function bear(extra: Partial<CardView> = {}): CardView {
  b++;
  return {
    instance_id: `b${b}`,
    name: "Grizzly Bears",
    owner: "me",
    controller: "me",
    type_line: "Creature — Bear",
    colors: ["G"],
    power: 2,
    toughness: 2,
    scryfall_id: "bears-1",
    battle_x: 100 + b,
    ...extra,
  };
}

describe("cardGroupKey (#3046)", () => {
  it("gives identical copies the same key, whatever their ID, position or tapped state", () => {
    const k = cardGroupKey(bear());
    expect(k).not.toBeNull();
    expect(cardGroupKey(bear({ battle_x: 7, battle_y: 3 }))).toBe(k);
    expect(cardGroupKey(bear({ tapped: true }))).toBe(k);
  });

  it("is null for a token, a commander, a face-down, phased-out or hidden card and a Ring-bearer", () => {
    expect(cardGroupKey(soldier())).toBeNull();
    expect(cardGroupKey(bear({ is_token: true, token_copy: true }))).toBeNull();
    expect(cardGroupKey(bear({ is_commander: true }))).toBeNull();
    expect(cardGroupKey(bear({ face_down: true }))).toBeNull();
    expect(cardGroupKey(bear({ phased_out: true }))).toBeNull();
    expect(cardGroupKey(bear({ known_by_you: false }))).toBeNull();
    expect(cardGroupKey(bear({ ring_bearer: true }))).toBeNull();
  });

  it("splits on everything the player sees: counters, damage, sickness, P/T, combat, goad, art", () => {
    const k = cardGroupKey(bear());
    const variants: Partial<CardView>[] = [
      { counters: { "+1/+1": 1 }, power: 3, toughness: 3 },
      { counters: { stun: 1 } },
      { damage_marked: 1 },
      { summoning_sick: true },
      { power: 4, toughness: 4 },
      { attacking_target: "bob" },
      { blocking_target: "ogre" },
      { goaded_by: "bob" },
      { scryfall_id: "bears-2" },
      { controller: "bob" },
      { monstrous: true },
      { chosen_name: "Lightning Bolt" },
    ];
    for (const extra of variants) {
      expect(cardGroupKey(bear(extra)), JSON.stringify(extra)).not.toBe(k);
    }
  });

  it("does not care about field order", () => {
    const a = bear({ instance_id: "x", battle_x: 0 });
    const reordered = Object.fromEntries(Object.entries(a).reverse()) as unknown as CardView;
    expect(cardGroupKey(reordered)).toBe(cardGroupKey(a));
  });

  it("ignores per-instance ability detail: cost options that name the other copy", () => {
    const ability = (other: string, unmet: boolean) => ({
      index: 0,
      ref: "own:0",
      label: "Tap another untapped creature you control: draw a card",
      tap_others_options: { cards: [other] },
      condition_unmet: unmet,
    });
    const x = bear({ instance_id: "x", activated_abilities: [ability("y", true)] as never });
    const y = bear({ instance_id: "y", activated_abilities: [ability("x", false)] as never });
    expect(cardGroupKey(x)).toBe(cardGroupKey(y));
    // ...but a spent exhaust ability marks a copy apart.
    const spent = bear({
      instance_id: "z",
      activated_abilities: [{ ...ability("x", false), exhausted: true }] as never,
    });
    expect(cardGroupKey(spent)).not.toBe(cardGroupKey(x));
  });
});

describe("rowEntries for printed cards (#3046)", () => {
  it("five identical untapped creatures draw as one group; a tapped one splits off", () => {
    const five = [bear(), bear(), bear(), bear(), bear()];
    const [only, ...rest] = rowEntries(five);
    expect(rest).toEqual([]);
    expect(only.kind === "group" && only.members).toEqual(five);
    const tapped = { ...five[2], tapped: true };
    const split = rowEntries([five[0], five[1], tapped, five[3], five[4]]);
    expect(split.map((e) => e.kind)).toEqual(["group", "card"]);
    expect(split[1].kind === "card" && split[1].card).toBe(tapped);
  });

  it("draws one card per highlight, so a ring never stands for a copy drawn elsewhere", () => {
    const [x, y, z, w] = [bear(), bear(), bear(), bear()];
    const ready = new Set([y.instance_id, w.instance_id]);
    const highlight = (c: CardView) => (ready.has(c.instance_id) ? "ready" : "");
    const entries = rowEntries([x, y, z, w], {}, { highlight });
    expect(entries.map((e) => e.kind === "group" && e.members)).toEqual([
      [x, z],
      [y, w],
    ]);
    // One group behind both cards: the member list still shows all four.
    expect(new Set(entries.map((e) => e.kind === "group" && e.groupKey)).size).toBe(1);
    expect(new Set(entries.map((e) => e.key)).size).toBe(2);
  });

  it("leaves token groups whole whatever the highlight (#1724)", () => {
    const [s1, s2] = [soldier(), soldier()];
    const highlight = (c: CardView) => (c === s1 ? "ready" : "");
    expect(rowEntries([s1, s2], {}, { highlight }).map((e) => e.kind)).toEqual(["group"]);
  });

  it("never mixes a token copy into a pile of the card it copies", () => {
    const card = bear();
    const copy = bear({ is_token: true, token_copy: true });
    expect(rowEntries([card, copy]).map((e) => e.kind)).toEqual(["card", "card"]);
  });
});

describe("isToken", () => {
  it("reads the Token supertype off the type line, as the engine does", () => {
    expect(isToken({ type_line: "Token Creature — Soldier" })).toBe(true);
    expect(isToken({ type_line: "Token Artifact — Treasure" })).toBe(true);
    expect(isToken({ type_line: "Creature — Soldier" })).toBe(false);
    expect(isToken({ type_line: undefined })).toBe(false);
  });

  it("ADR 0078: prefers the wire's own is_token when the server sends it", () => {
    expect(isToken({ type_line: "Creature — Soldier", is_token: true })).toBe(true);
    expect(isToken({ type_line: "Token Creature — Soldier", is_token: false })).toBe(false);
    // Undefined is_token (a pre-ADR-0078 view, or a fixture that never
    // set it) falls back to the type-line test unchanged.
    expect(isToken({ type_line: "Token Creature — Soldier", is_token: undefined })).toBe(true);
  });
});

describe("ptCounterDelta", () => {
  it("adds every P/T counter kind and ignores the rest", () => {
    expect(ptCounterDelta({ "+1/+1": 2, "-1/-1": 1, "+1/+0": 1, loyalty: 3 })).toEqual({
      power: 2,
      toughness: 1,
    });
    expect(ptCounterDelta(undefined)).toEqual({ power: 0, toughness: 0 });
  });
});

describe("tokenGroupKey", () => {
  it("gives identical tokens the same key", () => {
    expect(tokenGroupKey(soldier())).toBe(tokenGroupKey(soldier()));
  });

  it("separates different P/T, names, type lines and controllers", () => {
    const base = tokenGroupKey(soldier());
    expect(tokenGroupKey(soldier({ power: 2, toughness: 2 }))).not.toBe(base);
    expect(tokenGroupKey(soldier({ name: "Spirit" }))).not.toBe(base);
    expect(tokenGroupKey(soldier({ type_line: "Token Creature — Human Soldier" }))).not.toBe(base);
    expect(tokenGroupKey(soldier({ controller: "bob" }))).not.toBe(base);
  });

  it("never groups a nontoken, a face-down or a phased-out object", () => {
    expect(tokenGroupKey(soldier({ type_line: "Creature — Soldier" }))).toBeNull();
    expect(tokenGroupKey(soldier({ face_down: true }))).toBeNull();
    expect(tokenGroupKey(soldier({ phased_out: true }))).toBeNull();
  });

  it("never groups a Ring-bearer: its marker stays on a card of its own (ADR 0114)", () => {
    expect(tokenGroupKey(soldier({ ring_bearer: true }))).toBeNull();
  });

  it("keeps a token with +1/+1 counters in its group: the counters come back off the P/T", () => {
    expect(tokenGroupKey(soldier({ power: 3, toughness: 3, counters: { "+1/+1": 2 } }))).toBe(
      tokenGroupKey(soldier()),
    );
  });

  it("splits attackers by what they attack", () => {
    const home = tokenGroupKey(soldier());
    const atBob = tokenGroupKey(soldier({ attacking_target: "bob" }));
    expect(atBob).not.toBe(home);
    expect(tokenGroupKey(soldier({ attacking_target: "bob" }))).toBe(atBob);
    expect(tokenGroupKey(soldier({ attacking_target: "cat" }))).not.toBe(atBob);
  });
});

describe("groupTokens", () => {
  it("needs two members to make a group", () => {
    expect(groupTokens([soldier()]).size).toBe(0);
    expect(groupTokens([soldier(), soldier()]).size).toBe(1);
  });

  it("groups two identical copies of a printed card (#3046)", () => {
    expect(groupTokens([bear(), bear()]).size).toBe(1);
  });

  it("nontoken: false leaves printed cards alone, as the land strip needs", () => {
    expect(groupTokens([bear(), bear()], {}, { nontoken: false }).size).toBe(0);
    expect(groupTokens([soldier(), soldier()], {}, { nontoken: false }).size).toBe(1);
  });

  it("never piles a printed card that has anything attached", () => {
    const a = bear();
    const b = bear();
    const aura: CardView = { instance_id: "aura", name: "Pacifism", owner: "me", controller: "me" };
    expect(groupTokens([a, b], { [b.instance_id]: [aura] }).size).toBe(0);
  });

  it("keeps an equipped token in the group even though the Equipment changed its P/T", () => {
    const a = soldier();
    const b = soldier();
    const equipped = soldier({ power: 3, toughness: 1 });
    const sword: CardView = {
      instance_id: "sword",
      name: "Bonesplitter",
      owner: "me",
      controller: "me",
      type_line: "Artifact — Equipment",
      attached_to: { kind: "card", id: equipped.instance_id },
    };
    const groups = [...groupTokens([a, equipped, b], { [equipped.instance_id]: [sword] }).values()];
    expect(groups).toHaveLength(1);
    expect(groups[0].members.map((c) => c.instance_id)).toEqual([
      a.instance_id,
      equipped.instance_id,
      b.instance_id,
    ]);
  });

  it("does not fold an unequipped token with a different P/T into the group", () => {
    const groups = groupTokens([soldier(), soldier(), soldier({ power: 3, toughness: 1 })]);
    expect([...groups.values()].map((g) => g.members.length)).toEqual([2]);
  });
});

describe("rowEntries", () => {
  it("draws one card per state: untapped and tapped, each with its members", () => {
    const cards = [soldier(), soldier({ tapped: true }), soldier(), soldier({ tapped: true })];
    const entries = rowEntries(cards);
    expect(entries.map((e) => e.kind)).toEqual(["group", "group"]);
    const [u, t] = entries;
    if (u.kind !== "group" || t.kind !== "group") throw new Error("expected groups");
    expect(u.tapped).toBe(false);
    expect(u.members).toHaveLength(2);
    expect(t.tapped).toBe(true);
    expect(t.members).toHaveLength(2);
    expect(u.groupKey).toBe(t.groupKey);
  });

  it("draws ONE card when every member is in the same state", () => {
    const entries = rowEntries([soldier(), soldier(), soldier()]);
    expect(entries).toHaveLength(1);
    expect(entries[0].kind === "group" && entries[0].members.length).toBe(3);
  });

  it("draws a lone tapped member as a plain card beside the untapped group", () => {
    const tapped = soldier({ tapped: true });
    const entries = rowEntries([soldier(), soldier(), tapped]);
    expect(entries.map((e) => e.kind)).toEqual(["group", "card"]);
    expect(entries[1].kind === "card" && entries[1].card).toBe(tapped);
  });

  it("keeps the group where its first member stood, around other cards", () => {
    const bear: CardView = {
      instance_id: "bear",
      name: "Grizzly Bears",
      owner: "me",
      controller: "me",
      type_line: "Creature — Bear",
      battle_x: 0,
    };
    const s1 = soldier();
    const knight: CardView = { ...bear, instance_id: "knight", name: "Knight" };
    const s2 = soldier();
    const entries = rowEntries([bear, s1, knight, s2]);
    expect(entries.map((e) => e.key)).toEqual([
      "bear",
      expect.stringMatching(/^group:u:/),
      "knight",
    ]);
  });

  it("draws the plainest member, so one token's counters don't stand for the group", () => {
    const boosted = soldier({ power: 2, toughness: 2, counters: { "+1/+1": 1 } });
    const plain = soldier();
    const [e] = rowEntries([boosted, plain]);
    expect(e.kind === "group" && e.rep).toBe(plain);
  });
});

describe("the member list's selection", () => {
  it("attack: select N takes tokens that can attack now, in order, and skips the rest", () => {
    const sick = soldier({ summoning_sick: true });
    const ready1 = soldier();
    const tapped = soldier({ tapped: true });
    const ready2 = soldier();
    const already = soldier({ attacking_target: "bob" });
    const members = [sick, ready1, tapped, ready2, already];
    const blocker = (c: CardView) => (c.summoning_sick ? "sick" : c.tapped ? "tapped" : null);
    const cands = bulkCandidates(members, "attack", { attackBlocker: blocker });
    expect(cands).toEqual([ready1, ready2]);
    expect(selectFirstN(cands, 1)).toEqual([ready1.instance_id]);
    expect(selectFirstN(cands, 5)).toEqual([ready1.instance_id, ready2.instance_id]);
  });

  it("target: only the legal members", () => {
    const a = soldier();
    const b = soldier();
    expect(bulkCandidates([a, b], "target", { isLegalTarget: (c) => c === b })).toEqual([b]);
  });

  it("tap: untapped first", () => {
    const t = soldier({ tapped: true });
    const u = soldier();
    expect(bulkCandidates([t, u], "tap")).toEqual([u, t]);
  });

  it("drops a member that left while the list was open", () => {
    const a = soldier();
    const sel = [a.instance_id, "gone"];
    expect(liveSelection(sel, [a])).toEqual([a.instance_id]);
    const same = [a.instance_id];
    expect(liveSelection(same, [a])).toBe(same);
  });
});

describe("rowBadges", () => {
  it("shows counters, attachments, damage and summoning sickness", () => {
    const c = soldier({
      counters: { "+1/+1": 2, shield: 1 },
      damage_marked: 1,
      summoning_sick: true,
    });
    const sword = { ...soldier(), name: "Bonesplitter" };
    expect(rowBadges(c, [sword]).map((b) => b.text)).toEqual([
      "+1/+1 ×2",
      "shield",
      "Bonesplitter",
      "1 damage",
      "summoning sick",
    ]);
  });
});
