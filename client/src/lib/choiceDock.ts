// choiceDock.ts — the small pending choices, inline in the action dock
// (ADR 0111 §2, Delivery PR 5). Pure builders, like combatDock.ts and
// targetingDock.ts: ChoicePromptModal owns the open prompt, its state
// and its answer (and so the "Not accepted" refusal), Game.svelte owns
// the vote and the game's end, and these turn them into the
// DockRequest the dock draws.
//
// INLINE is a question plus at most about six short buttons, or a
// single number (§2): the yes/no family (trigger_prompt,
// optional_replacement, commander_return, confirm, may_cast, entry_pay_life,
// entry_riot), pay_amount's single number (ADR 0129 §3), pay_unless
// without card or tap picks, coin_call, loop_shortcut, mana_pick,
// choose_color, option_pick, entry_controller and entry_read_ahead when
// every option is a short label, and an open vote. Everything else is a sheet (PR 6)
// and stays in ChoicePromptModal until then.
//
// An inline prompt is not a modal: it blurs and blocks nothing, so the
// board can be read before answering. Its dialog keeps the name the
// modal had (`reason`, or the modal's fallback heading), because the
// e2e suite finds a trigger's buttons inside it, and the buttons keep
// their names: Yes, No, "Pay {2}", "Don't pay", Heads, Tails.
//
// Keys. None of these presses on Enter or Escape (§1: "a yes/no
// question does not take Enter"; a stray Enter must not accept an
// optional effect or pay a cost, and Escape is no "No"). They answer by
// the keys they had: Y / N, H / T / S, and 1-9 on the mana symbols.
// Each button names its key in aria-keyshortcuts and a cap. Focus goes
// to the dialog, not to the primary, so Enter on a focused "Yes" is not
// one keypress away when the question appears.

import type { Snippet } from "svelte";
import { guardedWritable } from "./guardedStore";

import type { DockAction, DockRequest } from "./dock";
import type { PendingChoiceView, PickOptionView, PlayerView, VoteView } from "./protocol";
import { colorPromptCopy } from "./manaPick";
import { mayCastCopy } from "./mayCast";
import { PhyrexianLifePerSymbol, maxPhyrexianLife, phyrexianLifeCost } from "./phyrexianLife";
import { doubledTriggerLabel } from "./triggerDoubling";
import { askedByHandText, canRemember } from "./autoAnswerPref";
import { L } from "./labels";
import { energyShortBy, energyShortReason, payAmountHint, payAmountResource } from "./payEnergy";

// "Short" for an option_pick / entry_controller (ADR 0111 §2: "Inline
// when every option is a short label; a sheet when an option embeds
// cards"; "at most about six short buttons"). Six options, none with
// cards, none longer than this many characters. The catalog's longest
// inline-worthy label today is "Return an enchantment card from your
// graveyard to your hand" (59); Seize the Spotlight's "Fame — its caster
// gains control of a creature you control until end of turn" (75) is a
// sentence and goes to a sheet.
export const INLINE_OPTION_MAX = 6;
export const INLINE_LABEL_MAX = 64;

const YES_NO_KINDS = new Set([
  "trigger_prompt",
  "optional_replacement",
  "commander_return",
  "confirm",
  "may_cast",
  "entry_pay_life",
  "entry_riot",
]);

// shortOptions reports whether an option_pick / entry_controller fits
// the dock: 1-6 options, no cards, every label short.
export function shortOptions(c: Pick<PendingChoiceView, "pick_options">): boolean {
  const opts = c.pick_options ?? [];
  if (opts.length === 0 || opts.length > INLINE_OPTION_MAX) return false;
  return opts.every(
    (o) => (o.cards?.length ?? 0) === 0 && (o.label ?? "").length <= INLINE_LABEL_MAX,
  );
}

// pickOptionText is an option_pick button's text: its label, and the
// mana it costs (#2854) when the label does not already say it —
// "Pay {2}" stays as it is, "Keep it" becomes "Keep it ({2})".
export function pickOptionText(o: Pick<PickOptionView, "label" | "mana_cost">): string {
  const cost = o.mana_cost ?? "";
  if (!cost || o.label.includes(cost)) return o.label;
  return `${o.label} (${cost})`;
}

// isInlineChoice reports whether a pending choice is answered inline in
// the dock rather than in ChoicePromptModal.
export function isInlineChoice(c: PendingChoiceView | null | undefined): boolean {
  if (!c) return false;
  if (YES_NO_KINDS.has(c.kind)) return true;
  switch (c.kind) {
    case "pay_unless":
      // With card picks (ADR 0108 §5) or a waterbend tap list (#1311)
      // the picks need room: a sheet, PR 6.
      return !c.pay_cards && !c.tap_cost;
    case "coin_call":
    case "loop_shortcut":
    case "mana_pick":
    case "choose_color":
    case "pay_amount":
      return true;
    case "option_pick":
    case "entry_controller":
    case "entry_read_ahead":
      return shortOptions(c);
    default:
      return false;
  }
}

// What the component knows that a pure builder cannot look up.
export interface ChoiceDockContext {
  // The trigger's source card name, wherever it is now ("Triggered
  // ability" when the viewer cannot see it).
  sourceName: string;
  // may_cast: the offered card's name, when it is visible.
  mayCastCardName?: string;
  // loop_shortcut: the count in the field, and whether it is legal.
  loopIterations?: number;
  loopAnswerable?: boolean;
  // The kind's own body (the mana symbols, the loop count field).
  body?: Snippet;
  // The server's refusal of the last answer to this prompt (#624).
  rejection?: string | null;
  // ADR 0129 §3: the viewer's energy, for an energy payment's Pay.
  energy?: number;
  // pay_amount: the stepper's value, and whether the server takes it.
  payAmount?: number;
  payAmountAnswerable?: boolean;
  // The viewer's life total: a mana pay_unless offers "pay with life" for
  // the symbols the chooser could pay 2 life each for (ADR 0131 §2),
  // bounded by CR 119.4. Undefined offers none.
  life?: number;
  // ADR 0127 §6: "Remember this answer", on a prompt that can take a
  // standing answer. `on` is the toggle's state (off each time a prompt
  // appears); the button pressed next also sets the rule. Undefined
  // draws no toggle.
  remember?: { on: boolean; onToggle: () => void };
}

export interface ChoiceDockHandlers {
  // The yes/no family and pay_unless: { apply }. `phyrexianLife` rides a
  // pay_unless "pay" that spends 2 life on that many symbols.
  onAnswer: (apply: boolean, phyrexianLife?: number) => void;
  onCoin: (call: "heads" | "tails" | "stop") => void;
  onOption: (index: number) => void;
  onLoop: (iterations: number) => void;
  // pay_amount (ADR 0129 §3): { amount }.
  onAmount?: (amount: number) => void;
}

// A key the inline prompt answers by itself (ChoicePromptModal's
// window handler): its aria-keyshortcuts and its cap. Never Enter or
// Escape, so the dock's one key handler never presses it.
function keyed(id: string, label: string, key: string, onPress: () => void): DockAction {
  return { id, label, keyShortcuts: key, cap: key, onPress };
}

interface Copy {
  title: string;
  tag: string;
  hint?: string;
  hintWarn?: boolean;
  detail?: string;
}

// payLifeMax is how many of a mana pay_unless's symbols the viewer can
// claim for 2 life each: the server's ceiling, capped by CR 119.4. Zero
// for every other prompt, and when the life total is unknown.
function payLifeMax(c: PendingChoiceView, ctx: ChoiceDockContext): number {
  if (c.kind !== "pay_unless" || ctx.life === undefined) return 0;
  return maxPhyrexianLife(c.phyrexian_symbols ?? 0, ctx.life);
}

// payLifeWords names what the life pays for in the hint: a granted {B}
// (K'rrik), or a printed Phyrexian symbol.
function payLifeWords(c: PendingChoiceView): string {
  const symbols = c.phyrexian_symbols ?? 0;
  if ((c.phyrexian_granted ?? 0) >= symbols) {
    return symbols === 1 ? "The {B}" : "Each {B}";
  }
  return symbols === 1 ? "The Phyrexian symbol" : "Each Phyrexian symbol";
}

function copyFor(c: PendingChoiceView, ctx: ChoiceDockContext): Copy {
  const reason = c.reason ?? "";
  switch (c.kind) {
    case "optional_replacement":
      if (c.entry_keyword === "unleash") {
        return {
          title: reason || "Unleash — enter with a +1/+1 counter?",
          tag: "unleash",
          hint: "It hasn't entered yet. With the counter it can't block for as long as it has one; without it, it enters plain.",
        };
      }
      return {
        title: reason || "Apply replacement?",
        tag: "replacement",
        hint: "You (the affected player) decide whether this substitution applies.",
      };
    case "commander_return":
      // ADR 0115 (CR 903.9a). The commander has already landed: its
      // dies and leaves triggers have seen it, and this only decides
      // whether it stays.
      return {
        title: reason || `${ctx.sourceName} — put it into the command zone?`,
        tag: "commander",
        hint:
          c.playable_from_zone === true
            ? "You could cast it from where it is now. Yes sends it to the command zone; No leaves it there."
            : "Yes sends it to the command zone; No leaves it where it is.",
      };
    case "entry_riot":
      return {
        title: reason || "Riot — a +1/+1 counter or haste?",
        tag: "riot",
        hint: "It hasn't entered yet: it enters with a +1/+1 counter, or with haste so it can attack this turn.",
      };
    case "trigger_prompt":
      return {
        title: reason || `${ctx.sourceName} triggered`,
        tag: "may",
        detail: doubledTriggerLabel(c.doubled_by, c.doubled_by_name) ?? undefined,
        hint:
          c.no_legal_target === true
            ? "No legal target — “Yes” passes without effect."
            : "Fire the ability, or let it pass without effect.",
        hintWarn: c.no_legal_target === true,
      };
    case "may_cast": {
      const words = mayCastCopy(c.may_cast_keyword, c.accept_label, c.decline_label);
      return {
        title: reason || "Cast it without paying its mana cost?",
        tag: words.source.split(" · ")[0] || "cast",
        hint:
          (ctx.mayCastCardName ? `${ctx.mayCastCardName} is exiled face up. ` : "") + words.hint,
      };
    }
    case "entry_pay_life":
      return {
        title: reason || `Pay ${c.pay_cost ?? ""} as it enters?`,
        tag: "as it enters",
        hint: `Pay ${c.pay_cost ?? "the life"} and it enters untapped; don't, and it enters tapped. Nothing has entered yet.`,
      };
    case "confirm":
      return {
        title: reason || "Choose one",
        tag: "choose",
        hint: "Both answers are legal — a choice between two things the card does. There may be another question after it.",
      };
    case "pay_unless": {
      if (c.pay_energy !== undefined && c.pay_energy !== null) {
        // ADR 0129 §3: an energy payment. Nothing taps; the counters
        // come off the seat.
        const have = ctx.energy ?? 0;
        const short = energyShortBy(c, have) > 0;
        return {
          title: reason || `${ctx.sourceName} — pay ${c.pay_cost ?? ""}?`,
          tag: "pay energy",
          hint: short
            ? `${energyShortReason(have, c.pay_energy)}: you can't pay, so ${ctx.sourceName} goes on as if you don't.`
            : `Pay ${c.pay_cost ?? "the energy"} from your ${have} energy, or don't.`,
          hintWarn: short,
        };
      }
      return {
        title: reason || `${ctx.sourceName} — pay ${c.pay_cost ?? ""}?`,
        tag: "pay unless",
        hint:
          `Pay ${c.pay_cost ?? "the cost"} from your pool (untapped sources auto-tap if it's short), or don't and let ${ctx.sourceName} do its thing.` +
          (payLifeMax(c, ctx) > 0
            ? ` ${payLifeWords(c)} can instead be paid with ${PhyrexianLifePerSymbol} life each.`
            : ""),
      };
    }
    case "pay_amount": {
      const pa = c.pay_amount;
      const resource = pa ? payAmountResource(pa) : "energy";
      const what = resource === "none" ? "choose a number" : `pay ${resource}`;
      return {
        title: reason || `${ctx.sourceName} — ${what}?`,
        tag: what,
        hint: pa ? payAmountHint(pa, ctx.sourceName) : undefined,
      };
    }
    case "coin_call": {
      const coins = c.coins ?? 1;
      const wins = c.wins ?? 0;
      return {
        title: reason || "Call the flip",
        tag: "coin flip",
        hint:
          `Call heads or tails for ${coins} ${coins === 1 ? "coin" : "coins"}.` +
          (wins > 0 ? ` You have won ${wins} ${wins === 1 ? "flip" : "flips"} so far.` : ""),
      };
    }
    case "loop_shortcut": {
      const n = c.loop_count ?? 0;
      return {
        title: reason || "This ability keeps resolving",
        tag: "loop",
        hint: `It has resolved ${n} ${n === 1 ? "time" : "times"} this turn with nobody doing anything in between. Say how many more times to run it; Stop here leaves auto-pass paused so you can step through by hand.`,
      };
    }
    case "mana_pick":
      return {
        title: reason || "Pick a color",
        tag: "mana",
        hint:
          Object.keys(c.color_amounts ?? {}).length > 0
            ? "Choose one color. All of this mana is added in that color."
            : "Choose a color to add to your mana pool.",
      };
    case "choose_color": {
      const copy = colorPromptCopy(c.color_purpose);
      return { title: reason || copy.title, tag: "color", hint: copy.hint };
    }
    case "option_pick":
      return {
        title: reason || "Choose one",
        tag: "choose one",
        hint: "Someone else's spell or ability is asking you. Every option is one you can take.",
      };
    case "entry_read_ahead":
      return {
        title: reason || "Read ahead — choose the starting chapter",
        tag: "read ahead",
        hint: "It hasn't entered yet: it enters with that many lore counters, and the chapters before it never happen.",
      };
    case "entry_controller":
      return {
        title: reason || "Choose an opponent",
        tag: "opponent",
        hint:
          "It hasn't entered yet: it enters under the control of the opponent you choose. " +
          (c.control_purpose === "benefit"
            ? "Whoever you choose will control it and get what it does."
            : "Whoever you choose will control it and live with what it does."),
      };
    default:
      return { title: reason || "Choose", tag: "choose" };
  }
}

// The action bar (and, for an option pick, the row) for each kind.
function answersFor(
  c: PendingChoiceView,
  ctx: ChoiceDockContext,
  h: ChoiceDockHandlers,
): Pick<DockRequest, "primary" | "secondary" | "row" | "rowLayout"> {
  const yes = (label: string) => keyed("yes", label, "Y", () => h.onAnswer(true));
  const no = (label: string) => keyed("no", label, "N", () => h.onAnswer(false));
  switch (c.kind) {
    case "optional_replacement":
    case "commander_return":
    case "trigger_prompt":
      return { primary: yes("Yes"), secondary: [no("No")] };
    case "may_cast": {
      const words = mayCastCopy(c.may_cast_keyword, c.accept_label, c.decline_label);
      return { primary: yes(words.accept), secondary: [no(words.decline)] };
    }
    case "entry_pay_life":
      return {
        primary: yes(`Pay ${c.pay_cost ?? ""}`.trim()),
        secondary: [no("Enter tapped")],
      };
    case "entry_riot":
      // Riot is a choice between two things, not a yes/no, so its keys
      // are the answers' initials: C for the counter, H for haste.
      return {
        primary: keyed("yes", c.accept_label || "+1/+1 counter", "C", () => h.onAnswer(true)),
        secondary: [keyed("no", c.decline_label || "Haste", "H", () => h.onAnswer(false))],
      };
    case "confirm":
      return {
        primary: yes(c.accept_label || "Yes"),
        secondary: [no(c.decline_label || "No")],
      };
    case "pay_unless": {
      const pay = yes(`Pay ${c.pay_cost ?? ""}`.trim());
      // ADR 0129 §8: greyed with the reason when the seat is short, so
      // the button, the key and the server agree (a short "Pay" would
      // be read as Don't pay).
      if (energyShortBy(c, ctx.energy ?? 0) > 0) {
        pay.disabled = true;
        pay.title = energyShortReason(ctx.energy ?? 0, c.pay_energy ?? 0);
      }
      // ADR 0131 §2: one "pay with life" answer per count the chooser may
      // claim, after the all-mana "Pay" and before "Don't pay". Mana is
      // never spent on a symbol the life pays for, and auto-tap never
      // claims it, so it is always the player's click.
      const lifeAnswers: DockAction[] = [];
      for (let k = 1; k <= payLifeMax(c, ctx); k++) {
        lifeAnswers.push({
          id: `pay-life-${k}`,
          label: `Pay with ${phyrexianLifeCost(k)} life`,
          onPress: () => h.onAnswer(true, k),
        });
      }
      return { primary: pay, secondary: [...lifeAnswers, no("Don't pay")] };
    }
    case "pay_amount": {
      // ADR 0129 §3 (owner decision 3): the stepper in the body sets the
      // amount; Pay sends it, and the decline sends 0. No keys: the
      // stepper's field takes digits, and Enter in it pays.
      // #1941: a life payment reads "Pay N life", and a number that is
      // not paid reads "Choose N" with no decline (0 is an ordinary
      // answer there, when the card allows it).
      const n = ctx.payAmount ?? 0;
      const onAmount = h.onAmount ?? (() => {});
      const min = c.pay_amount?.min ?? 0;
      const resource = c.pay_amount ? payAmountResource(c.pay_amount) : "energy";
      if (resource === "none") {
        return {
          primary: {
            id: "choose",
            label: L.chooseNumber(n),
            disabled: ctx.payAmountAnswerable === false,
            onPress: () => onAmount(n),
          },
        };
      }
      return {
        primary: {
          id: "pay",
          label: resource === "life" ? L.payLife(n) : L.payEnergy(n),
          disabled: n <= 0 || ctx.payAmountAnswerable === false,
          onPress: () => onAmount(n),
        },
        secondary: [
          {
            id: "none",
            label: min > 0 ? "Don't pay" : L.payNothing,
            onPress: () => onAmount(0),
          },
        ],
      };
    }
    case "coin_call":
      return {
        primary: keyed("tails", "Tails", "T", () => h.onCoin("tails")),
        secondary: [
          ...(c.allow_stop === true ? [keyed("stop", "Stop", "S", () => h.onCoin("stop"))] : []),
          keyed("heads", "Heads", "H", () => h.onCoin("heads")),
        ],
      };
    case "loop_shortcut": {
      const n = ctx.loopIterations ?? 0;
      return {
        primary: {
          id: "resolve",
          label: `Resolve ${n} more`,
          disabled: ctx.loopAnswerable === false,
          onPress: () => h.onLoop(n),
        },
        secondary: [{ id: "stop", label: "Stop here", onPress: () => h.onLoop(0) }],
      };
    }
    case "option_pick":
    case "entry_controller":
    case "entry_read_ahead":
      // No bar: a click on an option is the answer.
      return {
        primary: null,
        secondary: [],
        rowLayout: "stack",
        row: (c.pick_options ?? []).map((o, i) => ({
          id: `option-${i}`,
          label: pickOptionText(o),
          // ADR 0146: a ballot shows each option's votes so far.
          note: ballotNote(c, i),
          onPress: () => h.onOption(i),
        })),
      };
    default:
      // mana_pick, choose_color: the symbols in the body are the answer.
      return { primary: null, secondary: [] };
  }
}

// ballotNote is a ballot option's vote count so far (ADR 0146), or
// undefined on a prompt that is not a ballot and on "Don't vote again".
export function ballotNote(c: PendingChoiceView, i: number): string | undefined {
  const v = c.council_vote;
  if (!v) return undefined;
  const opt = v.offered[i] ?? -1;
  if (opt < 0) return undefined;
  return String(v.tally[opt] ?? 0);
}

// rememberTitle names the card and the question the toggle remembers an
// answer for: "Remember for Rhystic Study — pay {1}?".
function rememberTitle(c: PendingChoiceView): string {
  const card = c.auto_answer_card ?? "";
  const prompt = c.auto_answer_prompt ?? "";
  if (!prompt) return `Remember for ${card || "this card"}`;
  if (!card || prompt.startsWith(card)) return `Remember for ${prompt}`;
  return `Remember for ${card} — ${prompt}`;
}

// rememberRow is the "Remember this answer" toggle (ADR 0127 §6), as
// the request's row: a toggle button, drawn with aria-pressed. Absent
// on a prompt with no key.
function rememberRow(
  c: PendingChoiceView,
  ctx: ChoiceDockContext,
): Pick<DockRequest, "row" | "rowLayout"> {
  if (!ctx.remember || !canRemember(c)) return {};
  const remember = ctx.remember;
  return {
    row: [
      {
        id: "remember",
        label: L.rememberThisAnswer,
        title: rememberTitle(c),
        pressed: remember.on,
        alignEnd: true,
        onPress: () => remember.onToggle(),
      },
    ],
  };
}

// choiceRequest is the dock request for an inline pending choice.
export function choiceRequest(
  c: PendingChoiceView,
  ctx: ChoiceDockContext,
  handlers: ChoiceDockHandlers,
): DockRequest {
  const copy = copyFor(c, ctx);
  // ADR 0127 §4: a prompt with a rule that is asked anyway says why.
  const asked = askedByHandText(c);
  if (asked) {
    copy.hint = copy.hint ? `${asked} ${copy.hint}` : asked;
  }
  return {
    rank: "choice",
    // The modal's name: its heading, which was the reason (plus a
    // doubled trigger's note, which was part of the heading too).
    label: [copy.title, copy.detail].filter(Boolean).join(" "),
    tag: copy.tag,
    tone: "gold",
    question: copy.title,
    detail: copy.detail,
    hint: copy.hint,
    hintWarn: copy.hintWarn,
    body: ctx.body,
    refusal: ctx.rejection
      ? { tag: "Not accepted", text: ctx.rejection, actions: [], tone: "danger" }
      : null,
    focus: "dialog",
    ...rememberRow(c, ctx),
    ...answersFor(c, ctx, handlers),
  };
}

// inlineRefusal is the server error an inline prompt has claimed as the
// refusal of its own answer, and is showing as "Not accepted" in the
// dock. The strip's rejection toast stands down for exactly that error:
// with the modal's backdrop gone both would be on screen, and both are
// role="alert", so the refusal would be announced twice. Any other
// error is still the strip's. ChoicePromptModal sets and clears it.
export const inlineRefusal = guardedWritable<unknown>(null, "inlineRefusal");

// ---- the vote -------------------------------------------------------------
//
// The council's-dilemma vote (VotingPanel's open vote, top centre). It
// is a `step` request, not a `choice`: a vote never stops the game (the
// server waits on nobody's ballot and anyone can end it), so `next` and
// Pass turn stay in the action bar, as they did beside the floating
// panel, and any choice, flow or block declaration outranks it. Its
// dialog keeps the panel's name, "open vote"; each option is a button
// named by its text and its tally, pressed for the viewer's own ballot.

//
// A RULES vote (ADR 0146, CR 701.38: Council's Judgment, Plea for Power)
// is drawn with the same request for every seat that is not voting right
// now (councilVoteRequest): the options and their tallies as each vote
// is cast, read-only, with no end button. The voter answers their own
// ballot, an ordinary option_pick, whose buttons carry the same tallies
// (ballotNote).

export interface VoteRequestInput {
  vote: VoteView;
  viewerID: string | null;
  seats: Pick<PlayerView, "id" | "name">[];
  // Absent for a rules vote: nobody may cast out of turn, or end it.
  onCast?: (optionIndex: number) => void;
  onEnd?: () => void;
  // A rules vote's counts (a player may hold several votes) and its
  // own name and detail line.
  tally?: number[];
  label?: string;
  detail?: string;
}

export function voteRequest(input: VoteRequestInput): DockRequest {
  const { vote, viewerID, seats, onCast, onEnd } = input;
  let counts = input.tally;
  if (!counts) {
    counts = new Array(vote.options.length).fill(0) as number[];
    for (const opt of Object.values(vote.ballots ?? {})) {
      if (opt >= 0 && opt < counts.length) counts[opt]++;
    }
  }
  const mine = viewerID && onCast ? (vote.ballots?.[viewerID] ?? null) : null;
  const initiator = seats.find((s) => s.id === vote.initiator)?.name ?? "?";
  const row: DockAction[] = vote.options.map(
    (option, i): DockAction => ({
      id: `vote-${i}`,
      label: option,
      note: String(counts[i] ?? 0),
      pressed: onCast ? mine === i : undefined,
      emphasis: mine === i,
      disabled: !onCast,
      onPress: () => onCast?.(i),
    }),
  );
  if (onEnd) row.push({ id: "end-vote", label: "end vote", alignEnd: true, onPress: onEnd });
  return {
    rank: "step",
    label: input.label ?? "open vote",
    tag: "vote",
    tone: "plain",
    question: vote.topic || "(no topic)",
    detail: input.detail ?? `called by ${initiator}`,
    row,
  };
}

// councilVoteRequest is a rules vote as the seats that are not voting
// right now see it (ADR 0146): whose ballot it is, and each option's
// votes so far. Null when `c` is not a ballot or the viewer is its
// voter (their own prompt answers it).
export function councilVoteRequest(
  c: PendingChoiceView | null | undefined,
  viewerID: string | null,
  seats: Pick<PlayerView, "id" | "name">[],
): DockRequest | null {
  const v = c?.council_vote;
  if (!c || !v || c.chooser === viewerID) return null;
  const voter = seats.find((s) => s.id === c.chooser)?.name ?? "A player";
  return voteRequest({
    vote: {
      id: c.id,
      topic: c.reason ?? "",
      options: v.options,
      initiator: v.controller,
      ballots: {},
    },
    viewerID,
    seats,
    tally: v.tally,
    label: "vote",
    detail: `${voter} is voting`,
  });
}

// ---- the game's end ---------------------------------------------------------
//
// ADR 0111 §1, rule 4: once the game has ended, Back to lobby is the
// primary. The banner (who won, and how) stays in the attention strip.
// No Enter: leaving the table is not a confirm of something picked.
//
// GAME_OVER_GRACE_MS (#2919): how long Back to lobby stays disabled after
// the game ends. The request takes the bar, so the button lands in the
// corner `next` held a moment ago, and a click meant for `next` (the pass
// that resolved the lethal blow, a double click) would otherwise take the
// player off the table. `armed` is false until the grace has passed.
export const GAME_OVER_GRACE_MS = 1500;

export function gameOverRequest(onBack: () => void, armed = true): DockRequest {
  return {
    rank: "gameOver",
    label: "game over",
    // The press is refused too, not only the button disabled: a click
    // already dispatched at the old corner must not get through.
    primary: {
      id: "back",
      label: "Back to lobby",
      disabled: !armed,
      onPress: () => {
        if (armed) onBack();
      },
    },
    secondary: [],
  };
}
