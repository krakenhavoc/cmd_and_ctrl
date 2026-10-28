import { describe, expect, it } from "vitest";
import type { ActivatedAbilityView, CardView, GameView, LogEvent } from "../protocol";
import { createVoiceCueTracker, isUltimate } from "./voiceCues";

const ab = (label: string, loyalty_cost: number, extra: Partial<ActivatedAbilityView> = {}) =>
  ({ index: 0, label, loyalty_cost, ...extra }) as ActivatedAbilityView;

const walker = (abilities: ActivatedAbilityView[] = []) =>
  ({
    instance_id: "w1",
    name: "Test Walker",
    type_line: "Legendary Planeswalker — Test",
    activated_abilities: abilities,
  }) as CardView;

const ABILITIES = [ab("plus", 1), ab("minus", -2), ab("ult", -7)];

function view(seq: number, log: LogEvent[], bf: CardView[] = [], gy: CardView[] = []): GameView {
  return {
    seats: [{ graveyard: { cards: gy } }],
    battlefield: { cards: bf },
    log: log.map((e, i) => ({ ...e, seq: e.seq || seq - log.length + i + 1 })),
  } as unknown as GameView;
}
const zone = (seq: number, old_zone: string, new_zone: string): LogEvent =>
  ({ seq, kind: "zone", card_id: "w1", old_zone, new_zone, seat: 0, text: "" }) as LogEvent;
const activate = (seq: number, label: string): LogEvent =>
  ({ seq, kind: "activate", card_id: "w1", label, seat: 0, text: "" }) as LogEvent;

describe("isUltimate", () => {
  it("is the strictly largest minus cost", () => {
    expect(isUltimate(ABILITIES, "ult")).toBe(true);
    expect(isUltimate(ABILITIES, "minus")).toBe(false);
    expect(isUltimate(ABILITIES, "plus")).toBe(false);
  });
  it("is not a lone minus, a tie, a -X, or an unknown label", () => {
    expect(isUltimate([ab("plus", 1), ab("m", -3)], "m")).toBe(true);
    expect(isUltimate([ab("only", -3)], "only")).toBe(true);
    expect(isUltimate([ab("a", -3), ab("b", -3)], "a")).toBe(false);
    expect(isUltimate([ab("x", 0, { loyalty_cost_x: true }), ab("m", -2)], "x")).toBe(false);
    expect(isUltimate(ABILITIES, "nope")).toBe(false);
    expect(isUltimate(ABILITIES, undefined)).toBe(false);
  });
});

describe("createVoiceCueTracker", () => {
  it("baselines the first frame silently", () => {
    const t = createVoiceCueTracker();
    expect(t.observe(view(5, [zone(5, "hand", "battlefield")], [walker()]))).toEqual([]);
  });

  it("cues enter and leave for a planeswalker, once per entry", () => {
    const t = createVoiceCueTracker();
    t.observe(view(1, []));
    const enter = view(2, [zone(2, "stack", "battlefield")], [walker()]);
    expect(t.observe(enter)).toEqual([{ name: "Test Walker", moment: "enter" }]);
    expect(t.observe(enter)).toEqual([]);
    const leave = view(
      3,
      [zone(2, "stack", "battlefield"), zone(3, "battlefield", "graveyard")],
      [],
      [walker()],
    );
    expect(t.observe(leave)).toEqual([{ name: "Test Walker", moment: "leave" }]);
  });

  it("ignores non-walkers and tokens", () => {
    const t = createVoiceCueTracker();
    t.observe(view(1, []));
    const bear = { ...walker(), type_line: "Creature — Bear" } as CardView;
    const token = { ...walker(), is_token: true } as CardView;
    expect(t.observe(view(2, [zone(2, "hand", "battlefield")], [bear]))).toEqual([]);
    expect(t.observe(view(3, [zone(3, "hand", "battlefield")], [token]))).toEqual([]);
  });

  it("cues ultimate even when the walker died on the activation", () => {
    const t = createVoiceCueTracker();
    t.observe(view(1, [], [walker(ABILITIES)]));
    const after = view(4, [activate(4, "ult")], [], [walker()]);
    expect(t.observe(after)).toEqual([{ name: "Test Walker", moment: "ultimate" }]);
  });

  it("does not cue a non-ultimate activation, and a rewind cues nothing", () => {
    const t = createVoiceCueTracker();
    t.observe(view(5, [], [walker(ABILITIES)]));
    expect(t.observe(view(6, [activate(6, "plus")], [walker(ABILITIES)]))).toEqual([]);
    expect(t.observe(view(4, [activate(4, "ult")], [walker(ABILITIES)]))).toEqual([]);
    expect(t.observe(view(5, [activate(5, "ult")], [walker(ABILITIES)]))).toEqual([
      { name: "Test Walker", moment: "ultimate" },
    ]);
  });
});
