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
  // #1063: also carries the attack tax's whole price as a plain cost
  // string ("{2}{2}") when code === "attack_tax_unpaid" — a second
  // shape on the same key, not a second field, because the server's
  // ErrorPayload.Reason is a bare string on the wire either way.
  // #1507: and one of AttackRefusalReason when code === "illegal_attack".
  reason?: BlockRefusalReason | AttackRefusalReason | string;
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
  // CR 724.2 itself — the monarch's end-step draw, and the transfer
  // to whoever deals combat damage to them — so this field moves on
  // its own and the crown below follows it. The set_monarch action
  // stays as the way a card (or a table fixing the board) hands the
  // designation out in the first place. Added in S10.
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
  // Added in S11. Since ADR 0075 it mirrors `settings.undo_limit`:
  // -1 means unlimited, 0 means no undos.
  undo_limit?: number;
  // The table's settings (ADR 0075 §2.2). Public: every viewer,
  // spectators included, gets the same object. Added in S35 (#1032).
  settings?: TableSettingsView;
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
  // left lost at once, CR 104.4a — a draw).
  cause: "last_standing" | "effect" | "all_lost";
  // The object whose effect won, for an effect win.
  source?: string;
  source_name?: string;
}

// GameEndGateView is one "can't lose" / "can't win" gate on a seat
// (ADR 0057 Decision 7): where it comes from and what it stops.
export interface GameEndGateView {
  source?: string;
  source_name: string;
  cant_lose?: string[];
  cant_win?: boolean;
  // A gate a resolved spell granted until end of turn (Angel's Grace).
  this_turn?: boolean;
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
    | "special_action";
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
  // #1307: true on a cast or activation whose chosen targets include
  // an object on the stack — a counterspell, a Stifle, a redirect.
  // Smart autopass reads it to tell "can answer the thing on the
  // stack" from "has some instant". Absent on older servers, which
  // the client treats as a plain instant-speed move.
  targets_stack?: boolean;
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
  // ADR 0080 (#1063): a cost string the move charges that `params`
  // cannot name — today exactly one thing, the CR 508.1a attack tax
  // on a declare_attacker move ("{2}" for an attack into Propaganda).
  // A cast's mana is its printed cost and lives on the CardView; an
  // attack has no printed cost, so without this a consumer prices an
  // attack under Ghostly Prison exactly like a free one.
  mana?: string;
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
  // S30 random effects: the server-rendered public outcome of a die
  // roll or coin flip. The event text is already redacted and ready
  // for both the log and the attention strip.
  | "roll"
  | "flip"
  // #984: a player answered a "choose a ..." prompt out loud. The
  // chosen VALUE is `choice` on the first two ("G", "Elf"); a chosen
  // PLAYER is `target_seat`, like every other player in the log. All
  // three carry `card_id` — the card the answer was given for, and the
  // card whose later abilities read it back (CR 607.2d).
  | "choose_color"
  | "choose_type"
  | "choose_player"
  // A player answered an "as this enters, choose a card name" prompt
  // (CR 614.12): Pithing Needle, Phyrexian Revoker, Sorcerous
  // Spyglass. `choice` is the name as the player typed it, trimmed
  // and otherwise untouched — CR 201.2 admits any card name, so
  // there is no canonical spelling to report (#1210).
  | "choose_name"
  // #1214: a player answered one of the three resolution-time picks
  // (CR 608.2) — an opponent choosing from a revealed set, a seat
  // choosing among another player's permanents, a seat choosing N of
  // its own. `amount` is how many cards were chosen and is always
  // there; `card_id` names the one card for a single-card pick and is
  // absent for any other count; `target` is the card that ASKED.
  | "choose_cards"
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
  | "cycle"
  | "counters"
  | "scry"
  | "surveil"
  | "saga_chapter"
  | "class_level"
  // ADR 0075 §2.3: the host or the admin changed a table setting.
  // `label` is the setting's key ("undo_limit", "allow_spawn") and
  // `choice` its new value as text; `seat` is the host, or NoSeat when
  // the server admin made the change. The only kind that is about the
  // rules rather than about the game.
  | "settings"
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
  // #1199, ADR 0084: a permanent phased out or in (CR 702.26).
  // Narrated for `transform`'s reason and one more that is stronger
  // here — phasing out is not a zone change (CR 702.26d), so no
  // `zone` entry says it, and the board simply stops showing the
  // permanent, indistinguishable from a permanent that died unless
  // the log says which.
  | "phase_out"
  | "phase_in"
  // #1209, ADR 0082's 2026-09-23 amendment: a permanent that was
  // face up was turned face down (CR 708.2a). `card_id` is the
  // permanent; `target` is the object that did it (Ixidron, Cyber
  // Conversion). Names nobody — a CR 708.2 object has no name for
  // any viewer, its controller included — so `text` reads "a card"
  // and `card_id` is still present for the client to point at the
  // permanent on the board.
  | "turn_face_down";

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
  // Which combat damage step dealt a combat `damage` entry (CR 510.4).
  // Set only when that combat had a first-strike step; combat with no
  // first strike or double strike anywhere is untagged, so its
  // presence alone means there are two beats. Read it, don't derive it
  // from keywords (#187, ADR 0053 Decision 1).
  combat_step?: "first_strike" | "regular";
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
    // #1196, CR 115.7: "choose the new target for …" — Deflecting
    // Swat, Bolt Bend, Misdirection, Imp's Mischief. Carries the same
    // `pick_target` legal-target payload and is answered on the board
    // the same way; the KIND is what tells the server the answer
    // rewrites an item already on the stack. A prompt whose `min` is
    // 0 is a printed "you may", and the banner's Done button submits
    // the empty list to decline.
    | "retarget"
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
    // #804 CR 726: the loop breaker has fired and the repeating
    // ability's controller is asked how many more times it should
    // resolve. Answered with resolve_choice { iterations }, where 0
    // means "stop here" and leaves the table paused exactly where the
    // breaker put it. loop_count is how many times it has already
    // resolved this turn; loop_max_iterations is the ceiling the
    // engine will accept.
    | "loop_shortcut"
    // #764, CR 603.3c: a modal triggered ability's mode, chosen as
    // the ability is put on the stack. mode_options / mode_indexes /
    // mode_min / mode_max / mode_repeatable describe the offer.
    | "mode_pick"
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
  // #1311: populated for a "pay_unless" whose payment is a WATERBEND
  // cost ("Ward—Waterbend {4}", The Unagi of Kyoshi Island): the
  // chooser's untapped artifacts and creatures that may each pay {1}
  // of pay_cost's generic (CR 701.67a), and `max`, how many. The
  // "Pay" answer names them as `tap_ids` beside `apply: true`; the
  // rest is paid from the pool and auto-tap as usual. Absent for
  // every other pay-unless.
  tap_cost?: TapCostView;
  // S22: populated for kind "search_library" — how many of `options`
  // the searcher may take. The minimum is always zero, so the submit
  // button is live from the first render. Absent for every other
  // kind, and absent for non-chooser viewers, who are not told what
  // the search is for.
  search_max?: number;
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

  accept_label?: string;
  decline_label?: string;
  // #74: populated for kind "confirm" — the life the ACCEPT branch
  // charges (Sylvan Library's 4). Absent when the branch costs no
  // life. The label already says it; this is the number, for anything
  // that needs to reason about the price rather than print it.
  life_cost?: number;
  // #74: populated for kinds "choose_cards", "untap_choice",
  // "entry_reveal_from_hand" and #1214's three resolution-time picks
  // ("reveal_pick", "their_permanents", "own_permanents") — how few
  // and how many of `options` the chooser must pick. Absent for every
  // other kind, and (for the two kinds whose candidates are cards in a
  // hand — choose_cards and entry_reveal_from_hand) absent for
  // not told the size of a choice over someone else's hidden cards.
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
  // #804: populated for kind "loop_shortcut" — how many times the
  // repeating ability has already resolved this turn, and the largest
  // answer the engine accepts. `reason` carries "<card> — <ability>".
  loop_count?: number;
  loop_max_iterations?: number;
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
}

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
}

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
  owner: string;
  source_card_id: string;
  label?: string;
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
  // Table host (ADR 0075 §2.1), visible to every viewer. The host may
  // change table settings alongside the server admin.
  is_host?: boolean;
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
  // #500 (CR 305.2): how many lands this seat may play this turn,
  // and how many it already has. The engine REFUSES a land play past
  // the allowance, so a client should grey out the hand's lands when
  // lands_played_this_turn >= land_drops_per_turn rather than only
  // explain the rejection afterwards. land_drops_per_turn is the
  // EFFECTIVE allowance — a controlled Exploration or a one-turn
  // grant is already summed in. Normally 1 / 0.
  land_drops_per_turn?: number;
  lands_played_this_turn?: number;
  // S15: per-player mana pool. Each entry is an uppercase mana
  // letter ("W", "U", "B", "R", "G", "C") — order reflects
  // insertion order so the UI can highlight the most recent add.
  // Empties at every step boundary (CR 106.4), so this is absent
  // / empty in the common case outside an active cast sequence.
  mana_pool?: string[];
  // #623 (CR 114): the emblems this seat has, in creation order.
  // Absent for a seat with none, which is nearly every seat.
  //
  // Not a ZoneView — an emblem has no characteristics at all, so it
  // is not a card and there is nothing for the card renderer to draw.
  // The board shows these as chips beside the player identity, with
  // `text` as the hover. Public: every seat sees every emblem.
  emblems?: EmblemView[];
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
}

// One emblem (CR 114). `label` is the board name ("Elspeth, Sun's
// Champion emblem"); `text` is its printed ability, for the hover.
export interface EmblemView {
  instance_id: string;
  label: string;
  text: string;
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
  max: number;
  options: ModeOptionView[];
  // #764, CR 700.2d: "you may choose the same mode more than once"
  // (Mystic Confluence). The picker offers a count per option rather
  // than a toggle, and each occurrence is asked for its own targets.
  repeatable?: boolean;
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
  // S28: the cards that can pay the cost's card-shaped half — the
  // blue cards in your hand for Force of Will, the Islands you
  // control for Daze. The chosen instance ID rides back on cast_spell
  // as `alt_cost_ids`. Absent when the cost charges no cards (every
  // S22 keyword). #695: never present-and-empty any more — an offer
  // with nothing to pay it is not offered at all, for the same reason
  // one whose life half is unpayable is not.
  pay_options?: LegalTargetsView;
  // S28: the picker's prompt copy for `pay_options` ("a blue card").
  pay_label?: string;
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
}

// #1297: the pile an "Exile N cards from your …" cost reads — always
// the activator's own. The server stamps it with `exile_cost_n`.
export type ExileCostZone = "hand" | "graveyard";

// ActivatedAbilityView is one CR 602 activated ability on a
// battlefield permanent (S21 sub-PR 2). Public information, so it
// rides every viewer's snapshot; the client only offers the menu on
// permanents the viewer controls. `index` is what the
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
  discard_cost_n?: number;
  discard_cost_label?: string;
  discard_cost_options?: string[];
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
  // #1559: "with mana value X or less", X the announced X. Like
  // count_from_x, the server built this legal set before X was chosen,
  // so it is a superset: the picker drops every card whose
  // `mana_values` entry exceeds the X collected in the cost prompts,
  // and a card with no entry (an unreadable cost) meets no bound.
  mana_value_at_most_x?: boolean;
  mana_values?: Record<string, number>;
}

// TargetDifferenceView is a clause's set rule (#1559): `label`
// completes "those targets must …", and `keys` maps each legal card
// to the value no two picks may share.
export interface TargetDifferenceView {
  label: string;
  keys?: Record<string, string>;
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
  // Absent — every hand cast, every command-zone cast, and a
  // Gravecrawler whose graveyard permission carries no price — means
  // the printed cost is on the menu as usual. `castable_here` is one
  // bit and says only that a cast is possible from here; this is the
  // other half of the sentence, and the client must not infer it from
  // the shape of the offer list.
  alternative_cost_required?: boolean;
  // ADR 0073 (#664): the "you may pay an additional cost" offers this
  // card makes — kicker, multikicker, buyback. Rendered inside the
  // same picker the alternative costs open, because CR 601.2b
  // announces them together. Absent for nearly every card.
  optional_costs?: OptionalCostView[];
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
  // #746 (ADR 0048 addendum): the printed clauses of this card's own
  // cost modifiers whose price depends on its targets — Fireball's
  // "This spell costs {1} more to cast for each target beyond the
  // first", strive. The X picker opens before targeting and its
  // affordability readout is priced at one target, so it shows these
  // clauses under the readout. Absent for nearly every card and on
  // opponents' cards the viewer cannot read.
  target_cost_notes?: string[];
  // CR 107.4 (#916): how many symbols in the printed cost carry the
  // "or 2 life" option — 1 for Gitaxian Probe's "{U/P}", 2 for
  // Dismember's "{1}{B/P}{B/P}", 1 for a compleated planeswalker. The
  // cast flow opens a "pay N with life" stepper bounded by it and by
  // the caster's life total, and sends the answer as `phyrexian_life`.
  // Stamped with the other cast clauses on the viewer's own castable
  // cards and absent everywhere else, so its presence IS the question
  // "is there a life half to offer here".
  phyrexian_symbols?: number;
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
  // #1389: what THIS viewer would be charged to cast the card out of
  // EXILE right now — one entry per price the cast may claim, cheapest
  // first, each the total AFTER every CR 601.2f cost modifier (the
  // same pricer the cast path and the auto-tap preview use). Exile
  // only, and only on the frame of a seat holding a LIVE permission:
  // a warp or foretell grant whose later turn has not come carries
  // none. Read it through exileStrip.ts, which decides the badge.
  //
  // Since #1389 `castable_here` is stamped in exile too, with the same
  // meaning it has in a graveyard — "YOU may cast this from here NOW",
  // timing included — and is what the castable-from-exile strip
  // lights a card by.
  cast_prices?: CastPriceView[];
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
  // Regeneration shields on this permanent (CR 701.19a, #667). Each
  // one replaces the next destruction this turn: instead of dying it
  // is tapped, all damage is removed from it and it leaves combat.
  // Public state, like damage_marked; omitted when zero.
  regeneration_shields?: number;
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
  // Player ID who goaded this creature, or omitted when not goaded.
  // Cleared on zone exit. Added in S10; enforced server-side since
  // #1571 (CR 701.15b — attacks each combat if able, and a player
  // other than the goader if able).
  goaded_by?: string;
  // #1571 (ADR 0045 Decision 51): true on a creature the active player
  // owes an attack with right now — during declare_attackers, while the
  // declaration could still obey a CR 508.1d requirement it does not
  // (Zurgo, goad, Bident of Thassa). Server-computed; the client
  // renders it and never derives a requirement. Omitted otherwise.
  must_attack?: boolean;
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
  // #662 — this permanent's CR 702.16 protections, already PARSED by
  // the server. The raw "protection from red" tokens are in
  // `abilities` like every other keyword; this is the same list with
  // the quality pulled out, so the badge row renders "Protection from
  // Demons" without the client owning a copy of the grammar.
  protection?: ProtectionView[];
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

export interface ProtectionView {
  // The quality as the CARD prints it — "red", "Demons",
  // "artifacts", "everything". Badge tooltip text.
  printed: string;
  // Which characteristic of a source the quality is compared
  // against: "color", "card_type", "subtype", "everything" or
  // "player".
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
  value?: string;
}

// one entry per activated mana ability on a battlefield permanent.
// The client renders these as buttons in a right-click / long-press
// menu anchored to the card. Added in S15 sub-PR 2.
export interface ManaAbilityView {
  index: number;
  label?: string;
  // ADR 0093 Decision 5: the row's stable ref ("own:<i>",
  // "land:<colour>", "grant:<bundle>:<i>:<n>"), sent back as
  // activate_mana_ability's `ref`. See ActivatedAbilityView.ref.
  ref?: string;
  // ADR 0093 Decision 8: present on a mana ability another permanent
  // granted this one (Cryptolith Rite's "{T}: Add one mana of any
  // color." on every creature you control).
  granted_by?: GrantedByView;
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
  // S32 (#352): a mana component of the activation cost — the Signet
  // cycle's "{1}, {T}", Cabal Coffers' "{2}, {T}". Advisory like
  // life_cost. The server never auto-taps into a mana ability, so the
  // player has to float this mana before the entry will fire.
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
  adds_no_mana?: boolean;
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
export interface TableSettingsView {
  // Per-player per-turn undo budget. -1 is unlimited, 0 is no undos.
  undo_limit: number;
  // Whose undo entries a seat may take back.
  undo_scope: "own" | "host_any";
  // Each seat's life at game start. Fixed once the game is active.
  starting_life: number;
  // Damage from one commander that loses the game.
  commander_damage: number;
  // AI seat pacing preset.
  bot_pace: "fast" | "normal" | "slow";
  // Whether the host and admin may spawn cards and tokens on a live
  // table (every spawn is announced in the log).
  allow_spawn: boolean;
}
