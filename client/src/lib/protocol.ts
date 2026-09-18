// Mirror of the v0 protocol types from server/internal/protocol/protocol.go
// and server/internal/protocol/view.go. When the spec in docs/protocol.md
// evolves, update both sides in lockstep.

export const PROTOCOL_VERSION = 0;

export type Kind = "ping" | "pong" | "error" | "action" | "snapshot" | "chat";

export const ErrorCode = {
  BadVersion: "bad_version",
  BadJSON: "bad_json",
  BadRequest: "bad_request",
  Internal: "internal",
  // S15 sub-PR 3: the strict-mode mana-cost gate rejected a
  // cast_spell. The error payload carries `missing` (brace-formatted
  // unpaid symbols) and `card_id` (the rejected cast's instance) so
  // the client can render an "Override strict mode for this cast"
  // toast that re-fires the action with `force_cast: true`.
  InsufficientMana: "insufficient_mana",
  // #705 (ADR 0045 addendum Decision 8): a declare_blocker was refused
  // by the server's block-legality check. `message` is a server-built
  // sentence to show verbatim, `reason` the stable token and `card_id`
  // the blocker. The client never re-derives the rule.
  IllegalBlock: "illegal_block",
} as const;

export type ErrorCodeValue = (typeof ErrorCode)[keyof typeof ErrorCode];

export interface Frame<P = unknown> {
  v: number;
  kind: Kind;
  id: string;
  payload?: P;
}

export interface PingPayload {
  msg?: string;
}

export interface PongPayload {
  msg?: string;
  server_time: string;
}

export interface ErrorPayload {
  code: string;
  message: string;
  // S15: brace-formatted symbols the caster's pool can't cover
  // ("{R}", "{1}"). Populated when code === "insufficient_mana".
  missing?: string[];
  // S15: instance ID of the card whose cast was rejected. Lets the
  // client correlate the toast with the cast UI without keeping an
  // in-flight map. #705: the refused blocker on an illegal_block frame.
  card_id?: string;
  // #705: why a block was refused, populated when
  // code === "illegal_block". One of BlockRefusalReason.
  reason?: BlockRefusalReason;
}

// BlockRefusalReason mirrors game.BlockReason (server/internal/game/
// block_legality.go): the tokens the server sends today. Stable once
// shipped; new ones join in the change that first sends them.
export type BlockRefusalReason =
  | "cant_block"
  | "cant_be_blocked"
  | "flying"
  | "landwalk"
  | "fear"
  | "intimidate"
  | "shadow"
  | "horsemanship"
  | "skulk";

// ActionType is the string-literal union of every action name this
// client sends. Each literal is validated against the server's
// registry — the Type constants in server/internal/actions/actions.go
// plus the hub-level "undo" verb (server/internal/ws/hub.go) — so a
// typo'd action name is a compile error here instead of a runtime
// bad_request frame. The server accepts more action types than these
// (play_card, advance_step, set_poison, …); add literals as the UI
// grows call sites for them.
export type ActionType =
  | "activate_ability"
  | "activate_loyalty"
  | "activate_mana_ability"
  | "add_counter"
  | "add_player_counter"
  | "cast_spell"
  | "cast_vote"
  | "change_life"
  | "clear_combat"
  | "concede"
  | "counter_ability"
  | "counter_spell"
  | "declare_attacker"
  // Bulk attacking-set declaration (#318) — one action, one undo
  // entry, one broadcast, however wide the board.
  | "declare_attackers"
  | "declare_blocker"
  | "discard_selection"
  | "draw_card"
  | "end_vote"
  | "keep_hand"
  | "mark_damage"
  | "move_card"
  | "mulligan"
  | "pass_priority"
  | "pass_turn"
  | "resolve_choice"
  | "sacrifice_permanent"
  | "set_goaded"
  | "set_initiative"
  | "set_monarch"
  | "set_promise"
  | "set_undo_limit"
  | "shuffle_library"
  | "start_vote"
  | "tap"
  | "undo"
  | "untap"
  | "untap_all";

// ActionPayload is sent by the client to mutate game state. See
// docs/protocol.md for the full catalog of action types and their
// params shapes.
export interface ActionPayload {
  type: string;
  player?: string;
  params?: unknown;
}

// SnapshotPayload is the server's authoritative view of the game,
// broadcast after every successful action and sent once on connect.
export interface SnapshotPayload {
  seq: number;
  game: GameView;
}

// ChatPayload is the body of a Kind == "chat" frame in either
// direction. When the client sends one, only `text` is honoured —
// the server stamps `author_id`, `author_name`, and `timestamp` from
// the connection's principal before broadcasting. Spectator chat
// arrives with `author_id` empty and `author_name` == "spectator".
export interface ChatPayload {
  author_id?: string;
  author_name: string;
  text: string;
  timestamp: string;
  // kind classifies the line. Absent or "say" for anything a human
  // typed. The bot kinds are server-originated (S31 sub-PR 8) and
  // cannot be forged from a client — handleChat re-stamps every
  // authoritative field, so a player who types one gets "say".
  kind?: ChatKind;
  // reason carries the bot policy's rationale, split out of `text`
  // so the announcement can be shown while the reasoning stays
  // behind settings.gameplay.showBotReasoning.
  reason?: string;
}

// ChatKind mirrors protocol.ChatKind* in
// server/internal/protocol/protocol.go.
export type ChatKind = "say" | "bot_improvisation" | "bot_reasoning";

// A bot disclosing that it applied an effect by hand because the
// rules engine can't run the card. Never hidden: an unannounced
// improvisation is a bot cheating (ADR 0033 §8).
export const CHAT_BOT_IMPROVISATION = "bot_improvisation";
// A bot narrating why it chose a move. Debug output, hidden unless
// the viewer turns on "show bot reasoning".
export const CHAT_BOT_REASONING = "bot_reasoning";

// View types mirror server/internal/protocol/view.go.

export interface GameView {
  id: string;
  state: "lobby" | "active" | "ended";
  seats: PlayerView[];
  battlefield: ZoneView;
  stack: ZoneView;
  exile: ZoneView;
  turn: TurnView;
  // True between Start and the moment all seated, non-eliminated
  // players have committed to their opening hand via the keep_hand
  // action. Drives the keep / mulligan dialog and gates the normal
  // game UI. Added in S08.
  mulligans_open: boolean;
  // Player ID of the current monarch (Conspiracy mechanic). Empty /
  // omitted when no monarch is set. Sandbox marker; the must-attack
  // and combat-damage transfer rules are not enforced server-side.
  // Added in S10.
  monarch?: string;
  // Player ID currently holding the initiative (BG3 mechanic). Empty
  // when unassigned. Same sandbox posture as monarch. Added in S10.
  initiative?: string;
  // Per-pair "I owe you" promise tally as "{from}->{to}" string keys
  // → count. Sparse: zero entries are dropped server-side. Added in
  // S10.
  promises?: Record<string, number>;
  // Currently open council's-dilemma / politics vote, or omitted when
  // none is in progress. Added in S10.
  vote?: VoteView;
  // Per-player per-turn undo budget. Refreshed on each player's untap
  // step. Drives the "undos: N" indicator and gates the undo button.
  // Added in S11.
  undo_limit?: number;
  // Seat index that took the first turn. Used by the server to enforce
  // the CR 103.8a turn-1 skip-draw rule — in two-player games only,
  // since CR 103.8c has nobody skip at a larger table. Added in S13.
  // Pre-S13 replays
  // decode as 0 (Go's int zero), which matches the only seat games
  // ever started on before the field existed.
  starting_seat?: number;
  // Stack item announce-time metadata (S13.1). Indexed bottom..top —
  // the corresponding spell card (if any) lives in `stack` at the
  // same index. Empty when the stack is empty.
  stack_items?: StackItemView[];
  // APNAP-ordered queue of triggered abilities waiting to hit the
  // stack (S13.1, CR 603.3b). Empty when no triggers are pending.
  pending_triggers?: StackItemView[];
  // Queue of CR 603.7 delayed triggered abilities still owed — "at
  // the beginning of the next end step, return that card to the
  // battlefield" (S22). Public information: the ability was
  // announced when its source resolved and the cards it names sit in
  // the shared exile zone. Empty when nothing is pending.
  delayed_triggers?: DelayedTriggerView[];
  // Mirrors `Game.SplitSecondActive` — true while any item with
  // split second is on the stack (S13.1, CR 702.61). Drives the
  // client's "no responses allowed" UI gating.
  split_second_active?: boolean;
  // Cleanup-step pause map (S13.4, CR 402.2). Keys are player UUID
  // strings, values are the count each player must discard. Drives
  // DiscardPromptModal. Empty / absent when nobody owes discard.
  discard_pending?: Record<string, number>;
  // S14 generic "someone needs to pick" queue. Drives ChoicePromptModal
  // for effects like Thoughtseize where the chooser isn't the
  // discarder. Each entry carries its own options[] already filtered
  // per the viewer's visibility.
  pending_choices?: PendingChoiceView[];
  // S31: the closed list of moves THE VIEWER'S OWN SEAT may make
  // right now, enumerated server-side by `internal/legal`. The server
  // never ships another seat's list — an opponent's moves name the
  // cards in their hand — so this is always "mine".
  //
  // Absent means NO INFORMATION, not "nothing is legal": the field is
  // omitted whenever the seat owes no decision, and an older server
  // omits it entirely. Client predicates stay permissive when it is
  // missing and let the server do the rejecting.
  legal_moves?: LegalMoveView[];
  // The public game log (S31 sub-PR 0, ADR 0033 §4): the last ~200
  // table-visible events, oldest first. Every card reference in it has
  // been through the same visibility filter as the zones above, so an
  // entry naming a card is an entry this viewer is entitled to see
  // named. Absent on a game that has produced no events yet.
  log?: LogEvent[];
  // S22 broadcast reveals (CR 701.20): the cards players have shown
  // the WHOLE TABLE this turn, oldest first, at most 4 of them.
  //
  // The one field on this view that is byte-identical for every seat.
  // Everything else here has been through a per-viewer filter; a
  // reveal has not, because a reveal one seat could not see would not
  // be a reveal. Absent when nothing has been revealed this turn.
  //
  // Note what is NOT here: instance IDs. Neither the revealed cards'
  // nor the revealing card's. A reveal makes a card public for that
  // moment and to that extent — it does not hand out a handle that
  // outlives the moment, because the cards usually go straight back
  // into a hidden zone. Render from the printed identity.
  reveals?: RevealView[];
  // #628 (CR 726): set when the engine has watched the same triggered
  // ability resolve 25 times this turn with no player decision in
  // between, absent otherwise. Its presence is an instruction to this
  // client: STOP PASSING AUTOMATICALLY. Priority still rotates and
  // every pass the server is handed still works — the point is that a
  // person has to ask for the next iteration, with the loop's trigger
  // still on the stack. Table-wide and identical for every seat.
  loop_notice?: LoopNoticeView;
}

// LoopNoticeView is the CR 726 loop breaker's notice. `label` is the
// repeating ability's stack label, which by catalog convention reads
// "<card> — <what happens>", so it is the whole banner line.
// Mirrors `protocol.LoopNoticeView`.
export interface LoopNoticeView {
  source?: string;
  label: string;
  controller?: string;
  count: number;
}

// RevealView is one reveal: the cards a player showed the whole table
// at one moment, and why. Mirrors `protocol.RevealView`.
export interface RevealView {
  // Engine sequence number of the reveal's first event. Monotonic and
  // stable across frames — the key to dedupe and order on. It names an
  // event, not a card.
  seq: number;
  // Turn the reveal happened on. Present so a client that has just
  // reconnected can tell a live announcement from backlog.
  turn?: number;
  // Seat index of the player who revealed, or -1.
  seat: number;
  // NAME of the card whose effect revealed — never its instance ID.
  source?: string;
  // One-line label written by the effect, for the banner.
  reason?: string;
  // Zone the cards were revealed out of ("library", "hand"). The cards
  // did not move; a reveal is not a zone change.
  from?: string;
  // The revealed cards, in the order the table saw them. Truncated to
  // 8 — compare against `count`.
  cards: RevealedCardView[];
  // How many cards the reveal actually showed. Larger than
  // cards.length when the cap bit (Hermit Druid reveals a whole
  // library); render the difference as "+N more".
  count?: number;
}

// RevealedCardView is one card a reveal showed. Deliberately not a
// CardView: there is nowhere here to put an instance ID, which is the
// point. `scryfall_id` identifies a PRINTING, the same way the name
// does, so cardImageURL works from it without any handle on this
// instance of the card.
export interface RevealedCardView {
  name: string;
  mana_cost?: string;
  type_line?: string;
  scryfall_id?: string;
}

// LegalMoveView mirrors `legal.Move` server-side (ADR 0033 §1): one
// fully-specified thing the viewer's seat may do right now.
//
// `params` is EXACTLY the ActionPayload params that perform the move,
// so `{ type, player, params }` can be sent to the server unaltered
// and is contractually guaranteed to be accepted.
//
// Two caps apply, and both matter to anything reading this list.
// Target and mode expansion is capped at 12 moves per source card,
// and the whole list is capped at 48 — past which the server keeps
// one move per (source, kind) and drops the alternatives. The
// invariant you may rely on is "every card with a legal move has at
// least one entry here". Do NOT read it as the complete set of legal
// targets; that is what CardView.legal_targets is for.
export interface LegalMoveView {
  type: string;
  player: string;
  params?: Record<string, unknown>;
  kind: "pass" | "land" | "cast" | "activate" | "mana" | "attack" | "block" | "choice" | "mulligan";
  label: string;
  // Instance ID of the card the move is about, when there is one.
  // Moves with no card (pass_priority, keep_hand, mulligan) carry the
  // nil UUID rather than omitting the key — Go's omitempty does not
  // apply to a UUID array — so join on equality with a real instance
  // ID and never on presence.
  source?: string;
  // True on a move the server cannot refuse whatever else happens
  // between this frame and the click: passing priority, and a pending
  // choice's one unconditional answer (a search's "fail to find").
  // Absent means "legal right now", which is what every move in this
  // list already promises. Automated seats use it as the way out of a
  // prompt whose other answers keep being rejected.
  always_legal?: boolean;
  // What the move charges beyond its mana, in the components `params`
  // cannot name — the ones the engine reads off the ability rather
  // than off the payload. Absent for the overwhelming majority of
  // moves, which charge nothing but mana and the choices `params`
  // already lists.
  //
  // Advice about the move, NOT part of it: `params` stays exactly the
  // ActionPayload that performs the move, and this key is never sent
  // back.
  cost?: MoveCost;
}

export interface MoveCost {
  // Life paid at announce (CR 119.4) — Necropotence's 1,
  // Griselbrand's 7. Always positive when present, and payable: the
  // server only offers a cost the seat can meet. Note that "payable"
  // includes paying your last point, which is legal and lethal.
  life?: number;
  // A loyalty ability's counter delta (CR 606.4), signed as printed:
  // +1 adds one, -3 removes three.
  loyalty?: number;
  // #625: counters a "remove N counters" cost takes, and from which
  // permanent — often not the move's source (Heart of Kiran's crew
  // paid with a planeswalker's loyalty).
  counters?: { card_id: string; counter: string; n: number }[];
}

// LogKind mirrors `protocol.LogKind` server-side. Coarser than the
// engine's own event kinds — several engine events collapse into one
// line a player would read.
export type LogKind =
  | "step"
  | "cast"
  | "resolve"
  | "fizzle"
  | "counter"
  | "zone"
  | "draw"
  | "life"
  | "damage"
  | "attack"
  | "block"
  | "token"
  | "sacrifice"
  | "eliminated"
  // A player revealed cards (CR 701.20): one entry per reveal, however
  // many cards it showed. Never carries card_id; `amount` is the card
  // count, `old_zone` where they were revealed from, and `target_seat`
  // is set when the reveal was to one player only, in which case the
  // text names no card for anyone.
  | "reveal"
  // S30 random effects: the server-rendered public outcome of a die
  // roll or coin flip. The event text is already redacted and ready
  // for both the log and the attention strip.
  | "roll"
  | "flip";

// LogEvent mirrors `protocol.LogEvent` — one line of the public game
// log. `text` is the rendered, already-redacted sentence; the
// structured fields are the same facts for code that wants them
// (highlighting the card an entry names, filtering by seat).
export interface LogEvent {
  // Engine event sequence number. Monotonic, stable across frames —
  // use it as the keyed-each key.
  seq: number;
  kind: LogKind;
  // Turn number the entry happened on. Absent (0) for entries that
  // precede the first step announcement of a game.
  turn?: number;
  // Step name — present ONLY on `step` entries. Everything after a
  // step entry belongs to that step until the next one.
  step?: string;
  // Seat index of the responsible player, or -1 when the event has no
  // single actor.
  seat: number;
  // Seat index of the player the entry acts on (who was attacked, who
  // took the damage). Absent when the entry targets a card or nothing.
  target_seat?: number;
  // Instance ID of the card the entry is about. Absent when the entry
  // deliberately carries no card reference — a draw, or a zone change
  // with hidden zones at both ends.
  card_id?: string;
  // Instance ID of the CARD the entry acts on (a counterspell's
  // victim, a blocker's attacker). Player targets use target_seat.
  target?: string;
  // Signed / counting payload: life delta, damage dealt, cards drawn.
  amount?: number;
  old_zone?: string;
  new_zone?: string;
  // True when a `damage` entry is combat damage (CR 510).
  combat?: boolean;
  // Which combat damage step dealt a combat `damage` entry (CR 510.4).
  // Set only when that combat had a first-strike step; combat with no
  // first strike or double strike anywhere is untagged, so its
  // presence alone means there are two beats. Read it, don't derive it
  // from keywords (#187, ADR 0053 Decision 1).
  combat_step?: "first_strike" | "regular";
  // The rendered line. Already redacted for this viewer: a card the
  // viewer may not identify reads as "a card".
  text: string;
  // Random-effect details. These are optional so older log entries
  // and future effect families remain wire-compatible.
  sides?: number;
  results?: number[];
  faces?: string[];
  call?: "heads" | "tails";
  wins?: number;
}

// PendingChoiceView mirrors `protocol.PendingChoiceView` server-side.
// Drives the client picker modal for deferred-choice effects
// (Thoughtseize, future Vendilion-style cards). The Options slice
// is pre-filtered to what the viewer is legally allowed to see.
export interface PendingChoiceView {
  id: string;
  kind:
    | "discard_from_hand"
    | "mana_pick"
    | "replacement_order"
    | "optional_replacement"
    | "damage_assignment"
    | "trigger_prompt"
    | "pay_unless"
    | "trigger_order"
    | "pick_target"
    // S21: "each player sacrifices a creature of their choice"
    // (Grave Pact, Fleshbag Marauder). Answered with the generic
    // {choice_id, card_ids} payload — one entry — and rendered by the
    // shared card grid with sacrifice copy.
    | "sacrifice_choice"
    // S21: scry N (CR 701.22). Options carries the looked-at cards
    // top-first, redacted to the chooser alone — scry is "look at",
    // not "reveal". Answered with {bottom, top_order}.
    | "scry"
    // S22: surveil N (CR 701.25). Scry's frame with the
    // bottom-of-library leg replaced by the graveyard — same
    // chooser-only redaction on options, same top-first ordering.
    // Answered with {graveyard, top_order}, NOT {bottom, top_order}:
    // the server routes on which key is present, so sending scry's
    // payload for a surveil would bury the cards instead of binning
    // them.
    | "surveil"
    // S22: "look at the top N cards of your library, then put them
    // back in any order" (Ponder, Sensei's Divining Top). The family's
    // third member and the one with no away lane — answered with
    // {top_order} alone, naming every looked-at card exactly once.
    | "look_at_top"
    // Shocklands: "as this land enters, you may pay 2 life. If you
    // don't, it enters tapped." Answered with the shared yes/no
    // {choice_id, apply} payload — apply=true pays and the land
    // enters untapped. The permanent has NOT entered while this
    // prompt is open; the answer is what decides how it enters.
    // pay_cost carries the payment ("2 life").
    | "entry_pay_life"
    // S22: "search your library for ..." (CR 701.23). Options carries
    // the matching cards, sent ONLY to the chooser — a library is a
    // hidden zone and even the number of matches is private. Answered
    // with the generic {choice_id, card_ids} payload; an empty list
    // is a legal "fail to find" (CR 701.23b), so search_max is the
    // ceiling and the floor is zero.
    | "search_library"
    // S28 cascade (CR 702.85): "you may cast it without paying its
    // mana cost". Answered with the shared yes/no {choice_id, apply}
    // payload — apply=true takes the offer, and the server stamps a
    // free-cast permission on the card so it can be cast out of exile
    // this turn. `options` carries the one card being offered, so the
    // prompt shows the card rather than naming it in a sentence.
    | "may_cast"
    // S16.5: "you may have this creature enter as a copy of ..."
    // (Clone, Phyrexian Metamorph, Spark Double, Sakashima the
    // Impostor). Options carries the permanents that may be copied —
    // all public battlefield cards, so nothing is redacted. Answered
    // with the generic {choice_id, card_ids} payload; an EMPTY list
    // declines, because every printed card in this class says "you
    // may". The permanent has NOT entered while the prompt is open:
    // the answer decides what it enters AS, which is why its own ETB
    // trigger has not fired yet either.
    | "copy_target"
    // S26: "as this permanent enters, choose a creature type" (CR
    // 614.12) — Cavern of Souls, Door of Destinies, Vanquisher's
    // Banner, Adaptive Automaton. type_options carries the whole CR
    // 205.3m vocabulary for the picker to filter; answered with
    // resolve_choice { creature_type: "Elf" }.
    | "choose_creature_type"
    // #74 chained choices: the general two-way prompt, "do A, or do
    // B." Answered with the shared yes/no {choice_id, apply} payload —
    // apply=true takes the accept branch. accept_label / decline_label
    // are the card's own words for the two consequences ("Pay 4 life"
    // / "Put it on top"); absent means it really is a yes/no.
    //
    // It is what lets one prompt follow another: a card queues it from
    // inside the answer to an earlier prompt, so Ponder can ask "you
    // may shuffle" only AFTER the reorder, and Sylvan Library can ask
    // about the second card only after the first has been paid for.
    | "confirm"
    // #74 chained choices: the general card-set pick, "choose N of
    // these cards", with the card deciding what being chosen means.
    // Answered with the generic {choice_id, card_ids} payload, bounded
    // by choose_min / choose_max. Options and bounds reach the CHOOSER
    // ONLY — the candidates are usually cards in a hand, and their
    // number is as private as their faces.
    | "choose_cards"
    // #742: "choose a color" (CR 105.4) — as a permanent enters
    // (Coldsteel Heart, the Thriving lands; the answer is remembered on
    // the permanent) or while a spell resolves (Wash Out). color_options
    // carries the legal colours, a subset of W/U/B/R/G ("a color other
    // than blue" is four). Answered with the same {choice_id, color}
    // payload a mana_pick uses; the server routes the two by kind.
    | "choose_color"
    // S30 coin call: choose heads or tails for the pending flip. A
    // stop answer is offered only when allow_stop is true.
    | "coin_call"
    | string;
  chooser: string;
  from_player: string;
  count: number;
  source?: string;
  reason?: string;
  options?: CardView[];
  // S15: populated for kind "mana_pick" — the legal color buttons
  // the chooser's picker modal should render. Uppercase single-
  // character values ("W", "U", "B", "R", "G", "C"). Ordered server-
  // side with the chooser's commander colour identity first; render in
  // the order sent. Full 5-color for Birds of Paradise ("G" first in a
  // mono-green deck); narrowed to the identity only for Arcane Signet
  // and the other cards whose text says "in your commander's color
  // identity".
  color_options?: string[];
  // #742: on a "mana_pick" that adds more than one mana of the picked
  // colour ("{T}: Add three mana of any one color") — colour letter to
  // amount. A colour missing from the map adds one; absent on ordinary
  // picks. Also, choose_color reuses color_options above.
  color_amounts?: Record<string, number>;
  // S26: populated for kind "choose_creature_type" — every creature
  // type the engine knows, sorted. The list is long by design (the CR
  // 205.3m vocabulary is ~345 entries), so the picker filters it
  // rather than rendering it whole.
  type_options?: string[];
  // S17: populated for kind "replacement_order" — the CR 616
  // affected-player-chooses-order prompt. Client renders a drag-
  // reorder list of these entries and submits the IDs in the
  // chosen order as resolve_choice { order: string[] }. Absent
  // for other kinds. Wire modal lands in sub-PR 3 alongside
  // Doubling Season + Hardened Scales.
  replacement_options?: ReplacementOptionView[];
  // S19 sub-PR 8: populated for kind "trigger_order" — the CR 603.3b
  // "you choose the order of your simultaneous triggers" prompt.
  // Same {id, label, source_card_id} shape as replacement_options;
  // the client submits resolve_choice { order: string[] } of these
  // IDs in RESOLUTION order (first entry resolves first).
  trigger_options?: ReplacementOptionView[];
  // S20 sub-PR 2: populated for kind "pick_target" — a targeted
  // trigger's controller chooses its target. The client enters the
  // board-click targeting flow with this legal set (no modal) and
  // answers resolve_choice { target: {kind, id} }.
  pick_target?: LegalTargetsView;
  // S18: populated for kind "damage_assignment" — the CR 510.1c
  // multi-blocker combat damage prompt. Client renders a per-blocker
  // damage input panel (+ trample-to-player input when allow_trample
  // is set) and submits resolve_choice with
  // { assignments: [{blocker_id, amount}, ...], trample_to_player }.
  damage_assignment?: DamageAssignmentView;
  // S19 follow-up: populated for kind "trigger_prompt" — true when
  // the optional trigger has no legal target and answering "Yes"
  // will pass without effect (Reclamation Sage with no opponent
  // artifact, Eternal Witness with an empty graveyard, etc.). The
  // modal warns the chooser. Absent/false otherwise.
  no_legal_target?: boolean;
  // S19 sub-PR 6: populated for kind "pay_unless" — the cost the
  // chooser is being asked to pay ("{2}"). Same {choice_id, apply}
  // payload as the other yes/no kinds: apply=true pays (from pool,
  // auto-tapping if short), apply=false declines and the card's
  // "unless" consequence fires. Also carries the life payment ("2
  // life") for kind "entry_pay_life".
  pay_cost?: string;
  // S22: populated for kind "search_library" — how many of `options`
  // the searcher may take. The minimum is always zero, so the submit
  // button is live from the first render. Absent for every other
  // kind, and absent for non-chooser viewers, who are not told what
  // the search is for.
  search_max?: number;
  // #74: populated for kind "confirm" — the card's own words for the
  // accept and decline branches. Absent means the client renders Yes /
  // No, which is right for a prompt that really is a yes/no.
  accept_label?: string;
  decline_label?: string;
  // #74: populated for kind "confirm" — the life the ACCEPT branch
  // charges (Sylvan Library's 4). Absent when the branch costs no
  // life. The label already says it; this is the number, for anything
  // that needs to reason about the price rather than print it.
  life_cost?: number;
  // #74: populated for kind "choose_cards" — how few and how many of
  // `options` the chooser must pick. Absent for every other kind, and
  // absent for non-chooser viewers, who are not told the size of a
  // choice over someone else's hidden cards.
  choose_min?: number;
  choose_max?: number;
  // CR 603.2d: when this is a trigger_prompt or pick_target choice,
  // the public permanent that caused the additional trigger. The
  // server omits both fields for ordinary choices.
  doubled_by?: string;
  doubled_by_name?: string;
  // S30 coin call prompt metadata. `coins` is the number of coins
  // covered by one call; wins tracks an ongoing chain.
  allow_stop?: boolean;
  coins?: number;
  max_useful_wins?: number;
  wins?: number;
}

// ReplacementOptionView mirrors protocol.ReplacementOptionView —
// one entry in a replacement_order prompt's candidate list. ID is
// a decimal-string ReplacementEffectID; label is the prompt copy
// ("Doubling Season: double counters"); source_card_id is the card
// hosting the effect (empty for engine built-ins like commander-
// zone replacement). Added in S17 sub-PR 2.
export interface ReplacementOptionView {
  id: string;
  label?: string;
  source_card_id?: string;
}

// DamageAssignmentView mirrors protocol.DamageAssignmentView — the
// wire shape of a CR 510.1c multi-blocker combat damage prompt.
// The attacker's controller distributes attacker_power across the
// blockers (respecting at-least-lethal-in-order) and, if
// allow_trample, can overflow leftover to the defending player.
// Added in S18 sub-PR 3.
export interface DamageAssignmentView {
  attacker_card_id: string;
  blocker_card_ids: string[];
  attacker_power: number;
  allow_trample?: boolean;
  has_deathtouch?: boolean;
}

// DelayedTriggerView mirrors `protocol.DelayedTriggerView`
// server-side: one queued CR 603.7 delayed triggered ability. `at`
// is the step whose beginning fires it ("end", "upkeep"); `cards`
// are the instance IDs the effect acts on. Added in S22.
export interface DelayedTriggerView {
  id: string;
  controller: string;
  source?: string;
  label?: string;
  at: string;
  created_turn?: number;
  cards?: string[];
}

// StackItemView mirrors `protocol.StackItemView` server-side: the
// announce-time metadata for one item on the stack.
export interface StackItemView {
  id: string;
  kind: "spell" | "activated" | "triggered";
  controller: string;
  owner: string;
  source_card_id: string;
  label?: string;
  targets?: TargetRefView[];
  modes?: number[];
  x_value?: number;
  distribution?: Record<string, number>;
  hold_priority?: boolean;
  split_second?: boolean;
  // S22: the alternative cost this spell was cast for — "overload",
  // "evoke", "cleave" — absent for an ordinary cast. Load-bearing
  // for anyone deciding whether to respond: an overloaded Cyclonic
  // Rift is a one-sided wipe, a hard-cast one is a single bounce.
  alt_cost?: string;
  // S30: a CR 707.10 spell copy rather than a cast card. The copy
  // and its source look identical on the stack, and which is which
  // decides what countering one leaves behind.
  is_copy?: boolean;
  // CR 603.2d: public attribution for an additional triggered
  // ability created by a trigger-doubling permanent.
  doubled_by?: string;
  doubled_by_name?: string;
}

// TargetRefView mirrors `protocol.TargetRefView` server-side: a
// single announce-time target slot.
export interface TargetRefView {
  kind: "player" | "card" | "self" | "none";
  id?: string;
}

export interface VoteView {
  id: string;
  topic: string;
  options: string[];
  initiator: string;
  // voter player ID → option index
  ballots: Record<string, number>;
}

export interface PlayerView {
  id: string;
  name: string;
  seat: number;
  life: number;
  poison?: number;
  energy?: number;
  library: ZoneView;
  hand: ZoneView;
  graveyard: ZoneView;
  command: ZoneView;
  commander_damage: Record<string, number>;
  // Rolling per-player life-change log. Bounded server-side at
  // MaxLifeHistoryEntries (currently 50). Always emitted as an array
  // (possibly empty); each entry is server-stamped at the moment of
  // change. Added in S08.
  life_history: LifeChangeView[];
  // Set when the player has conceded (S08) or, in the future, lost
  // to a state-based action. Eliminated players still appear in the
  // seats list and may continue to spectate; the UI greys them out
  // and disables their action buttons. Omitempty on the wire — only
  // present when true.
  eliminated?: boolean;
  // True once the player has committed to their opening hand during
  // the mulligan window. Omitempty on the wire — absent means false.
  // Added in S08.
  hand_kept?: boolean;
  // Number of mulligans this player has taken in the current
  // opening-hand window. Omitempty on the wire — absent means 0.
  // Added in S08.
  mulligans_taken?: number;
  // True once the player has had a real deck installed via the
  // POST /games/{id}/decks endpoint. Absent / false means the seat
  // is still using the placeholder commander handed out at join
  // time. Drives the in-game deck-import modal. Added in S08.5.
  deck_imported?: boolean;
  // Per-turn undo budget left for this seat, refreshed on entering
  // their untap step. Drives the per-seat indicator on the toolbar.
  // Added in S11.
  undos_remaining?: number;
  // Discord identity (S12.5). Populated when the seat was claimed
  // via OAuth; absent for manual-name joins. The client builds
  // /avatars/{discord_id}/{discord_avatar_hash}.png to pull the
  // cached portrait, and prefers display_name over name for the
  // seat label.
  discord_id?: string;
  discord_avatar_hash?: string;
  display_name?: string;
  // Bot seat (S31, ADR 0033). is_bot marks a seat driven by a
  // server-side policy runner rather than a WebSocket client;
  // bot_tier is its difficulty tier ("random", "heuristic", …) and
  // bot_deck the curated deck it was seated with. The board renders
  // a BOT chip and a distinct avatar mark off these, and shows the
  // thinking pulse while such a seat holds priority.
  is_bot?: boolean;
  bot_tier?: string;
  bot_deck?: string;
  // Per-commander cast count for the Commander tax (S13.1, CR
  // 903.8). Keyed by commander instance UUID. Drives the "+N tax"
  // indicator next to the commander tile.
  commander_casts?: Record<string, number>;
  // Per-player named counter map (S13.2 — poison, energy,
  // experience, rad, plus homebrew). The legacy `poison` and
  // `energy` ints above stay populated for backwards compat.
  counters?: Record<string, number>;
  // Per-player cleanup-step hand-size cap (S13.4, CR 402.2).
  // 7 by default; -1 = no cap (Reliquary Tower / Thought Vessel).
  // Always present on the wire; the field is non-omitempty so
  // clients know the cap even when it's the default.
  max_hand_size?: number;
  // S15: per-player mana pool. Each entry is an uppercase mana
  // letter ("W", "U", "B", "R", "G", "C") — order reflects
  // insertion order so the UI can highlight the most recent add.
  // Empties at every step boundary (CR 106.4), so this is absent
  // / empty in the common case outside an active cast sequence.
  mana_pool?: string[];
}

export interface LifeChangeView {
  delta: number;
  new_total: number;
  at: string; // RFC3339
}

export interface ZoneView {
  kind: string;
  owner?: string;
  count: number;
  cards: CardView[];
}

// ModeSpecView is the wire shape of a modal spell's options
// (S20 sub-PR 4). Indexes into `options` ride cast_spell as `modes`.
export interface ModeSpecView {
  prompt: string;
  min: number;
  max: number;
  options: ModeOptionView[];
}

// AdditionalCostView is the "As an additional cost to cast this
// spell, discard a card" clause on a card in the viewer's own hand
// (S21 sub-PR 5). The picked instance IDs ride cast_spell as
// `discard_ids`; the server rejects a cast that arrives without
// them.
export interface AdditionalCostView {
  discard_cards?: number;
  // S21 sub-PR 6: the permanents that may pay a "sacrifice a
  // creature" clause, already filtered to the caster's own board.
  // The picked IDs ride cast_spell as `sacrifice_ids`.
  // Present-and-empty means the cost is unpayable, so the spell is
  // uncastable. #747: min / max are the clause's count ("sacrifice
  // two creatures" is 2 / 2) and the cards come in payment order —
  // see sacrificeCost.ts.
  sacrifice_options?: LegalTargetsView;
  // S23: a "pay X life" clause (Toxic Deluge). The X prompt has to
  // open for this card even though its printed mana cost has no {X},
  // and the announced X is both the life paid and the number the
  // spell's own text uses.
  demands_x?: boolean;
  label?: string;
}

// AlternativeCostView is one "you may cast this spell for its
// overload / evoke / cleave cost" offer on a card in the viewer's own
// hand (S22). Unlike an additional cost this is optional: the picker
// lists these alongside "pay the printed cost", and a cast that names
// none is the ordinary case. `key` rides cast_spell as
// `alternative_cost`.
export interface AlternativeCostView {
  key: string;
  label?: string;
  mana_cost?: string;
  // The target clause the spell has WHEN THIS COST IS PAID, already
  // resolved server-side against the cost's rewrite. Absent means the
  // spell has no targets under this cost (overload), so the cast
  // fires straight away.
  target_mode?: string;
  legal_targets?: LegalTargetsView;
  // S28: the "pay N life" half of the cost (Force of Will's 1, Snuff
  // Out's 4). Absent for the costs that charge none. The server
  // enforces the life total; this is for the label.
  life?: number;
  // S28: the cards that can pay the cost's card-shaped half — the
  // blue cards in your hand for Force of Will, the Islands you
  // control for Daze. The chosen instance ID rides back on cast_spell
  // as `alt_cost_ids`. Absent when the cost charges no cards (every
  // S22 keyword); present-and-empty means you have nothing that can
  // pay, so the offer is visible but unusable.
  pay_options?: LegalTargetsView;
  // S28: the picker's prompt copy for `pay_options` ("a blue card").
  pay_label?: string;
  // CR 107.3b (#831): the card prints an {X} in its mana cost and
  // this offer does not, so claiming it fixes X at 0 — the cast flow
  // skips the X picker and sends nothing. Absent for nearly every
  // offer, including one priced with an {X} of its own.
  x_locked_at_zero?: boolean;
}

// TapCostView is the "tap permanents you control to help pay for
// this" clause convoke and waterbend share, on a card in the
// viewer's own hand (S22). Like an alternative cost it is an OFFER —
// tapping nothing and paying the whole cost with mana is always
// legal — but unlike one it replaces nothing: it spends against a
// cost that is still owed. The picked instance IDs ride cast_spell
// as `tap_ids`.
export interface TapCostView {
  // "convoke" or "waterbend".
  key: string;
  // The clause as printed ("Convoke", "Waterbend {X}").
  label?: string;
  // The untapped permanents that may be tapped, already filtered to
  // the viewer's own board. Present-and-empty is not an error: the
  // caster simply has nothing to tap and pays in mana.
  options?: LegalTargetsView;
  // Convoke's "or one mana of that creature's color", shown in the
  // picker's hint. Absent for waterbend, where each permanent pays
  // exactly {1}.
  color_clause?: boolean;
  // How many permanents may be tapped — the mana value of what the
  // cast owes. 0 means "as many as the announced X", which is how a
  // waterbend {X} cost arrives (its size isn't chosen yet when the
  // snapshot is built).
  max?: number;
  // The keyword's own cost carries an {X} the caster must announce,
  // on a card whose PRINTED cost has none. Waterbender's Restoration
  // costs {U}{U} and still needs the X prompt.
  demands_x?: boolean;
}

// ExilePlayView is the impulse-exile grant on a card in exile —
// "exile the top card of that player's library, you may play it
// this turn" (S21 sub-PR 6). Public information; the client offers
// the action only when `player` is the viewer.
export interface ExilePlayView {
  player: string;
  // Casting only — a land under this grant is stranded (Ragavan).
  cast_only?: boolean;
  // "Spend mana as though it were mana of any color" (Breeches).
  any_color?: boolean;
  // S22 airbend: the mana cost the holder pays INSTEAD of the card's
  // printed one ("{2} rather than its mana cost"). Absent for
  // impulse exile, which charges the printed cost. Note that
  // `mana_cost` on the card still carries the printed value.
  cost_override?: string;
  // S29 warp: the earliest turn number the grant is live on — "you
  // may cast it from exile ON A LATER TURN". Absent for every grant
  // that is live as soon as it is made, which is all of impulse
  // exile and airbend.
  not_before_turn?: number;
  // S32: the printed face this grant opens, when it opens one —
  // a defeated Siege's "exile it, then cast it transformed", where
  // the card sitting in the exile pile still shows the battle and
  // the thing the button casts is `faces[face]`. Absent for every
  // grant that does not speak about faces (impulse exile, airbend,
  // warp, cascade), which is all of them before S32.
  //
  // Advisory only: the server settles the face from the grant rather
  // than from the request, so a client that ignores this labels the
  // button with the wrong name but cannot cast the wrong half.
  face?: number;
  // CR 107.3b (#831): the card prints an {X} in its mana cost and
  // this grant's price does not, so casting under it fixes X at 0 —
  // what a cascade hit carries. The cast flow skips the X picker.
  // Absent for a grant that charges the printed cost, which still
  // asks.
  x_locked_at_zero?: boolean;
}

// ActivatedAbilityView is one CR 602 activated ability on a
// battlefield permanent (S21 sub-PR 2). Public information, so it
// rides every viewer's snapshot; the client only offers the menu on
// permanents the viewer controls. `index` is what the
// activate_ability payload carries as `ability_index`.
export interface ActivatedAbilityView {
  index: number;
  label?: string;
  tap_cost?: boolean;
  sacrifice_self?: boolean;
  mana_cost?: string;
  life_cost?: number;
  sorcery_speed?: boolean;
  // #743: true while the ability's activation condition (CR 602.1b —
  // "Activate only if an opponent controls four or more lands",
  // "Activate only during your turn") is false. Absent when there is
  // no condition or it holds. Evaluated server-side for the
  // permanent's controller; the menu greys the row like
  // sorcery_speed, and the server refuses the activation regardless.
  condition_unmet?: boolean;
  // loyalty_cost is the +N / 0 / −N of a planeswalker's loyalty
  // ability (CR 606.4). Its PRESENCE, not its value, is what marks
  // the ability as a loyalty ability — 0 is a real printed cost —
  // so test for `!== undefined`, never for truthiness. Added with
  // #329 / #334.
  loyalty_cost?: number;
  // A "Sacrifice a creature"-style cost: the clause, and the
  // permanents the controller can pay it with right now. #747: min /
  // max are the clause's count ("Sacrifice two artifacts" is 2 / 2;
  // always equal) and the cards come in payment order — tokens first,
  // then lower mana value, then the source. See sacrificeCost.ts.
  sacrifice_label?: string;
  sacrifice_options?: LegalTargetsView;
  // S27: a Vehicle's crew cost (CR 702.122a). crew_cost is the
  // number that the tapped creatures' TOTAL POWER must reach;
  // crew_options lists the creatures that could pay it right now —
  // untapped creatures the controller controls, summoning-sick ones
  // INCLUDED, because tapping to crew is not paying a {T} cost.
  //
  // Unlike a sacrifice cost this is a many-pick prompt with a floor
  // rather than a count: any number of creatures is legal as long as
  // the running total reaches crew_cost, and overshooting is fine.
  // The picks ride activate_ability as `crew_ids`.
  crew_cost?: number;
  crew_options?: LegalTargetsView;
  // #625: a "remove N counters" cost component. counter_cost_n is how
  // many, and its presence marks the component.
  //
  //   - counter_cost_kind is the printed kind ("loyalty", "gold");
  //     ABSENT means "a counter" of any kind, and the picker asks for
  //     the kind as well as the permanent.
  //   - counter_cost_self: the counters come off the source itself
  //     ("Remove a gold counter from this artifact"); nothing is sent.
  //   - counter_cost_label: the "from" clause of the other form ("a
  //     planeswalker you control").
  //   - counter_cost_options: what could pay right now, most counters
  //     first — permanents the viewer controls holding enough of the
  //     kind, each with the kinds that could pay. A cost does not
  //     target, so hexproof permanents are included. Absent when
  //     nothing can pay.
  //
  // The choice rides activate_ability as `counter_source_ids` (not for
  // the self form) and `counter_kind` (only for the any-kind form).
  counter_cost_n?: number;
  counter_cost_kind?: string;
  counter_cost_self?: boolean;
  counter_cost_label?: string;
  counter_cost_options?: CounterCostOptionView[];
  // {X} in the ability's mana cost (CR 602.2b) — Helm of Obedience,
  // Treasure Vault, Soothsaying. demands_x opens the X picker before
  // the targeting step, and the answer rides activate_ability as
  // `x_value`.
  //
  // min_x is the floor the printed text puts on the announcement:
  // Helm of Obedience's "X can't be 0" ships 1, and absent means the
  // ordinary floor of zero. x_slots is how many {X} tokens the cost
  // carries — 2 for Treasure Vault's "{X}{X}" — so the picker can
  // say what a given X actually costs without re-parsing the string.
  demands_x?: boolean;
  min_x?: number;
  x_slots?: number;
  // Present when the ability targets. A full LegalTargetsView since
  // #334: the server now stamps the clause's min / max (it always
  // had them; abilityLegalTargets just never copied them across),
  // which is what lets an "up to one target" ability — Teferi's −3,
  // the Emperor's −2 — be confirmed with nothing picked.
  target_mode?: string;
  legal_targets?: LegalTargetsView;
}

// CounterCostOptionView is one permanent that could pay a "remove N
// counters" cost, with the kinds on it that could (#625).
export interface CounterCostOptionView {
  card_id: string;
  kinds: { kind: string; count: number }[];
}

export interface ModeOptionView {
  label: string;
  target_mode?: string;
  legal_targets?: LegalTargetsView;
}

// LegalTargetsView is a clause's legal set right now plus its
// target count (S20 sub-PR 5): min..max picks, max 0 = unbounded.
export interface LegalTargetsView {
  players?: string[];
  cards?: string[];
  min?: number;
  max?: number;
  // S22: the clause's count is the announced X, not a printed
  // constant ("Exile X target creatures you control"). min / max are
  // meaningless until X is chosen, so the picker substitutes the X
  // collected in the cost prompts.
  count_from_x?: boolean;
}

/**
 * One printed face of a multi-face card (ADR 0034). Enough to render
 * a picker row and a hover panel.
 */
export interface CardFaceView {
  name: string;
  type_line?: string;
  mana_cost?: string;
  oracle_text?: string;
  power?: number;
  toughness?: number;
  /** "/cards/{scryfall_id}/image?face=N", built server-side. */
  image?: string;
}

export interface NoUntapView {
  static?: boolean;
  next?: string[];
}

export interface CardView {
  instance_id: string;
  /**
   * The ACTIVE face's name. For the ~33,000 single-faced oracle IDs
   * this is simply the card's name, as it always was; for a modal
   * DFC or a transform card it is the name of the side that is
   * currently up — never Scryfall's "A // B" composite.
   */
  name: string;
  owner: string;
  controller: string;
  scryfall_id?: string;
  // Scryfall printed type line ("Legendary Creature — Human Wizard").
  // Used by the client to filter creature-only UIs (combat panel)
  // and to label cards. Omitted for placeholder demo cards. S08.
  type_line?: string;
  // Effective colors (W/U/B/R/G), including layer-5 changes. Omitted
  // means colorless; clients must never infer colors from mana_cost.
  colors?: string[];
  // Signed power when below zero, for comparisons such as skulk.
  // Otherwise use power, which retains its combat-damage zero clamp.
  negative_power?: number;
  // Parsed printed creature stats. Omitted (zero) for non-creatures
  // and for cards with non-numeric printed stats. S08.
  power?: number;
  toughness?: number;
  tapped?: boolean;
  // Untap-step restriction / one-shot marker state. `static` is omitted
  // for face-down cards; `next` contains player IDs and is public state.
  no_untap?: NoUntapView;
  counters?: Record<string, number>;
  is_commander?: boolean;
  // Damage marked on this creature for the lethal-damage SBA (S13.1,
  // CR 704.5g). Cleaned up in cleanup step (S13.2, CR 514.2). Only
  // meaningful on the battlefield; omitted when zero.
  damage_marked?: number;
  // S13.5 visual face-down flag (CR 708 — morph / manifest /
  // mutate-bottom, Necropotence's exile). Distinct from known_by_you:
  // a viewer who doesn't know a face-down card gets it redacted to
  // game state and Card.svelte draws a back (cardBack.ts, #95); a
  // viewer who knows it still sees the face.
  face_down?: boolean;
  // S13.5 per-viewer knowledge flag. True when the viewer is in the
  // server-side KnownBy set for this card. When false, printed
  // characteristics (name, type_line, scryfall_id, power, toughness,
  // counters, is_commander) are zero/empty.
  known_by_you?: boolean;
  // Normalised battlefield position in [0, 1], stamped by
  // `set_battlefield_position`.
  //
  // Sent for BATTLEFIELD CARDS ONLY, and for all of them (#29). So
  // absent means "not on the battlefield", never "at the origin" —
  // the pair used to be dropped whenever it was (0, 0), which hid a
  // deliberate origin stamp behind the same absence as a card that
  // had never been positioned. That matters for the sort in
  // BattlefieldRow: x=0 is "first in the row", and the server clamps
  // every out-of-range and NaN coordinate onto exactly 0.
  //
  // Still optional on a battlefield card in practice, so keep the
  // `?? 0` fallbacks: replays and bug-report frames captured before
  // this change omit the pair, and an unpositioned permanent reads
  // (0, 0) anyway — the server has no separate "unpositioned" state.
  battle_x?: number;
  battle_y?: number;
  // Player ID this card is currently declared to attack. Omitted
  // when not declared as attacker. Cleared on zone exit and by
  // clear_combat. Added in S08.
  attacking_target?: string;
  // S27: what attacking_target NAMES. An attacker may be declared
  // against a player, a planeswalker or a battle (CR 508.1d), so the
  // id is a seat id or an instance id and this says which. Absent
  // when nothing is declared.
  attacking_target_kind?: "player" | "planeswalker" | "battle";
  // S27: the seat protecting this battle (CR 310.9a). Absent for every
  // other card type and for a battle whose protector prompt has not
  // been answered. Public — it decides who may attack it.
  protector_player?: string;
  // S27: a battle's current defense counter total, lifted out of the
  // counters map the way loyalty is, because it is the card's life
  // total rather than one pip among several.
  defense?: number;
  // Attacker instance ID this card is currently declared to block.
  // Omitted when not declared as blocker. Cleared on zone exit and
  // by clear_combat. Added in S08.
  blocking_target?: string;
  // Player ID who goaded this creature, or omitted when not goaded.
  // Cleared on zone exit. Sandbox marker; must-attack-not-the-goader
  // is not enforced server-side. Added in S10.
  goaded_by?: string;
  // S24 (ADR 0036): the attachment relation for an Equipment or an
  // Aura — the permanent (`kind: "card"`) or player
  // (`kind: "player"`) this card is attached to. Omitted for every
  // card attached to nothing, which is nearly all of them.
  //
  // One direction only: "what is attached to this creature" is
  // derived by partitioning the battlefield on this field, so the
  // two directions cannot disagree.
  attached_to?: TargetRefView;
  // S14: card is in the server's effect catalog — when it resolves
  // (or ETBs), a registered effect fires automatically rather than
  // requiring manual sandbox clicks. Omitted when false so
  // non-catalog cards (the majority) don't carry the field. Drives
  // the gold-leaf "auto" badge on Card.svelte.
  auto?: boolean;
  // The honest inverse of `auto`, and deliberately not !auto. Most
  // cards have no catalog entry and don't need one — printed
  // keywords are enforced for every card in the dump, a vanilla
  // creature is complete, a basic land taps off its type line. This
  // is set only when the card prints rules the engine will not run,
  // which is the case behind reports #321 / #324 / #325 / #332 /
  // #333: five uncatalogued cards resolved into silence and the
  // player had no way to tell that from a defect.
  //
  // Surfaced at the moments a player forms an expectation — the
  // hover/inspect panel and the stack — and NOT as a board badge.
  // Most of a real battlefield would carry one, and a badge on
  // everything is a badge nobody reads.
  unimplemented?: boolean;
  // target_mode tells the cast-click flow what to prompt for at
  // announce time. Empty/absent ⇒ cast immediately with no target.
  // See client/src/lib/targeting.ts for the full enum.
  target_mode?: string;
  // S20: for a card in the viewer's own hand / command zone with a
  // structured TargetSpec — the players and card instance IDs its
  // target slot accepts right now. Absent for free-form cards. Both
  // lists empty = no legal target = not castable right now.
  legal_targets?: LegalTargetsView;
  // S20 sub-PR 4: for a modal card in the viewer's own hand — the
  // "Choose one" clause. Each option carries its label and, when it
  // targets, its own target_mode + legal set. The cast flow shows a
  // mode picker before targeting. Absent for non-modal cards and on
  // opponents' hands.
  modes?: ModeSpecView;
  // S21 sub-PR 5: for a card in the viewer's own hand with an
  // additional cost ("As an additional cost to cast this spell,
  // discard a card"). The cast flow collects the payment before
  // firing cast_spell. Absent for the vast majority of cards.
  additional_cost?: AdditionalCostView;
  // S22: for a card in the viewer's own hand that offers a cost paid
  // INSTEAD of its mana cost — overload, evoke, cleave. The cast flow
  // opens a picker before every other prompt, because the choice
  // changes what the rest of them ask. Absent for nearly every card.
  alternative_costs?: AlternativeCostView[];
  // S22: for a card in the viewer's own hand that lets you tap your
  // own permanents to help pay — convoke and waterbend. The cast
  // flow opens a picker after X and before targeting. Absent for
  // nearly every card.
  tap_cost?: TapCostView;
  // #746 (ADR 0048 addendum): the printed clauses of this card's own
  // cost modifiers whose price depends on its targets — Fireball's
  // "This spell costs {1} more to cast for each target beyond the
  // first", strive. The X picker opens before targeting and its
  // affordability readout is priced at one target, so it shows these
  // clauses under the readout. Absent for nearly every card and on
  // opponents' cards the viewer cannot read.
  target_cost_notes?: string[];
  // S29: set on a card sitting in a zone its own text opens as a
  // cast source — a flashback card in the graveyard. The zone
  // browser keys its cast button off this, the way exile keys its
  // impulse button off `exile_play`. Never set on hand or
  // command-zone cards: those surfaces are cast surfaces for
  // everything in them. The cost to pay rides `alternative_costs`,
  // already filtered to the offers claimable from this zone.
  castable_here?: boolean;
  // S21 sub-PR 6: present on a card in exile that someone may play
  // this turn. Absent for ordinary exile, which is nearly all of it.
  exile_play?: ExilePlayView;
  // S21 sub-PR 2: activated abilities offered by this permanent.
  activated_abilities?: ActivatedAbilityView[];
  // S21 sub-PR 2: CR 302.6 summoning sickness — entered this turn
  // without haste, so it can't attack or pay a {T} cost.
  summoning_sick?: boolean;
  // CR 606.3: a loyalty ability has already been activated on this
  // planeswalker this turn, so every loyalty row in its menu is
  // greyed until the turn cursor moves on. Before #334 this state
  // was server-only, which is why canActivateLoyalty had to take
  // the caller's guess as an argument.
  loyalty_activated?: boolean;
  // S15: raw Scryfall mana-cost string ("{1}{R}", "{W/U}", "{X}{B}"),
  // rendered as a read-only chip on hand-zone cards. Omitted for
  // lands and for placeholder / demo-seed cards. Also zeroed on the
  // per-viewer redaction path when the viewer is not a knower of the
  // card (so opponent hand-counts don't leak cost shapes).
  mana_cost?: string;
  // S15: activated mana abilities on this permanent. Populated for
  // catalog mana rocks (Sol Ring, Arcane Signet, Birds of Paradise)
  // and for every basic land (via the synthetic-ability fallback).
  // The ManaAbilityMenu renders one button per entry; the index
  // field in each entry is what the activate_mana_ability action
  // carries as `ability_index`.
  mana_abilities?: ManaAbilityView[];
  // S16/S18: effective keyword list — strings like "flying",
  // "first strike", "trample". Layered effects (Lord of Atlantis
  // grants "islandwalk" to other Merfolk) populate this alongside
  // printed keywords (S18 Spec.PrintedKeywords). S18 renders
  // keyword badges from this list via the KeywordBadgeRow component.
  abilities?: string[];
  // S24 — the restriction set the server computed for this
  // permanent: "cant_attack", "cant_block", "cant_be_blocked",
  // "cant_activate", "cant_activate_mana". Absent for the permanent
  // nothing is restricting, which is nearly all of them.
  //
  // Deliberately separate from `abilities`. A restriction is not a
  // keyword the permanent has — it is an effect something else has
  // (Pacifism, Arrest, a Whispersilk Cloak) — so it gets no badge;
  // it disables a control and supplies the reason.
  //
  // READ IT, DON'T DERIVE IT. Which creatures can attack is the
  // server's decision; this field is how it says so. #429 deleted a
  // pile of client-side rules re-derivation and this must not start
  // a new one.
  restrictions?: string[];
  // ADR 0034 — Scryfall's printing layout, absent for the ordinary
  // single-faced card. "modal_dfc" is the one the client acts on:
  // it means playing this card needs a face choice first.
  layout?: string;
  // ADR 0034 — every printed face, front first. PURELY ADDITIVE:
  // name / type_line / mana_cost / power / toughness above continue
  // to mean "the ACTIVE face's", which is why every existing type
  // check in cardTypes.ts, Card.svelte and timing.ts kept working
  // unchanged — they now receive one clean type line instead of a
  // "Sorcery // Land" concatenation. Absent for single-faced cards.
  faces?: CardFaceView[];
  // ADR 0034 — index into `faces`. Absent (0) is the front face.
  active_face?: number;
}

// ManaAbilityView mirrors `protocol.ManaAbilityView` server-side —
// one entry per activated mana ability on a battlefield permanent.
// The client renders these as buttons in a right-click / long-press
// menu anchored to the card. Added in S15 sub-PR 2.
export interface ManaAbilityView {
  index: number;
  label?: string;
  tap_cost?: boolean;
  sacrifice_cost?: boolean;
  produced?: string;
  // S21: "Sacrifice a creature: Add {C}{C}" (Ashnod's Altar) — a
  // mana ability whose cost sacrifices ANOTHER permanent. Mirrors
  // the identically-named fields on ActivatedAbilityView: the label
  // is the clause for the modal banner, and sacrifice_options lists
  // the legal choices already filtered to the controller
  // (CR 701.21a). Absent means the cost needs no extra choice. #747:
  // min / max are the count and the cards come in payment order, as
  // on ActivatedAbilityView.
  sacrifice_label?: string;
  sacrifice_options?: LegalTargetsView;
  // S22: a "Pay N life" component of the activation cost — Mana
  // Confluence's "{T}, Pay 1 life:". Advisory only; the server does
  // the real CR 119.4 check. A damage RIDER ("This land deals 1
  // damage to you", the painlands / Ancient Tomb) is NOT a cost and
  // never appears here — it is spelled out in `label` instead.
  life_cost?: number;
  // S32 (#352): a mana component of the activation cost — the Signet
  // cycle's "{1}, {T}", Cabal Coffers' "{2}, {T}". Advisory like
  // life_cost. The server never auto-taps into a mana ability, so the
  // player has to float this mana before the entry will fire.
  mana_cost?: string;
  // #743: true while the mana ability's activation condition is false
  // — Temple of the False God with four lands, Mox Opal without
  // metalcraft. Same flag and meaning as ActivatedAbilityView's.
  condition_unmet?: boolean;
  // #844, CR 903.4f: the ability says "any color in your commander's
  // color identity" (Command Tower, Arcane Signet, Commander's Sphere,
  // Path of Ancestry) and the controller has no commander, or one
  // whose colour identity is colourless. The quality is undefined or
  // empty, so the ability adds no mana at all and the row is greyed —
  // tapping the land would just lose it. Absent for every other
  // ability.
  adds_no_mana?: boolean;
  // S32 (#352): spend restrictions the produced mana will carry —
  // Ancient Ziggurat's "only to cast a creature spell", Eldrazi
  // Temple's "only colorless Eldrazi". Informational; the server's
  // pool solver is what actually refuses an illegal payment.
  restrictions?: string[];
}

// AttackTargetView is one legal attack target: the id to send as
// declare_attacker's `target`, and what it is. `id` is a seat id for
// a player and an instance id for a permanent. Added in S27.
export interface AttackTargetView {
  kind: "player" | "planeswalker" | "battle";
  id: string;
}

export interface TurnView {
  number: number;
  active_seat: number;
  // priority_holder is the seat index that currently holds priority
  // within the step. Equal to active_seat at every step boundary;
  // rotates on pass_priority. Added in S07.
  priority_holder: number;
  phase: string;
  step: string;
  // #328: seat indices that owe a declare-blockers decision — under
  // attack, holding at least one creature that could legally block
  // one of the attackers. Absent outside the declare_blockers step.
  //
  // The server computes this because block legality is a rules
  // question (CR 509.1a untapped, CR 509.1b evasion) that the client
  // must not re-derive in TypeScript. It exists because blocking is a
  // turn-based action rather than a response, so the auto-pass
  // "legal response?" predicate structurally could not see it.
  block_decision_seats?: number[];
  // S27: what the ACTIVE player's creatures may attack right now —
  // the other seats, the planeswalkers they don't control, and the
  // battles they don't protect (CR 506.2, 508.1d). Present only
  // during declare_attackers. Server-computed: the client must not
  // re-derive "who protects which battle".
  attack_targets?: AttackTargetView[];
}

// uuid generates a v4 UUID. Uses crypto.randomUUID when available (all
// modern browsers). Falls back to a simple implementation for older
// environments so tests can run under Node without polyfills.
export function uuid(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  // RFC 4122 v4, not cryptographically strong — only used as a fallback.
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
