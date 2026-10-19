// choiceDock.test.ts — ADR 0111 Delivery PR 5: which pending choices
// are answered inline in the dock, and what the dock request says.
// The rendering and the answers are pinned through the Game route in
// choiceDock.render.test.ts; this is the pure half.

import { describe, it, expect, vi } from "vitest";

import {
  anyPlayerEachBadge,
  anyPlayerEachHeader,
  anyPlayerEachPayers,
  anyPlayerEachPrice,
  anyPlayerEachSummary,
  anyPlayerEachWarning,
  anyPlayerOfferLine,
  anyPlayerOfferRequest,
  anyPlayerWaiting,
  assistRequest,
  assistWaiting,
  ballotNote,
  choiceRequest,
  councilVoteRequest,
  gameOverRequest,
  INLINE_LABEL_MAX,
  INLINE_OPTION_MAX,
  isAnyPlayerEachLeg,
  isInlineChoice,
  voteRequest,
  mayEndVote,
} from "./choiceDock";
import { dockKeyAction } from "./dock";
import type { PendingChoiceView } from "./protocol";

const choice = (over: Partial<PendingChoiceView>): PendingChoiceView =>
  ({
    id: "c1",
    kind: "trigger_prompt",
    chooser: "me",
    from_player: "me",
    count: 0,
    ...over,
  }) as PendingChoiceView;

const handlers = () => ({
  onAnswer: vi.fn(),
  onCoin: vi.fn(),
  onOption: vi.fn(),
  onLoop: vi.fn(),
});

describe("compound mana and life payments", () => {
  it("shows and charges the mandatory life beside mana", () => {
    const h = handlers();
    const r = choiceRequest(
      choice({ kind: "pay_unless", pay_cost: "{2}", life_cost: 2 }),
      { sourceName: "Gisa", life: 40 },
      h,
    );
    expect(r.primary?.label).toBe("Pay {2} and 2 life");
    expect(r.hint).toContain("and 2 life");
    r.primary!.onPress();
    expect(h.onAnswer).toHaveBeenCalledWith(true);
  });

  it("disables a life payment that is unavailable", () => {
    for (const [life, locked] of [
      [1, false],
      [40, true],
    ] as const) {
      const r = choiceRequest(
        choice({ kind: "pay_unless", pay_cost: "{2}", life_cost: 2, life_cost_unpayable: locked }),
        { sourceName: "Gisa", life },
        handlers(),
      );
      expect(r.primary?.disabled).toBe(true);
      expect(r.secondary?.at(-1)?.label).toBe("Don't pay");
    }
  });

  it("reserves fixed life before offering Phyrexian symbol payments", () => {
    const h = handlers();
    const r = choiceRequest(
      choice({
        kind: "pay_unless",
        pay_cost: "{B}{B}",
        life_cost: 2,
        phyrexian_symbols: 2,
        phyrexian_granted: 2,
      }),
      { sourceName: "Gisa", life: 5 },
      h,
    );
    expect(r.secondary?.map((a) => a.label)).toEqual(["Pay with 4 life", "Don't pay"]);
    r.secondary![0].onPress();
    expect(h.onAnswer).toHaveBeenCalledWith(true, 1);
  });
});

describe("isInlineChoice", () => {
  it("takes the yes/no family, coin, loop, mana and colour", () => {
    for (const kind of [
      "trigger_prompt",
      "optional_replacement",
      "commander_return",
      "confirm",
      "may_cast",
      "entry_pay_life",
      "entry_riot",
      "pay_unless",
      "coin_call",
      "loop_shortcut",
      "mana_pick",
      "choose_color",
    ]) {
      expect(isInlineChoice(choice({ kind })), kind).toBe(true);
    }
  });

  it("leaves every sheet kind to the modal", () => {
    for (const kind of [
      "scry",
      "surveil",
      "look_at_top",
      "put_in_library",
      "mode_pick",
      "trigger_order",
      "replacement_order",
      "damage_assignment",
      "choose_creature_type",
      "choose_card_name",
      "search_library",
      "sacrifice_choice",
      "choose_cards",
      "discard_from_hand",
      "revealed_hand_pick",
    ]) {
      expect(isInlineChoice(choice({ kind })), kind).toBe(false);
    }
    expect(isInlineChoice(null)).toBe(false);
  });

  it("pay_unless with card picks or a tap list is a sheet", () => {
    expect(
      isInlineChoice(
        choice({ kind: "pay_unless", pay_cards: { action: "discard", count: 1, options: [] } }),
      ),
    ).toBe(false);
    expect(isInlineChoice(choice({ kind: "pay_unless", tap_cost: { key: "waterbend" } }))).toBe(
      false,
    );
  });

  it("an option pick is inline only when short: six options, no cards, short labels", () => {
    const opts = (n: number, label = "Draw a card") => Array.from({ length: n }, () => ({ label }));
    for (const kind of ["option_pick", "entry_controller", "entry_read_ahead"]) {
      expect(isInlineChoice(choice({ kind, pick_options: opts(INLINE_OPTION_MAX) }))).toBe(true);
      expect(isInlineChoice(choice({ kind, pick_options: opts(INLINE_OPTION_MAX + 1) }))).toBe(
        false,
      );
      expect(isInlineChoice(choice({ kind, pick_options: [] }))).toBe(false);
      expect(
        isInlineChoice(choice({ kind, pick_options: opts(2, "x".repeat(INLINE_LABEL_MAX)) })),
      ).toBe(true);
      expect(
        isInlineChoice(choice({ kind, pick_options: opts(2, "x".repeat(INLINE_LABEL_MAX + 1)) })),
      ).toBe(false);
      expect(
        isInlineChoice(
          choice({
            kind,
            pick_options: [{ label: "Pile 1", cards: [{ instance_id: "a", name: "A" }] }],
          } as Partial<PendingChoiceView>),
        ),
      ).toBe(false);
    }
  });
});

describe("choiceRequest", () => {
  it("read ahead offers one button per chapter and answers with its index (#2123)", () => {
    const h = handlers();
    const r = choiceRequest(
      choice({
        kind: "entry_read_ahead",
        reason: "Read ahead — choose the chapter The Cruelty of Gix starts on",
        entry_keyword: "read ahead",
        pick_options: [{ label: "Chapter I" }, { label: "Chapter II" }, { label: "Chapter III" }],
      }),
      { sourceName: "The Cruelty of Gix" },
      h,
    );
    expect(r.label).toBe("Read ahead — choose the chapter The Cruelty of Gix starts on");
    expect(r.primary).toBeNull();
    expect(r.row?.map((a) => a.label)).toEqual(["Chapter I", "Chapter II", "Chapter III"]);
    r.row?.[2].onPress();
    expect(h.onOption).toHaveBeenCalledWith(2);
  });

  it("shows the mana an option costs when its label does not (#2854)", () => {
    const h = handlers();
    const r = choiceRequest(
      choice({
        kind: "option_pick",
        reason: "Winter's Chill — pay {1} or {2} for Bear?",
        pick_options: [
          { label: "Pay nothing: destroy Bear at end of combat" },
          { label: "Pay {1}: no combat damage", mana_cost: "{1}" },
          { label: "Keep fighting", mana_cost: "{2}" },
        ],
      }),
      { sourceName: "Winter's Chill" },
      h,
    );
    expect(r.row?.map((a) => a.label)).toEqual([
      "Pay nothing: destroy Bear at end of combat",
      "Pay {1}: no combat damage",
      "Keep fighting ({2})",
    ]);
    r.row?.[1].onPress();
    expect(h.onOption).toHaveBeenCalledWith(1);
  });

  it("is a choice, named by the reason, focusing its dialog", () => {
    const r = choiceRequest(
      choice({ reason: "Mulldrifter — draw two cards?" }),
      { sourceName: "x" },
      handlers(),
    );
    expect(r.rank).toBe("choice");
    expect(r.label).toBe("Mulldrifter — draw two cards?");
    expect(r.question).toBe("Mulldrifter — draw two cards?");
    expect(r.focus).toBe("dialog");
  });

  it("falls back to the modal's headings", () => {
    const ctx = { sourceName: "Smothering Tithe" };
    expect(choiceRequest(choice({}), ctx, handlers()).label).toBe("Smothering Tithe triggered");
    expect(
      choiceRequest(choice({ kind: "pay_unless", pay_cost: "{2}" }), ctx, handlers()).label,
    ).toBe("Smothering Tithe — pay {2}?");
    expect(choiceRequest(choice({ kind: "coin_call" }), ctx, handlers()).label).toBe(
      "Call the flip",
    );
  });

  it("never presses on Enter or Escape", () => {
    const kinds: Partial<PendingChoiceView>[] = [
      {},
      { kind: "optional_replacement" },
      { kind: "confirm" },
      { kind: "may_cast" },
      { kind: "entry_pay_life", pay_cost: "2 life" },
      { kind: "entry_riot" },
      { kind: "pay_unless", pay_cost: "{2}" },
      { kind: "coin_call", allow_stop: true },
      { kind: "loop_shortcut" },
      { kind: "mana_pick", color_options: ["G"] },
      { kind: "choose_color", color_options: ["G"] },
      { kind: "option_pick", pick_options: [{ label: "A" }] },
      { kind: "entry_read_ahead", pick_options: [{ label: "Chapter I" }] },
    ];
    for (const over of kinds) {
      const r = choiceRequest(
        choice(over),
        { sourceName: "x", loopIterations: 10, loopAnswerable: true },
        handlers(),
      );
      expect(dockKeyAction(r, "Enter"), over.kind).toBeNull();
      expect(dockKeyAction(r, "Escape"), over.kind).toBeNull();
    }
  });

  it("names Y and N on the yes/no buttons, and answers through onAnswer", () => {
    const h = handlers();
    const r = choiceRequest(
      choice({ kind: "pay_unless", pay_cost: "{2}" }),
      { sourceName: "x" },
      h,
    );
    expect(r.primary?.label).toBe("Pay {2}");
    expect(r.primary?.keyShortcuts).toBe("Y");
    expect(r.secondary?.map((a) => [a.label, a.keyShortcuts])).toEqual([["Don't pay", "N"]]);
    r.primary!.onPress();
    r.secondary![0].onPress();
    expect(h.onAnswer.mock.calls).toEqual([[true], [false]]);
  });

  // ADR 0109 §10: riot's two answers are a counter and haste, keyed C
  // and H, and the counter is the apply.
  it("asks riot as a counter or haste, keyed C and H", () => {
    const h = handlers();
    const r = choiceRequest(
      choice({
        kind: "entry_riot",
        reason: "Riot — Zhur-Taa Goblin enters with your choice of a +1/+1 counter or haste",
        accept_label: "+1/+1 counter",
        decline_label: "Haste",
        entry_keyword: "riot",
      }),
      { sourceName: "x" },
      h,
    );
    expect(r.question).toContain("Zhur-Taa Goblin");
    expect(r.tag).toBe("riot");
    expect([r.primary?.label, r.primary?.keyShortcuts]).toEqual(["+1/+1 counter", "C"]);
    expect(r.secondary?.map((a) => [a.label, a.keyShortcuts])).toEqual([["Haste", "H"]]);
    r.primary!.onPress();
    r.secondary![0].onPress();
    expect(h.onAnswer.mock.calls).toEqual([[true], [false]]);
  });

  it("words unleash's may as unleash, and keeps Yes and No", () => {
    const r = choiceRequest(
      choice({ kind: "optional_replacement", entry_keyword: "unleash" }),
      { sourceName: "x" },
      handlers(),
    );
    expect(r.tag).toBe("unleash");
    expect(r.hint).toContain("can't block");
    expect(r.primary?.label).toBe("Yes");
  });

  // ADR 0115 (CR 903.9a): the commander's owner answers Yes / No, and
  // the hint says when it could be cast from where it is.
  it("asks commander_return as Yes / No, naming a castable zone", () => {
    const h = handlers();
    const r = choiceRequest(
      choice({ kind: "commander_return", reason: "", source: "cmd" }),
      { sourceName: "Ezuri" },
      h,
    );
    expect(r.question).toBe("Ezuri — put it into the command zone?");
    expect(r.tag).toBe("commander");
    expect(r.hint).not.toContain("cast it");
    expect([r.primary?.label, r.primary?.keyShortcuts]).toEqual(["Yes", "Y"]);
    expect(r.secondary?.map((a) => [a.label, a.keyShortcuts])).toEqual([["No", "N"]]);
    r.primary!.onPress();
    r.secondary![0].onPress();
    expect(h.onAnswer.mock.calls).toEqual([[true], [false]]);

    const castable = choiceRequest(
      choice({ kind: "commander_return", playable_from_zone: true }),
      { sourceName: "Ezuri" },
      handlers(),
    );
    expect(castable.hint).toContain("You could cast it from where it is now");
  });

  it("carries a refusal as Not accepted", () => {
    const r = choiceRequest(choice({}), { sourceName: "x", rejection: "nope" }, handlers());
    expect(r.refusal).toMatchObject({ tag: "Not accepted", text: "nope", tone: "danger" });
  });
});

describe("voteRequest and gameOverRequest", () => {
  it("a vote is a step: it keeps next, with the tally on each option", () => {
    const r = voteRequest({
      vote: {
        id: "v",
        topic: "Monarchy?",
        options: ["yes", "no"],
        initiator: "b",
        ballots: { a: 1, b: 1 },
      },
      viewerID: "a",
      seats: [
        { id: "a", name: "Ann" },
        { id: "b", name: "Bob" },
      ],
      onCast: () => {},
      onEnd: () => {},
    });
    expect(r.rank).toBe("step");
    expect(r.label).toBe("open vote");
    expect(r.detail).toBe("called by Bob");
    expect(r.row?.map((a) => [a.label, a.note, a.pressed])).toEqual([
      ["yes", "0", false],
      ["no", "2", true],
      ["end vote", undefined, undefined],
    ]);
  });

  it("a rules vote (ADR 0146) is read-only for the seats not voting, with the tally", () => {
    const ballot = choice({
      kind: "option_pick",
      chooser: "b",
      reason: "Plea for Power — vote for time or knowledge",
      pick_options: [{ label: "time" }, { label: "knowledge" }, { label: "Don't vote again" }],
      council_vote: {
        controller: "a",
        voter: "b",
        options: ["time", "knowledge"],
        tally: [2, 1],
        offered: [0, 1, -1],
      },
    });
    const seats = [
      { id: "a", name: "Ann" },
      { id: "b", name: "Bob" },
    ];
    const r = councilVoteRequest(ballot, "a", seats);
    expect(r?.rank).toBe("step");
    expect(r?.label).toBe("vote");
    expect(r?.detail).toBe("Bob is voting");
    expect(r?.row?.map((a) => [a.label, a.note, a.disabled])).toEqual([
      ["time", "2", true],
      ["knowledge", "1", true],
    ]);
    // The voter answers their own ballot; its buttons carry the tally.
    expect(councilVoteRequest(ballot, "b", seats)).toBeNull();
    expect([0, 1, 2].map((i) => ballotNote(ballot, i))).toEqual(["2", "1", undefined]);
    expect(ballotNote(choice({ kind: "option_pick" }), 0)).toBeUndefined();
  });

  it("a secret vote (ADR 0180) shows who has voted, the viewer's own vote, and no counts", () => {
    const seats = [
      { id: "a", name: "Ann" },
      { id: "b", name: "Bob" },
      { id: "c", name: "Cy" },
    ];
    const secretBallot = (ballots: { voter: string; option: number }[]) =>
      choice({
        kind: "option_pick",
        chooser: "b",
        reason: "Truth or Consequences — secretly vote for truth or consequences",
        pick_options: [{ label: "truth" }, { label: "consequences" }],
        council_vote: {
          controller: "a",
          voter: "b",
          options: ["truth", "consequences"],
          tally: [0, 0],
          ballots,
          offered: [0, 1],
          secret: true,
          voted: ["a"],
        },
      });
    // Ann has voted; her wire carries her own ballot only.
    const ann = councilVoteRequest(secretBallot([{ voter: "a", option: 1 }]), "a", seats);
    expect(ann?.label).toBe("secret vote");
    expect(ann?.detail).toBe(
      "Bob is voting in secret · 1 voted so far · you voted for consequences, waiting for the others",
    );
    expect(ann?.row?.map((r) => [r.label, r.note])).toEqual([
      ["truth", undefined],
      ["consequences", "your vote"],
    ]);
    // Cy's wire carries no ballot: no counts, no marks.
    const cy = councilVoteRequest(secretBallot([]), "c", seats);
    expect(cy?.detail).toBe("Bob is voting in secret · 1 voted so far");
    expect(cy?.row?.map((r) => r.note)).toEqual([undefined, undefined]);
    // Bob's own ballot shows no tally either.
    expect([0, 1].map((i) => ballotNote(secretBallot([]), i))).toEqual([undefined, undefined]);
  });

  it("names who is choosing for whom when someone else casts the ballot (Illusion of Choice)", () => {
    const ballot = choice({
      kind: "option_pick",
      chooser: "a",
      reason: "Plea for Power — vote for time or knowledge — you choose Bob's vote",
      pick_options: [{ label: "time" }, { label: "knowledge" }],
      council_vote: {
        controller: "a",
        voter: "b",
        options: ["time", "knowledge"],
        tally: [0, 0],
        offered: [0, 1],
      },
    });
    const seats = [
      { id: "a", name: "Ann" },
      { id: "b", name: "Bob" },
      { id: "c", name: "Cy" },
    ];
    // A seat that is neither sees who is choosing for whom.
    expect(councilVoteRequest(ballot, "c", seats)?.detail).toBe("Ann is choosing Bob's vote");
    // The chooser answers the ballot itself; the voter does not, so Bob
    // is shown the same line as any watcher.
    expect(councilVoteRequest(ballot, "a", seats)).toBeNull();
    expect(councilVoteRequest(ballot, "b", seats)?.detail).toBe("Ann is choosing Bob's vote");
  });

  it("game over is Back to lobby, with no Enter", () => {
    const back = vi.fn();
    const r = gameOverRequest(back);
    expect(r.rank).toBe("gameOver");
    expect(r.primary?.label).toBe("Back to lobby");
    expect(dockKeyAction(r, "Enter")).toBeNull();
    r.primary!.onPress();
    expect(back).toHaveBeenCalledOnce();
  });

  it("game over before its grace is disabled and refuses a press (#2919)", () => {
    const back = vi.fn();
    const r = gameOverRequest(back, false);
    expect(r.primary?.disabled).toBe(true);
    r.primary!.onPress();
    expect(back).not.toHaveBeenCalled();
  });
});

describe("an unless-any-player-pays offer (ADR 0150 §5)", () => {
  const seats = [
    { id: "a", name: "Alice" },
    { id: "b", name: "Bob" },
    { id: "c", name: "Carol" },
    { id: "d", name: "Dan" },
  ];
  // Bob's Rhystic Tutor; Alice has declined, Carol is asked now.
  const link = (over: Partial<PendingChoiceView> = {}) =>
    choice({
      kind: "pay_unless",
      chooser: "c",
      reason: "Rhystic Tutor — pay {2} to stop the search?",
      pay_cost: "{2}",
      any_player_offer: { controller: "b", order: ["a", "b", "c", "d"], declined: ["a", "b"] },
      ...over,
    });

  it("tells the player asked whose effect it is and who else is asked", () => {
    const r = choiceRequest(link(), { sourceName: "Rhystic Tutor", seats }, handlers());
    expect(r.label).toBe("Rhystic Tutor — pay {2} to stop the search?");
    expect(r.tag).toBe("any player pays");
    expect(r.hint).toContain("Bob's Rhystic Tutor.");
    expect(r.hint).toContain("Any player may pay {2}: Alice and Bob declined; then Dan.");
    expect(r.primary?.label).toBe("Pay {2}");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Don't pay"]);
  });

  it("words the last link and an opponents-only offer", () => {
    expect(
      anyPlayerOfferLine(
        link({ chooser: "d", any_player_offer: { controller: "b", order: ["a", "b", "c", "d"] } }),
        seats,
      ),
    ).toBe("Any player may pay {2}: nobody after.");
    expect(
      anyPlayerOfferLine(
        link({ any_player_offer: { controller: "b", order: ["c", "d", "a"], opponents: true } }),
        seats,
      ),
    ).toBe("Any opponent may pay {2}: then Dan and Alice.");
  });

  it("is read-only for every other seat, and absent for the one asked", () => {
    const r = anyPlayerOfferRequest(link(), "a", seats, "Rhystic Tutor");
    expect(r?.rank).toBe("step");
    expect(r?.row).toBeUndefined();
    expect(r?.primary).toBeUndefined();
    expect(r?.detail).toBe("Carol is deciding whether to pay {2} (Bob's Rhystic Tutor)");
    expect(r?.hint).toBe("Any player may pay {2}: Alice and Bob declined; then Dan.");
    expect(anyPlayerOfferRequest(link(), "c", seats, "Rhystic Tutor")).toBeNull();
    expect(anyPlayerOfferRequest(choice({ kind: "pay_unless" }), "a", seats, "x")).toBeNull();
  });

  // ADR 0150 §3: life, damage, a sacrifice and an offer that costs nothing.
  it("words a life payment and greys it for a seat without the life", () => {
    const life = link({ pay_cost: "5 life", pay_life: 5, reason: "" });
    const ok = choiceRequest(life, { sourceName: "Dash Hopes", seats, life: 20 }, handlers());
    expect(ok.label).toBe("Dash Hopes — pay 5 life?");
    expect(ok.hint).toContain("Any player may pay 5 life:");
    expect(ok.hint).toContain("Pay 5 life, or don't.");
    expect(ok.primary?.label).toBe("Pay 5 life");
    expect(ok.primary?.disabled).toBeFalsy();
    const short = choiceRequest(life, { sourceName: "Dash Hopes", seats, life: 4 }, handlers());
    expect(short.primary?.disabled).toBe(true);
    expect(short.hint).toContain("You have 4 life, so you can't pay 5.");
  });

  it("words damage with the card's own answers", () => {
    const dmg = link({
      pay_cost: "5 damage",
      pay_damage: 5,
      reason: "",
      accept_label: "Take 5 damage",
      decline_label: "Don't",
    });
    const r = choiceRequest(dmg, { sourceName: "Browbeat", seats }, handlers());
    expect(r.label).toBe("Browbeat — take 5 damage?");
    expect(r.hint).toContain("Any player may take 5 damage:");
    expect(r.hint).toContain("Browbeat deals 5 damage to you if you take it.");
    expect(r.primary?.label).toBe("Take 5 damage");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Don't"]);
    expect(anyPlayerWaiting(dmg, seats, "Browbeat")).toBe(
      "Carol is deciding whether to take 5 damage (Bob's Browbeat)",
    );
  });

  it("words a sacrifice and an offer that costs nothing", () => {
    expect(
      anyPlayerOfferLine(
        link({
          pay_cost: "sacrifice a land",
          pay_cards: { action: "sacrifice", count: 1, options: [] },
        }),
        seats,
      ),
    ).toBe("Any player may sacrifice a land: Alice and Bob declined; then Dan.");
    const free = link({
      pay_cost: "",
      pay_nothing: true,
      accept_label: "Let them have the card",
      decline_label: "Let them draw three",
      any_player_offer: { controller: "b", order: ["c", "d", "a"], opponents: true },
    });
    expect(anyPlayerOfferLine(free, seats)).toBe(
      "Any opponent may let them have the card: then Dan and Alice.",
    );
    const r = choiceRequest(free, { sourceName: "Distant Memories", seats }, handlers());
    expect(r.primary?.label).toBe("Let them have the card");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Let them draw three"]);
    expect(r.hint).toContain("It costs you nothing.");
  });
});

describe("assist (ADR 0150 §5)", () => {
  const seats = [
    { id: "a", name: "Alice" },
    { id: "b", name: "Bob" },
    { id: "c", name: "Carol" },
  ];
  // Alice casts Fan Favorite and asks Bob, who can pay up to {2}.
  const ask = (over: Partial<PendingChoiceView> = {}) =>
    choice({
      kind: "pay_amount",
      chooser: "b",
      reason: "Alice asks you to help pay for Fan Favorite",
      pay_amount: { min: 0, max: 2, unit: "assist", resource: "mana" },
      assist: {
        caster: "a",
        spell: "s1",
        spell_name: "Fan Favorite",
        mana_cost: "{3}{B}",
        generic: 3,
      },
      ...over,
    });

  it("asks the helper with the caster, the spell and its cost, and pays from a stepper", () => {
    const h = { ...handlers(), onAmount: vi.fn() };
    const r = choiceRequest(ask(), { sourceName: "a card", seats, payAmount: 2 }, h);
    expect(r.label).toBe("Alice asks you to help pay for Fan Favorite ({3}{B})");
    expect(r.tag).toBe("assist");
    expect(r.hint).toContain("Pay up to {2} of its generic mana, or nothing.");
    expect(r.primary?.label).toBe("Pay {2}");
    expect(r.secondary?.map((a) => a.label)).toEqual(["Don't help"]);
    r.primary?.onPress();
    r.secondary?.[0].onPress();
    expect(h.onAmount.mock.calls).toEqual([[2], [0]]);
    const none = choiceRequest(ask(), { sourceName: "x", seats, payAmount: 0 }, h);
    expect(none.primary?.disabled).toBe(true);
  });

  it("gives the caster a waiting line with Cancel request", () => {
    const cancel = vi.fn();
    const r = assistRequest(ask(), "a", seats, cancel);
    expect(r?.rank).toBe("step");
    expect(r?.question).toBe("Waiting for Bob to help pay for Fan Favorite");
    expect(r?.row?.map((a) => a.label)).toEqual(["Cancel request"]);
    r?.row?.[0].onPress();
    expect(cancel).toHaveBeenCalledOnce();
  });

  it("is read-only for everyone else, and absent for the helper", () => {
    const r = assistRequest(ask(), "c", seats, vi.fn());
    expect(r?.question).toBe("Bob is deciding how much to help pay for Alice's Fan Favorite");
    expect(r?.row).toBeUndefined();
    expect(r?.primary).toBeUndefined();
    expect(assistRequest(ask(), "b", seats, vi.fn())).toBeNull();
    expect(assistRequest(choice({ kind: "pay_amount" }), "a", seats, vi.fn())).toBeNull();
    expect(assistWaiting(ask(), null, seats)).toBe(
      "Bob is deciding how much to help pay for Alice's Fan Favorite",
    );
  });
});

describe("a per-land any-player offer (ADR 0187 §8)", () => {
  const seats = [
    { id: "a", name: "Alice" },
    { id: "b", name: "Bob" },
    { id: "c", name: "Carol" },
    { id: "d", name: "Dan" },
  ];
  const land = (id: string) => ({ instance_id: id, type_line: "Basic Land — Forest" });
  // Bob's Cleansing; Alice chose two lands and Bob none, Carol chooses now,
  // Dan is after her.
  const leg = (over: Partial<PendingChoiceView> = {}) =>
    choice({
      kind: "choose_cards",
      chooser: "c",
      reason: "Cleansing — choose lands to pay 1 life for",
      choose_min: 0,
      choose_max: 3,
      options: [land("l1"), land("l2"), land("l3")] as never,
      any_player_offer: {
        controller: "b",
        order: ["a", "b", "c", "d"],
        declined: ["b"],
        each_life: 1,
      },
      earlier_choices: [
        { player_id: "a", card_ids: ["l1", "l2"] },
        { player_id: "b", card_ids: [] },
      ],
      ...over,
    });

  it("recognises only a choose_cards leg with each_life", () => {
    expect(isAnyPlayerEachLeg(leg())).toBe(true);
    expect(
      isAnyPlayerEachLeg(choice({ kind: "pay_unless", any_player_offer: leg().any_player_offer })),
    ).toBe(false);
    expect(
      isAnyPlayerEachLeg(leg({ any_player_offer: { controller: "b", order: ["a", "b", "c"] } })),
    ).toBe(false);
    expect(isAnyPlayerEachLeg(null)).toBe(false);
  });

  it("gives the chooser the rule, who is next, the choices so far and when it is paid", () => {
    expect(anyPlayerEachHeader(leg(), seats, "Cleansing")).toBe(
      "Bob's Cleansing. Each player in turn chooses lands to pay 1 life for; a land anyone pays for is saved. Then Dan. So far: Alice pays for 2 lands, Bob for none. Nothing is paid until the last player has chosen.",
    );
  });

  it("says the chooser is last, and leaves out a summary before anyone chose", () => {
    const last = leg({
      chooser: "d",
      any_player_offer: { controller: "b", order: ["a", "b", "c", "d"], each_life: 1 },
      earlier_choices: [],
    });
    expect(anyPlayerEachHeader(last, seats, "Cleansing")).toBe(
      "Bob's Cleansing. Each player in turn chooses lands to pay 1 life for; a land anyone pays for is saved. You are the last to choose. Nothing is paid until the last player has chosen.",
    );
  });

  it("counts a player given no leg (they could not pay) as none", () => {
    const c = leg({
      earlier_choices: [{ player_id: "a", card_ids: ["l1"] }],
      any_player_offer: {
        controller: "b",
        order: ["a", "b", "c", "d"],
        declined: ["b"],
        each_life: 1,
      },
    });
    expect(anyPlayerEachSummary(c, seats)).toBe("So far: Alice pays for 1 land, Bob for none.");
  });

  it("badges a land with every earlier payer and warns when it is selected", () => {
    const c = leg({
      earlier_choices: [
        { player_id: "a", card_ids: ["l1", "l2"] },
        { player_id: "b", card_ids: ["l2"] },
      ],
    });
    expect(anyPlayerEachPayers(c, "l1")).toEqual(["a"]);
    expect(anyPlayerEachPayers(c, "l2")).toEqual(["a", "b"]);
    expect(anyPlayerEachPayers(c, "l3")).toEqual([]);
    expect(anyPlayerEachBadge(["a"], seats)).toBe("Alice pays");
    expect(anyPlayerEachBadge(["a", "b"], seats)).toBe("Alice and Bob pay");
    expect(anyPlayerEachBadge([], seats)).toBe("");
    expect(anyPlayerEachWarning(c, ["a"], seats)).toBe(
      "Alice already pays for this: you'd both pay 1 life",
    );
    expect(anyPlayerEachWarning(c, ["a", "b"], seats)).toBe(
      "Alice and Bob already pay for this: you'd all pay 1 life",
    );
  });

  it("prices the pick as it is toggled", () => {
    const c = leg();
    expect(anyPlayerEachPrice(c, 0, 0)).toBe("Pay nothing");
    expect(anyPlayerEachPrice(c, 1, 0)).toBe("Pay 1 life for 1 land");
    expect(anyPlayerEachPrice(c, 3, 0)).toBe("Pay 3 life for 3 lands");
    expect(anyPlayerEachPrice(c, 3, 2)).toBe(
      "Pay 3 life for 3 lands (2 already paid for by others)",
    );
    const two = leg({ any_player_offer: { controller: "b", order: ["a", "c"], each_life: 2 } });
    expect(anyPlayerEachPrice(two, 2, 0)).toBe("Pay 4 life for 2 lands");
  });

  it("gives every other seat a read-only line with the choices so far", () => {
    const r = anyPlayerOfferRequest(leg(), "a", seats, "Cleansing");
    expect(r).not.toBeNull();
    expect(r?.tag).toBe("any player pays");
    expect(r?.question).toBe("Cleansing — choose lands to pay 1 life for");
    expect(r?.detail).toBe(
      "Carol is choosing lands to pay 1 life for (Bob's Cleansing). So far: Alice 2, Bob none.",
    );
    expect(r?.hint).toBe(
      "Each player in turn chooses lands to pay 1 life for; a land anyone pays for is saved. Then Dan. Nothing is paid until the last player has chosen.",
    );
    expect(r?.primary).toBeUndefined();
    // The chooser answers it in the sheet, not here.
    expect(anyPlayerOfferRequest(leg(), "c", seats, "Cleansing")).toBeNull();
  });

  it("names the noun from the board when the watcher has no options", () => {
    const watching = leg({ options: undefined });
    expect(anyPlayerWaiting(watching, seats, "Cleansing")).toContain("choosing permanents to pay");
    expect(anyPlayerWaiting(watching, seats, "Cleansing", [land("l1"), land("l2")])).toContain(
      "choosing lands to pay",
    );
  });

  it("leaves a pay_unless offer's lines alone", () => {
    const pu = choice({
      kind: "pay_unless",
      chooser: "c",
      pay_cost: "{2}",
      any_player_offer: { controller: "b", order: ["a", "b", "c", "d"], declined: ["a", "b"] },
    });
    expect(anyPlayerOfferLine(pu, seats)).toBe(
      "Any player may pay {2}: Alice and Bob declined; then Dan.",
    );
    expect(anyPlayerWaiting(pu, seats, "Rhystic Tutor")).toBe(
      "Carol is deciding whether to pay {2} (Bob's Rhystic Tutor)",
    );
  });
});

// #4131: the server lets the player who called a vote end it, and any seat
// once that player has left the table.
describe("mayEndVote", () => {
  const vote = { id: "v", topic: "?", options: ["a", "b"], initiator: "b", ballots: {} };
  it("is the initiator's alone while the initiator is at the table", () => {
    const seats = [{ id: "a" }, { id: "b" }];
    expect(mayEndVote(vote, "b", seats)).toBe(true);
    expect(mayEndVote(vote, "a", seats)).toBe(false);
    expect(mayEndVote(vote, null, seats)).toBe(false);
  });
  it("opens to any seat once the initiator is eliminated or gone", () => {
    expect(mayEndVote(vote, "a", [{ id: "a" }, { id: "b", eliminated: true }])).toBe(true);
    expect(mayEndVote(vote, "a", [{ id: "a" }])).toBe(true);
  });
});
