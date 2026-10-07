// autoAnswerPref.test.ts — ADR 0127 (#1961), the client half: the
// reconcile, the rule edits, "Remember this answer", the asked-by-hand
// lines and the notice.

import { describe, it, expect, vi } from "vitest";

import {
  AUTO_ANSWER_SETTINGS_BUDGET,
  askedByHandText,
  autoAnswerNotice,
  autoAnswerNoticeRequest,
  autoAnswersToSend,
  canRemember,
  newAutoAnswerEntries,
  newAutoAnswersPrefState,
  rememberedRule,
  ruleFor,
  withRule,
  withoutRule,
} from "./autoAnswerPref";
import { choiceRequest } from "./choiceDock";
import { L } from "./labels";
import type { GameView, LogEvent, PendingChoiceView } from "./protocol";
import {
  MAX_AUTO_ANSWERS,
  SYNCED_FIELDS,
  defaultSettings,
  normalizeAutoAnswers,
  type AutoAnswerRule,
} from "./settings";

const rule = (key: string, answer: "always" | "never" = "never"): AutoAnswerRule => ({
  key,
  card: "Rhystic Study",
  prompt: "Rhystic Study — pay {1}?",
  answer,
});

function view(
  rules: { key: string; answer: "always" | "never" }[] | undefined,
  extra: Partial<GameView> = {},
  me: Record<string, unknown> = {},
): GameView {
  return {
    seats: [
      { id: "me", seat: 0, auto_answers: rules, ...me },
      { id: "opp", seat: 1 },
    ],
    log: [],
    ...extra,
  } as unknown as GameView;
}

describe("the setting", () => {
  it("is synced and defaults to no rules", () => {
    expect(SYNCED_FIELDS.gameplay.autoAnswers).toBe("synced");
    expect(defaultSettings().gameplay.autoAnswers).toEqual([]);
  });

  it("keeps only well-formed rules, one per key, at most 100", () => {
    const raw = [
      rule("a"),
      { key: "a", card: "dup", prompt: "", answer: "always" },
      { key: "", answer: "never" },
      { key: "b", answer: "sometimes" },
      { key: "k".repeat(257), answer: "never" },
      "junk",
      { key: "c", answer: "always" },
    ];
    expect(normalizeAutoAnswers(raw).map((r) => r.key)).toEqual(["a", "c"]);
    expect(normalizeAutoAnswers("nope")).toEqual([]);
    const many = Array.from({ length: 150 }, (_, i) => rule(`k${i}`));
    expect(normalizeAutoAnswers(many)).toHaveLength(MAX_AUTO_ANSWERS);
  });
});

describe("autoAnswersToSend", () => {
  it("sends nothing when the seat already agrees", () => {
    const st = newAutoAnswersPrefState();
    expect(autoAnswersToSend(st, view(undefined), "me", [])).toBeNull();
    expect(
      autoAnswersToSend(st, view([{ key: "a", answer: "never" }]), "me", [rule("a")]),
    ).toBeNull();
  });

  it("sends the whole list, sorted, once per disagreement", () => {
    const st = newAutoAnswersPrefState();
    const want = [rule("b", "always"), rule("a")];
    expect(autoAnswersToSend(st, view(undefined), "me", want)).toEqual([
      { key: "a", answer: "never" },
      { key: "b", answer: "always" },
    ]);
    // The same stale frame: not re-sent.
    expect(autoAnswersToSend(st, view(undefined), "me", want)).toBeNull();
    // Caught up, then the server lost it (an older snapshot): again.
    autoAnswersToSend(
      st,
      view([
        { key: "a", answer: "never" },
        { key: "b", answer: "always" },
      ]),
      "me",
      want,
    );
    expect(autoAnswersToSend(st, view(undefined), "me", want)).not.toBeNull();
    // Clearing every rule sends an empty list.
    const st2 = newAutoAnswersPrefState();
    expect(autoAnswersToSend(st2, view([{ key: "a", answer: "never" }]), "me", [])).toEqual([]);
  });

  it("never sends for a spectator, an eliminated seat or a bot", () => {
    const st = newAutoAnswersPrefState();
    expect(autoAnswersToSend(st, view(undefined), null, [rule("a")])).toBeNull();
    expect(autoAnswersToSend(st, null, "me", [rule("a")])).toBeNull();
    expect(
      autoAnswersToSend(st, view(undefined, {}, { eliminated: true }), "me", [rule("a")]),
    ).toBeNull();
    expect(
      autoAnswersToSend(st, view(undefined, {}, { is_bot: true }), "me", [rule("a")]),
    ).toBeNull();
  });
});

describe("rule edits", () => {
  it("sets, replaces and removes a rule", () => {
    let list = withRule([], rule("a")) as AutoAnswerRule[];
    list = withRule(list, rule("a", "always")) as AutoAnswerRule[];
    expect(list).toHaveLength(1);
    expect(ruleFor(list, "a")?.answer).toBe("always");
    expect(withoutRule(list, "a")).toEqual([]);
    expect(ruleFor(list, undefined)).toBeUndefined();
  });

  it("refuses a 101st rule, and replacing one of 100 is fine", () => {
    const full = Array.from({ length: MAX_AUTO_ANSWERS }, (_, i) => rule(`k${i}`));
    expect(withRule(full, rule("new"))).toBe("too_many");
    expect(Array.isArray(withRule(full, rule("k3", "always")))).toBe(true);
  });

  it("refuses a rule that would overfill the account's settings", () => {
    const s = defaultSettings();
    const big = (i: number): AutoAnswerRule => ({
      key: `${"k".repeat(250)}${i}`,
      card: "c".repeat(120),
      prompt: "p".repeat(120),
      answer: "never",
    });
    let list: AutoAnswerRule[] = [];
    let refused = false;
    for (let i = 0; i < MAX_AUTO_ANSWERS; i++) {
      const next = withRule(list, big(i), { ...s, gameplay: { ...s.gameplay, autoAnswers: list } });
      if (next === "too_large") {
        refused = true;
        break;
      }
      list = next as AutoAnswerRule[];
    }
    expect(refused).toBe(true);
    expect(JSON.stringify(list).length).toBeLessThan(AUTO_ANSWER_SETTINGS_BUDGET);
  });
});

const prompt = (over: Partial<PendingChoiceView>): PendingChoiceView =>
  ({
    id: "c1",
    kind: "pay_unless",
    chooser: "me",
    from_player: "me",
    count: 1,
    pay_cost: "{1}",
    reason: "Rhystic Study — pay {1}?",
    auto_answer_key: "rs|triggered|row|pay_unless#1",
    auto_answer_card: "Rhystic Study",
    auto_answer_prompt: "Rhystic Study — pay {1}?",
    ...over,
  }) as PendingChoiceView;

const handlers = () => ({ onAnswer: vi.fn(), onCoin: vi.fn(), onOption: vi.fn(), onLoop: vi.fn() });

describe("Remember this answer", () => {
  it("maps the button pressed to Always or Never", () => {
    expect(rememberedRule(prompt({}), true)?.answer).toBe("always");
    expect(rememberedRule(prompt({}), false)).toEqual({
      key: "rs|triggered|row|pay_unless#1",
      card: "Rhystic Study",
      prompt: "Rhystic Study — pay {1}?",
      answer: "never",
    });
    expect(rememberedRule(prompt({ auto_answer_key: undefined }), true)).toBeNull();
  });

  it("is a toggle on a covered prompt, and absent with no key", () => {
    const onToggle = vi.fn();
    const req = choiceRequest(
      prompt({}),
      { sourceName: "Rhystic Study", remember: { on: false, onToggle } },
      handlers(),
    );
    const toggle = req.row?.find((a) => a.id === "remember");
    expect(toggle?.label).toBe(L.rememberThisAnswer);
    expect(toggle?.pressed).toBe(false);
    expect(toggle?.title).toBe("Remember for Rhystic Study — pay {1}?");
    toggle?.onPress();
    expect(onToggle).toHaveBeenCalled();
    // The answer buttons keep their names.
    expect(req.primary?.label).toBe("Pay {1}");

    const none = choiceRequest(
      prompt({ auto_answer_key: undefined }),
      { sourceName: "x", remember: { on: false, onToggle } },
      handlers(),
    );
    expect(none.row ?? []).toHaveLength(0);
    expect(canRemember(prompt({ kind: "commander_return" }))).toBe(false);
    expect(canRemember(prompt({ kind: "trigger_prompt" }))).toBe(true);
    expect(canRemember(prompt({ kind: "confirm" }))).toBe(true);
  });

  it("says why a prompt with a rule is asked anyway", () => {
    expect(askedByHandText(prompt({ asked_by_hand: "no_mana" }))).toMatch(/not enough mana/);
    expect(askedByHandText(prompt({ asked_by_hand: "empty_library" }))).toMatch(/library is empty/);
    expect(askedByHandText(prompt({ asked_by_hand: "loop" }))).toMatch(/loop/);
    expect(askedByHandText(prompt({ asked_by_hand: "undone" }))).toMatch(/undid/);
    expect(askedByHandText(prompt({}))).toBeNull();
    const req = choiceRequest(
      prompt({ asked_by_hand: "no_mana" }),
      { sourceName: "Rhystic Study" },
      handlers(),
    );
    expect(req.hint).toMatch(/^Always pay: not enough mana/);
  });
});

describe("the notice", () => {
  const line = (over: Partial<LogEvent>): LogEvent =>
    ({
      seq: 40,
      kind: "auto_answer",
      seat: 0,
      call: "pay",
      label: "{1}",
      card_id: "study",
      auto_answer_key: "k",
      text: "Me paid {1} for Rhystic Study (automatic)",
      ...over,
    }) as LogEvent;

  it("finds the viewer's own new lines", () => {
    const v = view(undefined, {
      log: [
        line({ seq: 10 }),
        line({ seq: 40 }),
        line({ seq: 41, seat: 1, auto_answer_key: undefined }),
      ],
    });
    expect(newAutoAnswerEntries(v, "me", 10).map((e) => e.seq)).toEqual([40]);
    expect(newAutoAnswerEntries(v, null, 0)).toEqual([]);
  });

  it("words the answer, and Undo follows the room's stamp", () => {
    const v = view(undefined, {}, { undo_auto_answer: 40 });
    const n = autoAnswerNotice(line({}), "Rhystic Study", v, "me");
    expect(n).toEqual({
      seq: 40,
      text: "Rhystic Study: paid {1} for you",
      key: "k",
      undoable: true,
    });
    expect(
      autoAnswerNotice(line({ call: "dont_pay", label: undefined }), "Rhystic Study", v, "me").text,
    ).toBe("Rhystic Study: didn't pay for you");
    expect(autoAnswerNotice(line({ call: "yes" }), "", v, "me").text).toBe(
      "A card: answered Yes for you",
    );
    // Someone has acted since: the stamp is gone.
    expect(autoAnswerNotice(line({}), "Rhystic Study", view(undefined), "me").undoable).toBe(false);
  });

  it("is a step request with Undo and Ask me next time", () => {
    const undo = vi.fn();
    const ask = vi.fn();
    const req = autoAnswerNoticeRequest(
      { seq: 40, text: "Rhystic Study: paid {1} for you", key: "k", undoable: false },
      undo,
      ask,
    );
    expect(req.rank).toBe("step");
    expect(req.label).toBe(L.automaticAnswer);
    const [u, a] = req.row ?? [];
    expect(u.label).toBe(L.undoAutomaticAnswer);
    expect(u.disabled).toBe(true);
    expect(u.title).toBe("Someone has acted since");
    expect(a.label).toBe(L.askMeNextTime);
    a.onPress();
    expect(ask).toHaveBeenCalled();
    // No key binding: a status takes no keys.
    expect(req.primary).toBeUndefined();
    expect(u.keyShortcuts).toBeUndefined();
  });
});
