// Mirror of the v0 protocol types from server/internal/protocol/protocol.go
// and server/internal/protocol/view.go. When the spec in docs/protocol.md
// evolves, update both sides in lockstep.

export const PROTOCOL_VERSION = 0;

// ADR 0122 §6: "ack" acknowledges an applied action to the client that
// sent it; "legal_moves_request" / "legal_moves" fetch a seat's whole
// move list. The browser handles "ack" and sends no requests.
// ADR 0162 §2: "visibility" tells the server whether this tab is in
// front, for Discord DM notifications. Sent, never answered.
export type Kind =
  | "ping"
  | "pong"
  | "error"
  | "action"
  | "snapshot"
  | "chat"
  | "ack"
  | "legal_moves_request"
  | "legal_moves"
  | "visibility";

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
  // #3855 (ADR 0068 amendment 2026-10-15): a waived payment (strict mana
  // off with a short pool, or Cast anyway) that something on the table
  // READS. Nothing was paid. `paper_payment` carries the shortfall slot
  // by slot and `card_id` the card cast (or the ability's source); the
  // client asks what was paid outside the pool and re-sends the action
  // with `paper_payment` (paperPayment.ts).
  PaperPaymentRequired: "paper_payment_required",
  // #705 (ADR 0045 addendum Decision 8): a declare_blocker was refused
  // by the server's block-legality check. `message` is a server-built
  // sentence to show verbatim, `reason` the stable token and `card_id`
  // the blocker. The client never re-derives the rule.
  IllegalBlock: "illegal_block",
  // #1063 (ADR 0080): a declare_attacker / declare_attackers refused
  // for want of the CR 508.1a attack tax. `reason` carries the whole
  // declaration's price as a cost string ("{2}{2}"), not a
  // BlockRefusalReason token — see ErrorPayload.reason. `missing` is
  // the symbols the pool and the tapper together could not cover.
  // No `card_id`: the refusal is about the declaration, not one card.
  AttackTaxUnpaid: "attack_tax_unpaid",
  // #1507 (ADR 0045 Decision 45): a declare_attacker / declare_attackers
  // refused because the declaration breaks a CR 508.1c count limit
  // (Silent Arbiter's "no more than one creature can attack each
  // combat", Crawlspace's "… can attack you …"). `reason` is one of
  // AttackRefusalReason, `message` a server-built sentence addressed to
  // the caller, `card_id` a creature from the refused declaration.
  // Nothing was declared, tapped or paid; a declare_attackers batch is
  // refused whole. #1533: an "attack with all" refused this way offers
  // the attackers picker, capped at `attack_targets[].attack_limit`.
  IllegalAttack: "illegal_attack",
  // ADR 0122 §6.1: answers to a legal_moves_request, which the browser
  // never sends — a seat that owes no decision, and a connection past
  // four requests a second. Listed so this table stays the server's.
  NoDecision: "no_decision",
  RateLimited: "rate_limited",
  // ADR 0177 §6: an action from a seat a bot is playing while its
  // player was away. Nothing was applied; the player takes the seat
  // back first (the Take it back banner).
  StandIn: "stand_in",
} as const;

export type ErrorCodeValue = (typeof ErrorCode)[keyof typeof ErrorCode];

export interface Frame<P = unknown> {
  v: number;
  kind: Kind;
  id: string;
  payload?: P;
}

/** The payload of a "visibility" frame (ADR 0162 §2). */
export interface VisibilityPayload {
  visible: boolean;
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
  // #1063: also carries the attack tax's whole price as a plain cost
  // string ("{2}{2}") when code === "attack_tax_unpaid" — a second
  // shape on the same key, not a second field, because the server's
  // ErrorPayload.Reason is a bare string on the wire either way.
  // #1507: and one of AttackRefusalReason when code === "illegal_attack".
  reason?: BlockRefusalReason | AttackRefusalReason | string;
  // #3855: the shortfall, when code === "paper_payment_required".
  paper_payment?: PaperPaymentPrompt;
}

// PaperPaymentPrompt is a paper_payment_required frame's detail (#3855):
// what the pool left unpaid, one slot per mana, coloured requirements
// first. `options` are the symbols that may pay the slot (W U B R G C
// for a generic one). `count`, when above one, is a long run of that
// many identical mana named once (#3904), answered by one mana entry
// with the same count. `artifact_sources` asks, per mana, whether an
// artifact made it (Coin of Mastery reads that).
export interface PaperPaymentPrompt {
  // `x` marks a slot of the cost's {X} part (#3941), the mana a "spend
  // only <colours> mana on X" clause restricts; the mana declared for it
  // carries `x: true`.
  slots: { symbol: string; options: string[]; count?: number; x?: boolean }[];
  artifact_sources?: boolean;
}

// BLOCK_REFUSAL_REASONS mirrors game.BlockReason (server/internal/game/
// block_legality.go): the tokens the server sends today. Stable once
// shipped; new ones join in the change that first sends them. A
// runtime list rather than a bare union so refusalTokens.test.ts can
// diff it against the Go const block (#1533) — the union below is
// derived from it, so the two cannot drift apart.
export const BLOCK_REFUSAL_REASONS = [
  "cant_block",
  "cant_be_blocked",
  "flying",
  "landwalk",
  "fear",
  "intimidate",
  "shadow",
  "horsemanship",
  "skulk",
  "protection",
  "cant_be_blocked_by",
  "cant_be_blocked_except_by",
  "cant_block_attacker",
  "too_few_blockers",
  "too_many_blockers",
  // #1339: the blocker's controller is not defending against that
  // attacker (CR 802.4a).
  "not_defending",
  // #1507 (ADR 0045 Decision 43): one more blocker would break a
  // whole-combat count limit — Silent Arbiter's "no more than one
  // creature can block each combat" (CR 509.1b).
  "declaration_limit",
  // #1597 (CR 509.1c): the declaration — or the defending player's
  // pass / finish_blocks that ends it — leaves a blocking requirement
  // unobeyed that it could obey (Lure, "blocks each combat if able",
  // "must be blocked if able"). `card_id` is the creature that could
  // obey it; the message names the requirement.
  "block_requirement",
  // #1706 (CR 509.1a/b): one more attacker than the blocker can block —
  // a creature that "can block an additional creature" already blocking
  // two. Only a creature that can block more than one is ever refused
  // this; an ordinary blocker's second block re-points it.
  "blocker_capacity",
  // #1501 (CR 509.1): the blocker's controller has already finished
  // declaring blockers this combat — a late block, such as one on a
  // ninja that entered attacking after the declaration.
  "blocks_declared",
] as const;
export type BlockRefusalReason = (typeof BLOCK_REFUSAL_REASONS)[number];

// ATTACK_REFUSAL_REASONS mirrors the AttackRefusal* constants in
// server/internal/protocol/protocol.go: the `reason` tokens an
// `illegal_attack` frame carries today. Stable once shipped; new ones
// join in the change that first sends them.
export const ATTACK_REFUSAL_REASONS = [
  // #1507 (ADR 0045 Decision 45): a CR 508.1c count limit refused the
  // declaration.
  "attack_limit",
  // #1571 (ADR 0045 Decision 51): a CR 508.1d requirement ("attacks
  // each combat if able", goad) — the declaration would make one
  // unobeyable, or the active player's pass in declare_attackers would
  // leave one unmet. `card_id` is the creature that carries it.
  "attack_requirement",
] as const;
export type AttackRefusalReason = (typeof ATTACK_REFUSAL_REASONS)[number];

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
  // #1279: complete the caller's CR 509.1 block declaration — the
  // "Done blocking" / "No blocks" button. Not priority-gated.
  | "finish_blocks"
  // Set-shaped block declaration (#750). Not a batching convenience:
  // a block COUNT (menace's minimum of two) is a property of the whole
  // declaration, so a two-creature menace block is legal only as a
  // pair and the single verb refuses either half of it. All or
  // nothing — a refused set stores none of itself.
  | "declare_blockers"
  | "discard_selection"
  | "draw_card"
  | "end_vote"
  // #672 (CR 506.4): take ONE creature out of combat; params
  // `{instance_id}`. The per-card sibling of clear_combat.
  | "remove_from_combat"
  // #671: reveal a HAND card to the table, or show it to one player;
  // params `{card_id, to?}`.
  | "reveal_card"
  | "keep_hand"
  // ADR 0121 §3: the opening roll. `roll_opening` rolls the caller's
  // d20; `host_roll_remaining` (host or admin) rolls for every seat
  // left in the round; `choose_starting_player` {seat} is the winner's
  // choice of who takes the first turn.
  | "roll_opening"
  | "host_roll_remaining"
  | "choose_starting_player"
  | "mark_damage"
  | "move_card"
  | "mulligan"
  | "pass_priority"
  | "pass_turn"
  | "end_turn"
  | "resolve_choice"
  // ADR 0121 §5: "Roll a die" — `{die: "d6" | "d20" | "coin"}`, a roll
  // at the table for fun. Never a game roll, never undoable, and the
  // server takes one per seat per 2 s.
  | "roll_table_die"
  // #1968: `{trigger_order: "when_it_matters" | "always" | "never"}` —
  // when the game asks this seat to order its triggers (the #1530
  // `{always_ask: boolean}` form is still read). A setting, not a play:
  // never undoable.
  | "set_trigger_order_preference"
  // ADR 0127 §3: `{rules: [{key, answer}]}` — the seat's standing
  // answers to repeated prompts, replacing the list ("always" or
  // "never"; Ask is no rule). A setting, not a play: never undoable.
  | "set_auto_answers"
  // ADR 0018's eighth amendment (#3016): `{key?: string}` — forget the
  // trigger order remembered under `key` (a trigger_order prompt's
  // auto_answer_key), or every one with no key. Never undoable.
  | "forget_trigger_order"
  // ADR 0150 §4 (owner decision 5): `{choice_id?}` — the caster
  // withdraws the request for help an assisted cast is parked on.
  | "cancel_assist"
  // #3955 (Wish): no params. Opens the outside pick over the card database
  // so the holder of a "play a card you own from outside the game" right
  // names the card to play now.
  | "outside_play"
  // ADR 0193 (CR 732.2): demonstrate a loop. `loop_record_start`,
  // `loop_record_stop` and `loop_record_cancel` take no params;
  // `loop_propose` takes `{count}` (1–1000); `resolve_loop_proposal`
  // answers a loop_proposal prompt with `{choice_id, accept: true}` or
  // `{choice_id, shorten: {iteration, step}}`. None is undoable; the
  // replay they lead to is one undo entry.
  | "loop_record_start"
  | "loop_record_stop"
  | "loop_record_cancel"
  | "loop_propose"
  | "resolve_loop_proposal"
  | "sacrifice_permanent"
  | "set_goaded"
  | "set_initiative"
  | "set_monarch"
  | "set_promise"
  // ADR 0075 §2.3. The params object IS a settings patch — only the
  // fields present are applied. Host or admin only, and deliberately
  // not undoable.
  | "set_table_settings"
  // Deprecated since S35: a one-field alias for set_table_settings
  // that cannot select an unlimited budget (a negative limit clamps
  // to 0). Kept so an old client and the gamecli scripts still work.
  | "set_undo_limit"
  | "shuffle_library"
  | "special_action"
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
//
// `seq` is non-decreasing only WITHIN one `generation` (#523, ADR
// 0044 decision 5) — a restart that rewinds the room to an earlier
// restore point bumps `generation` and can hand out a `seq` lower
// than one this client already rendered. ws.ts tracks `generation`
// and treats a change as "discard and re-render", never as a dropped
// or out-of-order frame. See docs/protocol.md.
export interface SnapshotPayload {
  seq: number;
  generation: number;
  game: GameView;
}

// OpeningRollView mirrors protocol.OpeningRollView (ADR 0121 §3): the
// open opening roll. Round 1 is every seat; each later round is the
// tied leaders of the one before. `chooser` is present once one leader
// remains. Everything in it is public.
export interface OpeningRollView {
  rounds: OpeningRollRoundView[];
  chooser?: number;
}

export interface OpeningRollRoundView {
  // The seats that roll in this round, in seat order.
  seats: number[];
  // Their dice, in the order rolled.
  rolls: OpeningRollDieView[];
}

export interface OpeningRollDieView {
  seat: number;
  result: number;
  // Present only when somebody other than the seat pressed the button:
  // the host's seat, or -1 for the server admin.
  by?: number;
}

// AckPayload is the body of a Kind == "ack" frame (ADR 0122 §6.4): sent
// to the client whose action was applied, carrying that action's frame
// `id`, after the snapshot of the state it names. `seq` and
// `generation` are that state's.
export interface AckPayload {
  seq: number;
  generation: number;
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
export type ChatKind = "say" | "bot_improvisation" | "bot_reasoning" | "bot_notice";

// A bot disclosing that it applied an effect by hand because the
// rules engine can't run the card. Never hidden: an unannounced
// improvisation is a bot cheating (ADR 0033 §8).
export const CHAT_BOT_IMPROVISATION = "bot_improvisation";
// A bot narrating why it chose a move. Debug output, hidden unless
// the viewer turns on "show bot reasoning".
export const CHAT_BOT_REASONING = "bot_reasoning";
// A bot telling the table about the seat itself, not a move: an
// own-key seat falling back to the heuristic (ADR 0182). Always shown,
// and dismissable.
export const CHAT_BOT_NOTICE = "bot_notice";

// View types mirror server/internal/protocol/view.go.

export interface GameView {
  id: string;
  state: "lobby" | "active" | "ended";
  seats: PlayerView[];
  battlefield: ZoneView;
  stack: ZoneView;
  exile: ZoneView;
  // The CR 702.26 phased-out permanents (#1199, ADR 0084). A shared,
  // owner-less, public zone like `exile`.
  //
  // They are ABSENT from `battlefield` rather than flagged inside it:
  // CR 702.26b says a phased-out permanent "is treated as though it
  // does not exist", and the server makes that true by keeping it out
  // of the slice every other reader walks. So anything that asks the
  // board a question — targeting, legal moves, combat, the bot — is
  // right without knowing phasing exists, and the client is the one
  // consumer that deliberately looks here (phasedOut.ts), because a
  // board that silently loses four permanents is indistinguishable
  // from a board that was wrathed.
  //
  // Optional on the type so a client built against a newer server
  // than it is talking to degrades to "nothing is phased out".
  phased_out?: ZoneView;
  turn: TurnView;
  // True between Start and the moment all seated, non-eliminated
  // players have committed to their opening hand via the keep_hand
  // action. Drives the keep / mulligan dialog and gates the normal
  // game UI. Added in S08.
  mulligans_open: boolean;
  // Player ID of the current monarch (Conspiracy mechanic). Empty /
  // omitted when no monarch is set. Since #375 the server enforces
  // CR 725.2 itself — the monarch's end-step draw, and the transfer
  // to whoever deals combat damage to them — so this field moves on
  // its own and the crown below follows it. The set_monarch action
  // stays as the way a card (or a table fixing the board) hands the
  // designation out in the first place. Added in S10.
  monarch?: string;
  // Player ID currently holding the initiative (BG3 mechanic). Empty
  // when unassigned. Same sandbox posture as monarch. Added in S10.
  initiative?: string;
  // The game's day/night designation (CR 731, ADR 0132): "day" or
  // "night", omitted while the game has neither. Public and the same
  // for every viewer; the action dock shows it beside the turn line.
  day_night?: "day" | "night";
  // Per-pair "I owe you" promise tally as "{from}->{to}" string keys
  // → count. Sparse: zero entries are dropped server-side. Added in
  // S10.
  promises?: Record<string, number>;
  // Currently open council's-dilemma / politics vote, or omitted when
  // none is in progress. Added in S10.
  vote?: VoteView;
  // Per-player per-turn undo budget. Refreshed on each player's untap
  // step. Drives the "undos: N" indicator and gates the undo button.
  // Added in S11. Since ADR 0075 it mirrors `settings.undo_limit`:
  // -1 means unlimited, 0 means no undos.
  undo_limit?: number;
  // The table's settings (ADR 0075 §2.2). Public: every viewer,
  // spectators included, gets the same object. Added in S35 (#1032).
  settings?: TableSettingsView;
  // The table's Discord voice room (ADR 0159 §6), reduced by the server
  // to this viewer's own access. Absent when the server offers none.
  voice?: VoiceView;
  // The host's rematch offer on an ended table (#3285, ADR 0164), reduced
  // by the server to this viewer's own part. Absent until one is offered,
  // and for a spectator.
  rematch?: RematchView;
  // Where this viewer dragged the stack panel to (#3429, ADR 0176), keyed
  // by seat id. The server keeps only the viewer's own entry; absent when
  // the panel is at its default and for a spectator.
  stack_positions?: Record<string, StackPositionView>;
  // Seat index that took the first turn. Used by the server to enforce
  // the CR 103.8a turn-1 skip-draw rule — in two-player games only,
  // since CR 103.8c has nobody skip at a larger table. Added in S13.
  // Pre-S13 replays
  // decode as 0 (Go's int zero), which matches the only seat games
  // ever started on before the field existed.
  starting_seat?: number;
  // ADR 0121 §3: the opening roll, present only while it is open
  // (between Start and the winner's choice). Public to every viewer.
  // While it is present `starting_seat` means nothing (it reads 0).
  opening_roll?: OpeningRollView;
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
  // ADR 0107 §5 (CR 615.12): the live "damage can't be prevented this
  // turn" grants, by their source's name (Skullcrack, Stomp), oldest
  // first. Absent on nearly every turn.
  damage_cant_be_prevented?: string[];
  // ADR 0108 §1: the live "if a creature would die this turn, exile it
  // instead" effects over a live set (Flaying Tendrils, Malicious
  // Eclipse), by their source's name, oldest first. Absent on nearly
  // every turn.
  exile_if_creatures_die?: string[];
  // ADR 0108 §7: the live "prevent all damage a source of your choice
  // would deal this turn" shields and their charged siblings, as
  // "<card> (<source>)" with "— N left" on a charged one, oldest first.
  // Absent on nearly every turn.
  damage_shields?: string[];
  // ADR 0108 §3: one line per live "it deals double (triple) that damage
  // instead" effect a resolved spell or ability made (Insult, Isengard
  // Unleashed, Lightning's Stagger), oldest first, already worded for the
  // banner. Absent on nearly every turn.
  damage_multipliers?: string[];
  // ADR 0108 §9: one line per live "that damage is dealt to <something>
  // instead" effect a resolved spell or ability made (Beacon of Destiny,
  // Harm's Way, Kor Chant), oldest first, already worded for the banner.
  // Absent on nearly every turn.
  damage_redirections?: string[];
  // ADR 0109 §6 (CR 601.2c): one line per live static that stops cards in
  // graveyards being targeted, the printed clause and its card ("Cards in
  // graveyards can't be the targets of spells or abilities. — Ground
  // Seal"). The graveyard viewer shows them as a banner; the legal target
  // sets already leave the refused cards out. Absent on nearly every turn.
  graveyard_target_bans?: string[];
  // Cleanup-step pause map (S13.4, CR 402.2). Keys are player UUID
  // strings, values are the count each player must discard. Drives
  // DiscardPromptModal. Empty / absent when nobody owes discard.
  discard_pending?: Record<string, number>;
  // S14 generic "someone needs to pick" queue. Drives ChoicePromptModal
  // for effects like Thoughtseize where the chooser isn't the
  // discarder. Each entry carries its own options[] already filtered
  // per the viewer's visibility.
  pending_choices?: PendingChoiceView[];
  // ADR 0166 (CR 723): who is making whose decisions right now. A
  // prompt the controlled player owes arrives with `chooser` set to the
  // controller and `on_behalf_of` to the controlled player.
  // ADR 0167: `until` is "resolution" (Word of Command) or "turn" (the
  // whole of a turn another player plays, Emrakul, the Promised End).
  // #3998: or "search" (Opposition Agent: only while the controlled
  // player's search prompt is open).
  // One entry per controlled player: the effect that applies now.
  player_controls?: PlayerControlView[];
  // ADR 0167 §8: turn-long controls still waiting for the controlled
  // player's next turn. Public; grants no visibility.
  scheduled_player_controls?: ScheduledPlayerControlView[];
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
  // ADR 0122 §6.1: true when the 48-move wire cap dropped anything
  // from legal_moves — the list then has one move per (source, kind,
  // targets_stack, has_targets, interacts), not every alternative.
  // Absent otherwise. Own seat only. The browser reads nothing from
  // legal_moves that the cap loses (legalActions.ts), so it only needs
  // to know.
  legal_moves_truncated?: boolean;
  // ADR 0105 (#1789): a per-card digest of the same enumeration, built
  // before the 48-move cap, so it stays exact down to the ability row.
  // Own seat only, like legal_moves. Absent means "highlight nothing",
  // never "nothing is legal". Read through lib/legalActions.ts, never
  // directly, so every highlight comes from one lookup.
  legal_actions?: LegalActionsView;
  // ADR 0167 §7: while the viewer plays another seat's turn,
  // legal_moves also holds that seat's moves (each with `player` set to
  // it) and legal_actions digests only the viewer's own. This maps each
  // controlled seat's player ID to the digest of its moves.
  controlled_legal_actions?: Record<string, LegalActionsView>;
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
  // #628 (CR 732): set when the engine has watched the same triggered
  // ability resolve 25 times this turn with no player decision in
  // between, absent otherwise. Its presence is an instruction to this
  // client: STOP PASSING AUTOMATICALLY. Priority still rotates and
  // every pass the server is handed still works — the point is that a
  // person has to ask for the next iteration, with the loop's trigger
  // still on the stack. Table-wide and identical for every seat.
  loop_notice?: LoopNoticeView;
  // ADR 0193 (CR 732.2): the loop a player is recording, has recorded or
  // has proposed. Table-wide and identical for every seat.
  loop_recording?: LoopRecordingView;
  // ADR 0057 (#749): the result of an ended game — who won, or a draw,
  // and why. Absent while the game is active and for a table an admin
  // closed with no result. An effect win ends the game with the other
  // seats still seated, so read the winner from here; gameOutcome.ts
  // falls back to "the one seat left standing" only when this is
  // absent on an ended view.
  outcome?: OutcomeView;
}

// OutcomeView is GameView.outcome (ADR 0057 Decisions 5 and 7).
export interface OutcomeView {
  kind: "win" | "draw";
  // The winner of a win.
  winner?: string;
  winner_seat?: number;
  // "last_standing" (every opponent left, CR 104.2a), "effect" (an
  // effect said this player wins, CR 104.2b) or "all_lost" (everyone
  // left lost at once, CR 104.4a — a draw) or "mandatory_loop" (a loop
  // of only mandatory actions repeated with no way to stop, CR 732.4 —
  // a draw).
  cause: "last_standing" | "effect" | "all_lost" | "mandatory_loop";
  // The object whose effect won, for an effect win; the loop's card,
  // for a mandatory loop's draw.
  source?: string;
  source_name?: string;
}

// GameEndGateView is one "can't lose" / "can't win" gate on a seat
// (ADR 0057 Decision 7): where it comes from and what it stops.
// One "can't be countered" grant or unspent promise on a seat (ADR
// 0106 §4, #1806). `text` is the clause as printed; `next_only` marks a
// promise the seat's next matching spell will spend.
export interface CounterShieldView {
  source?: string;
  source_name: string;
  text: string;
  next_only?: boolean;
}

export interface GameEndGateView {
  source?: string;
  source_name: string;
  cant_lose?: string[];
  cant_win?: boolean;
  // A gate a resolved spell granted until end of turn (Angel's Grace).
  this_turn?: boolean;
}

// LoopNoticeView is the CR 732 loop breaker's notice. `label` is the
// repeating ability's stack label, which by catalog convention reads
// "<card> — <what happens>", so it is the whole banner line.
// Mirrors `protocol.LoopNoticeView`.
export interface LoopNoticeView {
  source?: string;
  label: string;
  controller?: string;
  count: number;
  // #2506, CR 732.4: the loop has only mandatory actions and repeats
  // unchanged. If every seat passes it round once more, the game is a
  // draw.
  mandatory?: boolean;
}

// ADR 0193: GameView.loop_recording. `status` is "recording" (steps
// counts the actions so far), "ready" (it closed: net and iteration are
// set) or "proposed" (count more iterations, ending at end).
export interface LoopRecordingView {
  recorder: string;
  status: string;
  steps: number;
  name?: string;
  net?: string;
  iteration?: LoopStepView[];
  max_iterations: number;
  count?: number;
  end?: LoopPointView;
  shortened_by?: string;
}

// Before step `step` (0-based) of iteration `iteration` (1-based).
export interface LoopPointView {
  iteration: number;
  step: number;
}

export interface LoopStepView {
  actor?: string;
  label: string;
}

// ADR 0193: PendingChoiceView.loop_proposal, on a loop_proposal prompt.
// choice_points are the indices of steps that are the chooser's own:
// where CR 732.2b lets them shorten the proposal.
export interface LoopProposalView {
  proposer: string;
  name?: string;
  count: number;
  end: LoopPointView;
  net?: string;
  steps: LoopStepView[];
  choice_points?: number[];
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
  // Zone a library search is sending the cards to ("battlefield" for a
  // fetch). Absent for a reveal that is not part of a search: the
  // banner says "fetched" off this, never "revealed".
  to?: string;
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
// GameView.legal_actions (ADR 0105): what each of the viewer's cards
// may do right now, folded from the uncapped legal-move list. Keyed by
// card instance ID. See docs/protocol.md "LegalActionsView".
export interface LegalActionsView {
  // A pass_priority move exists.
  pass?: boolean;
  sources?: Record<string, LegalSourceView>;
}

export interface LegalSourceView {
  // Distinct move kinds, in enumeration order. Never pass, choice,
  // mulligan or opening_roll.
  kinds: Exclude<LegalMoveView["kind"], "pass" | "choice" | "mulligan" | "opening_roll">[];
  // Moves in the uncapped list that involve this card.
  moves: number;
  // ADR 0093 refs of live activated_abilities / zone_abilities rows.
  abilities?: string[];
  // Refs of live mana_abilities / zone_mana_abilities rows.
  mana_abilities?: string[];
  // Live special-action kinds: foretell, suspend, turn_face_up, plot.
  special_actions?: string[];
  // from_zone of each cast or land move.
  zones?: string[];
  // Printed faces the cast moves cast (0 is the front).
  faces?: number[];
  // Players, planeswalkers and battles this creature may attack.
  attack_targets?: string[];
  // ADR 0130 §5: the creature may be exerted as it attacks right now
  // (CR 701.43d) — an attack move with `exert: true` is offered beside
  // the plain one. Absent otherwise.
  exert_on_attack?: boolean;
  // Attackers this creature may block, alone or in a group.
  blocks?: string[];
  // #1918: present only when EVERY cast move for the card carries an
  // `idle_hint` — castable, but no cast would do anything right now
  // (an overloaded Counterflux with no spell to counter). The first
  // such hint; the hand draws a muted ring with it as the tooltip.
  cast_idle_hint?: string;
  // ADR 0122 §6.2: the enumerator's own caps that cut this card's
  // moves, and by how much. Absent when nothing was cut. `moves`
  // counts what survived; this says the card has more.
  truncated?: LegalCutView[];
}

// One entry of the enumerator's cut report (ADR 0122 §6.2,
// docs/protocol.md "The cut report"). In the digest it carries no
// `source`: the key is the card.
export interface LegalCutView {
  source?: string;
  choice?: string;
  cap:
    | "per_source"
    | "max_x"
    | "variable_counts"
    | "subset_scan"
    | "creature_types"
    | "card_names"
    | "cost_payments"
    | "repeats"
    | "ceiling";
  // Candidate moves the cap left out; a lower bound when at_least.
  omitted: number;
  at_least?: boolean;
}

export interface LegalMoveView {
  type: string;
  player: string;
  params?: Record<string, unknown>;
  kind:
    | "pass"
    | "land"
    | "cast"
    | "activate"
    | "mana"
    | "attack"
    | "block"
    | "choice"
    | "mulligan"
    | "special_action"
    // #1501: a declaring defender's finish_blocks ("No blocks" /
    // "Done blocking"). Not a pass: nobody holds priority while a
    // defender declares. Carries no card.
    | "finish_blocks"
    // ADR 0121 §4: roll_opening and choose_starting_player, the only
    // moves while `opening_roll` is open. Carries no card.
    | "opening_roll";
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
  // ADR 0166 §4: the controlled_play prompt this move answers — one way
  // of making Word of Command's forced play, or "it can't be played".
  // ControlledPlaySheet lists exactly these, and the server never folds
  // them into the capped list's one-per-card.
  controlled_play?: string;
  // #1307: true on a cast or activation whose chosen targets include
  // an object on the stack — a counterspell, a Stifle, a redirect.
  // Smart autopass reads it to tell "can answer the thing on the
  // stack" from "has some instant". Absent on older servers, which
  // the client treats as a plain instant-speed move.
  targets_stack?: boolean;
  // #2853: true on a cast or activation that chooses at least one
  // target, on the stack or anywhere else. Smart autopass reads it to
  // tell an ability that answers something (a pinger, a removal
  // activation) from an untargeted value ability (Mind Stone, a fetch
  // land, cycling). Absent on older servers, which the client treats
  // as untargeted.
  has_targets?: boolean;
  // #2853 (owner answer 2): true on an activation with no target, or a
  // mana ability, that can still answer something on the stack: a
  // sacrifice outlet, regeneration, protection or indestructible,
  // phasing, a blink, damage prevention, a pump. Smart autopass stops
  // for it as for a targeted ability. Absent on older servers. #3555:
  // protect, survive, pump, prevent and sac_outlet set it only while an
  // opponent's stack item threatens the viewer; otherwise they set
  // combat_interacts (the server applies the gate).
  interacts?: boolean;
  // #2871: true on an activation with no target that changes a fight
  // without answering a spell: crew, a manland, a granted flying or
  // menace, "can block an additional creature", a creature token.
  // Smart autopass counts it only in a combat window (inCombatWindow).
  // Never set beside `interacts`. Absent on older servers.
  combat_interacts?: boolean;
  // #2871 (owner answer): narrows combat_interacts. The activation
  // makes a creature token (or populates, or amasses), which matters
  // only as a blocker, so smart autopass counts it only while the
  // viewer is a defending player in this combat. Only ever set beside
  // combat_interacts.
  combat_defender_only?: boolean;
  // #1918: a player-facing sentence on a LEGAL cast that would do
  // nothing on the board as it stands ("Overloaded, this does nothing
  // right now: there's no spell you don't control."). Advice, never
  // sent back; only overload sets it today. Absent promises nothing.
  idle_hint?: string;
  // ADR 0122 §6.2: the move's answer is an open set — any card name
  // ("card_name"), or any X from `min` to `max` ("x") — with the rest
  // of `params` unchanged. Advice, never sent back.
  value?: { kind: "card_name" | "x"; min?: number; max?: number };
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
  // #1677: the part of `life` that pays Phyrexian mana symbols
  // instead of mana (CR 107.4f). Already INCLUDED in `life`, and in
  // life points: 2 per symbol, where `params.phyrexian_life` counts
  // symbols. Broken out because it buys nothing — the same move at
  // a different price — so it is no evidence of what the move does.
  phyrexian_life?: number;
  // A loyalty ability's counter delta (CR 606.4), signed as printed:
  // +1 adds one, -3 removes three.
  loyalty?: number;
  // #625: counters a "remove N counters" cost takes, and from which
  // permanent — often not the move's source (Heart of Kiran's crew
  // paid with a planeswalker's loyalty).
  counters?: { card_id: string; counter: string; n: number }[];
  // A mana cost string the move charges that `params` cannot name.
  // On a declare_attacker move (ADR 0080, #1063): the CR 508.1a attack
  // tax ("{2}" for an attack into Propaganda). On every cast_spell move
  // (ADR 0136 §2): the total mana the cast charges (CR 601.2f), with
  // the commander tax, X, alternative costs and the board's increases
  // and reductions in it ("{5}{G}{U}" for a twice-cast Tatyova), which
  // the printed mana_cost on the CardView is not. "{0}" when the cast
  // charges no mana.
  mana?: string;
  // #1600: how many cards a "Discard your hand" cost throws away if
  // the move is made now — Lion's Eye Diamond, Null Brooch, Slate of
  // Ancestry. Every other discard names its cards in
  // `params.discard_ids`; this one names none, so without it the
  // whole hand reads as free. Absent for an empty hand, which pays
  // the cost.
  hand?: number;
  // ADR 0129 §7: the energy counters the move removes from the seat
  // (CR 107.14) — "Pay N {E}", or N + X for "Pay X {E}" at the move's
  // X. Always positive when present, and payable.
  energy?: number;
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
  // #1279 / #1500: a defending player completed their block
  // declaration with zero blockers. The blocks a defender DID make
  // are already `block` entries above; this is the one fact those
  // can't carry — that the defender was asked and chose to take it.
  // Carries no `card_id` — `seat` is the defender, and the same
  // sentence for every viewer.
  | "no_blocks"
  | "token"
  | "sacrifice"
  // ADR 0130 §5: a player exerted a permanent (CR 701.43a). `seat` is
  // the player, `card_id` the permanent; server-rendered `text`.
  | "exert"
  // ADR 0127 §6: the server answered a prompt with its chooser's
  // standing answer ("Bob paid {1} for Rhystic Study (automatic)").
  // Server-rendered `text`.
  | "auto_answer"
  // ADR 0193: the demonstrated-loop conversation, and a replay's one
  // line ("P1 repeats Basalt Monolith ×500: P1: library -500").
  // Server-rendered `text`.
  | "loop"
  // A player left the game. `cause` says why ("life", "empty_draw",
  // "poison", "commander_damage", "effect", "concede"); `card_id` is
  // the source of an effect loss. One line per departure (ADR 0057).
  | "eliminated"
  // The game ended with a result: `seat` is the winner (-1 for a
  // draw), `cause` the outcome cause, `card_id` the winning source of
  // an effect win (ADR 0057).
  | "game_over"
  // An effect would have made `seat` win and a "can't win the game"
  // gate stopped it: `card_id` is the winning source, `target` the
  // gate's (ADR 0057).
  | "win_prevented"
  // A player revealed cards (CR 701.20): one entry per reveal, however
  // many cards it showed. Never carries card_id; `amount` is the card
  // count, `old_zone` where they were revealed from, and `target_seat`
  // is set when the reveal was to one player only, in which case the
  // text names no card for anyone.
  | "reveal"
  // A player milled cards from their library (CR 701.17a). The milled cards
  // are named in `text`; no card_id, and `amount` is the card count.
  | "mill"
  // A player searched a library, or shuffled one. `label` "shuffle" is a plain
  // shuffle; "search_shuffled" is a search that also shuffled. Names no card;
  // `target_seat` is set when the library was another player's.
  | "search"
  // S30 random effects: the server-rendered public outcome of a die
  // roll or coin flip. The event text is already redacted and ready
  // for both the log and the attention strip.
  | "roll"
  | "flip"
  // ADR 0121 §3: the opening roll moved on (a tie, a winner, or the
  // host rolling for everyone left), and the winner's choice of who
  // takes the first turn. Server-rendered `text`, like `roll`.
  | "opening_roll"
  | "starting_player"
  // ADR 0121 §5: a d6, a d20 or a coin rolled at the table for fun.
  // `sides` and `results`, or `faces` for a coin, and `roll_id`, which
  // an undo keeps when it writes the line again under a new `seq`.
  | "table_roll"
  // #984: a player answered a "choose a ..." prompt out loud. The
  // chosen VALUE is `choice` on the first two ("G", "Elf"); a chosen
  // PLAYER is `target_seat`, like every other player in the log. All
  // three carry `card_id` — the card the answer was given for, and the
  // card whose later abilities read it back (CR 607.2d).
  | "choose_color"
  | "choose_type"
  | "choose_player"
  // ADR 0102: a player chose the opponent a permanent enters under the
  // control of. The seat rides target_seat, as for choose_player.
  | "choose_controller"
  // A player answered an "as this enters, choose a card name" prompt
  // (CR 614.12): Pithing Needle, Phyrexian Revoker, Sorcerous
  // Spyglass. `choice` is the name as the player typed it, trimmed
  // and otherwise untouched — CR 201.2 admits any card name, so
  // there is no canonical spelling to report (#1210).
  | "choose_name"
  // A player answered an "as this enters, choose <A> or <B>" prompt
  // whose options are words printed on the card (CR 614.12): a
  // Siege's anchor word. `choice` is the word ("Khans", "Temur")
  // (#1572).
  | "choose_option"
  // #1214: a player answered one of the three resolution-time picks
  // (CR 608.2) — an opponent choosing from a revealed set, a seat
  // choosing among another player's permanents, a seat choosing N of
  // its own. `amount` is how many cards were chosen and is always
  // there; `card_id` names the one card for a single-card pick and is
  // absent for any other count; `target` is the card that ASKED.
  | "choose_cards"
  // ADR 0114 §9: the Ring tempted `seat`. `amount` is how many times it
  // has now; `card_id` is the creature chosen as their Ring-bearer
  // (absent when they controlled none); `cause` is "forced" when it was
  // their only creature and was chosen for them.
  | "ring_tempted"
  // ADR 0146: `seat` cast a vote (CR 701.38). `choice` is the option as
  // the card prints it (a word, or the name of the permanent or player
  // voted for) and `card_id` the card that called the vote.
  | "vote"
  // ADR 0180: `seat` cast a ballot in a secret vote; the line says
  // nothing about what they voted for. `card_id` is the card.
  | "vote_secret"
  // ADR 0180: a secret vote's ballots, revealed together on one line.
  // `seat` is the vote's controller, `card_id` the card, and `seats` and
  // `votes` are parallel: each voter and what they voted for.
  | "votes_revealed"
  // ADR 0150 §5: `seat` paid, or didn't pay, an "unless any player pays"
  // offer. `label` is the payment ("{2}") and `card_id` the card whose
  // effect it was.
  | "any_player_paid"
  | "any_player_declined"
  // #1021: six silences the log kept until they were written down.
  // `control` names two seats — `seat` gained control, `target_seat`
  // lost it (CR 613.1b). `special_action` carries the printed action
  // in `label` ("Foretell {2}", CR 116.2). `cycle` REPLACES the zone
  // line for the discard that paid for it (CR 702.29b). `counters`
  // carries the kind in `label` and the count AFTER the change in
  // `amount`. `scry` and `surveil` name NO card for anyone and carry
  // only the count that moved. `saga_chapter` and `class_level` carry
  // the chapter or level in `amount`.
  | "control"
  | "special_action"
  // ADR 0172 §5: a seat revealed its companion as the game began
  // ("Alice revealed Jegantha, the Wellspring as their companion").
  // `card_id` is the companion, public from then on.
  | "companion_revealed"
  // ADR 0106 §1 decision 6 (#1793): a player activated the "Any player
  // may activate this ability" row of a permanent another player
  // controls — "Bob activated Alice's Xantcha, Sleeper Agent".
  // `target_seat` is the permanent's controller.
  | "activate_across"
  // ADR 0119 §5: what goes on the stack, as it goes there. `trigger` is a
  // triggered ability triggering ("Alice's trigger: Mulldrifter — draw two
  // cards"); consecutive identical ones collapse, with `amount` the count
  // (absent for one). `activate` is an activated ability that uses the
  // stack (mana abilities never get one). Both carry the SOURCE in
  // `card_id` and the item's label in `label`, redacted together like an
  // ability's `resolve` line.
  | "trigger"
  | "activate"
  | "cycle"
  | "counters"
  // #3040: a count hit the table's ceiling and was pinned to it.
  // `label` is the quantity ("+1/+1 counters", "mana"), `amount` the
  // ceiling, `card_id` the object whose count was capped (or the mana's
  // source).
  | "count_capped"
  | "scry"
  | "surveil"
  // ADR 0099: a finished discover (CR 701.57b). `amount` is the N and
  // `card_id` the discovered card, absent when nothing was found.
  | "discover"
  // ADR 0082 amendment (#2570): a player manifested dread. Carries no card.
  | "manifest_dread"
  | "saga_chapter"
  | "class_level"
  // ADR 0103: a Room's door was unlocked or locked (CR 709.5c/g);
  // `label` is the door's name. `room_fully_unlocked` is the second
  // door (CR 709.5i), its own line by owner decision.
  | "door_unlocked"
  | "door_locked"
  | "room_fully_unlocked"
  // ADR 0075 §2.3: the host or the admin changed a table setting.
  // `label` is the setting's key ("undo_limit", "allow_spawn") and
  // `choice` its new value as text; `seat` is the host, or NoSeat when
  // the server admin made the change. The only kind that is about the
  // rules rather than about the game.
  | "settings"
  // ADR 0177: the table moved on without a player who was gone.
  // `label` is "stand_in", "concede" or "take_back"; `target_seat` the
  // seat; `seat` who decided (NoSeat for the admin); `amount` the
  // minutes the seat had been offline.
  | "away_seat"
  // ADR 0075 §2.4: the host or the admin put cards or tokens on the
  // table from nowhere. `label` is the card or token name, `amount`
  // the count, `new_zone` where they went and `target_seat` whose
  // zone it was. A spawn into a hidden zone names the zone and NOT
  // the card, so `label` is absent there for everyone.
  | "spawn"
  // ADR 0086 (#1238): a storm trigger settled on its count
  // (CR 702.40a). `card_id` is the storm spell and `amount` the
  // number of OTHER spells cast before it this turn, which is how
  // many copies are about to be created. Emitted at zero too, so a
  // trigger that found nothing to copy is distinguishable from one
  // that never fired. `text` reads "Grapeshot — storm count 3".
  // Unlike `counters` / `saga_chapter` / `class_level`, `amount`
  // survives redaction: the count is a fact about the turn's casts,
  // not a value read off the card.
  | "storm"
  // S46 (ADR 0079, #343): a permanent was turned over to its other
  // face (CR 701.27a). `label` is the name of the face it turned
  // FROM — the only place that name survives, since the card's own
  // name is already the new face by the time the entry is rendered.
  // Narrated, not silent: a card physically turning over is a thing
  // a player announces out loud, and a reader scrolling back wants
  // to know when it happened.
  | "transform"
  // ADR 0151 (#1892): a flip card on the battlefield was flipped
  // (CR 710). `label` is the name it flipped FROM. Its own kind,
  // never `flip`, which is a coin flip.
  | "flip_permanent"
  // ADR 0145: two cards were melded into one permanent (CR 701.42a).
  // `card_id` is the melded permanent and `label` names the two cards
  // it was melded from.
  | "meld"
  // #1199, ADR 0084: a permanent phased out or in (CR 702.26).
  // Narrated for `transform`'s reason and one more that is stronger
  // here — phasing out is not a zone change (CR 702.26d), so no
  // `zone` entry says it, and the board simply stops showing the
  // permanent, indistinguishable from a permanent that died unless
  // the log says which.
  | "phase_out"
  | "phase_in"
  // ADR 0132 (#2561): the game became day or night (CR 731.1). `label`
  // is the new designation. Narrated because the untap-step check
  // changes it with no spell or ability behind it.
  | "day_night"
  // #2696 (CR 702.131): a player got the city's blessing. `seat` is who,
  // `card_id` the ascend permanent or spell whose check granted it.
  | "citys_blessing"
  // ADR 0138 (#2122): a player's speed changed (CR 702.179). `seat` is
  // the player and `amount` the new speed. Narrated because the
  // start-your-engines state-based action has no card behind it.
  | "speed"
  // ADR 0059 Decision 11 (#753): an effect gave a player an extra turn
  // (CR 500.7). `seat` is who will take it and `card_id` the card whose
  // effect created it; one entry per turn.
  | "extra_turn"
  // #2529 (CR 614.10): a queued extra turn was skipped instead of
  // beginning (Trouble in Pairs). `seat` is who would have taken it,
  // `card_id` the card whose effect created it.
  | "extra_turn_skipped"
  // ADR 0188 Decision 1 (#3559, CR 614.10): an effect made a player
  // skip their next turns (Magosi, the Waterveil; Lethal Vapors). `seat`
  // is who will skip, `card_id` the card whose effect wrote the debt and
  // `amount` how many turns it adds.
  | "turn_skip"
  // ADR 0188 Decision 1 (CR 614.10a): a turn was skipped because its
  // player owed a skip. `seat` is who would have taken it, `card_id` the
  // card whose effect wrote the debt, and `label` is "extra" when the
  // turn skipped was an extra turn.
  | "turn_skipped"
  // ADR 0059 Decision 11 (#753): an effect added phases or a step to
  // the current turn (CR 500.8 / 500.9). `seat` is the active player,
  // `card_id` the card whose effect added them, and `label` what was
  // added: phase kinds in order, comma-separated ("combat,main"), or
  // "step:<step>" for one step ("step:end").
  | "extra_phase"
  // #2165, CR 724.1: an effect ended the turn (Sundial of the Infinite,
  // Time Stop). `seat` is the active player, whose turn it was, and
  // `card_id` the card whose effect ended it. The stack's exile lines
  // and a `step` line for the cleanup step follow it.
  | "turn_ended"
  // #3856, CR 727 (ADR 0201): the game was restarted (Karn Liberated's
  // −14). `seat` is the player who restarted it and takes the new
  // game's first turn, `card_id` / `label` the card, and `amount` how
  // many exiled cards will enter under their control after the
  // mulligans. Always the first line: the log begins again here.
  | "game_restarted"
  // #1209, ADR 0082's 2026-09-23 amendment: a permanent that was
  // face up was turned face down (CR 708.2a). `card_id` is the
  // permanent; `target` is the object that did it (Ixidron, Cyber
  // Conversion). Names nobody — a CR 708.2 object has no name for
  // any viewer, its controller included — so `text` reads "a card"
  // and `card_id` is still present for the client to point at the
  // permanent on the board.
  | "turn_face_down"
  // #2590, ADR 0082's 2026-10-07 second amendment: a face-down permanent was
  // turned face up by an EFFECT (CR 708.8). `card_id` is the permanent, public
  // again by then; `target` is the object that did it. The special action has
  // no line of this kind (its `special_action` line carries it).
  | "turn_face_up"
  // ADR 0167 (CR 723.1): a turn-long player control was created, began
  // with the controlled player's turn, and ended as the next turn
  // began. `seat` is the controller, `target_seat` the controlled
  // player, `card_id` the card whose effect it is (Emrakul).
  | "player_control_scheduled"
  | "player_control_began"
  | "player_control_ended";

// LogEvent mirrors `protocol.LogEvent` — one line of the public game
// log. `text` is the rendered, already-redacted sentence; the
// structured fields are the same facts for code that wants them
// (highlighting the card an entry names, filtering by seat).
export interface LogEvent {
  // Engine event sequence number. Monotonic, stable across frames —
  // use it as the keyed-each key.
  seq: number;
  kind: LogKind;
  // Turn sequence the entry happened on. Absent (0) for entries that
  // precede the first step announcement of a game.
  turn?: number;
  // Table-facing round, present on step entries. `turn` is the
  // per-turn sequence identity; this is the number shown to people.
  round?: number;
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
  // ADR 0119 §3: the stack item a `resolve`, `fizzle` or `counter`
  // entry is about, for spells and abilities alike. For a spell it
  // equals its card's ID; an ability's item has an ID no other field
  // carries. Absent on every other kind, and on entries written before
  // the field existed.
  stack_item_id?: string;
  // Instance ID of the CARD the entry acts on (a counterspell's
  // victim, a blocker's attacker). Player targets use target_seat.
  target?: string;
  // Signed / counting payload: life delta, damage dealt, cards drawn.
  amount?: number;
  // #1036: the SIZE of a `scry` / `surveil` entry's keyword action —
  // the "2" in "scry 2" — where `amount` is how many cards the table
  // then watched move. Absent on every other kind, and absent on a
  // scry recorded before the field existed. The rendered `text`
  // already says both; this is here for a client that wants the
  // numbers without parsing the sentence.
  looked_at?: number;
  old_zone?: string;
  new_zone?: string;
  // True when a `damage` entry is combat damage (CR 510).
  combat?: boolean;
  // ADR 0152 §2 (#1856): a `transform` entry whose turn-over was a
  // convert (CR 701.28a). The rendered `text` already says "converted".
  converted?: boolean;
  // Which combat damage step dealt a combat `damage` entry (CR 510.4).
  // Set only when that combat had a first-strike step; combat with no
  // first strike or double strike anywhere is untagged, so its
  // presence alone means there are two beats. Read it, don't derive it
  // from keywords (#187, ADR 0053 Decision 1).
  combat_step?: "first_strike" | "regular";
  // ADR 0118 §2 (#2188): a `cast` entry whose caster cast it without
  // paying its mana cost (`force_cast`). The `text` already ends
  // " without paying its mana cost"; this is for a client that wants
  // the bit. Public: never redacted with the card's name.
  unpaid?: boolean;
  // #3855: the mana paid outside the pool on a `cast` (or `activate`)
  // entry, as a cost string: declared at the "What did you pay?" picker
  // when `paper_declared`, or the shortfall a one-click Cast anyway
  // asserts. `unpaid` is now the picker's "cast without paying its mana
  // cost" answer. The `text` already says all of it.
  paper_paid?: string;
  paper_declared?: boolean;
  // #984: the value named at a "choose a ..." prompt — the colour
  // LETTER on a `choose_color` entry ("G"), the creature type on a
  // `choose_type` one ("Elf"). Absent on a `choose_player` entry,
  // whose answer is `target_seat`, and absent when the viewer is not a
  // knower of the card that asked: the answer identifies the card as
  // loudly as its name does, so it is redacted with it.
  choice?: string;
  // #1021: the printed name of the thing the entry is about when it is
  // not a card — the special action as the card prints it ("Foretell
  // {2}") on a `special_action` entry, the counter kind ("+1/+1") on a
  // `counters` one. Redacted with the card's name exactly as `choice`
  // is: both price or characterise the card the line no longer names.
  label?: string;
  // ADR 0057: why an `eliminated` player left, or how a `game_over`
  // game ended. Absent on every other kind.
  cause?: string;
  // ADR 0075 §2.4: the actor held the table when they took this
  // action. Set only on `spawn` entries, and stamped by the room
  // rather than the projection — the host is a room property, so the
  // engine cannot know it. Absent on a spawn by a non-host (the dev
  // route lets anyone at a preview table spawn) and on one by the
  // admin, who has no seat.
  actor_is_host?: boolean;
  // ADR 0127 §6: on an `auto_answer` entry, the key of the rule that
  // answered. Present only on the chooser's own view.
  auto_answer_key?: string;
  // The rendered line. Already redacted for this viewer: a card the
  // viewer may not identify reads as "a card".
  text: string;
  // Random-effect details. These are optional so older log entries
  // and future effect families remain wire-compatible.
  sides?: number;
  results?: number[];
  faces?: string[];
  // A coin call on a `flip`; on an `auto_answer` (ADR 0127), the answer,
  // "order" for a remembered trigger order (#3016, `label` the rows).
  // On a `fizzle`, "condition": an optional trigger's intervening "if"
  // no longer held as it resolved (CR 603.4, #3907).
  call?: "heads" | "tails" | "pay" | "dont_pay" | "yes" | "no" | "order" | "condition";
  wins?: number;
  // ADR 0121 §3: the seats an `opening_roll` entry names — the seats
  // that tied, or the seats the host rolled for. On `votes_revealed`
  // (ADR 0180), each vote's voter, parallel to `votes`.
  seats?: number[];
  // ADR 0180: what each seat in `seats` voted for, on `votes_revealed`.
  votes?: string[];
  // ADR 0121 §5: a `table_roll` entry's identity, the game's table-roll
  // count after it (1 for the first). An undo of an earlier action
  // writes the line again under a new `seq` and the same `roll_id`, so
  // the dice layer keys the roll on this and plays it once.
  roll_id?: number;
}

// PendingChoiceView mirrors `protocol.PendingChoiceView` server-side.
// Drives the client picker modal for deferred-choice effects
// (Thoughtseize, future Vendilion-style cards). The Options slice
// is pre-filtered to what the viewer is legally allowed to see.
// ADR 0188 Decision 12: the which-skip question's details
// (PendingChoiceView.turn_skip_order).
export interface TurnSkipOrderView {
  // The extra turn the question is about (an entry of turn.extra_turns).
  extra_turn_ref: number;
  // The card whose "skip your next turn" answer 0 spends, and its name
  // as the question was asked.
  skip_source?: string;
  skip_source_name?: string;
  // The extra-turn skips that apply: answer 1's effect.
  others?: string[];
  // The recommended answer's index: always 0, "use my skip".
  recommended: number;
}

export interface PendingChoiceView {
  id: string;
  kind:
    | "discard_from_hand"
    // #2115: the revealed-hand pick with a variant — the chosen card is
    // exiled (`pick_destination`), "you may choose" (`choose_min` 0),
    // or the revealing player's graveyard is offered too
    // (`pick_from_graveyard`). Answered with {choice_id, card_ids}, an
    // empty list when choosing nothing.
    | "revealed_hand_pick"
    // ADR 0166 §4: Word of Command's forced play. `from_player` is the
    // player you control, `options` the one card. Answered by one of
    // the legal moves naming it (`controlled_play`): a cast_spell sent
    // as that player, or {choice_id} with no cards, "it can't be
    // played", which the server takes only when that is true.
    | "controlled_play"
    | "mana_pick"
    | "replacement_order"
    | "optional_replacement"
    // ADR 0115 (CR 903.9a): "your commander was put into a graveyard or
    // exile; put it into the command zone?", asked of its owner as a
    // state-based action. `source` is the commander. Answered with the
    // shared yes/no {choice_id, apply} payload. `playable_from_zone`
    // says whether the owner could cast it from where it is now.
    | "commander_return"
    | "damage_assignment"
    | "trigger_prompt"
    | "pay_unless"
    | "trigger_order"
    | "pick_target"
    // #1196, CR 115.7: "choose the new target for …" — Deflecting
    // Swat, Bolt Bend, Misdirection, Imp's Mischief. Carries the same
    // `pick_target` legal-target payload and is answered on the board
    // the same way; the KIND is what tells the server the answer
    // rewrites an item already on the stack. A prompt whose `min` is
    // 0 is a printed "you may", and the banner's Done button submits
    // the empty list to decline.
    | "retarget"
    // #3158, CR 303.4f: "choose what this Aura enchants as it enters",
    // for an Aura put onto the battlefield without being cast (Sun
    // Titan's Animate Dead, Replenish). The same `pick_target` payload
    // (cards, a card in a graveyard included, and players; min 1 max 1),
    // answered on the board. It targets nothing.
    | "aura_host"
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
    // #996 / ADR 0088: "put these cards on top of / on the bottom of
    // the library in any order" (Brainstorm's put-back, "the rest on
    // the bottom in any order", Aetherspouts' "top or bottom"). The
    // family's fourth member: options are the cards, `placement` says
    // which lanes are open, and the answer is {top_order, bottom} with
    // BOTH lists top-first.
    | "put_in_library"
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
    // is a legal "fail to find" (CR 701.23b) unless search_min says the
    // search must find, so search_max is the ceiling and search_min
    // (default zero) the floor.
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
    // #1210: "as this permanent enters, choose a card name" (CR
    // 614.12) — Pithing Needle, Phyrexian Revoker, Sorcerous
    // Spyglass. Answered with resolve_choice { card_name: "Sol Ring" }.
    //
    // UNLIKE choose_creature_type there is no legal set: CR 201.2
    // admits any card name at all, so name_options is a SUGGESTION
    // list (the names visible in public zones) and the picker is a
    // filter over it beside a free-text box. The server accepts
    // whatever comes back, trimmed and non-empty.
    | "choose_card_name"
    // ADR 0198: "a card you own from outside the game" for a seat whose
    // deck has no sideboard — any card in the card database the effect's
    // filter accepts. collection_options are suggestions (no
    // instance_id); the picker also searches the database through GET
    // /games/{id}/outside/cards. Answered with resolve_choice
    // { card_names: ["Card Name"] }; [] declines when choose_min is 0.
    | "outside_collection"
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
    // #826 CR 502.3: the untap step's own determination — "choose
    // which of these untap", addressed to the active player over the
    // permanents actually in question under a cap ("can't untap more
    // than one land") or an opt-out ("you may choose not to untap
    // this"). Same {choice_id, card_ids} payload and the same
    // choose_min / choose_max bounds as choose_cards, and the same
    // picker renders it. UNLIKE choose_cards the options and bounds
    // reach every seat: the candidates are tapped permanents on the
    // battlefield, which everyone can already see.
    | "untap_choice"
    // #1198 CR 614.1c: "as this land enters, you may reveal an Island
    // or Swamp card from your hand. If you don't, it enters tapped."
    // The permanent is halfway onto the battlefield while this is
    // open — the answer decides HOW it enters, so the land is still in
    // hand and nothing has triggered. Same {choice_id, card_ids}
    // payload, the same choose_min / choose_max bounds and the same
    // picker as choose_cards, and — like choose_cards, unlike
    // untap_choice — options and bounds reach the CHOOSER ONLY: which
    // cards in a hand match the land's clause is exactly the hidden
    // information the question is about. An EMPTY list is the decline,
    // and it is always a legal answer. What was revealed reaches the
    // other seats afterwards, as an ordinary reveal in the log.
    | "entry_reveal_from_hand"
    // ADR 0102 CR 614.12a: "this enchantment enters under the control
    // of an opponent of your choice" (Captive Audience, Pendant of
    // Prosperity). The permanent is mid-entry while this is open. The
    // opponents ride pick_options, one per seat with `player` set, and
    // the answer is {option_index: N} as for option_pick.
    // control_purpose says whether the permanent hurts or helps the
    // seat that receives it.
    | "entry_controller"
    // ADR 0109 §10, CR 702.136a: riot's "a +1/+1 counter, or haste?",
    // asked before the permanent enters. The {choice_id, apply} payload
    // of the yes/no kinds: apply true takes the counter, false haste.
    // Mandatory — both answers are always accepted. accept_label and
    // decline_label name the two; source is the entering card.
    | "entry_riot"
    // ADR 0188 Decision 12, CR 616.1: the player owes "skip your next
    // turn" and an extra-turn skip (Trouble in Pairs, Ugin's Nexus)
    // applies to the same extra turn, so they choose which takes it.
    // The {choice_id, apply} payload: apply true uses their own skip
    // (answer 0, recommended), false lets the other effect skip it.
    // accept_label and decline_label name the two; turn_skip_order
    // carries the details. The turn change waits on it
    // (turn.between_turns).
    | "turn_skip_order"
    // #2123, CR 702.155b: read ahead's "choose the chapter this Saga
    // starts on", asked before the Saga enters. The chapters ride
    // pick_options in order ("Chapter I", "Chapter II", …) and the
    // answer is {option_index: N}, chapter N+1. source is the entering
    // Saga; entry_keyword is "read ahead".
    | "entry_read_ahead"
    // ADR 0098: Mox Diamond's "if this would enter, you may discard a
    // land card instead. If you don't, put it into its owner's
    // graveyard." The reveal's payload and bounds, and — like it — the
    // options and bounds reach the CHOOSER ONLY. An EMPTY list is the
    // decline, and it sends the entering card to the graveyard.
    | "entry_discard_from_hand"
    // ADR 0098 Decision 11: Heart of Yavimaya's "sacrifice a Forest
    // instead". Not a "may": choose_min == choose_max, and the
    // candidates are the chooser's own permanents, so every seat sees
    // them.
    | "entry_sacrifice"
    // ADR 0170: "as this enters, put a phylactery counter on an artifact
    // you control" (Phylactery Lich), "remove all counters from any
    // number of artifacts, creatures, and enchantments" (Sin). Public
    // permanents, answered with card_ids before the permanent is on the
    // battlefield; choose_suggested is the card's recommended pick.
    | "entry_choose_permanents"
    // #1214 CR 608.2 / CR 701.20: an opponent picks from a set you
    // revealed — "target opponent chooses two of those cards" (Gifts
    // Ungiven), and the first leg of a Fact or Fiction pile split.
    // Same {choice_id, card_ids} payload and the same choose_min /
    // choose_max bounds as choose_cards, and the same picker renders
    // it. UNLIKE choose_cards the options and the bounds reach every
    // seat: the card REVEALED them, and that reveal is the only thing
    // entitling the chooser to look at cards out of somebody else's
    // library at all.
    | "reveal_pick"
    // #1214: a seat choosing among the permanents somebody ELSE
    // controls — "for each player, you choose from among the
    // permanents that player controls …" (Tragic Arrogance). NOT
    // targeting: hexproof and shroud do not apply, because nothing is
    // targeted. Public like untap_choice; the candidates are
    // battlefield permanents. `from_player` is whose board is on
    // offer, which is not the chooser.
    | "their_permanents"
    // #1214: "choose N of your own permanents", made on resolution —
    // Scapeshift's "sacrifice any number of lands". The untargeted
    // sibling of a target clause, and the one of the three whose floor
    // is routinely zero.
    | "own_permanents"
    // ADR 0107 §6 (#1860), CR 609.7a: "a source of your choice" — the
    // Circles of Protection, Deflecting Palm. Exactly one of `options`,
    // answered with {choice_id, card_ids: [id]}, on the same grid. The
    // candidates span zones (permanents, spells on the stack, face-up
    // command-zone cards, and graveyard or exile cards something on the
    // stack still refers to), all public, so the options reach every
    // seat.
    | "choose_source"
    // ADR 0114 §4 (#2076), CR 701.54a: "the Ring tempts you — choose a
    // creature you control" as your Ring-bearer. Exactly one of
    // `options` (two or more of the chooser's creatures; with one the
    // server chooses it and logs it), answered with {choice_id,
    // card_ids: [id]} on the shared grid. Not targeting. Public: the
    // candidates are battlefield creatures. Its reason is "choose your
    // Ring-bearer".
    | "ring_bearer"
    // #2525, CR 701.34a: "choose any number of permanents and/or
    // players with counters on them" — a proliferate. The permanents
    // are `options`; the seats on offer are `choose_players`, and a
    // seat is picked by sending its player ID in the same `card_ids`
    // list. `choose_suggested` is the engine's beneficial pick, which
    // the modal pre-selects. Floor zero; public (counters are on the
    // table). Answered with {choice_id, card_ids}.
    | "proliferate"
    // ADR 0108 §7 (#1904), CR 615.7: a charged prevention shield ("the
    // next 3 damage") that meets several damage events at once, more
    // than it can cover — the protected player divides the charge among
    // them before any is dealt. `divide_shield` carries the charge and
    // the events; answered with {choice_id, distribution: {entry id:
    // share}}, the shares adding up to the charge.
    | "divide_shield"
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
    // #804 CR 732: the loop breaker has fired and the repeating
    // ability's controller is asked how many more times it should
    // resolve. Answered with resolve_choice { iterations }, where 0
    // means "stop here" and leaves the table paused exactly where the
    // breaker put it. loop_count is how many times it has already
    // resolved this turn; loop_max_iterations is the ceiling the
    // engine will accept.
    | "loop_shortcut"
    // ADR 0193, CR 732.2b: another player proposes repeating a loop
    // they recorded. Accept, or shorten it to one of your own steps;
    // answered with resolve_loop_proposal, not resolve_choice.
    | "loop_proposal"
    // #764, CR 603.3c: a modal triggered ability's mode, chosen as
    // the ability is put on the stack. mode_options / mode_indexes /
    // mode_min / mode_max / mode_repeatable describe the offer.
    | "mode_pick"
    | string;
  chooser: string;
  // ADR 0166, CR 723.5: the player who owes this prompt when another
  // player controls them. `chooser` is then the controller, who
  // answers it.
  on_behalf_of?: string;
  from_player: string;
  count: number;
  source?: string;
  reason?: string;
  // ADR 0133: the prompt's words name a card in the chooser's hidden
  // hand (an opening-hand action's question). Only the chooser gets
  // `source`, the question and the branch labels; everyone else gets the
  // prompt, its kind and its chooser, with a neutral `reason`.
  private_text?: boolean;
  options?: CardView[];
  // ADR 0116: for "discard_from_hand", the instance IDs among
  // `options` the chooser may pick ("you choose a nonland card from
  // it"). `options` stays the whole revealed hand. Absent means every
  // option (a prompt from before the filter).
  eligible?: string[];
  // ADR 0175: public permanent picks made by earlier selectors in an
  // ordered choice run. Resolve names through this viewer's filtered board.
  earlier_choices?: Array<{ player_id: string; card_ids: string[] }>;
  // ADR 0116: what `eligible` holds, as the card prints it — "nonland
  // card", "card with mana value 3 or greater".
  eligible_label?: string;
  // #2115: for "revealed_hand_pick", where the chosen card goes —
  // "discard", or "exile", which is not a discard.
  pick_destination?: "discard" | "exile" | "keep" | string;
  // #2115: for "revealed_hand_pick", `options` ends with the revealing
  // player's graveyard, every card of which may be chosen.
  pick_from_graveyard?: boolean;
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
  // #780: on a "choose_color" prompt — the card's own declaration of
  // what it will DO with the colour it is handed. Public (it is a
  // reading of the printed text) and absent on every other kind, and
  // on a prompt from a card nobody has annotated yet.
  //
  // It exists because CR 105.4 makes all five colours a legal answer,
  // so nothing else on the prompt says which question is being asked:
  // Coldsteel Heart and Wash Out send the identical five buttons. The
  // picker reads it for its WORDING (colorPromptCopy in manaPick.ts);
  // the ORDER of color_options is already the server's answer to the
  // same question (#986), so nothing here re-sorts.
  color_purpose?: "mana" | "benefit" | "harm" | "filter" | "protect" | string;
  // ADR 0102: populated for kind "entry_controller" — what giving the
  // entering permanent away does to the seat that receives it. The
  // picker reads it only for its wording.
  control_purpose?: "harm" | "benefit" | string;
  // #3309: populated for kind "aura_host" — what the entering Aura does
  // to the permanent it enchants. Absent when the card declares none.
  // The bot reads it; the client ignores it.
  aura_purpose?: "harm" | "benefit" | string;
  // ADR 0109 §10: on an "entry_riot" or "optional_replacement" prompt,
  // the entry keyword it asks about — "riot" or "unleash" — and on an
  // "entry_read_ahead" prompt "read ahead" (#2123). Absent on every
  // other prompt, and on a "may" that is not unleash's.
  entry_keyword?: "riot" | "unleash" | "read ahead" | string;
  // ADR 0188 Decision 12: populated for kind "turn_skip_order".
  turn_skip_order?: TurnSkipOrderView;
  // ADR 0104: on a "trigger_prompt" whose yes TRADES the source for a
  // spell (Perplexing Chimera) — that spell's instance ID. The client
  // does not read it; the bot weighs the trade with it.
  trade_for?: string;
  // S26: populated for kind "choose_creature_type" — every creature
  // type the engine knows, sorted. The list is long by design (the CR
  // 205.3m vocabulary is ~345 entries), so the picker filters it
  // rather than rendering it whole.
  type_options?: string[];
  // #1210: populated for kind "choose_card_name" — the distinct card
  // names visible in a PUBLIC zone (the battlefield, every graveyard,
  // the stack), sorted. A SUGGESTION list and not a legal set: the
  // picker filters it and also offers a free-text box, because CR
  // 201.2 lets a player name a card nobody at the table is holding.
  name_options?: string[];
  // ADR 0198: populated for kind "outside_collection" — the cards the
  // prompt suggests from the card database, each with an empty
  // instance_id (no card exists until one is taken). Sent to the chooser
  // only, with choose_min / choose_max.
  collection_options?: CardView[];
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
  // #3015: for kind "trigger_order", one sentence in the player's
  // terms that says why the game asked instead of ordering the batch
  // itself ("Ajani's Pridemate watches life totals."). Display text,
  // shown under the heading; absent when there is no reason.
  trigger_order_why?: string;
  // #3016 (ADR 0018's eighth amendment): on a keyed trigger_order
  // prompt, true when the chooser has an order remembered for this
  // batch — the "Use this order again" toggle's starting state. The
  // chooser's alone.
  trigger_order_remembered?: boolean;
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
  // ADR 0108 §7: populated for kind "divide_shield".
  divide_shield?: DivideShieldView;
  // S19 follow-up: populated for kind "trigger_prompt" — true when
  // the optional trigger has no legal target and answering "Yes"
  // will pass without effect (Reclamation Sage with no opponent
  // artifact, Eternal Witness with an empty graveyard, etc.). The
  // modal warns the chooser. Absent/false otherwise.
  no_legal_target?: boolean;
  // ADR 0197 §10 (#2731): populated for kind "trigger_prompt" — true
  // when the ability itself says "Yes" would leave the chooser no
  // better off as things stand (Arni Brokenbrow's new base power would
  // be no higher than its current one). The dock warns. Absent/false
  // otherwise.
  no_gain?: boolean;
  // S19 sub-PR 6: populated for kind "pay_unless" — the cost the
  // chooser is being asked to pay ("{2}"). Same {choice_id, apply}
  // payload as the other yes/no kinds: apply=true pays (from pool,
  // auto-tapping if short), apply=false declines and the card's
  // "unless" consequence fires. Also carries the life payment ("2
  // life") for kind "entry_pay_life".
  pay_cost?: string;
  // ADR 0131 §2 (#2531): the "or 2 life" half of a mana "pay_unless".
  // `phyrexian_symbols` is how many symbols of `pay_cost` the chooser
  // could pay 2 life each for (a printed {B/P}, or a {B} under K'rrik);
  // `phyrexian_granted` how many of those are the grant's. The answer
  // is {apply: true, phyrexian_life: n}. Absent at zero.
  phyrexian_symbols?: number;
  phyrexian_granted?: number;
  // ADR 0127 §2, §8: the key a standing answer to this prompt is filed
  // under (absent when it can take none), and its display copies — the
  // card's name and the question — which Settings shows beside the rule.
  // `asked_by_hand` says why a prompt that has a rule is asked anyway.
  // All four are on the chooser's own view only.
  auto_answer_key?: string;
  auto_answer_card?: string;
  auto_answer_prompt?: string;
  asked_by_hand?: AskedByHand;
  // #1311: populated for a "pay_unless" whose payment is a WATERBEND
  // cost ("Ward—Waterbend {4}", The Unagi of Kyoshi Island): the
  // chooser's untapped artifacts and creatures that may each pay {1}
  // of pay_cost's generic (CR 701.67a), and `max`, how many. The
  // "Pay" answer names them as `tap_ids` beside `apply: true`; the
  // rest is paid from the pool and auto-tap as usual. Absent for
  // every other pay-unless.
  tap_cost?: TapCostView;
  // ADR 0108 §5: a "pay_unless" whose payment is not mana ("Echo—
  // Discard a card", "Cumulative upkeep—Sacrifice a land"). pay_cost
  // carries the payment's words. The "Pay" answer names exactly
  // `count` of `options` as `card_ids` beside `apply: true`. A
  // discard's options are the chooser's hand and reach the chooser
  // only. Absent for a mana payment.
  pay_cards?: PayCardsView;
  // ADR 0129 §3: the energy a "pay_unless" asks for ("you may pay
  // {E}{E}", "unless you pay {E}"). pay_cost carries the symbols. Paying
  // 0 {E} is a payment too, so 0 is present. A "Pay" from a seat with
  // less energy is a decline.
  pay_energy?: number;
  // ADR 0150 §3: the other payments of a "pay_unless" that name nothing,
  // each answered with { apply: true } alone. pay_life: pay that much
  // life (a seat with less, or whose life can't change, can't pay).
  // pay_damage: the prompt's source deals that much damage to the
  // chooser, always a payment. pay_nothing: "yes" costs nothing
  // (Distant Memories). At most one is present.
  pay_life?: number;
  pay_damage?: number;
  pay_nothing?: boolean;
  // ADR 0187 §5: an either payment — one offer with two or more ways to
  // pay (Worms of the Earth: "sacrifice two lands or take 5 damage").
  // One entry per branch, in the card's order; pay_cost holds the whole
  // offer's words. The "Pay" answer names the branch as `pay_option` (an
  // index here) beside `apply: true`, with `card_ids` for a branch that
  // names cards. Instead of pay_cards, pay_life, pay_damage, pay_energy
  // and pay_nothing.
  pay_either?: PayOptionView[];
  // ADR 0129 §3: a "pay_amount" prompt, "you may pay any amount of
  // {E}". Answered with { amount }: 0, or min..max.
  pay_amount?: PayAmountView;
  // S22: populated for kind "search_library" — how many of `options`
  // the searcher may take. Absent for every other kind, and absent for
  // non-chooser viewers, who are not told what the search is for.
  search_max?: number;
  // #3053: the fewest the searcher may take. Absent (zero) for a search
  // for a card with a stated quality (CR 701.23b — "Find nothing" is a
  // legal answer even against exactly one match). At least one only for
  // an unconditional "a card" / "three cards" search (CR 701.23d), where
  // an empty answer is refused and the client offers no such button.
  search_min?: number;
  // #3217: on a "search your library and/or graveyard" prompt, the
  // instance IDs among `options` that are in the searcher's graveyard.
  // Everything else is a library card. Absent for a library-only search.
  search_graveyard?: string[];
  // #3988: on a "search that player's graveyard, hand, and library"
  // prompt (Kotose, the Silent Spider), the instance IDs among `options`
  // that are in the searched player's hand, and `search_all_zones` set:
  // every zone is searched, so the library is shuffled whatever is taken.
  search_hand?: string[];
  search_all_zones?: boolean;
  // ADR 0160: a "copy_target" prompt whose pick is the creature to
  // exchange text boxes with rather than the object to copy (Deadpool,
  // Trading Card). Same {card_ids} answer, same empty decline.
  exchanges_text?: boolean;
  // #996 / ADR 0088: populated for kind "put_in_library" — which lanes
  // the answer may use. "top" answers {top_order} alone, "bottom"
  // answers {bottom} alone (top-first), "top_or_bottom" answers both.
  placement?: "top" | "bottom" | "top_or_bottom";
  // #1298: refine a put_in_library's top lane. `top_count` is EXACTLY
  // how many cards `top_order` must hold (Cream of the Crop's "put one
  // of those cards on top"; absent is any number). `top_depth` is where
  // the top lane lands, counted from the top (2 is Temporal Cleansing's
  // "second from the top"; absent is the top). Both public.
  top_count?: number;
  top_depth?: number;
  // #74: populated for kind "confirm" — the card's own words for the
  // accept and decline branches. Absent means the client renders Yes /
  // No, which is right for a prompt that really is a yes/no.
  /**
   * pick_options populates the "option_pick" kind (#568): one entry
   * per branch of "choose one of the following", in the card's
   * printed order. Answered with `{option_index: N}` — the INDEX,
   * because an option is a consequence and not always a set of cards.
   *
   * An option's own `cards` are context the client renders beside the
   * label (a Fact or Fiction pile); they are already redacted
   * per-viewer by the server, and an option over cards this seat may
   * not see arrives with the label and no cards at all.
   */
  pick_options?: PickOptionView[];
  /**
   * ADR 0146: present when this option_pick is one player's BALLOT in a
   * vote (CR 701.38), carrying the vote so far. Public: every seat sees
   * each vote as it is cast.
   */
  council_vote?: CouncilVoteView;
  /**
   * ADR 0150: present when this pay_unless is one link of an "unless any
   * player pays" offer: whose effect it is, everyone it may reach in the
   * order asked, and who has declined. Public on every seat's view.
   */
  any_player_offer?: AnyPlayerOfferView;
  /**
   * ADR 0150 §5: present when this pay_amount asks an assist helper how
   * much of another player's spell to pay: who is casting what. Public on
   * every seat's view; the caster's dock shows a waiting line from it.
   */
  assist?: AssistPromptView;

  accept_label?: string;
  decline_label?: string;
  // ADR 0099 §7: populated for kind "may_cast". The rule the offer is
  // asked under ("cascade", "discover", "suspend", "madness") and the
  // instance id of the exiled card it offers. A may_cast also carries
  // accept_label / decline_label when the server named its branches.
  may_cast_keyword?: string;
  may_cast_card?: string;
  // ADR 0115 decision 2: populated for kind "commander_return" — the
  // owner could cast or play the commander from the graveyard or exile
  // it is in now (escape, flashback, an adventure), ignoring timing and
  // mana. Computed by the server on every view. #2390: also set on the
  // "optional_replacement" that is CR 903.9b's commander question when
  // the commander is headed for its owner's hand.
  playable_from_zone?: boolean;
  // #2390: populated for an "optional_replacement" that offers a dredge
  // (CR 702.52a) — N, the number of cards the yes mills. The card it
  // returns rides `source`. Computed by the server on every view.
  dredge?: number;
  // #2419: populated for an "entry_sacrifice" that is devour (CR
  // 702.82a) — N, the +1/+1 counters each sacrificed creature buys, and
  // what the entering creature's own "for each creature it devoured"
  // ability pays per creature (cards drawn, life gained). Absent on the
  // fixed-count sacrifice lands. Computed by the server on every view.
  devour?: number;
  devour_draw?: number;
  devour_life?: number;
  // #74: populated for kind "confirm" — the life the ACCEPT branch
  // charges (Sylvan Library's 4). Absent when the branch costs no
  // life. The label already says it; this is the number, for anything
  // that needs to reason about the price rather than print it.
  // For pay_unless, mandatory life paid alongside pay_cost.
  life_cost?: number;
  // Computed affordability of the mandatory life, including life locks.
  life_cost_unpayable?: boolean;
  // #74: populated for kinds "choose_cards", "untap_choice",
  // "entry_reveal_from_hand", #1214's three resolution-time picks
  // ("reveal_pick", "their_permanents", "own_permanents") and ADR 0107's
  // "choose_source" (one of one) — how few
  // and how many of `options` the chooser must pick. Absent for every
  // other kind, and (for the two kinds whose candidates are cards in a
  // hand — choose_cards and entry_reveal_from_hand) absent for
  // not told the size of a choice over someone else's hidden cards.
  choose_min?: number;
  choose_max?: number;
  // #2680 / #3270: kind "choose_cards" only, and only to its chooser.
  // Where the named card goes when the card says so: "battlefield",
  // "battlefield_tapped", or "cast_free" ("you may cast a spell from
  // your hand without paying its mana cost" — answering with a card
  // opens its cast at once), "cast_free_library" (the same over cards the
  // chooser looked at on top of their own library, #2110), "outside" (ADR 0191: the named cards
  // come in from the chooser's sideboard), "cast_free_outside" / "play_outside" (#3955: the named
  // sideboard cards are cast free, or played this turn, from outside the game; the pick opens the
  // sideboard tab), or "manifest" (ADR 0189: the named hand
  // cards are put onto the battlefield face down as 2/2 creatures).
  // Absent on a plain discard or look.
  choose_destination?: string;
  // #2525: kind "proliferate" (and ADR 0170's "entry_choose_permanents",
  // suggestion only). The seats on offer beside
  // `options` (player IDs), and the engine's suggested answer
  // (instance IDs and player IDs) for the modal to pre-select. A
  // default, not a rule: any subset is a legal answer.
  choose_players?: string[];
  choose_suggested?: string[];
  // CR 603.2d: when this is a trigger_prompt or pick_target choice,
  // the public permanent that caused the additional trigger. The
  // server omits both fields for ordinary choices.
  doubled_by?: string;
  doubled_by_name?: string;
  // #3817: on a trigger's pick_target, the trigger as card and ability.
  // Every copy of one card shares it, two abilities do not, so it groups
  // "same target for all of this card's triggers". Opaque, the
  // chooser's alone, and absent for a trigger with no catalog identity.
  shared_target_key?: string;
  // S30 coin call prompt metadata. `coins` is the number of coins
  // covered by one call; wins tracks an ongoing chain.
  allow_stop?: boolean;
  coins?: number;
  max_useful_wins?: number;
  wins?: number;
  // #804: populated for kind "loop_shortcut" — how many times the
  // repeating ability has already resolved this turn, and the largest
  // answer the engine accepts. `reason` carries "<card> — <ability>".
  loop_count?: number;
  loop_max_iterations?: number;
  // ADR 0193: populated for kind "loop_proposal".
  loop_proposal?: LoopProposalView;
  // #764, CR 603.3c: populated for kind "mode_pick" — the bullets a
  // modal TRIGGER offers its controller as the ability is put on the
  // stack. Only the choosable ones are listed (a bullet whose clause
  // has no legal target is dropped), so mode_indexes carries the
  // ModeSpec index each label belongs to and that is what the answer
  // sends back: `resolve_choice {choice_id, modes: [i, …]}`.
  // Repeats are legal only when mode_repeatable (CR 700.2d).
  mode_options?: string[];
  mode_indexes?: number[];
  mode_min?: number;
  mode_max?: number;
  mode_repeatable?: boolean;
  // ADR 0097 (#1749): the bullets this object's ability has ALREADY
  // chosen under "choose one that hasn't been chosen [this turn]", in
  // printed order and parallel to each other. Beside the offer, not in
  // it — `mode_indexes` is still what may be answered. The picker
  // shows them disabled; `mode_not_chosen` words the note.
  mode_used_options?: string[];
  mode_used_indexes?: number[];
  mode_not_chosen?: ModeNotChosen;
}

// ModeNotChosen is a "that hasn't been chosen" restriction (ADR 0097):
// once per turn, or once ever, for one object's ability.
export type ModeNotChosen = "this_turn" | "ever";

// ReplacementOptionView mirrors protocol.ReplacementOptionView —
// one entry in a replacement_order prompt's candidate list. ID is
// a decimal-string ReplacementEffectID; label is the prompt copy
// ("Doubling Season: double counters"); source_card_id is the card
// hosting the effect (empty for engine built-ins like commander-
// zone replacement). Added in S17 sub-PR 2.
/**
 * PickOptionView is one branch of an "option_pick" prompt (#568):
 * the card's own words for it, the cards it is about (a pile, or
 * nothing), and the life it charges.
 */
export interface PickOptionView {
  label: string;
  cards?: CardView[];
  life_cost?: number;
  /**
   * The seat this option is about, for the prompts whose branches ARE
   * players — "choose a player" / "choose an opponent" (#929) and
   * True-Name Nemesis's as-enters sibling (#980). Absent on every
   * other option, which is all of them.
   *
   * `label` is still what the player reads; this is the identity, so
   * the client can render a seat chip rather than parse a name back
   * out of the text. Public by CR 400.2 — who is seated is not hidden
   * — so it is never redacted. #994.
   */
  player?: string;
  /**
   * The mana this option costs its chooser ("{1}", "{B}"), paid by the
   * server when the option is chosen (#2854): Winter's Chill's "may pay
   * {1} or {2}". Absent on an option that costs no mana. Only options
   * the chooser can pay are listed; an answer the board can no longer
   * pay is refused and the prompt stays open.
   */
  mana_cost?: string;
  /**
   * What this option costs its chooser besides life and mana, in cards
   * (#3140): the hint a bot reads on a villainous choice (CR 701.55).
   * The client shows the option's label and nothing of this.
   */
  harm?: number;
}

export interface ReplacementOptionView {
  id: string;
  label?: string;
  source_card_id?: string;
  // #3016: on a keyed trigger_order prompt, which of the batch's rows
  // this trigger is (1, 2, …; copies of one ability share it). Only an
  // order that keeps each row together can be remembered. The
  // chooser's alone.
  row?: number;
}

// DamageAssignmentView mirrors protocol.DamageAssignmentView — the
// wire shape of a CR 510.1c multi-blocker combat damage prompt.
// The attacker's controller distributes attacker_power across the
// blockers (respecting at-least-lethal-in-order) and, if
// allow_trample, can overflow leftover to the defending player.
// Added in S18 sub-PR 3.
// ADR 0108 §7 (#1904), CR 615.7: a divide_shield prompt. `charge` is
// what the shield has left; each entry is one damage event it meets —
// `source_name` would deal `amount` to `target_name` (a player when
// target_is_player). The answer gives each entry 0..amount, adding up
// to `charge`.
export interface DivideShieldView {
  label?: string;
  charge: number;
  entries: DivideShieldEntryView[];
}

export interface DivideShieldEntryView {
  id: string;
  source_id: string;
  source_name?: string;
  target_id: string;
  target_name?: string;
  target_is_player?: boolean;
  amount: number;
  combat?: boolean;
}

export interface DamageAssignmentView {
  attacker_card_id: string;
  blocker_card_ids: string[];
  attacker_power: number;
  allow_trample?: boolean;
  has_deathtouch?: boolean;
  // #1706, CR 510.1d: the card in attacker_card_id is a BLOCKER that
  // blocks every creature in blocker_card_ids, and its controller
  // divides its damage among them freely — no order, no trample.
  blocker_divides?: boolean;
  // #2956 (ADR 0147): the server's canonical split
  // (legal.CanonicalDamageSplit), the one the bots answer with. The
  // prompt pre-fills it, and with "Auto-assign combat damage" on the
  // client sends it unasked when covers_lethal is set. `lethal` is the
  // lethal damage of each blocker, indexed like blocker_card_ids.
  suggested?: DamageSplitView;
  lethal?: number[];
  covers_lethal?: boolean;
}

// DamageSplitView is a damage-assignment answer in the resolve_choice
// payload's own shape (#2956).
export interface DamageSplitView {
  assignments: { blocker_id: string; amount: number }[];
  trample_to_player?: number;
}

// DelayedTriggerView mirrors `protocol.DelayedTriggerView`
// server-side: one queued CR 603.7 delayed triggered ability. `at`
// is the step whose beginning fires it ("end", "upkeep"); `cards`
// are the instance IDs the effect acts on. `on` (#663) is the event
// condition of a "when you next cast …" trigger, which is owed on the
// next matching event rather than at a step — such a trigger carries
// `on` and an empty `at`. Added in S22.
export interface DelayedTriggerView {
  id: string;
  controller: string;
  source?: string;
  label?: string;
  at: string;
  created_seq?: number;
  cards?: string[];
  on?: string[];
}

// StackItemView mirrors `protocol.StackItemView` server-side: the
// announce-time metadata for one item on the stack.
export interface StackItemView {
  id: string;
  kind: "spell" | "activated" | "triggered";
  controller: string;
  // ADR 0104: the player a stolen spell was put on the stack by — its
  // caster (CR 110.2b). Present only while it differs from
  // `controller`, i.e. while somebody else has taken the spell.
  default_controller?: string;
  owner: string;
  source_card_id: string;
  label?: string;
  // #3830: the printed Oracle text of the ability this item is — a
  // triggered ability's clause, an activated ability's whole
  // "cost: effect" line. Absent on a spell, a copy, an item with no
  // single clause of its own (granted, delayed, rule-made) and any
  // item whose source the table cannot see; show `label` then.
  ability_text?: string;
  targets?: TargetRefView[];
  modes?: number[];
  // #764: the oracle bullet of each chosen mode, in announce order
  // and with repeats. The caster's hand card is gone by the time the
  // spell is on the stack, so the overlay reads the labels off the
  // item rather than printing raw indexes.
  mode_labels?: string[];
  x_value?: number;
  distribution?: Record<string, number>;
  hold_priority?: boolean;
  split_second?: boolean;
  // #1553: the spell can't be countered — its own printed rider or the
  // mana that paid for it (Cavern of Souls). Read from the engine's one
  // counter gate; absent when false and on ability items.
  cant_be_countered?: boolean;
  // ADR 0107 §5: the spell's own "the damage can't be prevented", under
  // its condition as it stands now (Combust; Banefire with X of 5 or
  // more). Absent when false.
  damage_cant_be_prevented?: boolean;
  // S22: the alternative cost this spell was cast for — "overload",
  // "evoke", "cleave" — absent for an ordinary cast. Load-bearing
  // for anyone deciding whether to respond: an overloaded Cyclonic
  // Rift is a one-sided wipe, a hard-cast one is a single bounce.
  alt_cost?: string;
  // CR 702.174 (#1267): the player the gift was promised to. Absent
  // when no gift was promised, which includes every spell without
  // one. Public — the promise is announced at CR 601.2b, and whether
  // it was made changes what the spell does.
  gift_to?: string;
  // S30: a CR 707.10 spell copy rather than a cast card. The copy
  // and its source look identical on the stack, and which is which
  // decides what countering one leaves behind.
  is_copy?: boolean;
  // CR 603.2d: public attribution for an additional triggered
  // ability created by a trigger-doubling permanent.
  doubled_by?: string;
  doubled_by_name?: string;
  // #761: what paid for this spell — how many mana, and the distinct
  // COLOURS among them in WUBRG order (colourless is not a colour, so
  // it never appears here even though it counts in mana_spent). Mana
  // is spent face up, so this is public, and a responder to a
  // converge spell needs to see how wide it converged.
  //
  // mana_spent_unknown means the cast went through permissive mode or
  // a strict-mode override: the engine never took the mana and has no
  // record of what it was. Render that as unknown, never as zero —
  // "nothing was spent" is a different and much stronger claim, and
  // the one Vexing Bauble punishes.
  mana_spent?: number;
  colors_spent?: string[];
  mana_spent_unknown?: boolean;
}

// TargetRefView mirrors `protocol.TargetRefView` server-side: a
// single announce-time target slot.
export interface TargetRefView {
  kind: "player" | "card" | "self" | "none";
  id?: string;
  // #764: the target CLAUSE this pick answered — the clause index,
  // and the index into the item's `modes` whose clause list that is.
  // Both omitted at zero, which is every single-clause non-modal
  // announcement.
  slot?: number;
  mode?: number;
}

// CouncilVoteView is a rules vote in progress (ADR 0146, CR 701.38), as
// the open ballot carries it. Not the sandbox VoteView below: nobody can
// re-cast or end it, and the ability that called it reads the tally.
export interface CouncilVoteView {
  // The controller of the ability that called the vote, and the player
  // voting now.
  controller: string;
  voter: string;
  // Where the cards the options name are, when not the battlefield
  // ("graveyard": Custodi Squire). Absent otherwise.
  zone?: string;
  // The vote's choices in printed order, and the votes each has so far.
  options: string[];
  tally: number[];
  // In a secret vote, only the viewer's own ballots (option -2 is a vote
  // for nothing).
  ballots?: { voter: string; option: number }[];
  // ADR 0180: a secret vote. The tally is all zeros and `ballots` holds
  // the viewer's own; `voted` lists who has voted so far. The votes are
  // revealed together in the log (`votes_revealed`) when the last is in.
  secret?: boolean;
  voted?: string[];
  // ADR 0180: the "vote for nothing" option of a vote for up to one.
  abstain?: string;
  // ADR 0180, a bot hint: a voter wants the option naming themselves.
  self_best?: boolean;
  // Each of the prompt's pick_options as an index into `options`; -1 is
  // the "Don't vote again" option of an extra vote the voter may decline,
  // -2 the "vote for nothing" option of a vote for up to one.
  offered: number[];
  optional?: boolean;
  votes_left?: number;
  // Bot hints: how much the controller, and an opponent, wants each
  // option to win.
  for_controller?: number[];
  for_opponents?: number[];
}

// AnyPlayerOfferView is an "unless any player pays" offer in progress
// (ADR 0150 §5), as its open link carries it. The link's chooser is the
// player asked now.
// AssistOfferView is a spell with assist's offer to its caster (ADR 0150
// §5): "up to {max}", and the players who may be asked.
export interface AssistOfferView {
  // The generic part of the total cost for the default announcement out
  // of this zone, commander tax and cost modifiers included.
  max: number;
  // Every other player still in the game, in seat order.
  players: string[];
}

// AssistPromptView is an assist helper's prompt (ADR 0150 §5), as the
// pay_amount carries it. The prompt's chooser is the helper.
export interface AssistPromptView {
  caster: string;
  // The card being cast, its name and its total cost as announced.
  spell: string;
  spell_name: string;
  mana_cost?: string;
  // The generic part of the total cost; the prompt's max is the most
  // the helper can pay, which may be less.
  generic: number;
}

export interface AnyPlayerOfferView {
  // The controller of the effect being paid against ("you" on the card).
  controller: string;
  // Everyone the offer may reach, in the order asked (APNAP from the
  // active player), and those who have declined so far.
  order: string[];
  declined?: string[];
  // "Any opponent may": the controller is not asked.
  opponents?: boolean;
  // Paying is what the controller wants (Icy Prison: it keeps the
  // permanent), so the words say "keep", not "stop".
  helps_controller?: boolean;
  // ADR 0187 §2: the life one object costs on a per-object offer. Set,
  // the prompt is a choose_cards whose answer is the objects the chooser
  // will pay that much for; earlier_choices carries each earlier player's
  // picks. Absent on every other any-player offer.
  each_life?: number;
}

export interface VoteView {
  id: string;
  topic: string;
  options: string[];
  initiator: string;
  // voter player ID → option index
  ballots: Record<string, number>;
}

// ManaPoolRun is one run of PlayerView.mana_pool (#3041).
export interface ManaPoolRun {
  color: string;
  count: number;
}

export interface PlayerView {
  id: string;
  name: string;
  seat: number;
  life: number;
  poison?: number;
  energy?: number;
  // ADR 0138 (CR 702.179): the player's speed, 1 to 4; absent while
  // they have none. 4 is max speed (CR 702.179e). Public.
  speed?: number;
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
  // Set when the player has left the game: a concession (S08), a
  // state-based loss, or an effect loss (ADR 0057). An effect WIN
  // eliminates nobody. Eliminated players still appear in the
  // seats list and may continue to spectate; the UI greys them out
  // and disables their action buttons. Omitempty on the wire — only
  // present when true.
  eliminated?: boolean;
  // True once the player has committed to their opening hand during
  // the mulligan window. Omitempty on the wire — absent means false.
  // Added in S08.
  hand_kept?: boolean;
  // CR 103.5 (#2237): true on the one seat whose turn it is to keep or
  // mulligan. Mulligan decisions go in turn order, starting player
  // first. Omitempty: absent on every other seat.
  mulligan_turn?: boolean;
  // ADR 0168: how many cards this seat puts on the bottom of its library
  // if it keeps now (the London mulligan, CR 103.5). Omitempty: absent
  // when zero, outside the window, and once the seat has kept.
  mulligan_bottom?: number;
  // #1968: this seat's trigger-order mode, "always" or "never"; absent
  // means the default, "when_it_matters". Present only in the seat's OWN
  // view; the server blanks it for everyone else.
  trigger_order?: "always" | "never";
  // #1530: true when trigger_order is "always", kept for older clients.
  // Own view only, like trigger_order.
  trigger_order_always_ask?: boolean;
  // ADR 0127 §3: this seat's standing answers, sorted by key. Own view
  // only; the client reconciles it with gameplay.autoAnswers.
  auto_answers?: AutoAnswerRuleView[];
  // ADR 0127 §6 (owner amendment 2026-10-07): the seq of this seat's
  // automatic answer while it is the top undo entry, so `undo` takes it
  // back. Own view only; absent once anything else sits on top.
  undo_auto_answer?: number;
  // #3016: how many trigger orders this seat has remembered this game,
  // for the game menu's "Forget remembered trigger orders (N)". Own
  // view only.
  trigger_order_memory_count?: number;
  // Why an `undo` from this seat would be refused by the undo stack
  // right now (#2963): "nothing" (nothing on the stack), "not_yours"
  // (the most recent action is another seat's), "game_over" or
  // "secret_vote" (it revealed a secret vote, ADR 0180), "past_bot"
  // (a bot has acted since) or "too_deep" (past the last eight
  // actions), the last two refused for everyone, admin included
  // (ADR 0194). Absent means the stack allows it; the per-turn budget
  // is separate (undos_remaining). Own view only.
  undo_blocked?: "nothing" | "not_yours" | "game_over" | "secret_vote" | "past_bot" | "too_deep";
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
  // AI agent seat (ADR 0122 §7): an outside model joined as a guest
  // through the MCP seat binary. NOT a bot. Declared at join, never
  // cleared, identical for every viewer. agent_client is the MCP
  // client's name, [a-z0-9._-] cut to 32 characters, or "unknown".
  is_agent?: boolean;
  agent_client?: string;
  // #2696 (CR 702.131c): the city's blessing, a player designation
  // ascend gives and nothing takes away. Public, identical for every
  // viewer; absent until the seat has it.
  citys_blessing?: boolean;
  // ADR 0188 Decision 9 (#3559, CR 614.10): how many of this player's
  // next turns they will skip. Public, identical for every viewer; absent
  // at zero.
  turn_skips?: number;
  // The playmat the seat's signed-in owner chose (ADR 0128): a
  // same-origin path, /playmats/<uuid>, drawn behind that seat's
  // battlefield. Public and identical for every viewer. Absent for a
  // guest, a bot, an agent and a person with none. Use playmatSrc()
  // (lib/playmat.ts) to turn it into something an <img> can load.
  playmat_url?: string;
  // How strongly the owner darkens that playmat under the cards, in
  // percent, 30 to 90 (ADR 0128 amendment). The same for every viewer.
  // Absent with no playmat; draw 58 when absent.
  playmat_wash?: number;
  // Table host (ADR 0075 §2.1), visible to every viewer. The host may
  // change table settings alongside the server admin.
  is_host?: boolean;
  // ADR 0177 §1: the seat's player is not at the table. Absent while
  // present, and always on a bot seat or one that has left the game.
  // Stamped by the hub on every frame; public.
  presence?: PresenceView;
  // ADR 0177 §4: a bot is playing this person's seat while they were
  // away. is_bot stays false: the seat is still theirs. Public.
  stand_in?: StandInView;
  // Per-commander cast count for the Commander tax (S13.1, CR
  // 903.8). Keyed by commander instance UUID. Drives the "+N tax"
  // indicator next to the commander tile.
  commander_casts?: Record<string, number>;
  // Per-player named counter map (S13.2 — poison, energy,
  // experience, rad, plus homebrew). The legacy `poison` and
  // `energy` ints above stay populated for backwards compat.
  counters?: Record<string, number>;
  // Per-player cleanup-step hand-size cap (S13.4, CR 402.2).
  // 7 by default; -1 = no cap (Reliquary Tower / Thought Vessel);
  // otherwise 0 or more (ADR 0113 §3: Null Profusion's 2,
  // Jin-Gitaxias's 0). The EFFECTIVE value. playerKeywordBadges.ts
  // shows a badge on a seat whose value is not 7.
  // Always present on the wire; the field is non-omitempty so
  // clients know the cap even when it's the default.
  max_hand_size?: number;
  // #500 (CR 305.2): how many lands this seat may play this turn,
  // and how many it already has. The engine REFUSES a land play past
  // the allowance, so a client should grey out the hand's lands when
  // lands_played_this_turn >= land_drops_per_turn rather than only
  // explain the rejection afterwards — timing.ts landDropsSpent does,
  // for a frame with no move list (#2203). land_drops_per_turn is the
  // EFFECTIVE allowance — a controlled Exploration or a one-turn
  // grant is already summed in. Normally 1 / 0.
  land_drops_per_turn?: number;
  lands_played_this_turn?: number;
  // S15: per-player mana pool, in insertion order, as runs of one
  // colour (#3041): `color` is an uppercase mana letter ("W", "U",
  // "B", "R", "G", "C") and `count` how many of it sit together.
  // Read the amount with manaPoolTotal (lib/manaPool.ts), never
  // `.length`, which counts runs. Empties at every step boundary
  // (CR 106.4), so this is absent / empty in the common case
  // outside an active cast sequence.
  mana_pool?: ManaPoolRun[];
  // #623 (CR 114): the emblems this seat has, in creation order.
  // Absent for a seat with none, which is nearly every seat.
  //
  // Not a ZoneView — an emblem has no characteristics at all, so it
  // is not a card and there is nothing for the card renderer to draw.
  // The board shows these as chips beside the player identity, with
  // `text` as the hover. Public: every seat sees every emblem.
  emblems?: EmblemView[];
  // ADR 0172 §5: the companion this seat revealed, while it waits
  // outside the game (CR 702.139a). Absent for a seat with none and
  // once it has been put into the hand. Public: every seat sees the
  // card. Its `special_actions` carry the {3} row ("Put into your hand
  // (companion)", kind "companion") for the owner only, and the row
  // sends `special_action { card_id, kind: "companion" }`.
  companion?: CardView;
  // ADR 0191 §7: the cards this seat owns outside the game, its deck's
  // sideboard. OWNER ONLY: present on your own seat and absent on every
  // other, for a spectator and for the controller of your turn. Only an
  // effect that brings a card in from outside the game (a wish,
  // Extrapolate the Impossible) can reach them; the zone browser shows
  // them read-only.
  sideboard?: CardView[];
  // How many cards this seat owns outside the game. Public, like a
  // hand's size. Absent for a seat with none.
  sideboard_count?: number;
  // #3955 (Wish): the seat's live "you may play a card you own from outside
  // the game this turn" right, OWNER-ONLY. `can_name` is true when the seat's
  // pool is the card database, so the `outside_play` action is how it names
  // the card to play; with a sideboard its cards carry the cast surface
  // directly.
  outside_play?: { source_name: string; can_name?: boolean };
  // #1197/#1201 (CR 702.11d, CR 702.16i): the abilities this SEAT has
  // right now — bare engine tokens, "hexproof" or a general
  // "protection from <quality>" ("protection from everything" is the
  // only quality a catalogued card grants a player today — Teferi's
  // Protection, The One Ring, Leyline of Sanctity, Aegis of the
  // Gods). Derived grants (a controlled Leyline) come first, then
  // ones granted for a duration (Teferi's Protection).
  //
  // EFFECTIVE, like max_hand_size and land_drops_per_turn — computed
  // on every projection, not read off stored state. PUBLIC and
  // unredacted, like emblems: protection and hexproof on a seat are
  // facts about the board, and `legal_targets` already excludes a
  // protected seat for a viewer who could not otherwise tell why.
  // Absent for a seat with none, which is nearly every seat.
  //
  // Unlike CardView.protection, the wire does NOT parse the quality
  // into a structured ProtectionView here — there is exactly one
  // production quality today, so a second structured field for one
  // value wasn't worth shipping (server/internal/protocol/view.go).
  // playerKeywordBadges.ts title-cases the parsed quality itself
  // rather than switching on today's two spellings, so a future card
  // granting a player some other quality reaches a badge without a
  // wire change.
  keywords?: string[];
  // #1200 (CR 119.7, CR 119.8): "your life total can't change" on
  // this seat — Platinum Emperion's printed static, or a grant that
  // lasts until this player's next turn (Teferi's Protection,
  // Teferi's Reproach). Absent for a seat with none, which is nearly
  // every seat.
  //
  // EFFECTIVE and PUBLIC, same posture as `keywords` above. It is a
  // separate field rather than another token in that list because a
  // life-total lock is not an ability the player HAS — the list is
  // engine ability tokens and the client parses "protection from
  // <quality>" out of it. See ADR 0085 Decision 7.
  //
  // What it means at the table: no life gain, no life loss, no life
  // PAYMENT (so a Phyrexian symbol, a shockland's 2 life and Snuff
  // Out are all off the table for this seat). Damage is still dealt
  // and still triggers; it just moves no life. Poison still lands.
  life_total_locked?: boolean;
  // ADR 0107 §5 (CR 119.7): this seat can't gain life — a battlefield
  // static (Leyline of Punishment), a turn grant (Skullcrack) or the
  // rest of the game (Screaming Nemesis). Absent when false.
  cant_gain_life?: boolean;
  // ADR 0109 §4 (CR 101.2): the clause that stops this seat playing ANY
  // land from its hand right now ("Players can't play lands — Territorial
  // Dispute", "You can't play lands this turn — Turf Wound"). Absent when
  // nothing does. A ban that names particular lands (City in a Bottle)
  // rides on the land's own `cant_cast` instead.
  cant_play_lands?: string;
  // #2559 (CR 101.2): the clause that stops this seat playing ANY card
  // from its hand right now, a spell or a land ("You can't play cards
  // from your hand — Memory Vessel"). While it is set `cant_play_lands`
  // carries the same clause; the badge row shows one NO HAND badge.
  cant_play_from_hand?: string;
  // ADR 0199 (#3953, CR 701.20a): why this seat plays with its hand
  // revealed right now ("Telepathy", "Stromgald Spy"). While it is set the
  // hand's cards reach every seat face up, cards drawn later included.
  // Absent when the hand is as hidden as any other.
  hand_revealed?: string[];
  // ADR 0057 (#749, CR 104.3): the "can't lose the game" / "can't win
  // the game" gates on this seat. `cant_lose` lists the causes that
  // can't make this player lose right now ("life", "empty_draw",
  // "poison", "commander_damage", "effect") — all five under a
  // Platinum Angel; concession is never in it. `cant_win` is true when
  // an effect can't make this player win. `end_gates` names the
  // sources for the badge's tooltip. Derived, public, and absent for
  // nearly every seat.
  cant_lose?: string[];
  cant_win?: boolean;
  end_gates?: GameEndGateView[];
  // ADR 0106 §4 decision 6 (#1806): this seat's live "can't be
  // countered" grants and unspent one-use promises, in the order they
  // were made — Veil of Summer's "Spells you control can't be countered
  // this turn", an unused Insist. Public, and absent for nearly every
  // seat.
  counter_shields?: CounterShieldView[];
}

// One emblem (CR 114). `label` is the board name ("Elspeth, Sun's
// Champion emblem"); `text` is its printed ability, for the hover.
export interface EmblemView {
  instance_id: string;
  label: string;
  text: string;
  // ADR 0114 §2: how many times the Ring has tempted the emblem's owner.
  // Absent on every other emblem. For the Ring, `text` is the lines it
  // has gained so far.
  level?: number;
  // ADR 0114 owner decision 1: every line of the Ring with the count it
  // is gained at, so the table can show the lines still to come.
  // Absent on every other emblem.
  lines?: EmblemLineView[];
}

export interface EmblemLineView {
  text: string;
  at: number;
}

export interface LifeChangeView {
  delta: number;
  new_total: number;
  at: string; // RFC3339
  // Per-player counter, starting at 1, stamped server-side on every
  // recorded change. Monotonic and stable across frames, and it keeps
  // climbing after `life_history` stops growing at its cap — the only
  // field on the entry that identifies it (#703). Neither the array
  // index nor `at` can: the array is trimmed from the front, and `at`
  // is RFC3339 SECONDS, so two changes in one second collide. Absent
  // (0) only on entries from a snapshot taken before #703.
  seq: number;
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
  // #1590: the bound THIS caster is held to right now — the printed
  // max, raised while a conditional mode count holds (Jeska's Will's
  // "if you control a commander as you cast this spell, you may choose
  // both instead"). The server stamps it per caster, so a picker that
  // offers up to `max` never offers a selection the gate refuses.
  max: number;
  options: ModeOptionView[];
  // #764, CR 700.2d: "you may choose the same mode more than once"
  // (Mystic Confluence). The picker offers a count per option rather
  // than a toggle, and each occurrence is asked for its own targets.
  repeatable?: boolean;
  // #1655: the bounds with the card's optional costs announced —
  // Inscription of Ruin's "if this spell was kicked, choose any number
  // instead", Depth Defiler-shaped "choose both instead". `min` / `max`
  // above are the bounds with none announced; the picker switches to
  // these when the caster ticked an optional cost. Absent when paying
  // changes nothing.
  if_optional_paid?: { min: number; max: number };
  // ADR 0097 (#1749): "choose one that hasn't been chosen [this
  // turn]" on an activated ability. Each option this object's ability
  // has already chosen carries `used: true`.
  not_chosen?: ModeNotChosen;
  // #2126, CR 702.120a: "Escalate [cost]" — what each mode beyond the
  // first costs. Present only on an escalate spell. `max` above is
  // already clamped to 1 + `max_extra`, so the picker never offers a
  // count whose cards or creatures the viewer lacks.
  escalate?: EscalateView;
}

// EscalateView is ModeSpecView.escalate. Discards ride cast_spell's
// `discard_ids` and taps ride `teamwork_ids`, (modes - 1) times the
// per-mode count each.
export interface EscalateView {
  label: string;
  mana_cost?: string;
  discard_cards?: number;
  tap_creatures?: number;
  // The viewer's untapped creatures that could pay the taps.
  tap_options?: string[];
  // Per viewer: the most extra modes the non-mana half can be paid for.
  max_extra: number;
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
  // see sacrificeCost.ts. ADR 0100 §3: "sacrifice any number of
  // creatures" is min 0 / max 0, an open count from zero, and
  // "sacrifice X lands" sets count_from_x — the number picked is the
  // x_value. Either may be paid with none (castSacrificeRange).
  sacrifice_options?: LegalTargetsView;
  // S23: a "pay X life" clause (Toxic Deluge). The X prompt has to
  // open for this card even though its printed mana cost has no {X},
  // and the announced X is both the life paid and the number the
  // spell's own text uses.
  demands_x?: boolean;
  label?: string;
  // ADR 0100 §2: an either/or additional cost — "sacrifice an artifact
  // or discard a card". Each branch is one of these views, with its
  // own `key` and whether the viewer could pay it right now
  // (`payable`; absent means it cannot be taken). The chosen branch's
  // index rides cast_spell as `cost_branch`, and its cards ride the
  // usual lists. A cost with branches carries no components of its own.
  branches?: AdditionalCostView[];
  // Branches only: the branch's identity, its mana ("{5}", joins the
  // total the preview prices once `cost_branch` is sent), its fixed
  // life, its blight N and the creatures that could take the counters.
  key?: string;
  mana_cost?: string;
  pay_life?: number;
  blight?: number;
  blight_options?: LegalTargetsView;
  // #2174: "blight X" — the announced X is the counter count, put on the
  // one creature in `blight_options`; `blight_x_max` (absent at 0) is the
  // greatest toughness among your creatures, the most X may be.
  blight_x?: boolean;
  blight_x_max?: number;
  // ADR 0100 amendment 2026-10-07: a branch that reveals a card from your
  // hand (`reveal`) or, to behold (`behold`), may choose a permanent you
  // control instead. `reveal_options` lists the cards that could pay it,
  // hand cards first; the one pick rides cast_spell as `reveal_ids`.
  reveal?: boolean;
  behold?: boolean;
  reveal_options?: LegalTargetsView;
  // #3402: an "exile a creature you control" additional cost (Soul
  // Exchange). `exile_permanent_options` lists the permanents that could
  // pay it, in payment order, min / max the clause's count; the picks ride
  // cast_spell as `exile_permanent_ids`. Present-and-empty means the cost
  // cannot be paid right now.
  exile_permanent_label?: string;
  exile_permanent_options?: LegalTargetsView;
  // #3868: a "tap an untapped Vampire you control" additional cost (New
  // Blood), on the cost or on one of its branches. `tap_others_options`
  // lists the untapped permanents that could pay it, in payment order,
  // min / max the clause's count; the picks ride cast_spell as
  // `teamwork_ids`. Fewer cards than `min` means the cost cannot be paid
  // right now. A creature that arrived this turn may pay: it is not {T}.
  // #3961: min 0 (max every option) is "tap any number of untapped
  // creatures you control" (Burn at the Stake); none is a legal payment.
  tap_others_label?: string;
  tap_others_options?: LegalTargetsView;
  // #3175: a "you may exile any number of blue cards from your hand"
  // additional cost (March of Swirling Mist). `exile_from_hand_options`
  // lists the caster's hand cards that could pay it, min 0: none is always
  // legal. The picks ride cast_spell as `exile_from_hand_ids`, and their
  // number is the card's per-card discount. The options reach the card's
  // owner alone.
  exile_from_hand_label?: string;
  exile_from_hand_options?: LegalTargetsView;
  // #3536: an "exile N cards from your graveyard" additional cost
  // (Abhorrent Oculus), the activated ability's four fields of the same
  // names. `exile_cost_options` lists the cards that could pay it, in pile
  // order; the picks ride cast_spell as `exile_ids`, exactly
  // `exile_cost_n` of them. Fewer options than `exile_cost_n` means the
  // cost cannot be paid right now (CR 118.3).
  exile_cost_n?: number;
  exile_cost_label?: string;
  exile_cost_options?: string[];
  exile_cost_zone?: ExileCostZone;
  // #1948: "exile X cards from your graveyard" (Skeletal Scrying,
  // Harvest Pyre). `demands_x` is set too, so the X prompt opens; the
  // picks on `exile_ids` must number exactly the announced X
  // (`exile_cost_n` is absent). `exile_cost_x_max` is the largest X the
  // graveyard pays, absent at 0.
  exile_cost_x?: boolean;
  exile_cost_x_max?: number;
  payable?: boolean;
}

// OptionalCostView is one "you may pay an additional cost as you
// cast this spell" offer — kicker, multikicker, buyback (CR 601.2b,
// ADR 0073). Like an alternative cost it is an OFFER; unlike one,
// the offers COMPOSE: a cast may claim one alternative cost and any
// number of these, which is why they render as toggles beside the
// alternative-cost radio list rather than as a modal of their own.
//
// `index` is what rides back on cast_spell in `optional_costs`,
// repeated once per payment for a repeatable cost. It is a POSITION
// and not a key, because a position is what the server's paid record
// holds; `key` is here for labelling only.
export interface OptionalCostView {
  index: number;
  key: string;
  label?: string;
  mana_cost?: string;
  // How many times this cost may be paid for one cast: 1 for kicker
  // and buyback, the multikicker cap above that. 1 renders a
  // checkbox, more renders a stepper.
  max_times?: number;
  // #3232: the floor on the announced X while this cost is paid —
  // Thieving Skydiver's "Kicker {X}. X can't be 0" is 1. Absent when the
  // cost has no floor; meaningful only beside a `mana_cost` containing
  // {X}. Read by optionalCostXFloor (targeting.ts).
  x_min?: number;
  // ADR 0129 §5: the energy each payment charges (Reiterating Bolt's
  // replicate, 3). `max_times` is then capped at what the viewer's energy
  // pays for, and `energy_short` says it pays for none.
  energy?: number;
  energy_short?: boolean;
  // The card-shaped halves, in the same shape and with the same
  // meaning AdditionalCostView gives them: a present-and-empty
  // sacrifice_options means the offer cannot be taken right now.
  discard_cards?: number;
  sacrifice_options?: LegalTargetsView;
  // CR 702.174a (#1267): a gift offer. Taking it means naming one
  // opponent to promise the gift to, and that choice rides cast_spell
  // as `gift_opponent`. `opponent_options` is who may be named — the
  // opponents still in the game. On a gift offer an absent or empty
  // list means nobody is left to promise it to, so the offer cannot
  // be taken right now (the server's `omitempty` drops an empty
  // list).
  chooses_opponent?: boolean;
  opponent_options?: string[];
  // #1703: teamwork (CR 702.194a) — tap any number of your untapped
  // creatures with total power `teamwork` or more. `teamwork_options`
  // is who may be tapped; present-and-empty means the whole set falls
  // short and the offer cannot be taken. The picks ride cast_spell as
  // `teamwork_ids`.
  teamwork?: number;
  teamwork_options?: LegalTargetsView;
  // #1703: blight N (CR 701.68a) — put `blight` -1/-1 counters on one
  // creature you control. `blight_options` is your creatures;
  // present-and-empty means you control none. The pick rides
  // cast_spell as `blight_ids`.
  blight?: number;
  blight_options?: LegalTargetsView;
  // #1267: the target clause the spell has WHEN THIS COST IS PAID —
  // Long River's Pull counters any spell once the gift is promised.
  // Same shape and meaning as AlternativeCostView's trio; absent when
  // the cost leaves the card's own clause alone (every kicker).
  target_mode?: string;
  legal_targets?: LegalTargetsView;
  clauses?: LegalTargetsView[];
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
  // ADR 0152 §5: the printed face this offer casts, when it is not the
  // face the card shows — disturb's and More Than Meets the Eye's back
  // face. The picker names it ("casts Goldbug, Scrappy Scout").
  casts_face?: number;
  mana_cost?: string;
  // The target clause the spell has WHEN THIS COST IS PAID, already
  // resolved server-side against the cost's rewrite. Absent means the
  // spell has no targets under this cost (overload), so the cast
  // fires straight away.
  target_mode?: string;
  legal_targets?: LegalTargetsView;
  // #764: the offer's clause list when its rewritten statement has
  // more than one clause.
  clauses?: LegalTargetsView[];
  // S28: the "pay N life" half of the cost (Force of Will's 1, Snuff
  // Out's 4). Absent for the costs that charge none. This is for the
  // label — #695: the OFFER ITSELF is absent when the caster's life
  // total is below it (CR 119.4), so a rendered offer is always one
  // the server will accept. Exactly N still appears; paying down to
  // zero is legal.
  life?: number;
  // ADR 0129 §5: the "pay N {E}" half of the cost (Nissa, Worldsoul
  // Speaker's eight, Primal Prayers' one, Amped Raptor's "equal to its
  // mana value"). Absent for the costs that charge none. Like `life`, for
  // the label: an offer the caster is short of energy for is not offered.
  energy?: number;
  // S28: the cards that can pay the cost's card-shaped half — the
  // blue cards in your hand for Force of Will, the Islands you
  // control for Daze. The chosen instance ID rides back on cast_spell
  // as `alt_cost_ids`. Absent when the cost charges no cards (every
  // S22 keyword). #695: never present-and-empty any more — an offer
  // with nothing to pay it is not offered at all, for the same reason
  // one whose life half is unpayable is not.
  pay_options?: LegalTargetsView;
  // #1727: the cost's card-shaped half when it is a SACRIFICE — Dread
  // Return's "Flashback—Sacrifice three creatures", Fireblast's two
  // Mountains. Set instead of `pay_options`, in the same shape (and
  // payment order) as the additional cost's `sacrifice_options`, so
  // the cast flow opens SacrificeCostModal for it. The picked IDs
  // still ride `alt_cost_ids`, not `sacrifice_ids`.
  sacrifice_options?: LegalTargetsView;
  // ADR 0135 §2: the `pay_options` cards are DISCARDED, not exiled or
  // returned (retrace's land, Snag's Forest card, Foil's two cards). The
  // picker words its prompt "Discard". A discard with a set rule (Foil's
  // "an Island card and another card") carries `pay_options.each_of`.
  discards?: boolean;
  // ADR 0135 §1: the cost's card-shaped half when it TAPS permanents —
  // Orim's Cure's "tap an untapped creature you control", Battle
  // Screech's three white creatures. The untapped permanents you
  // control that match, min == max == the count, in the shape an
  // ability's `tap_others_options` has, so the cast flow opens the tap
  // picker. Set instead of `pay_options` and `sacrifice_options`; the
  // picked IDs ride `alt_cost_ids`. #3931: harmonize's "up to one
  // untapped creature" is min 0, max 1 — tapping none is a payment.
  tap_options?: LegalTargetsView;
  // S28: the picker's prompt copy for `pay_options` ("a blue card"),
  // for `sacrifice_options` ("three creatures") or for `tap_options`
  // ("an untapped creature you control").
  pay_label?: string;
  // ADR 0135 §4: an emerge offer (CR 702.119a) — the permanent picked
  // from `sacrifice_options` reduces the cost by its mana value.
  reduces_by_mana_value?: boolean;
  // ADR 0135 §4: an emerge offer's price per candidate, keyed by the
  // instance IDs in `sacrifice_options`: the candidate's mana value and
  // the mana the cast pays with it sacrificed, from the server's pricer.
  sacrifice_prices?: Record<string, AltCostPriceView>;
  // #3931: a harmonize offer (CR 702.180a) — the creature picked from
  // `tap_options` reduces the cost by its power.
  reduces_by_power?: boolean;
  // #3931: a harmonize offer's price per candidate, keyed by the
  // instance IDs in `tap_options`: the creature's power and the mana the
  // cast pays with it tapped, from the server's pricer.
  tap_prices?: Record<string, AltCostTapPriceView>;
  // CR 107.3b (#831): the card prints an {X} in its mana cost and
  // this offer does not, so claiming it fixes X at 0 — the cast flow
  // skips the X picker and sends nothing. Absent for nearly every
  // offer, including one priced with an {X} of its own.
  x_locked_at_zero?: boolean;
  // CR 107.4 (#916): how many symbols in THIS offer's cost carry the
  // "or 2 life" option. Claiming the offer replaces the mana cost, so
  // it replaces the ceiling on `phyrexian_life` too — a picker that
  // read the printed count while paying an alternative cost would
  // offer a payment the announce gate rejects. Absent for every offer
  // that prints none, which is all of them today.
  phyrexian_symbols?: number;
  // ADR 0131 (#2531): how many of `phyrexian_symbols` are payable with
  // life only because the viewer controls a permanent that grants it —
  // K'rrik's "for each {B} in a cost, you may pay 2 life rather than pay
  // that mana" — rather than because the cost prints a Phyrexian symbol.
  // Absent for every viewer without such a permanent. When every symbol
  // is granted the cast asks about life only if mana falls short, and
  // the card's menu keeps a "Pay life for {B}…" row.
  phyrexian_granted?: number;
  // #1686: the engine will refuse THIS offer right now for timing (CR
  // 307.1) even though its zone and payability both check out. Every
  // S22 keyword here answers to the card's own printed timing (or a
  // wider per-player grant) the same way the printed cost does, so
  // this stays absent for them; a live miracle grant is the first
  // offer with a clock of its own that can disagree with the printed
  // cost's. Read by the cost picker via castableAlternativeCostsOf
  // (targeting.ts), which drops an offer this is set on.
  timing_closed?: boolean;
  // #3225: the largest X the viewer may announce while claiming THIS
  // offer — "you may cast this spell as though it had flash if X is 3
  // or less" (CR 601.3c) — already folded with the card's printed
  // `x_max`. The server refuses a larger X under the claim; the cast for
  // the printed cost is not bound by it. Present only on an offer that
  // carries a ceiling of its own. Read by xCeilingFor (targeting.ts).
  x_max?: number;
  // ADR 0126 §6: what the spell does when cast for this cost, where
  // that differs from the card's own purpose (overload). Bot data.
  purpose?: PurposeView;
}

// TapCostView is the "tap permanents you control to help pay for
// this" clause convoke and waterbend share, on a card in the
// viewer's own hand (S22). Like an alternative cost it is an OFFER —
// tapping nothing and paying the whole cost with mana is always
// legal — but unlike one it replaces nothing: it spends against a
// cost that is still owed. The picked instance IDs ride cast_spell
// as `tap_ids`.
// DelveView is a card's delve (CR 702.66, ADR 0100 §4): the graveyard
// cards the caster may exile, each paying for {1} of the generic mana.
export interface DelveView {
  // The caster's own graveyard, in the server's payment order: lands
  // first, then cards with no graveyard cast surface, then the rest,
  // oldest first. "Choose for me" fills the budget from the front.
  options?: LegalTargetsView;
  // The budget for the default announcement (this zone, X = 0, nothing
  // tapped). A hint only — the preview's `delve_budget` is the cap.
  max?: number;
}

// ADR 0108 §5: a pay-unless prompt's non-mana payment. `options` are
// instance IDs (the chooser's hand for a discard, the permanents of
// the clause's kind they control for a sacrifice); one payment names
// exactly `count` of them.
// ADR 0129 §3: the bounds of a pay_amount prompt. goal is the card's
// own threshold (the stepper starts there); unit is what one point buys.
// ADR 0129's amendment of 2026-10-09 (#1941): `resource` is what each
// point costs — energy (also when absent), life, or none for a number
// that is chosen and not paid ("an amount of damage of your choice").
// A payment may always be declined with 0; a number that is not paid is
// min..max alone. With `no_max` the number has no ceiling and `max` is
// only the server's overflow guard. `marks` are other numbers the card
// names (the bot's offers); `self_damage` says the chooser is dealt the
// same amount.
export interface PayAmountView {
  min: number;
  max: number;
  goal?: number;
  // ADR 0150 §4: "assist" is {1} of another player's spell.
  unit: "damage" | "counters" | "cards" | "power" | "tax" | "other" | "assist";
  // ADR 0150 §4: "mana" is generic mana an assist helper pays from their
  // pool and then their untapped sources.
  resource?: "energy" | "life" | "none" | "mana";
  no_max?: boolean;
  marks?: number[];
  self_damage?: boolean;
}

export interface PayCardsView {
  // "exile_graveyard" (ADR 0187 §3): options are the chooser's own
  // graveyard, public to every seat.
  action: "discard" | "sacrifice" | "exile_graveyard";
  count: number;
  options: string[];
}

// One branch of an either payment (ADR 0187 §5).
export interface PayOptionView {
  // "sacrifice", "discard", "exile_graveyard", "life", "damage",
  // "energy" or "nothing".
  action: string;
  // How many cards one payment names, or the amount of life, damage or
  // energy. Absent for a branch with no amount.
  count?: number;
  // The branch as a verb phrase: "sacrifice two lands", "take 5 damage".
  words: string;
  // Whether the chooser can make this branch right now (CR 118.3). A
  // branch they can't is shown greyed.
  payable: boolean;
  // The instance IDs that could pay a branch that names cards, as
  // pay_cards.options are (a discard's reach the chooser only).
  options?: string[];
}

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

// SpecialActionView is one CR 116.2 special action offered on a card
// in the viewer's own hand — foretell, suspend, plot — and, since
// #1391, on the top card of their own library when a permanent grants
// one there (Fblthp, Lost on the Range's plot). A row and nothing
// more: no targets, no modes, no cost picker, so the client sends
// `special_action { card_id, kind, strict, auto_tap }` straight from
// it, plus `cost` when the card offers the same kind twice (a plot
// card under Fblthp: its own plot cost or its mana cost). `available` is the server's own per-kind timing answer, so the
// client greys the row rather than re-deriving a rule it would get
// backwards (foretell is legal under split second; suspend is not).
export interface SpecialActionView {
  kind: string;
  label: string;
  cost?: string;
  // #1319: what the engine actually charges for `cost` right now,
  // after every CR 601.2f cost modifier on the battlefield — Ranar
  // the Ever-Watchful's "The first card you foretell each turn costs
  // {0} to foretell". Same contract as ActivatedAbilityView's
  // charged_mana_cost: absent means "not priced" (fall back to
  // `cost`), an empty string is a real "this now costs nothing".
  charged_cost?: string;
  available?: boolean;
  // ADR 0103: the door an "unlock" row unlocks — "left" or "right" —
  // and what the special_action payload names back. Absent on every
  // other kind.
  door?: "left" | "right";
  // ADR 0189: which rule a "turn_face_up" row is taken under —
  // "mana_cost", "morph" or "disguise" — and what the special_action
  // payload names back. A manifested or cloaked card whose card also
  // prints morph or disguise lists two rows, possibly at the same
  // cost, and this is what tells them apart. Absent on every other
  // kind.
  route?: "mana_cost" | "morph" | "disguise";
}

// RoomDoorsView is a Room's two CR 709.5c designations (ADR 0103):
// `true` is an unlocked door. Present only on a face-up Room on the
// battlefield, so `false` is a locked door.
export interface RoomDoorsView {
  left: boolean;
  right: boolean;
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
  // #1573: "mana of any TYPE can be spent" (Hostage Taker) — colorless
  // counts too, so a {C} in the cost is payable with anything. Set
  // with `any_color` alongside it. A label only: `cast_prices` and
  // `castable_here` already reflect it.
  any_type?: boolean;
  // S22 airbend: the mana cost the holder pays INSTEAD of the card's
  // printed one ("{2} rather than its mana cost"). Absent for
  // impulse exile, which charges the printed cost. Note that
  // `mana_cost` on the card still carries the printed value.
  cost_override?: string;
  // S29 warp: the earliest turn sequence the grant is live on — "you
  // may cast it from exile ON A LATER TURN". Absent for every grant
  // that is live as soon as it is made, which is all of impulse
  // exile and airbend.
  not_before_seq?: number;
  // S32: the printed faces this grant opens, when it opens any.
  // Absent for every grant that does not speak about faces (impulse
  // exile, airbend, warp, cascade), which is all of them before S32.
  //
  // Two grants name one face each, in opposite directions: a defeated
  // Siege's "exile it, then cast it transformed" names the BACK face,
  // where the card sitting in the exile pile still shows the battle
  // and the thing the button casts is `faces[1]`; CR 715.4's
  // Adventure grant names the CREATURE face, face 0. That second one
  // is why this is a list rather than the number it was until #719 —
  // "absent" and "face 0" are different facts.
  //
  // Advisory only: the server settles the face from the grant rather
  // than from the request, so a client that ignores this labels the
  // button with the wrong name but cannot cast the wrong half.
  faces?: number[];
  // CR 107.3b (#831): the card prints an {X} in its mana cost and
  // this grant's price does not, so casting under it fixes X at 0 —
  // what a cascade hit carries. The cast flow skips the X picker.
  // Absent for a grant that charges the printed cost, which still
  // asks.
  x_locked_at_zero?: boolean;
  // #2559: when the grant ends, and the player the window is counted
  // against for the three windows that name one ("until Bo's next
  // turn"). Read by exileGrants.ts for the tooltip and a bystander's
  // label. Absent: say "until end of turn", as before the field.
  until?: ExilePlayUntil;
  until_player?: string;
  // #3192: the clause a CONDITIONAL grant is live under ("during a turn
  // you attacked with a token"), and whether it is false right now. The
  // card keeps its grant while the condition is false, so the client
  // withholds the play button instead of hiding the grant. Absent for an
  // unconditional grant.
  condition?: string;
  condition_unmet?: boolean;
  // #3490: the holder's next pass of priority ends this grant (rebound,
  // suspend, discover and cascade's free casts; ADR 0099 section 4).
  // Passing is the decline, so autopass holds while the viewer may cast
  // a card that carries it (priority.ts owesFreeCast). Absent otherwise.
  closes_on_pass?: boolean;
}

// #2559: the windows `exile_play.until` names.
export type ExilePlayUntil =
  | "end_of_turn"
  | "end_of_next_turn"
  | "next_turn"
  | "next_end_step"
  | "while_exiled"
  | "until_another";

// #1297: the pile an "Exile N cards from your …" cost reads — always
// the activator's own. The server stamps it with `exile_cost_n`.
export type ExileCostZone = "hand" | "graveyard";

// ActivatedAbilityView is one CR 602 activated ability on a
// battlefield permanent (S21 sub-PR 2). Public information, so it
// rides every viewer's snapshot; the client only offers the menu on
// permanents the viewer controls, plus a permanent's `any_player` rows
// to every other seat (ADR 0106 §1). `index` is what the
// activate_ability payload carries as `ability_index`.
export interface ActivatedAbilityView {
  index: number;
  label?: string;
  // ADR 0093 Decision 5: the row's stable ref ("own:<i>",
  // "grant:<bundle>:<i>:<n>"). Sent back as activate_ability's `ref`
  // so a grant that appeared or vanished since this snapshot is
  // refused rather than fired on whatever moved to `index`. Optional
  // only because a server older than #1551 sends none.
  ref?: string;
  // ADR 0093 Decision 8: present on a row ANOTHER permanent granted
  // this one (Necrotic Sliver's "{3}, Sacrifice this permanent: …" on
  // every Sliver). Absent on the permanent's own abilities.
  granted_by?: GrantedByView;
  // ADR 0155 Decision 10: present on a row ANOTHER object's abilities were
  // lent as ("has all activated abilities of …", Necrotic Ooze): the lender
  // by id and printed name. `granted_by` is then the borrowing permanent
  // itself. Absent on every row that is not lent, and when the lender sits
  // in a hidden zone.
  borrowed_from?: GrantedByView;
  tap_cost?: boolean;
  sacrifice_self?: boolean;
  mana_cost?: string;
  // #1190: what the engine will actually charge for mana_cost right
  // now, after every CR 601.2f cost modifier on the battlefield —
  // Boom Scholar's "Exhaust abilities of other permanents you control
  // cost {2} less to activate" turns a printed `{3}{R}` into a
  // charged `{1}{R}`. Present whenever mana_cost is, and EQUAL to it
  // when no modifier reaches this ability, which is nearly every
  // ability in the game. Prefer this field for the row's own display;
  // show mana_cost as a tooltip only when the two differ (see
  // chargedCostNote in contextMenu.logic.ts).
  charged_mana_cost?: string;
  // #1296: what the mana component costs if the ability targets each
  // legal target, keyed by the target's ID (card instance or player),
  // valued as charged_mana_cost is ("" = free). Present only when the
  // PRICE reads the target — Dragonfire Blade's "{1} less for each
  // color of the creature it targets" — on a one-target, non-modal
  // ability; charged_mana_cost is then the price before a target is
  // chosen. See targetPrices.ts.
  target_charged_mana_costs?: Record<string, string>;
  life_cost?: number;
  // ADR 0129 §8: "Pay N {E}" — the energy counters an activation
  // removes from the activator (CR 107.14). energy_cost_x marks "Pay X
  // {E}": the announced X is added, and demands_x is set. When the
  // controller is short, cant_activate carries the server's refusal
  // ("Not enough energy (have 2, need 3)").
  energy_cost?: number;
  energy_cost_x?: boolean;
  sorcery_speed?: boolean;
  // #1208: true when the engine will refuse this activation RIGHT NOW
  // for timing (CR 602.5d, CR 606.3), as modified by any per-player
  // statement on the board — The Wandering Emperor's "you may
  // activate her loyalty abilities any time you could cast an
  // instant", Leonin Shikari's "you may activate equip abilities any
  // time you could cast an instant".
  //
  // It is the ROW's verdict where sorcery_speed is the ability's
  // printed clause, and it is the one to grey on: the client no
  // longer derives the window for a catalogued ability. Absent means
  // the engine has no timing objection, which is every instant-speed
  // ability, always.
  timing_closed?: boolean;
  // #743: true while the ability's activation condition (CR 602.1b —
  // "Activate only if an opponent controls four or more lands",
  // "Activate only during your turn") is false. Absent when there is
  // no condition or it holds. Evaluated server-side for the
  // permanent's controller; the menu greys the row like
  // sorcery_speed, and the server refuses the activation regardless.
  condition_unmet?: boolean;
  // #1181: true when this is an exhaust ability ("Activate each
  // exhaust ability only once") that this permanent has already
  // activated. Absent otherwise. Unlike condition_unmet it does not
  // come back: only a new object (CR 400.7 — a flicker, not an untap)
  // clears it, which is why the menu says something different.
  exhausted?: boolean;
  // #2697, CR 702.142a: why a boast ability ("Activate only if this
  // creature attacked this turn and only once each turn") cannot be
  // activated right now. "not_attacked" recovers when the creature
  // attacks; "used" at the next turn (Birgi, God of Storytelling raises
  // the number of uses the server counts). Absent when the ability is not
  // a boast ability or nothing objects.
  boast_blocked?: "not_attacked" | "used";
  // #1210: the printed clause of a board-wide "can't be activated"
  // static that refuses THIS ability right now ("Activated abilities
  // of creatures can't be activated", Cursed Totem). Absent, which is
  // nearly always, means nothing refuses it. Distinct from
  // CardView.restrictions' `cant_activate` token, the per-permanent
  // Arrest bit. ADR 0117 §2 greys the row on it and shows the clause.
  cant_activate?: string;
  // loyalty_cost is the +N / 0 / −N of a planeswalker's loyalty
  // ability (CR 606.4). Its PRESENCE, not its value, is what marks
  // the ability as a loyalty ability — 0 is a real printed cost —
  // so test for `!== undefined`, never for truthiness. Added with
  // #329 / #334.
  loyalty_cost?: number;
  // #1944: a loyalty cost of −X. loyalty_cost is 0, demands_x is set,
  // and the announced X is removed, at most the permanent's loyalty
  // (CR 606.6).
  loyalty_cost_x?: boolean;
  // A "Sacrifice a creature"-style cost: the clause, and the
  // permanents the controller can pay it with right now. #747: min /
  // max are the clause's count ("Sacrifice two artifacts" is 2 / 2;
  // always equal) and the cards come in payment order — tokens first,
  // then lower mana value, then the source. See sacrificeCost.ts.
  sacrifice_label?: string;
  sacrifice_options?: LegalTargetsView;
  // #1213: a "Return a permanent you control to its owner's hand"
  // cost (Quirion Ranger's Forest, Master Transmuter's artifact,
  // Meloku's land). `return_label` is the clause as printed and
  // `return_options` the permanents that could pay it right now, in
  // the same payment order sacrifice_options uses — so one picker
  // serves both. min / max are the clause's count (1 on every printed
  // card). A TAPPED permanent is a legal pick, and so is the ability's
  // own source when the clause admits it. The picks ride
  // activate_ability as `return_ids`; an absent or empty list means
  // the cost cannot be paid (CR 118.3) and the server refuses.
  return_label?: string;
  return_options?: LegalTargetsView;
  // #1600: an "Exile a creature you control" cost (The Soul Stone's
  // harness, Altar of Bhaal, City of Shadows). `return_options` one
  // destination over: the clause as printed and the permanents that
  // could pay right now, in payment order, min / max the clause's count.
  // The picks ride activate_ability as `exile_permanent_ids` — NOT
  // `exile_ids`, which names cards in a hand or a graveyard. An absent
  // or short list means the cost cannot be paid (CR 118.3).
  // ADR 0137: on a craft ability the options also include cards in
  // the activator's own graveyard (CR 702.167b), after the permanents.
  exile_permanent_label?: string;
  exile_permanent_options?: LegalTargetsView;
  // #1310: the CR 701.67 clause of a "Waterbend {N}:" cost (Aang,
  // Swift Savior; Katara, Water Tribe's Hope), in the same TapCostView
  // shape a hand card's convoke / waterbend ships as `tap_cost`, so
  // TapCostModal serves both. `options` are the untapped artifacts and
  // creatures that could pay (the source among them unless the cost
  // also prints {T}); `max` is how many — 0 with `demands_x` means
  // "as many as the X you announce". Optional in both directions:
  // tapping none pays the whole cost with mana. The picks ride
  // activate_ability as `waterbend_ids`.
  waterbend?: TapCostView;
  // #759: a "Tap another untapped creature you control" cost — the
  // station ability's (CR 702.184a), #758's tap-another component.
  // `tap_others_label` is the clause without the verb and
  // `tap_others_options` the untapped permanents that could pay it
  // right now, min / max the clause's count. Not the {T} symbol, so a
  // creature that arrived this turn is on the list; not a target, so
  // a hexproof one is too. The picks ride activate_ability as
  // `tap_ids`; fewer options than `min` means the server refuses
  // (CR 118.3). #1421: `count_from_x` means the number picked is the
  // activation's `x_value`; min / max are then unset (0 / 0).
  tap_others_label?: string;
  tap_others_options?: LegalTargetsView;
  // #660: the discard cost components (CR 702.29a and the general
  // "Discard a creature card" clause). `discard_self` is cycling's
  // "Discard this card" — advisory only, there is nothing to pick,
  // because the source IS the payment. `discard_cost_n` is the count
  // of the general clause and its presence marks that component;
  // `discard_cost_label` is the clause as printed ("a creature
  // card"), and `discard_cost_options` the cards in hand that could
  // pay it right now. The picks ride activate_ability as
  // `discard_ids`; exactly `discard_cost_n` options means there is
  // nothing to ask and the client skips its picker.
  discard_self?: boolean;
  // #1221: scavenge's and embalm's "Exile this card from your
  // graveyard" (CR 702.96a / CR 702.128a) — `discard_self` one zone
  // over, advisory for the same reason and sending nothing for the
  // same reason: the source IS the payment.
  exile_self?: boolean;
  // #2028: "Return this enchantment to its owner's hand" (Gossamer
  // Chains, Shigeki). Advisory like `exile_self`: the source is the
  // payment, so nothing is picked and nothing is sent.
  return_self?: boolean;
  // #2726: "Put this creature on the bottom of its owner's library"
  // (Timestream Navigator). Advisory like `return_self`.
  bottom_self?: boolean;
  // ADR 0130 §4: "Exert this creature" — the source won't untap during
  // the activator's next untap step. Always payable, so it never greys
  // the row, and nothing is sent.
  exert?: boolean;
  // #3845: "Mill N cards" as a cost (The Warring Triad). Nothing is
  // sent. A library too short to pay it makes the server stamp
  // cant_activate, which greys the row.
  mill_cost?: number;
  discard_cost_n?: number;
  discard_cost_label?: string;
  discard_cost_options?: string[];
  // ADR 0109 §7, owner decision 3: the discard is "at random"
  // (Pyromancy, Meteor Storm). The engine draws the cards; there are no
  // `discard_cost_options` and nothing is sent, so the client confirms
  // the cost instead of opening the picker.
  discard_cost_random?: boolean;
  // #2527: "Discard X cards" (Gix, Yawgmoth Praetor). The count is the
  // activation's X, so `discard_cost_n` is absent: the client always
  // opens the picker (zero up to every card in `discard_cost_options`),
  // sends the picks as `discard_ids` and their number as `x_value`, and
  // skips the X stepper. `demands_x` is set beside it.
  discard_cost_count_from_x?: boolean;
  // #2190: "Discard a card with mana value X" (Kozilek, the Great
  // Distortion). `discard_cost_n` is 1 and `demands_x` is set, but X is
  // not asked for: the card picked IS the announcement, so the client
  // sends its mana value as `x_value` (the engine refuses any other) and
  // narrows the ability's target clause by it (`mana_value_equals_x`).
  discard_cost_mana_value_x?: boolean;
  // #2598: "Reveal X black cards from your hand" (Martyr of Bones). The
  // clause as printed without the verb, the printed count (absent for the
  // X form) and the matching hand cards — the controller's alone.
  // `reveal_cost_count_from_x` marks the X form: the number of cards
  // picked IS the announced X, so the client opens its picker (zero up to
  // every card in `reveal_cost_options`), sends the picks as `reveal_ids`
  // and their number as `x_value`, and skips the X stepper. `demands_x`
  // is set beside it.
  reveal_cost_n?: number;
  reveal_cost_label?: string;
  reveal_cost_options?: string[];
  reveal_cost_count_from_x?: boolean;
  // #3250: "Reveal your hand" as a cost (Sasaya). Names no card and no
  // count: the whole hand is shown when the ability is activated.
  reveal_hand_cost?: boolean;
  // ADR 0109 §7 (#1902): "Put a card from your hand on top of your
  // library" (Penance, Leashling). The count, the clause as printed and
  // the cards in the viewer's hand that could pay; the picks ride
  // activate_ability as `top_ids`.
  top_cost_n?: number;
  top_cost_label?: string;
  top_cost_options?: string[];
  // ADR 0109 §7: "Exile the top N cards of your library" (Seasoned
  // Tactician, Arc-Slogger). Nothing to pick or send; confirmed.
  library_exile_cost_n?: number;
  // #1297: an "Exile N cards from your graveyard" / "… from your hand"
  // component — Grim Lavamancer's "Exile two cards from your graveyard",
  // Holistic Wisdom's "Exile a card from your hand". The mana ability's
  // four exile fields (#1283) under the same names: the count, the
  // clause as printed, the cards that could pay right now, and the pile
  // they are in. NOT a discard — the picks ride activate_ability as
  // `exile_ids`, never `discard_ids`.
  exile_cost_n?: number;
  exile_cost_label?: string;
  exile_cost_options?: string[];
  exile_cost_zone?: ExileCostZone;
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
  // ADR 0186 Decision 6 (#3549): what each creature in crew_options counts
  // for on THIS cost, keyed by card ID. Differs from `power` when a rule
  // says so ("crews Vehicles as though its power were 2 greater"). Absent
  // on a teamwork prompt, which counts plain power; when absent a client
  // sums `power`.
  crew_power?: Record<string, number>;
  // #2695: a Mount's saddle ability (CR 702.171a) rides the two crew
  // fields above — the same many-pick, floor-on-total-power prompt and
  // the same `crew_ids` payload — and sets this flag so the prompt says
  // "Saddle" rather than "Crew". crew_options then leaves out the Mount
  // itself: the cost taps OTHER creatures. Absent — not `false` — for
  // every other ability.
  saddle?: boolean;
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
  //   - counter_cost_among (#789): "from AMONG artifacts you control".
  //     The N counters may be split across any number of the listed
  //     permanents, so the picker is many-pick with a running total
  //     and the payload carries a count per permanent.
  //   - counter_cost_variable (#789): "Remove X counters" / "any
  //     number". counter_cost_n is then the FLOOR rather than the
  //     amount, and counter_cost_max is the most the viewer could
  //     name right now — the stepper's ceiling.
  //   - counter_cost_add / counter_cost_add_kind (#789): a cost that
  //     PUTS counters on the source (Devoted Druid's -1/-1). Nothing
  //     is chosen; counter_add_blocked is CR 118.3 saying the
  //     permanent can't have them, which greys the row.
  //
  // The choice rides activate_ability (or activate_mana_ability) as
  // `counter_source_ids` (not for the self form), `counter_counts`
  // (only for the among and variable forms) and `counter_kind` (only
  // for the any-kind form) — or, when an any-kind AMONG payment mixes
  // kinds, `counter_kinds`, one per permanent (#943, Tekuthal).
  counter_cost_n?: number;
  counter_cost_kind?: string;
  counter_cost_self?: boolean;
  counter_cost_label?: string;
  counter_cost_among?: boolean;
  counter_cost_variable?: boolean;
  counter_cost_max?: number;
  counter_cost_options?: CounterCostOptionView[];
  counter_cost_add?: number;
  counter_cost_add_kind?: string;
  counter_add_blocked?: boolean;
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
  // CR 107.4f (#917, #916): how many symbols in the ability's mana
  // component carry the "or 2 life" option — 1 for Birthing Pod's
  // "{1}{G/P}", 2 for Solphim's "{1}{R/P}{R/P}". It is the ceiling on
  // the `phyrexian_life` the activation may claim, and the reason the
  // menu knows to open the stepper at all. A COUNT rather than
  // something the client derives, for the reason demands_x is one.
  phyrexian_symbols?: number;
  // ADR 0131 (#2531): how many of `phyrexian_symbols` are payable with
  // life only because the viewer controls a permanent that grants it —
  // K'rrik's "for each {B} in a cost, you may pay 2 life rather than pay
  // that mana" — rather than because the cost prints a Phyrexian symbol.
  // Absent for every viewer without such a permanent. When every symbol
  // is granted the cast asks about life only if mana falls short, and
  // the card's menu keeps a "Pay life for {B}…" row.
  phyrexian_granted?: number;
  // Present when the ability targets. A full LegalTargetsView since
  // #334: the server now stamps the clause's min / max (it always
  // had them; abilityLegalTargets just never copied them across),
  // which is what lets an "up to one target" ability — Teferi's −3,
  // the Emperor's −2 — be confirmed with nothing picked.
  target_mode?: string;
  legal_targets?: LegalTargetsView;
  // #764: every clause of the ability's statement when it has more
  // than one, and its CR 700.2 mode clause when it is modal
  // (Aetheric Amplifier). A modal ability announces its modes with
  // its targets in the one activate_ability (CR 602.2b), so the menu
  // shows the same mode picker a modal spell's hand card gets.
  clauses?: LegalTargetsView[];
  modes?: ModeSpecView;
  // ADR 0106 §1 decision 5 (#1793): the ability's own "Any player may
  // activate this ability" (CR 602.2). Every seat may open this row on
  // the permanent, not only its controller, and each seat's copy of
  // the row is stamped with THAT seat as the activator: the charged
  // price, life_cost, condition_unmet, timing_closed and the cost
  // options are the viewer's (CR 602.1a, CR 109.5). Whether the row is
  // live for the viewer right now is still the digest's answer
  // (`legal_actions`), never the row's. Absent on every other row.
  any_player?: boolean;
  // ADR 0106 §1 amendment 2026-10-07 (#1947): "Only your opponents may
  // activate this ability" (Clergy of the Holy Nimbus) and "Only this
  // creature's owner may activate this ability" (Personal Incarnation).
  // Read with the card's `controller` and `owner` to say who the row is
  // open to; the controller's own opponents-only row is greyed.
  opponents_only?: boolean;
  owner_only?: boolean;
  // ADR 0106 §1 amendment 2026-10-09 (#1947): a granted row's "Only you
  // may activate this ability" (Martyrdom). `activator` is the one player
  // who may: the "you" of the effect that granted it, whoever controls the
  // permanent now. Absent on every other row.
  grantor_only?: boolean;
  activator?: string;
  // #2449: a CR 702.6 equip ability. Bot data; the client does not
  // read it.
  equip?: boolean;
  // #2500: a row whose whole effect is untapping its own source
  // (Basalt Monolith). Bot data; the client does not read it.
  untap_self?: boolean;
  // ADR 0126 §6 (and ADR 0106 §1 decision 8 on an any-player row):
  // what the row does. Bot data; the client does not read it.
  purpose?: PurposeView;
}

// PurposeView is what a card, a mode, an alternative cost or an
// ability does, as printed amounts the catalog declares (ADR 0126 §6;
// docs/protocol.md "Declared purpose"). Bot data; the client does not
// read it. Every field is absent when zero.
export interface PurposeView {
  // #3309: on an Aura card, what it does to the permanent it enchants.
  aura?: "benefit" | "harm" | string;
  // #3712: the card may be cast as though it had flash and is sacrificed
  // at the next cleanup step when it was. Absent otherwise.
  flash_cleanup_sacrifice?: boolean;
  // ADR 0126 (amendment of 2026-10-15, #3762): what the card does only
  // when its kicker is paid, and what turning it face up does. Bot data.
  kicked?: PurposeView;
  turned_face_up?: PurposeView;
  draws?: number;
  controller_loses_life?: number;
  discards?: number;
  lands?: number;
  // ADR 0136 §2: how many of `lands` enter untapped (Harrow 2, Nature's
  // Lore 1). Never more than `lands`; absent when they enter tapped.
  lands_untapped?: number;
  tutors?: number;
  self_mill_tutor?: number;
  tokens?: number;
  // #3165: how many of `tokens` are creature tokens, and the lowercase
  // creature subtype whose controlled count is how many it makes.
  creature_tokens?: number;
  tokens_per_controlled?: string;
  // ADR 0129 §7: energy counters it gives its controller.
  energy?: number;
  sweep?: SweepView;
  death_payoff?: boolean;
  // ADR 0130 (amendment of 2026-10-07): what an exert row does.
  pump?: { power?: number; toughness?: number; keywords?: string[] };
  extra_combat?: number;
  // ADR 0188 Decision 10 (#3566): extra turns its controller takes
  // after this one (CR 500.7) and next turns it makes them skip
  // (CR 614.10).
  extra_turns?: number;
  skips_turns?: number;
  prevent_combat_damage_to_self?: boolean;
  damage_to_creature?: number;
  damage_each_opponent?: number;
  life_gain?: number;
  // ADR 0135 §3: the N of an awaken offer.
  awaken_land?: number;
  // #2480: cards it exiles that its controller may play this turn.
  impulse?: number;
  // #3072: on a landfall row, what it pays for each land that enters
  // under its controller's control.
  landfall_payoff?: LandfallPayoffView;
  // #3146: on the card of a landfall doubler (Ancient Greenwarden), how
  // many additional times each landfall trigger of its controller's fires.
  landfall_extra_triggers?: number;
  // #3165: on a row that pays for each creature that enters under its
  // controller's control.
  creature_enter_payoff?: LandfallPayoffView;
  // ADR 0126 (amendment of 2026-10-08): what happens to each target,
  // one entry per target clause of the statement this purpose rides
  // on, keyed by the slot a move's `targets[].slot` names.
  targets?: TargetPurposeView[];
}

// TargetPurposeView is what a spell or an ability does to the target
// picked for clause `slot` (ADR 0126, amendment of 2026-10-08). The
// player amounts are that player's. Bot data; every amount is absent
// when zero.
export interface TargetPurposeView {
  slot: number;
  draws?: number;
  discards?: number;
  // ADR 0188 Decision 10: the target player takes this many extra turns.
  extra_turns?: number;
  tokens?: number;
  life_gain?: number;
  life_loss?: number;
  damage?: number;
  // #1944: the damage is the announced X; `damage` is then absent.
  damage_is_x?: boolean;
  // #3961: the damage is this many times the number of creatures the
  // cast's "tap any number" cost taps; `damage` is then absent.
  damage_per_tapped?: number;
  // #3161: the spell puts the picked graveyard card onto the battlefield under its caster.
  reanimates?: boolean;
  // #2679: what the target's controller is given when it is removed.
  returns?: TargetReturnView;
}

// TargetReturnView is what a removal spell hands the controller of the
// permanent it removes (#2679). Bot data; every field is absent when
// zero.
export interface TargetReturnView {
  creature_tokens?: number;
  token_power?: number;
  token_toughness?: number;
  life_equal_to_power?: boolean;
  lands?: number;
  lands_untapped?: number;
}

// LandfallPayoffView is what a landfall row pays its controller for each
// land that enters under their control (#3072). Bot data; every field is
// absent when zero.
export interface LandfallPayoffView {
  draws?: number;
  life_gain?: number;
  tokens?: number;
  mana?: number;
  counters?: number;
  energy?: number;
  damage_each_opponent?: number;
}

// ActivationPurposeView is ADR 0106's name for PurposeView.
export type ActivationPurposeView = PurposeView;

// SweepView is a board wipe's description (ADR 0126 §4, §6).
export interface SweepView {
  matches: string;
  how: string;
  amount?: number;
  amount_is_x?: boolean;
  opponents_only?: boolean;
  partial?: boolean;
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
  // #764: every clause of the bullet's statement when it has more
  // than one, in printed order. Absent for the one-clause bullet
  // that is nearly every bullet, where legal_targets is the whole
  // answer.
  clauses?: LegalTargetsView[];
  // Spree (CR 702.172a, ADR 0065's 2026-09-23 amendment): this
  // bullet's own additional mana cost, in brace notation, paid only
  // if it is chosen — on top of the card's printed cost and every
  // OTHER chosen bullet's. Absent for an ordinary modal bullet, which
  // is every modal card before S45.
  cost?: string;
  // ADR 0097 (#1749): this object's ability has already chosen this
  // bullet, so it cannot be chosen again (this turn, or ever — see
  // ModeSpecView.not_chosen). The server refuses it.
  used?: boolean;
  // ADR 0126 §6: what this bullet does. Bot data.
  purpose?: PurposeView;
}

// LegalTargetsView is a clause's legal set right now plus its
// target count (S20 sub-PR 5): min..max picks, max 0 = unbounded.
// #2526: one part of a sacrifice clause's set rule ("a Swamp") and the
// candidates that could fill it. A candidate that fits two parts (a
// Swamp Forest) is listed in both.
export interface SacrificeGroupView {
  label: string;
  cards?: string[];
}

export interface LegalTargetsView {
  // #1943: compatible candidates keyed by the earlier clause's pick.
  relation?: { slot: number; pairs: Record<string, string[]> };
  // #2526: a SACRIFICE clause's set rule — the picks must fill every
  // group with a different permanent. See sacrificeCost.ts.
  each_of?: SacrificeGroupView[];
  // ADR 0137's amendment: a craft clause's "two that share a card
  // type" — every pick must have a key every other pick also has. See
  // sacrificeCost.ts (fitsShares).
  shares?: TargetSharesView;
  // #2097: a SACRIFICE clause that takes every permanent in `cards`
  // ("sacrifice all creatures you control"); min and max are their
  // number. The client confirms rather than picks, and sends
  // `sacrifice_ids` empty for the server to fill.
  all?: boolean;
  players?: string[];
  cards?: string[];
  min?: number;
  max?: number;
  // S22: the clause's count is the announced X, not a printed
  // constant ("Exile X target creatures you control"). min / max are
  // meaningless until X is chosen, so the picker substitutes the X
  // collected in the cost prompts.
  count_from_x?: boolean;
  // With count_from_x: the announced X is a ceiling ("up to X target
  // cards"), so the picker confirms with any number from 0 to X.
  up_to_x?: boolean;
  // #764: the clause's printed wording, shown in the picker banner
  // when a statement has more than one clause and the banner has to
  // say which question it is asking.
  label?: string;
  // #764: this clause's picks must differ from every EARLIER
  // clause's ("a second target permanent you control").
  distinct?: boolean;
  // #1559: the clause's rule over the chosen SET (CR 601.2c) — no two
  // picks may share a key ("that each have a different mana value",
  // "controlled by different players"). The picker greys a candidate
  // whose key a pick of THIS clause already holds, and says the rule.
  // A card missing from `keys` collides with nothing.
  different?: TargetDifferenceView;
  // #1807: the opposite rule — every pick must SHARE one key ("from a
  // single graveyard", keyed on the card's owner). The picker greys a
  // candidate whose key differs from a pick of THIS clause, and says
  // the rule. A card missing from `keys` fits any group.
  same?: TargetDifferenceView;
  // #1559: "with mana value X or less", X the announced X. Like
  // count_from_x, the server built this legal set before X was chosen,
  // so it is a superset: the picker drops every card whose
  // `mana_values` entry exceeds the X collected in the cost prompts,
  // and a card with no entry (an unreadable cost) meets no bound.
  mana_value_at_most_x?: boolean;
  mana_values?: Record<string, number>;
  // #3860, CR 601.2f: the extra life each legal card costs the caster
  // to name, keyed by card id. Terror of the Peaks: "spells your
  // opponents cast that target this creature cost an additional 3 life".
  // Present only on a spell's clause, only for taxed cards.
  life_taxes?: Record<string, number>;
  // The cards in `life_taxes` whose tax the viewer cannot pay right now
  // (CR 119.4). A player who can't pay can't target them: the picker
  // drops them from the legal set and the dock says why.
  life_taxes_unpayable?: string[];
  // ADR 0109 §9: the rest of the X bounds, on the same superset rule.
  // `mana_value_equals_x` narrows from `mana_values` by an EXACT match
  // (Lazav); `power_at_most_x` and `toughness_at_most_x` narrow from
  // `powers` and `toughnesses`. With `x_from_counters_removed` the X is
  // the number of counters the activation's cost is removing (Simic
  // Manipulator), not the one collected in the X prompt. A card with no
  // entry in the map meets no bound.
  mana_value_equals_x?: boolean;
  power_at_most_x?: boolean;
  toughness_at_most_x?: boolean;
  x_from_counters_removed?: boolean;
  powers?: Record<string, number>;
  toughnesses?: Record<string, number>;
  // #1563, CR 601.2d: the clause's effect is "divided as you choose"
  // among its picks. The picker asks for a share per pick once a step
  // has two or more — each at least 1, summing to the amount — and
  // sends them as the action's `distribution`. Absent on every clause
  // that divides nothing.
  divide?: DivideView;
}

// DivideView is the amount a clause divides (#1563): a fixed `total`,
// or — with `from_x` — the X collected in the cost prompts, doubled
// once X reaches `double_from_x` (Shatterskull Smashing). A pick_target
// prompt's amount is always a fixed `total`.
export interface DivideView {
  total?: number;
  from_x?: boolean;
  double_from_x?: number;
  // #1657: "distribute UP TO that many" (Lathiel) — the shares may add
  // up to less than the amount, each pick still at least 1. An amount
  // read off the board (Ureni's lands, Orca's power, Avacyn's
  // Judgment's madness X) arrives already resolved into the fields
  // above; there is nothing for the client to compute.
  up_to?: boolean;
}

// TargetDifferenceView is a clause's set rule (#1559): `label`
// completes "those targets must …", and `keys` maps each legal card
// to the value no two picks may share.
export interface TargetDifferenceView {
  label: string;
  keys?: Record<string, string>;
}

// TargetSharesView is LegalTargetsView.shares (ADR 0137's amendment):
// `label` is the rule's words ("share a card type"), and `keys` maps
// each candidate to every key it has (an artifact creature has two).
// A candidate missing from `keys` can be part of no payment.
export interface TargetSharesView {
  label: string;
  keys?: Record<string, string[]>;
}

/**
 * CastSurfaceView is the announce surface of ONE CASTABLE OBJECT:
 * everything the cast chain asks about before `cast_spell` goes out,
 * for one half of one card out of one zone.
 *
 * It is carried twice (#992). `CardView` extends it for the face that
 * is UP, which is what every reader in this client has always read.
 * `CardFaceView` extends it for each face a cast may CHOOSE — both
 * halves of a modal DFC (CR 712.11b) and of an adventure card
 * (CR 715.3) — so `cardAsFace` can swap the block in when the player
 * picks a half instead of clearing what the front published. Clearing
 * it is why casting Stomp from this client never opened a target
 * picker.
 *
 * Same field names, same wire keys, both places: the picker, the X
 * stepper, the cost modal and the targeting flow read one shape and
 * do not know which of the two they were handed.
 */
export interface CastSurfaceView {
  // target_mode tells the cast-click flow what to prompt for at
  // announce time. Empty/absent ⇒ cast immediately with no target.
  // See client/src/lib/targeting.ts for the full enum.
  target_mode?: string;
  // S20: for a card in the viewer's own hand / command zone with a
  // structured TargetSpec — the players and card instance IDs its
  // target slot accepts right now. Absent for free-form cards. Both
  // lists empty = no legal target = not castable right now.
  legal_targets?: LegalTargetsView;
  // #764: every clause of a MULTI-clause target statement, in
  // printed order, each with its own legal set, count and printed
  // wording — Bite Down's "target creature you control" then "target
  // creature or planeswalker you don't control". Absent for the
  // single-clause card that is nearly every card. The picker walks
  // them one prompt at a time.
  clauses?: LegalTargetsView[];
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
  // #1012: the PRINTED mana cost is not one of the prices this cast
  // may claim out of the zone the card is in, so the caster must name
  // one of `alternative_costs`. A Faithless Looting in the graveyard
  // is castable at its flashback cost and at nothing else; a card a
  // permission PRICES is the same shape.
  //
  // Absent — every ordinary hand cast, every command-zone cast, and a
  // Gravecrawler whose graveyard permission carries no price — means
  // the printed cost is on the menu as usual. `castable_here` is one
  // bit and says only that a cast is possible from here; this is the
  // other half of the sentence, and the client must not infer it from
  // the shape of the offer list.
  alternative_cost_required?: boolean;
  // #1686: the printed cost IS one of the prices (`alternative_cost_required`
  // is false) but the engine will refuse it RIGHT NOW for timing (CR
  // 307.1) — the one hand-zone exception to "absent means the printed
  // cost is on the menu as usual" above. A live miracle grant opens
  // its OWN claim at instant speed and says nothing about the printed
  // one, which stays whatever timing the card prints; drawn on another
  // player's turn, the printed sorcery cast is still on the menu but
  // not choosable until the caster's own main phase. Read via
  // printedCostCastableNow (targeting.ts), not printedCostClaimable —
  // the zone browser's button label and canCastFromHand's tooltip stay
  // on the zone-and-payability question alone.
  printed_cost_timing_closed?: boolean;
  // ADR 0073 (#664): the "you may pay an additional cost" offers this
  // card makes — kicker, multikicker, buyback. Rendered inside the
  // same picker the alternative costs open, because CR 601.2b
  // announces them together. Absent for nearly every card.
  optional_costs?: OptionalCostView[];
  // ADR 0150 §5 (CR 702.132a): a spell with assist's offer to this
  // caster — who may be asked to help pay its generic mana, and how
  // much generic there is. The cast modal's Assist row reads it and the
  // pick rides cast_spell as `assist_player`. Absent for every card
  // without assist.
  assist?: AssistOfferView;
  // #760 (ADR 0073 §7): the printed clause that stops this card being
  // cast from the zone it is in right now — "Each player can't cast
  // more than one spell each turn", "Cast this spell only if you
  // control a legendary creature or planeswalker". Absent, which is
  // nearly always, means nothing refuses the cast.
  //
  // The server's own cast gate answered this, so a card carrying it
  // is one the server WILL refuse: grey it and show the clause rather
  // than dispatching cast_spell and surfacing a toast.
  cant_cast?: string;
  // S22: for a card in the viewer's own hand that lets you tap your
  // own permanents to help pay — convoke and waterbend. The cast
  // flow opens a picker after X and before targeting. Absent for
  // nearly every card.
  tap_cost?: TapCostView;
  // ADR 0100: delve (CR 702.66) — the caster's graveyard cards that may
  // be exiled to pay generic mana, in the server's payment order, and a
  // hint of how many. The picker's real cap is the auto-tap preview's
  // `delve_budget`. Absent for every card without delve.
  delve?: DelveView;
  // #746 (ADR 0048 addendum): the printed clauses of this card's own
  // cost modifiers whose price depends on its targets — Fireball's
  // "This spell costs {1} more to cast for each target beyond the
  // first", strive. The X picker opens before targeting and its
  // affordability readout is priced at one target, so it shows these
  // clauses under the readout. Absent for nearly every card and on
  // opponents' cards the viewer cannot read.
  target_cost_notes?: string[];
  // #2167: the card types a cast or play of this face may spend under a
  // permission that opens one of each type — Muldrotha, the Gravetide's
  // "a permanent spell of each permanent type", Aminatou's Augury's "for
  // each nonland card type". The viewer's own answer, ranked: the first
  // is the type the permission's other cards need least, which the type
  // picker selects by default. Two or more entries mean the cast chain
  // asks which one (PermissionTypeModal) and sends it as
  // `permission_type`; one needs no question. Absent on every other cast.
  permission_types?: string[];
  // CR 107.4 (#916): how many symbols in the printed cost carry the
  // "or 2 life" option — 1 for Gitaxian Probe's "{U/P}", 2 for
  // Dismember's "{1}{B/P}{B/P}", 1 for a compleated planeswalker. The
  // cast flow opens a "pay N with life" stepper bounded by it and by
  // the caster's life total, and sends the answer as `phyrexian_life`.
  // Stamped with the other cast clauses on the viewer's own castable
  // cards and absent everywhere else, so its presence IS the question
  // "is there a life half to offer here".
  phyrexian_symbols?: number;
  // ADR 0131 (#2531): how many of `phyrexian_symbols` are payable with
  // life only because the viewer controls a permanent that grants it —
  // K'rrik's "for each {B} in a cost, you may pay 2 life rather than pay
  // that mana" — rather than because the cost prints a Phyrexian symbol.
  // Absent for every viewer without such a permanent. When every symbol
  // is granted the cast asks about life only if mana falls short, and
  // the card's menu keeps a "Pay life for {B}…" row.
  phyrexian_granted?: number;
  // S29: set on a card sitting in a zone its own text opens as a
  // cast source — a flashback card in the graveyard. The zone
  // browser keys its cast button off this, the way exile keys its
  // impulse button off `exile_play`. Never set on hand or
  // command-zone cards: those surfaces are cast surfaces for
  // everything in them. The cost to pay rides `alternative_costs`,
  // already filtered to the offers claimable from this zone.
  //
  // #1015: the server derives it from that offer list and its own
  // cast gate — "nothing refuses this cast, and at least one price is
  // claimable". An escape card in a graveyard too small to pay for it
  // is NOT castable here, and used to render a button with no offer
  // behind it. Whether the printed cost is one of those prices is
  // `alternative_cost_required`, not this bit.
  //
  // WHOSE ANSWER IT IS: yours, always (#1055). The bit is stamped
  // only on the frame of a seat that may actually make the cast — the
  // pile's owner for a printed flashback, the holder of a permission
  // over the card for a granted one, both of them on their own frames
  // when both are true — and is absent on everybody else's copy of the
  // same card, spectators included.
  //
  // It was public until #1055, and it meant the PILE OWNER's answer,
  // so a reader had to pair it with `exile_play` to find out which of
  // the two it was holding. Both readers — zoneBrowser.logic and
  // libraryTop — now ask the bit alone. What is still public is the
  // half that is a fact about the CARD rather than about a player:
  // `alternative_costs`, `alternative_cost_required`, `modes`,
  // `additional_cost`, `optional_costs`, `tap_cost`,
  // `target_cost_notes`, `phyrexian_symbols` and `cant_cast`, because
  // a card in a graveyard is a card every player may read.
  castable_here?: boolean;
  // #3625: a HAND card the viewer holds a pass-closed free cast over
  // (Buster Sword, Rishkar's Expertise; #3270): their next pass ends the
  // window and the card stays in hand. The hand twin of
  // `exile_play.closes_on_pass`; per viewer, absent otherwise.
  closes_on_pass?: boolean;
  // #1389: what THIS viewer would be charged to cast the card out of
  // EXILE right now — one entry per price the cast may claim, cheapest
  // first, each the total AFTER every CR 601.2f cost modifier (the
  // same pricer the cast path and the auto-tap preview use). Exile
  // only, and only on the frame of a seat holding a LIVE permission:
  // a warp or foretell grant whose later turn has not come carries
  // none. Read it through exileStrip.ts, which decides the badge.
  //
  // #2202: also on the cards in the viewer's OWN command zone, priced
  // for a cast from there — the commander tax (CR 903.8) plus every
  // cost modifier. `printed` is false once the tax is above 0. Read it
  // through castStrip.ts (commanderCostBadge).
  //
  // Since #1389 `castable_here` is stamped in exile too, with the same
  // meaning it has in a graveyard — "YOU may cast this from here NOW",
  // timing included — and is what the castable-from-exile strip
  // lights a card by.
  cast_prices?: CastPriceView[];
  // #2581: the largest X THIS viewer may announce for the spell right
  // now, under its printed "X can't be greater than <count>" (Winter's
  // Chill's snow lands, Open the Way's players). The server refuses a
  // larger X, so the X picker stops here. Absent for every card with no
  // printed ceiling; present-and-0 means X = 0 is the only announcement.
  x_max?: number;
}

// CastPriceView is one price a cast out of exile may claim (#1389).
export interface CastPriceView {
  // The `alternative_costs[i].key` this price claims — "foretell", a
  // granted "flashback" — and the value the cast sends as
  // `alternative_cost`. Absent for the path that claims none: the
  // printed cost, or a permission's own flat price (airbend's {2}, a
  // plotted card's {0}).
  alternative_cost?: string;
  label?: string;
  // Brace notation, never empty: a free cast reads "{0}".
  cost: string;
  // Life charged on top (CR 119.4). Absent for every exile price today.
  life?: number;
  // ADR 0129 §5: energy charged on top (Amped Raptor's "an amount of {E}
  // equal to its mana value"). Absent for every price that charges none.
  energy?: number;
  // True when this price IS the printed mana cost, untouched — the
  // strip draws no badge for it.
  printed?: boolean;
}

/**
 * One printed face of a multi-face card (ADR 0034). Enough to render
 * a picker row and a hover panel — and, since #992, enough to CAST:
 * it extends CastSurfaceView, so a face the card offers a cast of
 * carries the same announce block the card carries for the face that
 * is up.
 *
 * The block is present only on a face a cast may actually choose:
 * both halves of a modal DFC and of an adventure card, the front
 * alone of a transform card, and exactly the faces a grant names when
 * one does (CR 715.4's Adventure permission opens the creature and no
 * other). On every other face the fields are simply absent, which
 * reads correctly as "this half announces nothing".
 */
export interface CardFaceView extends CastSurfaceView {
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

// One entry of CardView.land_type_effects (ADR 0109 §1): "Island until
// end of turn — Tidal Warrior".
export interface LandTypeEffect {
  // The land types the effect gives ("Island").
  types: string[];
  // "In addition to its other types" (CR 205.1b). Absent is CR 305.7's
  // replacement: the old land types and rules-text abilities are gone.
  in_addition?: boolean;
  // ADR 0109 §2: "loses all land types" (Ultima's blight); types is
  // empty. With it, loses_abilities is the same effect's "and
  // abilities" and gains the texts of the abilities it gives.
  loses_all?: boolean;
  loses_abilities?: boolean;
  gains?: string[];
  // The duration in the card's words; absent for an effect with none.
  until?: string;
  // The card whose effect it is.
  source?: string;
}

export interface CardView extends CastSurfaceView {
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
  // ADR 0078: true for a token (CR 111), absent for a real card.
  // Public — survives the non-knower redaction — because everyone at
  // a paper table can see a token is a token. When present alongside
  // `scryfall_id`, that id is a resolved Scryfall TOKEN PRINTING
  // chosen for its art, not a card this player owns.
  is_token?: boolean;
  // #3043: a token that is a copy of a real card (carries the original's
  // art and frame). Always alongside is_token; public.
  token_copy?: boolean;
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
  // ADR 0172 §3: a deck's revealed companion, outside the game or after
  // it has come into the game. Public; drives the "Companion" badge.
  is_companion?: boolean;
  // Damage marked on this creature for the lethal-damage SBA (S13.1,
  // CR 704.5g). Cleaned up in cleanup step (S13.2, CR 514.2). Only
  // meaningful on the battlefield; omitted when zero.
  damage_marked?: number;
  // Regeneration shields on this permanent (CR 701.19a, #667). Each
  // one replaces the next destruction this turn: instead of dying it
  // is tapped, all damage is removed from it and it leaves combat.
  // Public state, like damage_marked; omitted when zero.
  regeneration_shields?: number;
  // ADR 0108 §1: the effects that would exile this permanent instead if
  // it died now (Lava Coil's "if that creature would die this turn,
  // exile it instead"), by source name. Public; battlefield only.
  exiled_if_it_dies?: string[];
  // ADR 0108 §2 (CR 701.19c): this permanent can't be regenerated this
  // turn (Incinerate, Whippoorwill). Public; battlefield only.
  cant_be_regenerated?: boolean;
  // ADR 0187 §4 (CR 510.1a): this creature assigns no combat damage this
  // turn (Carrion Rats, Carrion Wurm). Public; battlefield only.
  assigns_no_combat_damage?: boolean;
  // ADR 0109 §10 (CR 702.136a): this permanent's haste came from its
  // own riot, so its haste chip is labelled "Riot". Public;
  // battlefield only.
  riot_haste?: boolean;
  // ADR 0109 §1: the resolved effects changing this permanent's land
  // types, oldest first — Tidal Warrior's "becomes an Island until end
  // of turn", Navigator's Compass's "in addition to its other types".
  // The type line already shows the result. Public; battlefield only.
  land_type_effects?: LandTypeEffect[];
  // S13.5 visual face-down flag (CR 708 — morph / manifest /
  // mutate-bottom, Necropotence's exile). Distinct from known_by_you:
  // a viewer who doesn't know a face-down card gets it redacted to
  // game state and Card.svelte draws a back (cardBack.ts, #95); a
  // viewer who knows it still sees the face.
  face_down?: boolean;
  // WHY it is face down (ADR 0069): "exiled" (CR 406.3, Necropotence),
  // "foretold" (CR 702.143b), "hideaway" (CR 702.75a, ADR 0091 — the
  // controller of the permanent that hid it may look), "permitted"
  // (#1573 — the holder of the cast permission over it may look:
  // Gonti, Night Minister; Outrageous Robbery), or one of the
  // CR 708.2 permanent states
  // "manifested" / "morphed" / "disguised" / "cloaked". PUBLIC —
  // everyone can see that a permanent is a morph — so it survives the
  // non-knower redaction and labels the card back.
  face_down_kind?: string;
  // Whether THIS viewer may look at the face of a face-down object:
  // its controller for a CR 708.5 permanent, its owner for a foretold
  // card (CR 702.143d), nobody for a plain face-down exile (CR 406.3).
  // Stamped per-viewer by the server; equals `face_down && known_by_you`.
  // True means "draw the real face plus a face-down badge".
  face_visible?: boolean;
  // A phased-out permanent (CR 702.26, #1199, ADR 0084). Always true
  // on a card in `GameView.phased_out` and absent everywhere else, so
  // it is redundant with the zone the card arrived in — carried so a
  // CardView pulled out of that zone into a list still says what it
  // is. PUBLIC, like face_down_kind: it survives the non-knower
  // redaction, because everyone can see the board stop showing a
  // permanent and everyone needs to be able to tell "phased out" from
  // "died".
  phased_out?: boolean;
  // S13.5 per-viewer knowledge flag. True when the viewer is in the
  // server-side KnownBy set for this card. When false, printed
  // characteristics (name, type_line, scryfall_id, power, toughness,
  // counters, is_commander) are zero/empty.
  //
  // A face-down PERMANENT is the one exception to "zero/empty": its
  // CR 708.2 body (type line "Creature", 2/2, no name, no colours) is
  // public and reaches every viewer, because an opponent has to see
  // the 2/2 to block it.
  known_by_you?: boolean;
  // #671: the OTHER seats (player IDs, table order) that know this
  // hand card. Sent on the viewer's own hand cards only.
  known_by_seats?: string[];
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
  // #1339: the seat defending against this attack — the only seat
  // whose creatures may block it (CR 802.4a): the player attacked, the
  // planeswalker's controller, or the battle's PROTECTOR. Absent when
  // nothing is declared. #1364: once the attacked planeswalker or
  // battle has left, it still names the player who was defending it
  // (CR 506.4c "it may be blocked"), and attacking_target_kind is
  // absent. #1376: the same holds when the planeswalker or battle is
  // removed from combat WITHOUT leaving — a control change or phasing
  // out — and attacking_target is then the reserved id
  // "00000000-0000-0000-0000-000000000506", which names no seat or
  // card. Read it through defendingPlayerOf (attackTargets.ts), which
  // covers older frames.
  defending_player?: string;
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
  // #1706: every attacker this card blocks, in declaration order, when
  // it blocks two or more (High Ground, Palace Guard). Omitted for an
  // ordinary blocker — read blockedAttackersOf rather than either field.
  blocking_targets?: string[];
  // #1706: how many attackers this creature can block when an effect
  // lets it block more than one; blocks_any_number instead when it can
  // block any number. Both omitted for the ordinary one-attacker case.
  block_capacity?: number;
  blocks_any_number?: boolean;
  // Player ID whose goad on this creature is the most recent, or
  // omitted when not goaded. Cleared on zone exit. Added in S10;
  // enforced server-side since #1571 (CR 701.15b — attacks each combat
  // if able, and a player other than the goader if able).
  goaded_by?: string;
  // #1598 (CR 701.15c): every player whose goad is on this creature,
  // oldest first; its last entry equals goaded_by. Omitted when not
  // goaded, and by a server from before #1598.
  goaders?: string[];
  // #1571 (ADR 0045 Decision 51): true on a creature the active player
  // owes an attack with right now — during declare_attackers, while the
  // declaration could still obey a CR 508.1d requirement it does not
  // (Zurgo, goad, Bident of Thassa). Server-computed; the client
  // renders it and never derives a requirement. Omitted otherwise.
  must_attack?: boolean;
  // #1597 (CR 509.1c): true on a creature a defending player owes a
  // block with right now — during declare_blockers, while their
  // declaration is pending and could still obey a requirement it does
  // not (Lure, Grand Melee, "blocks each combat if able", "must be
  // blocked if able"). Server-computed; the client renders it and never
  // derives a requirement. Omitted otherwise.
  must_block?: boolean;
  // #3898: on a creature a defending player who is still declaring
  // could put into the declaration, each attacker it could not be
  // declared blocking on its own right now, with the server's reason
  // token and sentence (addressed to its controller). A lone block on a
  // menace attacker reads too_few_blockers; the attacker's
  // `min_blockers` says how many it takes. Pairs already declared and
  // anything face down are left out. Only during declare_blockers. The
  // board greys those attackers while this creature is selected.
  block_refusals?: BlockRefusalView[];
  // #3898: the fewest creatures that may block this attacker (CR
  // 509.1b: 2 for menace), while its defending player is declaring.
  // Omitted when it is less than 2.
  min_blockers?: number;
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
  // #2258 (ADR 0118 decision 8): the card prints a mana ability the
  // catalog does not run (unimplemented AND the printed text has an
  // "Add {X}" clause). Smart autopass reads it to hold the main phase
  // for "Cast anyway". Absent when false.
  may_make_mana_unautomated?: boolean;
  // ADR 0153 §5: on a hand card that does something when a spell or
  // ability an opponent controls makes its owner discard it. Declared by
  // the card, never inferred: "battlefield" (put onto the battlefield
  // instead), "library" (on top of its owner's library instead) or
  // "trigger" (the discard triggers an ability of the card). Absent
  // otherwise.
  discard_by_opponent?: "battlefield" | "library" | "trigger";
  // #992: the announce surface of the face that is UP. Inherited
  // from CastSurfaceView, so every existing `card.legal_targets`
  // read is unchanged and `cardAsFace` has an identically shaped
  // block on each face to swap in.
  // S21 sub-PR 6: present on a card in exile that someone may play
  // this turn — and, since ADR 0066, on a card in a graveyard or on a
  // library top that a permission opens. Absent for ordinary exile,
  // which is nearly all of it.
  //
  // PUBLIC: the trigger that granted it resolved in the open, so every
  // viewer gets one. #1037: a viewer who holds a permission over the
  // card gets THEIR OWN rather than whichever live permission the
  // server found first, so two seats that may both cast one card each
  // see the grant they would cast under — its cost override, its
  // faces, its any-color clause. Read it, never guess from the zone.
  exile_play?: ExilePlayView;
  // S21 sub-PR 2: activated abilities offered by this permanent.
  activated_abilities?: ActivatedAbilityView[];
  // #660: activated abilities this card offers while it is IN HAND —
  // cycling and typecycling (CR 702.29). A separate field from
  // `activated_abilities` because the two are read by different UI
  // and because a hand, unlike the battlefield, is not public: the
  // server strips this from every seat but the hand's owner. `index`
  // is the ability's index in the card's FULL list, so the same
  // activate_ability payload works for both.
  zone_abilities?: ActivatedAbilityView[];
  // #1228: MANA abilities this card offers while it is IN HAND —
  // "Exile this card from your hand: Add {R}" (the Spirit Guides,
  // CR 113.6). `zone_abilities`' twin one ability kind over, and a
  // separate field for the reason `mana_abilities` is separate from
  // `activated_abilities`: the wire payload differs
  // (`activate_mana_ability`, not `activate_ability`), so a client
  // sends the verb that matches the row it read. `index` is the
  // ability's index in the card's FULL mana-ability list. Stripped
  // from every seat but the hand's owner, like `zone_abilities`.
  zone_mana_abilities?: ManaAbilityView[];
  // #658 / #659: CR 116.2 special actions this card offers while it
  // is IN HAND — "Foretell {2}", "Suspend 1—{R}". Not abilities and
  // not casts: they use no stack and there is nothing to respond to,
  // so a row fires `special_action` directly with no picker in
  // between. Hidden from every seat but the hand's owner, like
  // `zone_abilities`.
  special_actions?: SpecialActionView[];
  // S21 sub-PR 2: CR 302.6 summoning sickness — entered this turn
  // without haste, so it can't attack or pay a {T} cost.
  summoning_sick?: boolean;
  // #3928: a summoning-sick creature whose controller may still pay its
  // {T} costs, under an "as though those creatures had haste" static
  // (Thousand-Year Elixir, CR 609.4). It still can't attack. Only ever
  // set beside `summoning_sick`; read through `activationSick`.
  activates_as_though_haste?: boolean;
  // ADR 0108 §5: this permanent's echo triggers at its controller's
  // next upkeep (CR 702.30a). Derived on the server; never set on a
  // face-down permanent.
  echo_due?: boolean;
  // CR 606.3: a loyalty ability has already been activated on this
  // planeswalker this turn, so every loyalty row in its menu is
  // greyed until the turn cursor moves on. Before #334 this state
  // was server-only, which is why canActivateLoyalty had to take
  // the caller's guess as an argument.
  loyalty_activated?: boolean;
  // ADR 0071 (CR 716.2): a Class permanent's level designation — 1
  // for a Class nobody has levelled, up from there. Present only for
  // a Class on the battlefield, so the badge renders on presence
  // rather than on parsing the type line. An uncatalogued Class
  // carries it too: the level is engine state, not catalog state.
  class_level?: number;
  // ADR 0071 (CR 719.3): this Case is solved, and its "Solved —"
  // lines are on. Absent — not `false` — for everything else.
  //
  // There is no field for "which printed abilities are active": an
  // inactive ACTIVATED ability is already missing from
  // `activated_abilities`, and an inactive static or trigger has no
  // per-ability representation here to grey out.
  solved?: boolean;
  // ADR 0090 (CR 722.3a): this permanent is prepared — its controller
  // may cast the copy of its prepare spell that sits in exile, which
  // arrives as an ordinary exile card with an `exile_play` stamp
  // naming face 1. Absent — not `false` — for everything else.
  prepared?: boolean;
  // ADR 0103 (CR 709.5): a Room's doors. Present only for a face-up
  // Room on the battlefield. A fully locked Room has no name (CR 709.5),
  // so `name` is empty and the client labels it from `faces`.
  doors?: RoomDoorsView;
  // ADR 0071 amendment, #1321 (CR 701.64): this permanent is
  // harnessed — the marker that switches its printed "∞ — [ability]"
  // lines on. Unlike `class_level` / `solved` it carries no subtype
  // gate: any permanent can print "Harness [this permanent]", so the
  // field is read straight off the card once it is on the
  // battlefield. Absent — not `false` — for everything else.
  harnessed?: boolean;
  // ADR 0071 amendment, #1700 (CR 701.37b): this permanent is
  // monstrous — set by the monstrosity keyword action ("{cost}:
  // Monstrosity N.") and kept until the permanent leaves the
  // battlefield. Switches on its "as long as this creature is
  // monstrous" lines. Read straight off the card the same way
  // `harnessed` is — no card type owns monstrosity. Absent — not
  // `false` — for everything else.
  monstrous?: boolean;
  // ADR 0071 amendment 2026-10-09, #2049 (CR 702.112b): this permanent
  // is renowned — one of its renown triggers resolved, and it stays
  // renowned until it leaves the battlefield. Switches on its "as long
  // as this creature is renowned" lines. Read straight off the card,
  // like `monstrous`. Absent — not `false` — for everything else.
  renowned?: boolean;
  // ADR 0071 amendment 2026-10-08, #2695 (CR 702.171): this Mount is
  // saddled — until end of turn, or until it leaves the battlefield.
  // Switches on its "while saddled" lines and is what its "attacks
  // while saddled" triggers read. Public, read straight off the card.
  // Absent — not `false` — for everything else.
  saddled?: boolean;
  // ADR 0186 §6 (#3543, CR 702.122c / 702.171c): the card IDs of the
  // creatures tapped this turn to pay for this Vehicle's crew ability
  // (`crewed_by`) or this Mount's saddle ability (`saddled_by`), in the
  // order they were tapped, that are still on the battlefield as the same
  // objects. Public. Absent — not `[]` — when nobody paid, and absent from
  // the next turn on.
  crewed_by?: string[];
  saddled_by?: string[];
  // ADR 0071 amendment, #2698 (CR 701.60): this permanent is
  // suspected — it has menace and can't block for as long as it is.
  // Public, and kept on a face-down permanent: the designation was
  // given to the object in public, and it is the reason the creature
  // cannot block. Absent — not `false` — for everything else.
  suspected?: boolean;
  // ADR 0114 §3, §9 (CR 701.54b): this permanent is its controller's
  // Ring-bearer. Public, and kept on a face-down permanent: the
  // designation was chosen in public and says nothing about the card.
  // Absent — not `false` — for everything else.
  ring_bearer?: boolean;
  // #781 (CR 105.4 / CR 614.12): the answers this permanent's
  // controller gave to its "as this enters, choose a color" and "as
  // this enters, choose a creature type" instructions — one uppercase
  // colour letter ("G") and one canonical creature type ("Elf").
  // Absent when the card asks no such question, and absent between
  // the permanent entering and the prompt being answered.
  //
  // PUBLIC. The choice is announced at the table, and CR 607.2d makes
  // it the only way to read the card's other lines: "creatures you
  // control of the chosen color" names a set nobody can compute
  // without it. Cleared with the other type-derived bits for a card
  // the viewer is not a knower of — "Elf" names Cavern of Souls.
  //
  // RENDER THEM THROUGH `chosenValueChips` (chosenValues.ts). That
  // module is the one place either letter becomes a word, so the card
  // tile, the hover panel and the zone browser cannot disagree about
  // what "G" means.
  chosen_color?: string;
  named_tribe?: string;
  // #1210 (CR 614.12): the CARD NAME this permanent's "as this enters,
  // choose a card name" instruction was answered with — Pithing
  // Needle, Phyrexian Revoker, Sorcerous Spyglass. Same lifecycle,
  // PUBLIC/redaction and rendering module as chosen_color above.
  chosen_name?: string;
  // #1572 (ADR 0071): the NAMED OPTION this permanent's "as this
  // enters, choose <A> or <B>" instruction was answered with — a
  // Siege's anchor word, "Khans" or "Temur". Decides which of the
  // permanent's OWN printed abilities exists. Same lifecycle,
  // PUBLIC/redaction and rendering module as chosen_color above.
  chosen_option?: string;
  // #1941: the number chosen as this permanent entered — the life
  // Phyrexian Processor's controller paid, the size of its tokens.
  // Absent at zero. Same lifecycle, PUBLIC/redaction and rendering
  // module as chosen_color above.
  chosen_number?: number;
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
  // ADR 0083 — a TOKEN's printed ability text, verbatim ("When this
  // token dies, you gain 1 life."). Absent for every printed card and
  // for a vanilla token.
  //
  // A token has no printing behind it, so there is no scryfall_id to
  // resolve oracle text from and the card renders through the
  // name-fallback path. An activated ability still reaches the player
  // as a row in the right-click menu; a TRIGGER or a STATIC has no
  // control, so without this the text would be invisible. Newlines
  // separate printed lines.
  token_text?: string;
  // ADR 0093 Decision 8 — the abilities OTHER effects gave this
  // permanent, each with the granting card's printed text: Cryptolith
  // Rite's "{T}: Add one mana of any color." on a creature, a copy's
  // CR 707.9a grant (no source). The one place a granted TRIGGER is
  // visible at all — it has no menu row. One entry per grantor.
  granted_abilities?: GrantedAbilityView[];
  // ADR 0155 owner decision 2: the objects this permanent "has all activated
  // abilities of" that the catalog does not automate. They lend nothing; the
  // menu names each in a disabled "<lender>: not automated" line.
  unautomated_lenders?: GrantedByView[];
  // ADR 0160 (CR 612.5) — the object whose text box this permanent has
  // when it is not its own (Exchange of Words, Deadpool, Trading Card).
  // Its abilities, keywords and rows already are that text's; the
  // printed oracle text the inspector fetches is not. Absent when the
  // permanent has its own text.
  text_box_from?: string;
  // ADR 0160's 2026-10-15 amendment (CR 612.2) — the words of this
  // object's text a text-changing effect replaced, one line each:
  // "black → blue", "Swamp → Island" (Sleight of Mind, Magical Hack).
  // The printed oracle text uses the old words. Absent when none.
  text_changes?: string[];
  // #2219 — the card's non-keyword triggered, static and activated
  // abilities right now, triggered first, each with a short label the
  // SERVER wrote (the stack's label for a triggered or activated row).
  // The art tile counts them in one chip per kind (⚡ ◆ ↻) and lists
  // the labels on hover or focus. Battlefield and hand only; absent
  // for an uncatalogued card, a keyword-only card and a face-down
  // permanent. The client never reads oracle text for these.
  ability_rows?: AbilityRowView[];
  // ADR 0126 §6: what the card does, as the catalog declares it. Hand,
  // battlefield and stack; cleared with ability_rows for a viewer who
  // cannot see the card. Bot data; the client does not read it.
  purpose?: PurposeView;
  // #662 — this permanent's CR 702.16 protections, already PARSED by
  // the server. The raw "protection from red" tokens are in
  // `abilities` like every other keyword; this is the same list with
  // the quality pulled out, so the badge row renders "Protection from
  // Demons" without the client owning a copy of the grammar.
  protection?: ProtectionView[];
  // S24 — the restriction set the server computed for this
  // permanent: "cant_attack", "cant_block", "cant_be_blocked",
  // "cant_activate", "cant_activate_mana", and (#2733)
  // "cant_become_suspected". Absent for the permanent
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
  // ADR 0106 §2 (#1794) — this creature's restrictions on WHOM it may
  // attack: Xantcha's "can't attack its owner or planeswalkers its
  // owner controls", with the owner already resolved to a seat id.
  // Draws the card's "CAN'T ATTACK <name>" chip (cantAttack.ts) and
  // nothing else: which targets an attacker may pick is the
  // legal_actions digest's `attack_targets`. Absent for nearly every
  // card, and for a restriction naming the creature's own controller.
  attack_target_restrictions?: AttackTargetRestrictionView[];
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
  // ADR 0103 (CR 702.102): a split card with fuse in its owner's hand
  // also offers the FUSED cast of both halves — this is its announce
  // surface (both names, both costs, the left half's clauses then the
  // right half's). The face picker offers it as a third choice and the
  // cast sends `fuse: true`.
  fused?: CardFaceView;
  // ADR 0145 (CR 712.4a): a MELDED permanent's two cards, carrier
  // first. Everything above describes the combined back face, which
  // is the permanent; this is what it is made of, and what its owner
  // gets back when it leaves. Absent on every other card.
  melded_from?: CardFaceView[];
  // ADR 0145 (CR 712.4): a meld card's combined back face — the
  // permanent it and its partner become when they meld. Absent on
  // every other card and on a melded permanent.
  melds_into?: CardFaceView;
}

// ManaAbilityView mirrors `protocol.ManaAbilityView` server-side —
// ProtectionView is one "protection from <quality>" on a permanent
// (CR 702.16), parsed by the server. #662.
//
// The client never parses a protection token. Protection is the only
// keyword whose ability carries a parameter, and the engine keeps
// exactly one closed grammar for it (server/internal/game/
// protection.go); a second copy here would be free to disagree about
// what "protection from Demons" means.
// GrantedByView names the permanent that granted an ability row (ADR
// 0093). `name` is absent when the grantor is no longer there to name.
export interface GrantedByView {
  id: string;
  name?: string;
}

// GrantedAbilityView is one ability another effect gave a permanent,
// as its granting card prints it (ADR 0093 Decision 8). A copy's
// CR 707.9a grant has no source.
export interface GrantedAbilityView {
  text: string;
  source_id?: string;
  source_name?: string;
}

// #2219: one non-keyword ability of a card, for the art tile's chips.
// Mirrors protocol.AbilityRowView.
export type AbilityRowKind = "triggered" | "static" | "activated";
export interface AbilityRowView {
  kind: AbilityRowKind;
  label: string;
  // ADR 0126 §6: the row's declared purpose (death_payoff). Bot data.
  purpose?: PurposeView;
  // ADR 0130 (amendment of 2026-10-07): "linked" on exert's "when you
  // do" trigger, "payoff" on "Whenever you exert a creature". Bot data.
  exert?: "linked" | "payoff";
}

// #3898: one entry of CardView.block_refusals, an attacker this creature
// could not block on its own right now. Mirrors protocol.BlockRefusalView.
export interface BlockRefusalView {
  attacker: string;
  // One of BlockRefusalReason (the illegal_block reasons).
  reason: BlockRefusalReason | string;
  // The server's sentence, for the hover.
  message: string;
}

// ADR 0106 §2 (#1794): one CR 508.1c restriction on whom a creature may
// attack. Mirrors protocol.AttackTargetRestrictionView.
export interface AttackTargetRestrictionView {
  // The seat the creature can't attack (its owner).
  player: string;
  // Also "or planeswalkers that player controls".
  planeswalkers?: boolean;
  // The card whose text imposes it.
  source?: string;
  // ADR 0107 §2 (#1879): on a "can't attack unless defending player
  // controls <X>" row, the X this player controls none of ("Island").
  // The row then names an opponent rather than the owner.
  unless?: string;
  // #1879: on every row a defending-player clause adds, the clause this
  // player doesn't meet, in full: "Bob controls no Island", "Bob isn't
  // poisoned", "Alice doesn't control more creatures than Bob".
  reason?: string;
}

export interface ProtectionView {
  // The quality as the CARD prints it — "red", "Demons",
  // "artifacts", "everything". Badge tooltip text.
  printed: string;
  // Which characteristic of a source the quality is compared
  // against: "color", "card_type", "subtype", "everything",
  // "player", "mana_value_at_most" (value = the bound N, #2181),
  // "ring_bearer" (#2145) or "opponents" (#2745).
  kind: string;
  // What the rules compare — the wire colour ("R"), the lowercase
  // card type ("artifact"), the canonical singular subtype
  // ("Demon"). Absent for "everything".
  //
  // For "player" (CR 702.16k, #980 — True-Name Nemesis) it is the
  // chosen SEAT'S ID, and the comparison is against the source's
  // controller rather than against any characteristic of it. An id,
  // like `CardView.controller`; `printed` stays "the chosen player",
  // so the display string never holds a UUID. Absent while the
  // permanent's as-enters prompt is still open, which reads correctly
  // as "protected from nobody".
  //
  // For "opponents" (#2745 — "protection from each of your opponents")
  // it is the seat that HAS the protection, and a source matches when
  // its controller is any other seat.
  value?: string;
}

// one entry per activated mana ability on a battlefield permanent.
// The client renders these as buttons in a right-click / long-press
// menu anchored to the card. Added in S15 sub-PR 2.
export interface ManaAbilityView {
  index: number;
  label?: string;
  // ADR 0131 §2 (#2531): ActivatedAbilityView's pair for this mana
  // ability's own mana component — how many symbols the viewer could pay
  // 2 life each for, and how many of them are a grant's. The answer
  // rides activate_mana_ability as `phyrexian_life`. Absent at zero.
  phyrexian_symbols?: number;
  phyrexian_granted?: number;
  // ADR 0093 Decision 5: the row's stable ref ("own:<i>",
  // "land:<colour>", "grant:<bundle>:<i>:<n>"), sent back as
  // activate_mana_ability's `ref`. See ActivatedAbilityView.ref.
  ref?: string;
  // ADR 0093 Decision 8: present on a mana ability another permanent
  // granted this one (Cryptolith Rite's "{T}: Add one mana of any
  // color." on every creature you control).
  granted_by?: GrantedByView;
  // ADR 0155 Decision 10: present on a row ANOTHER object's abilities were
  // lent as ("has all activated abilities of …", Necrotic Ooze): the lender
  // by id and printed name. `granted_by` is then the borrowing permanent
  // itself. Absent on every row that is not lent, and when the lender sits
  // in a hidden zone.
  borrowed_from?: GrantedByView;
  tap_cost?: boolean;
  sacrifice_cost?: boolean;
  // #1228: the "Exile this card from your hand" component of a mana
  // ability that functions from a hand (CR 113.6) — the Spirit
  // Guides. The same wire name ActivatedAbilityView carries it under,
  // so one cost chip serves both ability kinds. Advisory; the server
  // validates the zone and pays the exile.
  exile_self?: boolean;
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
  // #758: "Tap an untapped creature you control" on a mana ability
  // (Springleaf Drum). Same fields as an activated ability's
  // TapOthers component; the answer rides activate_mana_ability as
  // `tap_ids`. This surface is fixed-count: CR 605.3b gives a mana
  // ability no X announcement, so `count_from_x` is never set here.
  tap_others_label?: string;
  tap_others_options?: LegalTargetsView;
  // #1600: "Exile a creature you control" on a mana ability (Food
  // Chain). Same fields as ActivatedAbilityView's; the answer rides
  // activate_mana_ability as `exile_permanent_ids`.
  exile_permanent_label?: string;
  exile_permanent_options?: LegalTargetsView;
  // #1213: a "Discard N cards" component on a MANA ability — Skirge
  // Familiar's "Discard a card: Add {B}". Exactly the three fields
  // ActivatedAbilityView carries under exactly the same names,
  // because it is the same component with a second owner: the count,
  // the clause as printed, and the cards in hand that could pay it
  // right now. The picks ride activate_mana_ability as `discard_ids`.
  discard_cost_n?: number;
  discard_cost_label?: string;
  discard_cost_options?: string[];
  // #1283: an "Exile N cards from your hand" component — Cadaverous
  // Bloom's "Exile a card from your hand: Add {B}{B} or {G}{G}". The
  // discard triple's shape under its OWN names, because an exiled
  // card is not discarded (no discard event, nothing for madness to
  // see). The picks ride activate_mana_ability as `exile_ids`.
  // #1297: `exile_cost_zone` names the pile the options are in.
  exile_cost_n?: number;
  exile_cost_label?: string;
  exile_cost_options?: string[];
  exile_cost_zone?: ExileCostZone;
  // S22: a "Pay N life" component of the activation cost — Mana
  // Confluence's "{T}, Pay 1 life:". Advisory only; the server does
  // the real CR 119.4 check. A damage RIDER ("This land deals 1
  // damage to you", the painlands / Ancient Tomb) is NOT a cost and
  // never appears here — it is spelled out in `label` instead.
  life_cost?: number;
  // ADR 0129 §5: a "Pay N {E}" component — Aether Hub's "{T}, Pay
  // {E}:". When the controller is short, cant_activate says so.
  energy_cost?: number;
  // ADR 0130 §4: "Exert this land" / "Exert this creature" (Arena of
  // Glory, Oasis Ritualist). The auto-tapper never pays such a row; the
  // player activates it from this menu.
  exert?: boolean;
  // #3845: "Mill N cards" as a cost (Millikin's "{T}, Mill a card").
  // The auto-tapper never pays such a row. A library too short to pay
  // it makes the server stamp cant_activate, which greys the row.
  mill_cost?: number;
  // S32 (#352): a mana component of the activation cost — the Signet
  // cycle's "{1}, {T}", Cabal Coffers' "{2}, {T}". Advisory like
  // life_cost. The client sends every activate_mana_ability with
  // `auto_tap` (#2215), so the server taps the player's other sources
  // for whatever the floating pool is missing.
  mana_cost?: string;
  // #1191, #1190: ActivatedAbilityView.charged_mana_cost for a mana
  // ability — CR 605.1a makes a mana ability an activated ability, so
  // Boom Scholar's discount reaches Loot, the Pathfinder's "{G}, {T}"
  // exactly as it reaches a CR 602 ability, and the row says so
  // through the same field. Present whenever mana_cost is, equal to
  // it absent a modifier.
  charged_mana_cost?: string;
  // #743: true while the mana ability's activation condition is false
  // — Temple of the False God with four lands, Mox Opal without
  // metalcraft. Same flag and meaning as ActivatedAbilityView's.
  condition_unmet?: boolean;
  // #1183: an exhaust MANA ability ("Activate each exhaust ability
  // only once") this permanent has already used — Loot, the
  // Pathfinder's "Exhaust — {G}, {T}: Add three mana of any one
  // color". Same flag, same name and same meaning as
  // ActivatedAbilityView's, which is why `abilityBlocked` greys both
  // kinds of row with one predicate and one string. Absent for every
  // other mana ability, which is all of them.
  //
  // It recovers differently from `condition_unmet`: a condition may
  // hold again next turn, an exhaust only if the permanent becomes a
  // new object (CR 400.7 — a flicker, not an untap).
  exhausted?: boolean;
  // #844, CR 903.4f: the ability says "any color in your commander's
  // color identity" (Command Tower, Arcane Signet, Commander's Sphere,
  // Path of Ancestry) and the controller has no commander, or one
  // whose colour identity is colourless. The quality is undefined or
  // empty, so the ability adds no mana at all and the row is greyed —
  // tapping the land would just lose it. Absent for every other
  // ability.
  //
  // ADR 0117 §5 widens it to "adds nothing right now": a computed
  // output that comes to no mana (a power-0 Vivi Ornitier, an Exotic
  // Orchard with nothing to copy, CR 106.5 and 106.7). The client
  // greys the row either way, with the generic reason.
  adds_no_mana?: boolean;
  // #1210: ActivatedAbilityView.cant_activate for a mana ability (a
  // Cursed Totem reaches a mana creature's mana ability). Same key, so
  // ADR 0117 §2's one predicate reads both kinds of row.
  cant_activate?: string;
  // #1443: for each slot of the output that asks for a colour, the
  // colours that pick would offer — one list per picking slot, in
  // output order, narrowed (Command Tower, CR 903.4f) and ordered
  // (#843, commander identity first) by the SAME server function the
  // `mana_pick` prompt uses. A painland's "{R|W}" is [["R","W"]], a
  // filter land's "{W|U}{W|U}" two lists. Absent when nothing is
  // picked. The picker at the card offers these directly and sends the
  // answer up front as `activate_mana_ability`'s `color` / `colors`, so
  // nothing is tapped until the player has chosen.
  color_options?: string[][];
  // #2558: the ability adds "N mana of different colors" (Firemind
  // Vessel, "{W|U|B|R|G:2}"). Its N lists in color_options are one per
  // mana and the answer must name N DIFFERENT colours; the server
  // refuses a repeated one before anything is paid.
  different_colors?: boolean;
  // S32 (#352): spend restrictions the produced mana will carry —
  // Ancient Ziggurat's "only to cast a creature spell", Eldrazi
  // Temple's "only colorless Eldrazi". Informational; the server's
  // pool solver is what actually refuses an illegal payment.
  restrictions?: string[];
  // #789: the counter half of the activation cost — Vivid Creek's
  // "Remove a charge counter from this land", Ramos's five +1/+1
  // counters, Mage-Ring Network's "any number of storage counters".
  // The SAME field names an activated ability carries, and the same
  // meanings, because it is the same component: counterCost.ts reads
  // both through one structural type and builds one payload.
  counter_cost_n?: number;
  counter_cost_kind?: string;
  counter_cost_self?: boolean;
  counter_cost_label?: string;
  counter_cost_among?: boolean;
  counter_cost_variable?: boolean;
  counter_cost_max?: number;
  counter_cost_options?: CounterCostOptionView[];
  counter_cost_add?: number;
  counter_cost_add_kind?: string;
  counter_add_blocked?: boolean;
}

// AttackTargetView is one legal attack target: the id to send as
// declare_attacker's `target`, and what it is. `id` is a seat id for
// a player and an instance id for a permanent. Added in S27.
export interface AttackTargetView {
  kind: "player" | "planeswalker" | "battle";
  id: string;
  // ADR 0080 (#1063): the CR 508.1a price ONE creature pays to attack
  // this target — "{2}" against a seat with Propaganda out, "{2}{2}"
  // against one with Propaganda and Ghostly Prison, absent when
  // attacking it is free. A declaration's real price is this once per
  // attacking creature.
  //
  // READ IT, DON'T DERIVE IT. The server prices it for the active
  // seat through the same function the engine charges with; nothing
  // in the client re-derives what an attack costs, for the same
  // reason nothing re-derives who may attack (#429, ADR 0045 §6).
  tax?: string;
  // #1533 (ADR 0045 Decision 46): how many MORE creatures may be
  // declared attacking this target this combat under a CR 508.1c count
  // limit (Silent Arbiter, Crawlspace), after the ones already
  // attacking. Absent when no limit counts this target; 0 is a real
  // answer (the limit is used up), so test with `!== undefined`, never
  // truthiness. Engine-computed: n new attackers at this target are
  // accepted exactly when n <= attack_limit. Read it, don't derive it.
  attack_limit?: number;
}

// One step still to come this turn (TurnView.upcoming).
export interface PlannedStepView {
  step: string;
  phase_id: number;
}

export interface TurnView {
  seq: number;
  number: number;
  active_seat: number;
  // priority_holder is the seat index that currently holds priority
  // within the step. Equal to active_seat at every step boundary;
  // rotates on pass_priority. Added in S07.
  priority_holder: number;
  phase: string;
  step: string;
  // ADR 0059 Decision 11 (#753): this turn was created by an effect
  // (CR 500.7). `number` is still the round, so the board keeps "R3"
  // and adds an "Extra turn" mark. Absent on a normal turn.
  extra?: boolean;
  // Seat indices of queued extra turns, in the order they will be
  // taken (next first). Turns of players who have left are omitted.
  extra_turns?: number[];
  // ADR 0188 Decision 9 (#3566): the seat index whose turn begins after
  // this one, as the engine decides it: queued extra turns first, then
  // normal rotation, passing a seat that has left and spending skip
  // charges. Seat 0 is a real answer. Absent when the game is not in
  // play. It cannot see a replacement skip (Trouble in Pairs).
  next_seat?: number;
  // #3826: what the viewer's own next pass hands over. "ends_turn": it is
  // the viewer's turn, they hold priority on an empty stack and no later
  // step of the turn grants priority. "your_turn_next": the same in
  // another player's turn, with the viewer's turn next (next_seat counts
  // extra turns and skips). Absent for a routine pass. Per viewer; it
  // ignores the viewer's own auto-pass and stops.
  pass_hint?: "ends_turn" | "your_turn_next";
  // ADR 0188 Decision 12: set while the turn change waits on a
  // "turn_skip_order" question. The old turn is over and the next has
  // not begun; step still names the old turn's step, and nobody holds
  // priority.
  between_turns?: boolean;
  // ADR 0059 Decision 11 (#753): the phase in progress. `phase_id` is
  // unique within the turn (the template's phases are 1..5, an added
  // phase gets the next id); `phase_ordinal` is the nth phase of its
  // family this turn ("Combat 2"). Precombat and postcombat main are one
  // family. Both absent before the game's first phase is planned.
  phase_id?: number;
  phase_ordinal?: number;
  // The rest of this turn, in order: every step still to come, with the
  // phase it belongs to, including added phases and steps. Absent at
  // cleanup. `first_strike_damage` is listed in every combat and walked
  // through when no combatant has first or double strike (#717).
  upcoming?: PlannedStepView[];
  // #328: seat indices that owe a declare-blockers decision — under
  // attack, holding at least one creature that could legally block
  // one of the attackers. Absent outside the declare_blockers step.
  //
  // The server computes this because block legality is a rules
  // question (CR 509.1a untapped, CR 509.1b evasion) that the client
  // must not re-derive in TypeScript. It exists because blocking is a
  // turn-based action rather than a response, so the auto-pass
  // "legal response?" predicate structurally could not see it.
  //
  // #1279: a seat leaves this list once its block declaration is
  // COMPLETE (it passed, sent finish_blocks, or had no legal block as
  // the step began), even while it still has a creature that could
  // block.
  block_decision_seats?: number[];
  // #1279: where each DEFENDING seat's block declaration stands.
  // `block_pending_seats` — still declaring; `blocks_declared_seats` —
  // finished, with or without blocks. A defending seat is in exactly
  // one; a seat nothing is attacking is in neither. Both absent
  // outside declare_blockers. The "Done blocking" / "No blocks"
  // control shows while the viewer's seat is pending.
  block_pending_seats?: number[];
  blocks_declared_seats?: number[];
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

// TableSettingsView is a table's configuration (ADR 0075 §2.2). The
// keys match the settings patch the server accepts, so a client can
// send back exactly the fields it read.
// VoiceView is the table's Discord voice room on the game view (ADR 0159
// §6). `url` reaches only a viewer who was let in, while the room is
// ready.
export interface VoiceView {
  enabled: boolean;
  status: VoiceStatus;
  you?: VoiceAccess;
  url?: string;
  // #3311: seated viewers only. `connected` maps a seat's player id to
  // its state in the room; a seat not in it is not connected. `granted`
  // and `in_room` are the chip's "3 of 4".
  connected?: Record<string, VoiceConn>;
  granted?: number;
  in_room?: number;
}

export type VoiceConn = "live" | "muted";

export type VoiceStatus = "off" | "pending" | "ready" | "idle" | "full" | "unavailable";

export type VoiceAccess = "granted" | "not_in_server" | "no_discord" | "spectator" | "pending";

// RematchView is the host's offer of the same game again (#3285, ADR
// 0164). `game_id` is the new table, in the lobby until the host starts
// it. `open` stays true while at least one other human is still on the
// results screen without having answered. `you` is this viewer's part.
export interface RematchView {
  game_id: string;
  host: string;
  host_name?: string;
  open: boolean;
  you?: RematchPart;
}

// StackPositionView is the stack panel's top-left corner as a fraction of
// the board's width and height, each 0..1 (#3429, ADR 0176).
export interface StackPositionView {
  x: number;
  y: number;
}

export type RematchPart = "host" | "undecided" | "accepted" | "declined";

// ADR 0177 §1: PlayerView.presence.
export interface PresenceView {
  // "offline": no connection is bound to the seat. "hidden": every one
  // that is says its tab is hidden.
  state: "offline" | "hidden";
  // RFC 3339: when the seat stopped being present.
  since: string;
  // The table may move on without the seat (§2): offline for the away
  // time while the table waits on it, and nobody chose to keep waiting.
  movable?: boolean;
}

// ADR 0177 §4: PlayerView.stand_in.
export interface StandInView {
  // The bot tier playing the seat: "heuristic".
  tier: string;
  // RFC 3339: when the bot took the seat.
  since: string;
}

export interface TableSettingsView {
  // Per-player per-turn undo budget. -1 is unlimited, 0 is no undos.
  undo_limit: number;
  // Whose undo entries a seat may take back.
  undo_scope: "own" | "host_any";
  // Each seat's life at game start. Fixed once the game is active.
  starting_life: number;
  // Damage from one commander that loses the game.
  commander_damage: number;
  // The table pace (ADR 0143 §2.6): bot think time, and the stack hold
  // every seat waits before auto-passing another player's spell.
  bot_pace: "fast" | "normal" | "slow";
  // Whether the host and admin may spawn cards and tokens on a live
  // table (every spawn is announced in the log).
  allow_spawn: boolean;
  // Free mulligans, the house rule (#3387): 0 or absent is the rules
  // (one free in a multiplayer game, none at two players), N gives N
  // free at any table, -1 makes every mulligan free. Fixed once the
  // game is active.
  free_mulligans?: number;
  // ADR 0177 §7: minutes a seat the table waits on must be offline
  // before the table is offered a way on without it, 0 for never.
  // Absent from an older server, which reads as the default, 3.
  away_after?: number;
}

// ADR 0127: why a prompt with a standing answer is asked by hand.
// "controlled" (ADR 0167 §4): another player controls the chooser, so
// their standing answer does not apply.
export type AskedByHand =
  | "no_mana"
  | "empty_library"
  | "loop"
  | "undone"
  | "controlled"
  // #3016: a remembered trigger order whose batch changed, or whose
  // source is now hidden.
  | "changed"
  | "hidden";

/** ADR 0166 / 0167: one player making another's decisions right now. */
export interface PlayerControlView {
  controller: string;
  controlled: string;
  source?: string;
  source_name?: string;
  until: "resolution" | "turn" | "search";
}

/** ADR 0167 §8: a turn-long control waiting for the controlled player's next turn. */
export interface ScheduledPlayerControlView {
  controller: string;
  controlled: string;
  source_name?: string;
}

// ADR 0127 §3: one of a seat's standing answers, as the server holds it.
export interface AutoAnswerRuleView {
  key: string;
  answer: "always" | "never";
}

// ADR 0135 §4: one emerge candidate's price — the permanent's mana value,
// and the cast's mana cost with it sacrificed ("{1}{U}{U}").
export interface AltCostPriceView {
  mana_value: number;
  price: string;
}

// #3931: one harmonize candidate's price — the creature's power, and the
// cast's mana cost with it tapped ("{2}{R}{R}").
export interface AltCostTapPriceView {
  power: number;
  price: string;
}
