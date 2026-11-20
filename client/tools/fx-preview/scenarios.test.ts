// The fx preview harness keeps a scenario for every cue (#4868). A new
// EffectKind with no scenario fails here, and so does a scenario whose log
// would not plan the cue it says it shows, so a cue issue's "the harness
// has a scenario for it" is checked, not just remembered.

import { describe, expect, it } from "vitest";
import { isExtortLabel } from "../../src/lib/extort";
import { cardView } from "./table";
import {
  EFFECT_BASE_MS,
  deckedSeats,
  planEffects,
  type EffectKind,
} from "../../src/lib/tableEffects";
import { SCENARIOS, type Scenario } from "./scenarios";
import { stamp, zoneOf, type FxCard } from "./table";

const KINDS = Object.keys(EFFECT_BASE_MS) as EffectKind[];
const isKind = (k: string): k is EffectKind => (KINDS as string[]).includes(k);

function plan(s: Scenario) {
  const log = stamp(s.log);
  const find = (cards: FxCard[], id: string) => cards.find((c) => c.id === id);
  const typeOf = (id: string) => {
    const c = find(s.before, id) ?? find(s.after, id);
    return c ? cardView(c).type_line : undefined;
  };
  return {
    log,
    plan: planEffects(log, log, {
      isPlaneswalker: (id) => /\bplaneswalker\b/i.test(typeOf(id) ?? ""),
      combatMotion: false,
      isCommander: (id) => {
        const c = find(s.after, id);
        return !!c && zoneOf(c) === "command" && c.commander === true;
      },
      typeLineOf: typeOf,
    }),
  };
}

describe("the fx preview scenarios", () => {
  it("cover every table-effect cue", () => {
    const shown = new Set(SCENARIOS.flatMap((s) => s.shows));
    const missing = KINDS.filter((k) => !shown.has(k));
    expect(missing, "add a scenario to tools/fx-preview/scenarios.ts").toEqual([]);
    expect(shown.has("board_wipe")).toBe(true);
  });

  it("have unique ids, and every card is on the table once", () => {
    const ids = SCENARIOS.map((s) => s.id);
    expect(new Set(ids).size).toBe(ids.length);
    for (const s of SCENARIOS) {
      for (const cards of [s.before, s.after]) {
        const cardIDs = cards.map((c) => c.id);
        expect(new Set(cardIDs).size, s.id).toBe(cardIDs.length);
      }
    }
  });

  it("name only cards their tables hold", () => {
    for (const s of SCENARIOS) {
      const known = new Set([...s.before, ...s.after].map((c) => c.id));
      for (const e of s.log) {
        for (const id of [e.card_id, e.target, ...(e.card_ids ?? [])]) {
          if (id) expect(known.has(id), `${s.id}: ${id}`).toBe(true);
        }
        if (e.stack_item_id) {
          const items = s.before.filter((c) => zoneOf(c) === "stack").map((c) => c.item ?? c.id);
          expect(items, s.id).toContain(e.stack_item_id);
        }
      }
    }
  });

  for (const s of SCENARIOS) {
    it(`${s.id}: its log plays ${s.shows.join(", ")}`, () => {
      const { log, plan: p } = plan(s);
      const kinds = new Set(p.cues.map((c) => c.kind));
      for (const want of s.shows) {
        if (isKind(want)) expect(kinds.has(want), want).toBe(true);
        else if (want === "board_wipe") expect(p.wipes.length).toBeGreaterThan(0);
        else if (want === "decked_out") expect(deckedSeats(log).length).toBeGreaterThan(0);
        else if (want === "eliminated")
          expect(log.some((e) => e.kind === "eliminated" && e.cause !== "empty_draw")).toBe(true);
        else if (want === "extort")
          expect(log.some((e) => e.kind === "resolve" && isExtortLabel(e.label))).toBe(true);
        else if (want === "dice_roll") expect(log.some((e) => e.kind === "roll")).toBe(true);
        else if (want === "coin_flip") expect(log.some((e) => e.kind === "flip")).toBe(true);
      }
    });
  }
});
