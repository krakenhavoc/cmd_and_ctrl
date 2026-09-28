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

describe("isToken", () => {
  it("reads the Token supertype off the type line, as the engine does", () => {
    expect(isToken({ type_line: "Token Creature — Soldier" })).toBe(true);
    expect(isToken({ type_line: "Token Artifact — Treasure" })).toBe(true);
    expect(isToken({ type_line: "Creature — Soldier" })).toBe(false);
    expect(isToken({ type_line: undefined })).toBe(false);
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

  it("never groups two copies of a printed card", () => {
    const bear = (id: string): CardView => ({
      instance_id: id,
      name: "Grizzly Bears",
      owner: "me",
      controller: "me",
      type_line: "Creature — Bear",
      power: 2,
      toughness: 2,
    });
    expect(groupTokens([bear("b1"), bear("b2")]).size).toBe(0);
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
