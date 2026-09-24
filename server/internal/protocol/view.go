package protocol

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// view.go holds the wire-format projection of the server's
// authoritative game state. These types are what clients see; they
// are decoupled from the internal game package types so that the
// domain model can evolve without breaking the wire.
//
// The conversion functions (ViewOfGame, viewOfPlayer, etc.) are the
// single place where internal representation maps to the wire. When
// the game package grows new fields, they do NOT automatically show
// up on the wire until this file adds them — which is exactly the
// kind of explicit boundary we want.

// GameView is the complete JSON-serialisable snapshot of a Game.
type GameView struct {
	ID          string       `json:"id"`
	State       string       `json:"state"`
	Seats       []PlayerView `json:"seats"`
	Battlefield ZoneView     `json:"battlefield"`
	Stack       ZoneView     `json:"stack"`
	Exile       ZoneView     `json:"exile"`
	// PhasedOut is the CR 702.26 phased-out permanents (#1199,
	// ADR 0084). A shared, owner-less, public zone beside the other
	// three.
	//
	// They are a SEPARATE ZONE rather than battlefield entries
	// carrying a flag, and that is load-bearing in two directions.
	// stampNoUntap and stampCombatTargets index
	// view.Battlefield.Cards[i] positionally against
	// g.Battlefield.Cards[i], so merging would desynchronise every
	// positional stamp. And internal/aiseat never touches game.Game —
	// its sixteen battlefield walks are over this view — so a zone the
	// bot does not read is a bot that cannot count a phased-out
	// creature as a blocker, with nothing to teach and nothing to
	// forget. The same argument ADR 0084 Decision 1 makes for the
	// engine, one layer up.
	PhasedOut ZoneView `json:"phased_out"`
	Turn      TurnView `json:"turn"`
	// MulligansOpen reflects Game.MulligansOpen — true between Start
	// and the moment all seated players have committed to their
	// opening hand via the keep_hand action. Clients render the
	// keep / mulligan dialog while open. Added in S08.
	MulligansOpen bool `json:"mulligans_open"`
	// Monarch is the player ID currently designated as the monarch,
	// or empty string if no monarch is set. Since #375 the engine
	// moves this itself: CR 724.2's two inherent triggered abilities
	// (the end-step draw, and the transfer to whoever deals combat
	// damage to the monarch) are enforced server-side, so a client
	// that renders this field renders a crown that moves on its own.
	// Added in S10.
	Monarch string `json:"monarch,omitempty"`
	// Initiative is the player ID currently holding the initiative
	// (BG3 mechanic), or empty if unassigned. Same sandbox posture as
	// Monarch. Added in S10.
	Initiative string `json:"initiative,omitempty"`
	// Promises is the per-pair "I owe you" token tally as
	// "{from}->{to}" string keys → count. Zero entries are dropped on
	// the wire so the map stays small. Added in S10.
	Promises map[string]int `json:"promises,omitempty"`
	// Vote is the currently open vote (council's dilemma /
	// politics), or nil when no vote is in progress. Added in S10.
	Vote *VoteView `json:"vote,omitempty"`
	// UndoLimit is the per-player per-turn undo budget. Drives the
	// client's "undos remaining" indicator. Added in S11.
	//
	// Kept for wire compatibility since ADR 0075: it mirrors
	// Settings.UndoLimit, and -1 (game.UndoUnlimited) means no budget.
	// No omitempty any more: 0 is a real limit ("no undos") since the
	// Start-time rewrite of 0 to the default was removed, and a client
	// reading an absent key falls back to 1.
	UndoLimit int `json:"undo_limit"`
	// Settings is the table's configuration (ADR 0075 §2.2). Public on
	// purpose — every viewer, spectators included, sees the same
	// object, because undo rules and the spawn switch are things the
	// other players should be able to see. Added in S35 (#1032).
	Settings *TableSettingsView `json:"settings,omitempty"`
	// StartingSeat is the seat index that took the first turn. Used
	// by the server to enforce the CR 103.8a turn-1 skip-draw rule
	// (two-player games only — CR 103.8c);
	// surfaced on the wire so spectators / reconnects can render
	// "first player" UI affordances. Pre-S13 replays decode as 0,
	// which matches the only seat games started on before this field
	// existed. Added in S13.
	StartingSeat int `json:"starting_seat"`
	// StackItems is the announce-time metadata for every item
	// currently on the stack — caster, target list, modes, X,
	// distribution, hold-priority, split-second flags. Indexed in
	// stack order (bottom..top); the corresponding card (for spell
	// items) lives in the existing `stack` ZoneView. The dual
	// representation keeps existing zone plumbing intact while the
	// new cast/resolve UI reads StackItems for its rendering. Empty
	// when the stack is empty. Added in S13.1.
	StackItems []StackItemView `json:"stack_items,omitempty"`
	// PendingTriggers is the APNAP-ordered queue of triggered
	// abilities waiting to hit the stack (CR 603.3b). Drained on
	// next priority-grant boundary. Empty when no triggers are
	// pending. Added in S13.1.
	PendingTriggers []StackItemView `json:"pending_triggers,omitempty"`
	// DelayedTriggers is the queue of CR 603.7 delayed triggered
	// abilities still owed — "at the beginning of the next end step,
	// return that card to the battlefield". Public information: the
	// ability was announced when its source resolved, and the cards
	// it names sit in the shared exile zone, so no per-viewer
	// redaction applies. Empty when nothing is pending. Added in
	// S22.
	DelayedTriggers []DelayedTriggerView `json:"delayed_triggers,omitempty"`
	// SplitSecondActive mirrors `Game.SplitSecondActive` — true
	// while any item with split second is on the stack (CR 702.61).
	// Drives the client's "no responses allowed" UI gating. Added
	// in S13.1.
	SplitSecondActive bool `json:"split_second_active,omitempty"`
	// DiscardPending is the cleanup-step pause map (S13.4): keys
	// are player UUID strings, values are the count each player
	// must discard. Drives the client's discard-prompt modal.
	// Empty when nobody owes discard. Added in S13.4.
	DiscardPending map[string]int `json:"discard_pending,omitempty"`
	// PendingChoices is the S14 generic "someone needs to pick"
	// queue. Populated by effect cards that defer a decision
	// (Thoughtseize-style where the caster picks from the
	// target's revealed hand). Each entry has an ID, the chooser,
	// the discarder / source-zone owner, the count, a reason
	// label, and the candidate cards (redacted per viewer so
	// every client sees what its viewer is legally allowed to).
	// Drained by resolve_choice actions. Added in S14 sub-PR 5.
	PendingChoices []PendingChoiceView `json:"pending_choices,omitempty"`
	// LegalMoves is the closed list of things the VIEWER'S OWN seat
	// may do right now, straight out of internal/legal. Empty
	// whenever that seat owes no decision — which is most frames,
	// because the enumerator returns nothing for a seat that holds
	// neither priority nor a pending choice nor a combat
	// declaration.
	//
	// OWN SEAT ONLY, and that is a hard rule rather than a
	// nicety: an opponent's move list names the cards in their hand
	// they could cast, which is the whole of hidden information.
	// The field is therefore never populated by ViewOfGame — it is
	// projected out of the unexported legalBySeat map by
	// FilterViewFor, the same shape the knowers map uses, so the
	// unfiltered view that goes to the crash dump and the replay log
	// carries no seat's moves at all. Added in S31 sub-PR 2.
	LegalMoves []LegalMoveView `json:"legal_moves,omitempty"`
	// legalBySeat is the per-seat enumeration, keyed by player UUID
	// string. Unexported, so encoding/json never writes it: the only
	// way a move list reaches a client is through FilterViewFor
	// picking out that client's own key. Cleared by construction on
	// every filtered copy, which keeps repeated FilterViewFor calls
	// idempotent.
	legalBySeat map[string][]LegalMoveView
	// Log is the public game log: the last PublicLogMax table-visible
	// events, oldest first. A projection of game.Game.Events, not a
	// stored buffer — see log.go. Every card reference in it goes
	// through the same S13.5 knower redaction as every CardView, in
	// FilterViewFor. Added in S31 sub-PR 0 (ADR 0033 §4).
	Log []LogEvent `json:"log,omitempty"`
	// Reveals is the broadcast reveal window: the cards players have
	// shown the whole table this turn (CR 701.20), oldest first, at
	// most PublicRevealMax of them. The counterpart to the
	// controller-only look-at-cards prompt the scry family rides, and
	// the one field on this view that is identical for every seat —
	// FilterViewFor passes it through untouched, because a reveal one
	// seat could not see would not be a reveal.
	//
	// It carries printed identity and no instance IDs whatsoever, for
	// either the revealed cards or the card that revealed them. See
	// reveal_frame.go for why that is the design rather than an
	// omission. Added in S22.
	Reveals []RevealView `json:"reveals,omitempty"`
	// LoopNotice is the CR 726 loop breaker's flag: set when the
	// engine has seen the same triggered ability resolve
	// game.DefaultLoopThreshold times this turn with no player
	// decision in between, nil otherwise. Its presence is the
	// instruction to every client on the table: stop passing
	// AUTOMATICALLY. Priority still rotates, every pass_priority the
	// server is handed still works, and the loop's trigger is still
	// on the stack — the point is only that a person has to ask for
	// the next iteration. Public, like the stack it describes: a loop
	// is something the whole table can see running. Added for #628
	// (ADR 0055).
	LoopNotice *LoopNoticeView `json:"loop_notice,omitempty"`
}

// LoopNoticeView is the wire shape of game.LoopNotice. Label is the
// repeating ability's stack label, which by catalog convention reads
// "<card> — <what happens>", so a client has the whole banner line
// without resolving Source against the board.
type LoopNoticeView struct {
	Source     string `json:"source,omitempty"`
	Label      string `json:"label"`
	Controller string `json:"controller,omitempty"`
	Count      int    `json:"count"`
}

// viewOfLoopNotice projects the engine's loop notice, or nil.
func viewOfLoopNotice(n *game.LoopNotice) *LoopNoticeView {
	if n == nil {
		return nil
	}
	return &LoopNoticeView{
		Source:     uuidStringOrEmpty(n.Source),
		Label:      n.Label,
		Controller: uuidStringOrEmpty(n.Controller),
		Count:      n.Count,
	}
}

// LegalMoveView is one entry of the viewer's legal-move list. It is
// legal.Move verbatim rather than a parallel struct: the
// enumerator's JSON tags ARE the wire contract (ADR 0033 §1 — "the
// same list is what the client consumes"), and a copy here would be
// one more thing to keep in sync for no gain.
type LegalMoveView = legal.Move

// PendingChoiceView is the wire shape of a PendingChoice.
// Serialised per-viewer with Options pre-filtered to the cards
// the viewer is legally allowed to see. Chooser-viewer sees
// revealed cards in full; others see either backs (for
// discard_from_hand targeting an opponent's hand) or the raw
// IDs plus redacted characteristics.
type PendingChoiceView struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	Chooser    string     `json:"chooser"`
	FromPlayer string     `json:"from_player"`
	Count      int        `json:"count"`
	Source     string     `json:"source,omitempty"`
	Reason     string     `json:"reason,omitempty"`
	Options    []CardView `json:"options,omitempty"`
	// ColorOptions populates the S15 "mana_pick" kind: one entry per
	// legal color button the chooser's picker modal should render.
	// Uppercase single-character values ("W", "U", "B", "R", "G",
	// "C"). Absent for non-mana choices. Ordered server-side: the
	// commander's colour identity first, then the rest; only a source
	// whose printed text says "in your commander's color identity"
	// (Command Tower, Arcane Signet) is narrowed to it. Render in the
	// order given. Added in S15 sub-PR 2.
	//
	// A "choose_color" prompt reuses the field, and since #986 it is
	// ordered too — by the prompt's ColorPurpose, through the same
	// legal.OrderColorOptionsLocked the bot seat's move list is built
	// with, so the first button and the first offered move are the
	// same colour. Battlefield counts only, so the order is public and
	// identical for every viewer.
	ColorOptions []string `json:"color_options,omitempty"`

	// ColorAmounts populates a "mana_pick" that adds more than one mana
	// of the picked colour (#742: Gilded Lotus's "three mana of any one
	// color", Nyx Lotus's devotion) — colour letter to amount. A colour
	// missing from the map adds one; absent on every ordinary pick.
	ColorAmounts map[string]int `json:"color_amounts,omitempty"`

	// ColorPurpose populates the "choose_color" kind: what the card
	// asking will DO with the answer — "mana", "benefit", "harm",
	// "filter" or "protect" (game.ColorPurpose). Public information,
	// because it is a reading of the card's own printed text, and
	// carried because CR 105.4 makes all five colours legal so nothing
	// else on the prompt says which one the effect wants. Absent on
	// every other kind. Added by #780.
	ColorPurpose string `json:"color_purpose,omitempty"`

	// TypeOptions populates the S26 "choose_creature_type" kind: every
	// creature type the engine knows (CR 205.3m), for the picker to
	// filter. Materialised here from game.AllCreatureTypes rather than
	// stored on the PendingChoice, because the legal set is the same
	// for every such prompt and the server can always rebuild it.
	TypeOptions []string `json:"type_options,omitempty"`

	// NameOptions populates the #1210 "choose_card_name" kind: the
	// distinct card names visible in a PUBLIC zone right now — the
	// battlefield, every graveyard, the stack.
	//
	// A SUGGESTION LIST, never a legal set, and that is the whole
	// difference from TypeOptions above. CR 201.2 lets a player name
	// any card name at all, so the client renders this as a filter
	// list beside a free-text box and the server accepts whatever
	// comes back. Public zones only, because this goes to every
	// viewer the prompt reaches and a convenience list is not worth
	// a hidden-information leak.
	NameOptions []string `json:"name_options,omitempty"`

	// ReplacementOptions populates the S17 "replacement_order" kind:
	// one entry per applicable CR 614 replacement effect the
	// chooser is ordering. The client renders a drag-reorder list
	// with label + source-card context and returns the IDs in the
	// chosen order as an `order []string` payload. Absent for
	// non-replacement choices. Added in S17 sub-PR 2.
	ReplacementOptions []ReplacementOptionView `json:"replacement_options,omitempty"`

	// DamageAssignment populates the S18 "damage_assignment" kind:
	// the attacker's controller assigns combat damage across an
	// ordered blocker list (CR 510.1c). Absent for non-assignment
	// choices. Added in S18 sub-PR 3.
	DamageAssignment *DamageAssignmentView `json:"damage_assignment,omitempty"`

	// TriggerOptions populates the S19 "trigger_order" kind: one
	// entry per pending trigger the chooser is ordering (CR
	// 603.3b), with label + source card for context. The client
	// renders the same reorder list as replacement_order and
	// submits `order []string` of these IDs in RESOLUTION order
	// (first entry resolves first). Added in S19 sub-PR 8.
	TriggerOptions []ReplacementOptionView `json:"trigger_options,omitempty"`

	// PickTarget populates the S20 "pick_target" kind: the legal
	// players / cards a targeted trigger's controller may choose
	// from, computed when the trigger fired. The client enters its
	// board-click targeting flow with this set and answers with
	// resolve_choice {target: {kind, id}}. Added in S20 sub-PR 2.
	PickTarget *LegalTargetsView `json:"pick_target,omitempty"`

	// NoLegalTarget marks a "trigger_prompt" whose effect has no
	// legal target and will pass without effect if the chooser
	// answers "Yes" (Reclamation Sage with no opponent artifact,
	// Eternal Witness with an empty graveyard, etc.). The client
	// warns the chooser. Absent (false) for prompts with a legal
	// target or no target requirement. Added in S19 follow-up.
	NoLegalTarget bool `json:"no_legal_target,omitempty"`

	// PayCost populates the S19 "pay_unless" kind: the printed cost
	// the chooser is being asked to pay ("{2}"). `{apply: true}`
	// pays, `{apply: false}` declines. Absent for other kinds.
	// Added in S19 sub-PR 6.
	PayCost string `json:"pay_cost,omitempty"`

	// TapCost is the waterbend clause of a "pay_unless" whose payment
	// is a waterbend cost (#1311, "Ward—Waterbend {4}"): the untapped
	// artifacts and creatures the CHOOSER could tap, each paying {1}
	// of PayCost's generic, and Max, how many. The same TapCostView a
	// hand card's convoke / waterbend ships. `{apply: true, tap_ids:
	// [...]}` pays with those taps plus mana for the rest. Absent for
	// every other pay-unless.
	TapCost *TapCostView `json:"tap_cost,omitempty"`

	// AcceptLabel / DeclineLabel populate the "confirm" kind: the
	// card's own words for the two branches ("Pay 4 life" / "Put it on
	// top"). Absent means the client renders Yes / No, which is what a
	// prompt that really is a yes/no wants. Added with the chained
	// choice queue (#74).
	AcceptLabel  string `json:"accept_label,omitempty"`
	DeclineLabel string `json:"decline_label,omitempty"`

	// LifeCost is the life a "confirm" prompt's ACCEPT branch charges
	// (Sylvan Library's 4). Zero for a branch that costs no life.
	// Carried for the same reason legal.MoveCost.Life is (#547): a
	// client — human or bot — holding only this payload would
	// otherwise price "pay 4 life to keep it" like "shuffle your
	// library".
	LifeCost int `json:"life_cost,omitempty"`

	// Coin-call answers and the public progress of a flip chain.
	AllowStop     bool `json:"allow_stop,omitempty"`
	Coins         int  `json:"coins,omitempty"`
	MaxUsefulWins int  `json:"max_useful_wins,omitempty"`
	Wins          int  `json:"wins,omitempty"`

	// ChooseMin / ChooseMax populate the "choose_cards" kind: how few
	// and how many of Options the chooser must pick. Both are sent —
	// including a zero Min, which is why the client reads Max to tell
	// this kind's grid from a fixed-count discard. Added with the
	// chained choice queue (#74).
	ChooseMin int `json:"choose_min,omitempty"`
	ChooseMax int `json:"choose_max,omitempty"`

	// SearchMax populates the S22 "search_library" kind: how many of
	// Options the searcher may take. The minimum is always zero —
	// CR 701.23b permits failing to find — so the client's submit
	// button is live from the first render. Absent for other kinds.
	SearchMax int `json:"search_max,omitempty"`
	// Placement populates the ADR 0088 "put_in_library" kind: which
	// lanes the answer may use — "top" ({top_order} alone), "bottom"
	// ({bottom} alone, top-first) or "top_or_bottom" (both). Absent
	// for other kinds.
	Placement string `json:"placement,omitempty"`
	// LoopCount / LoopMaxIterations populate the #804 "loop_shortcut"
	// kind (CR 726): how many times the repeating ability has already
	// resolved this turn, and the ceiling the engine will accept on
	// the answer. The client renders a number field between 0 and the
	// max; `reason` carries "<card> — <ability>". Answered with
	// `{choice_id, iterations}`, where 0 means "stop here".
	LoopCount         int `json:"loop_count,omitempty"`
	LoopMaxIterations int `json:"loop_max_iterations,omitempty"`
	// PickOptions populates the #568 "option_pick" kind: one entry
	// per branch of "choose one of the following", in the card's
	// printed order. The client renders a button per option and
	// answers with `{option_index: N}` — the INDEX, not a card list,
	// because an option is a consequence and not always a set of
	// cards. Each option's own Cards ride through the same per-viewer
	// redaction as Options above. Absent for other kinds.
	PickOptions []PickOptionView `json:"pick_options,omitempty"`

	// ModeOptions / ModeIndexes / ModeMin / ModeMax / ModeRepeatable
	// populate the #764 "mode_pick" kind (CR 603.3c): the bullets a
	// modal TRIGGER offers its controller as it goes on the stack,
	// how many to choose, and whether the same one may be chosen
	// more than once (CR 700.2d). Only the choosable bullets are
	// listed — an option whose target clause has no legal target is
	// dropped before the prompt is queued — so mode_indexes carries
	// the ModeSpec index each label belongs to, which is what the
	// answer sends back. Answered with `{choice_id, modes: [i, …]}`.
	ModeOptions    []string `json:"mode_options,omitempty"`
	ModeIndexes    []int    `json:"mode_indexes,omitempty"`
	ModeMin        int      `json:"mode_min,omitempty"`
	ModeMax        int      `json:"mode_max,omitempty"`
	ModeRepeatable bool     `json:"mode_repeatable,omitempty"`
	// DoubledBy / DoubledByName identify the public permanent that
	// caused this additional trigger (CR 603.2d). They are
	// present only on trigger_prompt and pick_target choices.
	DoubledBy     string `json:"doubled_by,omitempty"`
	DoubledByName string `json:"doubled_by_name,omitempty"`
}

// PickOptionView is one branch of an "option_pick" prompt (#568):
// the card's own words for it, the cards it is about (a Fact or
// Fiction pile, or nothing at all), and the life it charges.
//
// Cards is projected per viewer exactly like PendingChoiceView.Options
// — see redactChoiceCards. An option over a pool the chooser does not
// own shows that chooser only the cards they are entitled to see,
// which on every printed card of this family means the revealed ones.
type PickOptionView struct {
	Label    string     `json:"label"`
	Cards    []CardView `json:"cards,omitempty"`
	LifeCost int        `json:"life_cost,omitempty"`
	// Player is the seat this option is about, for the prompts whose
	// branches ARE players — "choose a player", "choose an opponent"
	// (#929) and True-Name Nemesis's as-enters sibling (#980).
	// Absent on every other option, which is all of them.
	//
	// Public by CR 400.2: who is seated is not hidden information, so
	// this rides no redaction. The client may render a seat chip with
	// it instead of only the rendered Label — and may do nothing with
	// it, which is what it does today. #994.
	Player string `json:"player,omitempty"`
}

// LegalTargetsView is the wire shape of game.LegalTargets: player
// UUID strings and card instance-ID strings the spell's target slot
// accepts right now. Added in S20 sub-PR 1.
type LegalTargetsView struct {
	Players []string `json:"players,omitempty"`
	Cards   []string `json:"cards,omitempty"`
	// Min / Max are the clause's target count (S20 sub-PR 5): the
	// client's picker completes on the first click at 1 / 1, and
	// otherwise accumulates picks until the player confirms with at
	// least Min. Max 0 means unbounded.
	Min int `json:"min"`
	Max int `json:"max"`
	// CountFromX marks a clause whose count is the announced X
	// rather than a printed constant — "Exile X target creatures you
	// control". Min / Max are meaningless while X is unknown, so the
	// client substitutes the X it collected in the cost prompts.
	// Added in S22 alongside convoke / waterbend.
	CountFromX bool `json:"count_from_x,omitempty"`

	// Label is the clause's printed wording — "target creature you
	// control". Absent for a single-clause statement, where the
	// banner reads the card's target_mode as it always has; present
	// on every entry of `clauses`, where the picker has to say which
	// of several questions it is asking. Added by #764.
	Label string `json:"label,omitempty"`

	// Distinct marks a clause whose picks must differ from every
	// EARLIER clause's ("a second target permanent you control"), so
	// the picker can grey what is already taken. Added by #764.
	Distinct bool `json:"distinct,omitempty"`
}

// clausesView projects every clause of a multi-clause statement,
// each with its own legal set, count and label (#764, ADR 0065 §7).
// Nil for the single-clause statement that is nearly every card —
// `legal_targets` alone says everything there, and every client path
// that predates #764 keeps reading it.

// ModeSpecView / ModeOptionView are the wire shape of game.ModeSpec
// for the owner's hand cards. Added in S20 sub-PR 4.
type ModeSpecView struct {
	Prompt  string           `json:"prompt"`
	Min     int              `json:"min"`
	Max     int              `json:"max"`
	Options []ModeOptionView `json:"options"`
	// Repeatable is CR 700.2d, "you may choose the same mode more
	// than once" (Mystic Confluence). The picker offers a count per
	// option instead of a toggle, and each occurrence is asked for
	// its own targets. Added by #764.
	Repeatable bool `json:"repeatable,omitempty"`
}

type ModeOptionView struct {
	Label string `json:"label"`
	// TargetMode / LegalTargets mirror CardView.target_mode /
	// legal_targets for the option's FIRST target clause; both absent
	// for untargeted options.
	//
	// And they mirror their per-viewer scope too (#1172). `modes` is
	// public on a public pile — it is the card's printed text — but
	// `legal_targets` inside it is the ASKING SEAT's answer, for the
	// reason the card's own is: hexproof, shroud, protection and
	// "target opponent" narrow a target set by who is asking. It rides
	// castStamps into one seat's frame and is absent for everybody
	// else, bystanders and spectators included, while `target_mode` —
	// the printed prompt shape — stays public beside `label`.
	TargetMode   string            `json:"target_mode,omitempty"`
	LegalTargets *LegalTargetsView `json:"legal_targets,omitempty"`
	// Clauses is every clause of the option's statement when it has
	// more than one, in printed order. Absent for the one-clause
	// bullet that is nearly every bullet. Added by #764.
	//
	// Per viewer with `legal_targets`, and for the same reason
	// (#1172): each entry is a legal set.
	Clauses []LegalTargetsView `json:"clauses,omitempty"`
	// Cost is CR 702.172a's Spree: this bullet's own additional mana
	// cost, in brace notation, paid only if it is chosen — on top of
	// the card's printed cost and every OTHER chosen bullet's. Empty
	// for an ordinary modal bullet, which is every modal card before
	// S45. Public alongside `label`: it is the printed clause, the
	// same reason `target_mode` is public while `legal_targets` is
	// not. Added by ADR 0065's 2026-09-23 amendment.
	Cost string `json:"cost,omitempty"`
}

// AdditionalCostView is the wire shape of game.AdditionalCost — the
// "As an additional cost to cast this spell, discard a card" clause
// on a card in the viewer's own hand. The client opens its cost
// picker on this before the mode / X / target prompts, because
// that's the order the costs are actually locked in (CR 601.2f
// precedes 601.2h). Added in S21 sub-PR 5.
type AdditionalCostView struct {
	// DiscardCards is how many cards the caster must discard. The
	// picked instance IDs ride back on cast_spell's discard_ids.
	DiscardCards int `json:"discard_cards,omitempty"`
	// SacrificeOptions lists the permanents that may pay a
	// "sacrifice a creature" clause, already filtered to the
	// caster's own board (CR 701.21a). Its min / max are how many
	// the clause sacrifices ("sacrifice two creatures" is 2 / 2,
	// #747), and its cards come in payment order — see
	// sacrificeCostOptions. The picked instance IDs ride back on
	// cast_spell's sacrifice_ids. Absent when the cost has
	// no sacrifice component; present-and-empty means the cost is
	// unpayable, which makes the spell uncastable.
	SacrificeOptions *LegalTargetsView `json:"sacrifice_options,omitempty"`
	// DemandsX marks a "pay X life" clause (Toxic Deluge). The
	// client must open its X prompt for this card even though the
	// printed mana cost has no {X} in it, and the announced X is
	// both the life paid and the number the spell's own text uses.
	// Added in S23.
	DemandsX bool `json:"demands_x,omitempty"`
	// Label is the clause as printed ("Discard a card"), shown
	// above the picker.
	Label string `json:"label,omitempty"`
}

// OptionalCostView is the wire shape of one optional additional cost
// — kicker, multikicker, buyback (CR 601.2b, ADR 0073) — on a card
// the viewer could cast.
//
// Like AlternativeCostView it is an OFFER rather than a demand, and
// unlike it the offers COMPOSE: a cast may claim one alternative cost
// and any number of these, which is why the client renders them as
// toggles beside the alternative-cost radio list rather than as a
// fourth modal in the chain. They are the same question asked at the
// same moment.
//
// `index` is what rides back on cast_spell as an entry in
// `optional_costs`, repeated once per payment for a repeatable one.
// It is a POSITION and not a key, because that is what the engine's
// paid record holds; `key` is here for the client's own labelling.
type OptionalCostView struct {
	// Index is the cost's position in the card's OptionalCosts slice
	// — the value the announcement names.
	Index int `json:"index"`
	// Key is "kicker", "multikicker" or "buyback".
	Key string `json:"key"`
	// Label is the clause as printed ("Kicker {4}").
	Label string `json:"label,omitempty"`
	// ManaCost is the mana half, in brace notation. Empty for a
	// purely non-mana cost (Constant Mists' "Buyback—Sacrifice a
	// land").
	ManaCost string `json:"mana_cost,omitempty"`
	// MaxTimes is how many times this cost may be paid for one cast:
	// 1 for kicker and buyback, the multikicker cap for multikicker.
	// The client renders a checkbox at 1 and a stepper above it.
	MaxTimes int `json:"max_times,omitempty"`
	// DiscardCards and SacrificeOptions are the card-shaped halves,
	// in the same shape and with the same meaning AdditionalCostView
	// gives them — present-and-empty SacrificeOptions means this
	// offer cannot be taken right now.
	DiscardCards     int               `json:"discard_cards,omitempty"`
	SacrificeOptions *LegalTargetsView `json:"sacrifice_options,omitempty"`

	// ChoosesOpponent marks gift's cost (CR 702.174a, ADR 0089): taking
	// this offer means naming one of OpponentOptions, sent back on
	// cast_spell as `gift_opponent`. OpponentOptions is who may be
	// chosen right now — every other player still in the game. On a
	// gift offer an empty list (which the wire omits) means nobody is
	// left to promise it to and the offer cannot be taken.
	ChoosesOpponent bool     `json:"chooses_opponent,omitempty"`
	OpponentOptions []string `json:"opponent_options,omitempty"`

	// TargetMode / LegalTargets / Clauses are the target clause the
	// spell has WHEN THIS COST IS PAID — a gift's "if the gift was
	// promised, instead … target …" (CR 702.174m) — in exactly the
	// shape AlternativeCostView gives cleave. All absent when paying
	// the cost leaves the card's own clause alone, which is every
	// kicker and buyback. Per viewer, like every legal set here.
	TargetMode   string             `json:"target_mode,omitempty"`
	LegalTargets *LegalTargetsView  `json:"legal_targets,omitempty"`
	Clauses      []LegalTargetsView `json:"clauses,omitempty"`
}

// AlternativeCostView is the wire shape of one game.AlternativeCost
// — "you may cast this spell for its overload / evoke / cleave cost"
// — on a card in the viewer's own hand or command zone. Added in
// S22.
//
// Unlike AdditionalCostView this is an OFFER, not a demand: the
// client shows a picker with "pay the printed cost" alongside each
// entry, and a cast that names none is the ordinary case. `key` is
// what rides back on cast_spell as `alternative_cost`.
type AlternativeCostView struct {
	// Key is the stable identifier the cast names to claim this
	// cost.
	Key string `json:"key"`
	// Label is the clause as printed ("Overload {4}{R}").
	Label string `json:"label,omitempty"`
	// ManaCost is what this cost charges, in brace notation. Empty
	// means free.
	ManaCost string `json:"mana_cost,omitempty"`
	// TargetMode / LegalTargets describe the target clause the spell
	// has WHEN THIS COST IS PAID, already resolved against the
	// alternative cost's rewrite — cleave's wider clause, or the
	// card's own when the cost leaves it alone. Both absent when
	// paying the cost leaves the spell with no targets at all
	// (overload), which is why the client can key off their presence
	// rather than reasoning about the rewrite itself.
	//
	// `legal_targets` is PER VIEWER (#1172), like every other legal
	// set on this surface: the offer itself is a printed price and
	// stays public on a public pile, but the clause it leaves the
	// spell with is resolved against the board for the seat the stamp
	// was built for. `target_mode` is the printed shape of that clause
	// and stays public with the rest of the offer.
	TargetMode   string            `json:"target_mode,omitempty"`
	LegalTargets *LegalTargetsView `json:"legal_targets,omitempty"`

	// Life is the "pay N life" half of the cost (Force of Will's 1,
	// Snuff Out's 4). Zero — absent — for the costs that charge none.
	// The server enforces the life total; this is for the label.
	Life int `json:"life,omitempty"`

	// PayOptions is the set of cards that can pay the cost's
	// card-shaped half: the blue cards in the caster's hand for Force
	// of Will, the Islands they control for Daze. The chosen instance
	// ID rides back on cast_spell as `alt_cost_ids`.
	//
	// Absent when the cost charges no cards, which is every S22
	// keyword. Present-and-empty means the caster has nothing that
	// can pay — the offer is visible but unusable, which is the
	// honest thing to show for a Force of Will held with no other
	// blue card.
	//
	// Like the additional cost's sacrifice_options this is NOT a
	// target list: a cost does not target (CR 601.2h), so hexproof
	// and shroud never narrow it.
	//
	// PER VIEWER all the same (#1172). The list is picked out by "you
	// control" / "your hand" / "your graveyard" predicates, so it
	// answers "what may YOU pay" and is the asking seat's answer even
	// though nothing about targeting narrowed it. #1169 said as much
	// about a hand, where the whole offer list is dropped for a viewer
	// who was merely shown one card; this is the same sentence on a
	// public pile, where the offer stays and the list does not.
	PayOptions *LegalTargetsView `json:"pay_options,omitempty"`

	// PayLabel is the picker's prompt copy for PayOptions ("a blue
	// card", "an Island you control"). Absent when there is nothing
	// to pick.
	PayLabel string `json:"pay_label,omitempty"`

	// XLockedAtZero is CR 107.3b for THIS offer: the card prints an
	// {X} in its mana cost and this cost does not, so claiming it
	// fixes X at 0 and the client must not open its X picker. Absent
	// — the overwhelming majority — means X is announced as usual.
	//
	// Server-computed (game.CastCost.LocksXAtZero) rather than
	// re-derived from `mana_cost` on the client, so the rule has one
	// statement and the picker cannot disagree with the announce
	// gate that would reject what it collected. Added for #831.
	XLockedAtZero bool `json:"x_locked_at_zero,omitempty"`

	// PhyrexianSymbols is CR 107.4's "or 2 life" count for THIS
	// offer's cost, the same field CardView carries for the printed
	// one (#916): claiming the offer replaces the mana cost, so it
	// replaces the ceiling on `phyrexian_life` too. Absent for every
	// offer that prints no Phyrexian symbol, which is all of them
	// today — shipped because the field the client reads must not
	// depend on which cost is being paid.
	PhyrexianSymbols int `json:"phyrexian_symbols,omitempty"`
}

// TapCostView is the wire shape of game.TapPermanentsCost — the
// "tap permanents you control to help pay for this" clause that
// convoke and waterbend share — on a card in the viewer's own hand
// or command zone. Added in S22.
//
// Like an alternative cost this is an OFFER, not a demand: "you MAY
// tap any number", and a cast that taps nothing and pays the whole
// cost with mana is always legal. Unlike one, it does not replace
// anything — it spends against a cost that is still owed. The picked
// instance IDs ride back on cast_spell as `tap_ids`.
type TapCostView struct {
	// Key names the mechanic on the wire ("convoke", "waterbend").
	Key string `json:"key"`
	// Label is the clause as printed ("Convoke", "Waterbend {X}").
	Label string `json:"label,omitempty"`
	// Options are the untapped permanents that may be tapped, already
	// filtered to the viewer's own board. Present-and-empty means
	// there is nothing to tap, which is not an error — the caster
	// pays the whole cost with mana.
	Options *LegalTargetsView `json:"options,omitempty"`
	// ColorClause is convoke's "or one mana of that creature's
	// color". The client shows it in the picker's hint so a player
	// can see why tapping a white creature at a {W} is worth more
	// than tapping a colorless one. Absent for waterbend, where every
	// permanent pays exactly {1}.
	ColorClause bool `json:"color_clause,omitempty"`
	// Max is how many permanents may be tapped: the mana value of
	// what the cast owes. 0 means "derive it from the announced X" —
	// a waterbend {X} cost, whose size the caster has not chosen yet
	// when this snapshot is built.
	Max int `json:"max,omitempty"`
	// DemandsX marks a keyword cost that carries its own {X}
	// (waterbend {X}) on a card whose PRINTED cost has none.
	// Waterbender's Restoration costs {U}{U} and still needs the X
	// prompt; without this the client would never open it.
	DemandsX bool `json:"demands_x,omitempty"`
}

// DamageAssignmentView is the wire shape of the CR 510.1c
// multi-blocker damage-assignment prompt. The attacker's
// controller orders the blockers and assigns damage across them
// (with at-least-lethal-in-order enforcement); with trample, the
// leftover spills to the defending player. Added in S18 sub-PR 3.
type DamageAssignmentView struct {
	AttackerCardID string   `json:"attacker_card_id"`
	BlockerCardIDs []string `json:"blocker_card_ids"`
	AttackerPower  int      `json:"attacker_power"`
	AllowTrample   bool     `json:"allow_trample,omitempty"`
	HasDeathtouch  bool     `json:"has_deathtouch,omitempty"`
}

// ReplacementOptionView is one entry in a PendingChoiceView's
// ReplacementOptions slice — the wire shape of one CR 616 order-
// prompt candidate. ID is the server-side ReplacementEffectID
// serialised as a decimal string so JSON round-trips cleanly;
// Label is the prompt copy ("Doubling Season: double counters");
// SourceCardID points at the card hosting the effect (empty for
// engine built-ins like commander-zone). Added in S17 sub-PR 2.
type ReplacementOptionView struct {
	ID           string `json:"id"`
	Label        string `json:"label,omitempty"`
	SourceCardID string `json:"source_card_id,omitempty"`
}

// StackItemView is the wire shape of a stack-item's announce-time
// metadata. Mirrors `game.StackItem` with UUIDs serialised as
// strings. See server/internal/game/stack.go for field semantics.
type StackItemView struct {
	ID           string          `json:"id"`
	Kind         string          `json:"kind"`
	Controller   string          `json:"controller"`
	Owner        string          `json:"owner"`
	SourceCardID string          `json:"source_card_id"`
	Label        string          `json:"label,omitempty"`
	Targets      []TargetRefView `json:"targets,omitempty"`
	Modes        []int           `json:"modes,omitempty"`
	// ModeLabels is the oracle bullet of each chosen mode, in
	// announce order and with repeats — what "modes: 0, 2" used to
	// make the table guess. The caster's hand card is gone by the
	// time the spell is on the stack, so the labels have to travel
	// with the item. Added by #764 (ADR 0065 §7).
	ModeLabels   []string       `json:"mode_labels,omitempty"`
	XValue       int            `json:"x_value,omitempty"`
	Distribution map[string]int `json:"distribution,omitempty"`
	HoldPriority bool           `json:"hold_priority,omitempty"`
	SplitSecond  bool           `json:"split_second,omitempty"`
	// AltCost is the key of the alternative cost this spell was cast
	// for — "overload", "evoke", "cleave" — empty for an ordinary
	// cast (S22). Public information the moment it is announced, and
	// load-bearing for the table: an overloaded Cyclonic Rift is a
	// one-sided board wipe and a hard-cast one bounces a single
	// permanent, so a responder needs to see which is on the stack.
	AltCost string `json:"alt_cost,omitempty"`

	// GiftTo is the player a gift was promised to (CR 702.174a/k, ADR
	// 0089) — absent when the spell promised none. Public: the promise
	// is made aloud as the spell is cast, and it is load-bearing for a
	// responder, because a promised Long River's Pull counters any
	// spell and an unpromised one only a creature spell.
	GiftTo string `json:"gift_to,omitempty"`

	// IsCopy marks a CR 707.10 spell copy — Reverberate's output,
	// not a cast card (S30). Public and worth showing: the copy and
	// the spell it came from are two identical-looking entries on
	// the stack, and which one is the copy decides what a responder
	// gets by countering it (countering the copy leaves the
	// original; countering the original leaves the copy, because a
	// copy is independent of its source once created).
	IsCopy bool `json:"is_copy,omitempty"`
	// DoubledBy / DoubledByName identify the public permanent whose
	// effect caused this additional trigger (CR 603.2d).
	DoubledBy     string `json:"doubled_by,omitempty"`
	DoubledByName string `json:"doubled_by_name,omitempty"`

	// ManaSpent / ColorsSpent are what paid for this spell (#761):
	// how many mana, and the distinct COLOURS among them in WUBRG
	// order. Public information — mana is spent face up — and
	// load-bearing for the table: a responder looking at a converge
	// spell on the stack needs to see how wide it converged before
	// deciding whether to answer it.
	//
	// ManaSpentUnknown is PaidCost.OnPaper: the cast went through
	// permissive mode or a strict-mode override, so the engine did
	// not take the mana and has no record of what it was. The client
	// greys the pill rather than rendering "0 mana", which would be
	// a claim nobody made. Absent on an ability item and on a copy
	// (CR 707.10 — nothing was spent to cast a copy, and that zero
	// is real).
	ManaSpent        int      `json:"mana_spent,omitempty"`
	ColorsSpent      []string `json:"colors_spent,omitempty"`
	ManaSpentUnknown bool     `json:"mana_spent_unknown,omitempty"`
}

// DelayedTriggerView is the wire shape of one queued CR 603.7
// delayed triggered ability. Mirrors `game.DelayedTrigger` with
// UUIDs serialised as strings; the Effect closure has no wire form
// (Label is what the client renders). See
// server/internal/game/delayed.go. Added in S22.
type DelayedTriggerView struct {
	ID          string   `json:"id"`
	Controller  string   `json:"controller"`
	Source      string   `json:"source,omitempty"`
	Label       string   `json:"label,omitempty"`
	At          string   `json:"at"`
	CreatedTurn int      `json:"created_turn,omitempty"`
	Cards       []string `json:"cards,omitempty"`
	// On is the #663 EVENT condition — "when you next cast an
	// instant or sorcery spell this turn". Present instead of a
	// meaningful `at` for such a trigger, so a client can tell "owed
	// at a step" from "owed on the next matching event" rather than
	// rendering an empty step.
	On []string `json:"on,omitempty"`
}

// TargetRefView is the wire shape of a single announce-time target
// slot. Kind is one of "player", "card", "self", "none"; ID is the
// referenced UUID (zero string for "self" / "none"). See
// server/internal/game/stack.go for the canonical taxonomy.
type TargetRefView struct {
	Kind string `json:"kind"`
	ID   string `json:"id,omitempty"`
	// Slot / Mode name the target CLAUSE this pick answered: the
	// clause index within the clause list, and the index into the
	// item's `modes` whose clause list that is. Both omitted at zero,
	// which is what every single-clause non-modal announcement has
	// always meant. Added by #764 (ADR 0065 §2).
	Slot int `json:"slot,omitempty"`
	Mode int `json:"mode,omitempty"`
}

// VoteView is the wire form of game.Vote. Ballots is keyed by voter
// player ID (string UUID) for trivial JSON serialisation.
type VoteView struct {
	ID        string         `json:"id"`
	Topic     string         `json:"topic"`
	Options   []string       `json:"options"`
	Initiator string         `json:"initiator"`
	Ballots   map[string]int `json:"ballots"`
}

// PlayerView is the wire representation of a Player. Full-fidelity
// views come out of ViewOfGame; FilterViewFor then zeroes out any
// zones that should be hidden from a specific viewer (opponent hand
// cards, opponent library cards) while preserving the `count` so the
// UI can still render a placeholder stack.
type PlayerView struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Seat      int      `json:"seat"`
	Life      int      `json:"life"`
	Poison    int      `json:"poison,omitempty"`
	Energy    int      `json:"energy,omitempty"`
	Library   ZoneView `json:"library"`
	Hand      ZoneView `json:"hand"`
	Graveyard ZoneView `json:"graveyard"`
	Command   ZoneView `json:"command"`
	// CommanderDamage maps commander card instance ID → total damage
	// that commander has dealt to this player (CR 903.10a). Keyed by
	// COMMANDER, not by opposing player, since S25 (#77) — which is
	// the shape the client's per-commander hover rows were already
	// written against.
	CommanderDamage map[string]int   `json:"commander_damage"`
	LifeHistory     []LifeChangeView `json:"life_history"`
	// Eliminated reflects Player.Eliminated. Set when the player
	// concedes (S08); future state-based action work in S13+ may
	// also flip it. An eliminated player still appears in seats[],
	// still spectates, but the UI greys them out and disables their
	// quick-action buttons. Added in S08.
	Eliminated bool `json:"eliminated,omitempty"`
	// HandKept reflects Player.HandKept — true once the player has
	// committed to their opening hand. Drives the per-seat
	// "kept ✓" / "deciding…" indicator during the mulligan window.
	// Added in S08.
	HandKept bool `json:"hand_kept,omitempty"`
	// MulligansTaken reflects Player.MulligansTaken. Surfaced so the
	// UI can show "mulligans taken: N". Added in S08.
	MulligansTaken int `json:"mulligans_taken,omitempty"`
	// DeckImported reflects Player.DeckImported — true once the
	// player has had a real deck installed via ReplaceDeck (vs. the
	// 1-card placeholder commander handed out at AddPlayer time).
	// Drives the in-game deck-import modal in Game.svelte (S08.5
	// wave 1) — a card-count check is unreliable because the
	// placeholder also produces non-zero library / command counts.
	DeckImported bool `json:"deck_imported,omitempty"`
	// UndosRemaining is the per-turn undo budget left for this seat,
	// refreshed when the cursor enters their untap step. Drives the
	// "undos: N" indicator on the toolbar so the viewer can see at
	// a glance whether their next undo will be allowed. Added in S11.
	UndosRemaining int `json:"undos_remaining,omitempty"`

	// Discord identity (S12.5). Populated when the seat was claimed
	// via OAuth; zero values for manual-name joins. The client
	// builds /avatars/{discord_id}/{discord_avatar_hash}.png to
	// pull the cached portrait, and prefers display_name over name
	// for the seat label.
	DiscordID         string `json:"discord_id,omitempty"`
	DiscordAvatarHash string `json:"discord_avatar_hash,omitempty"`
	DisplayName       string `json:"display_name,omitempty"`

	// Bot seat (S31 sub-PR 4). IsBot marks a seat driven by an
	// aiseat runner; BotTier is its policy tier and BotDeck the
	// curated deck it was seated with. The client renders a BOT chip
	// and a distinct avatar mark off these, and shows the thinking
	// pulse while such a seat holds priority.
	IsBot   bool   `json:"is_bot,omitempty"`
	BotTier string `json:"bot_tier,omitempty"`
	BotDeck string `json:"bot_deck,omitempty"`

	// IsHost marks the table host (ADR 0075 §2.1) — the seat that may
	// change table settings alongside the server admin. Public to
	// every viewer. Not read from the engine: the room stamps it on
	// each capture from the host the lobby designated (ws/host.go).
	IsHost bool `json:"is_host,omitempty"`

	// CommanderCasts is the per-commander cast count from the
	// command zone (S13.1, CR 903.8). Keyed by commander instance
	// UUID string. Drives the "+N tax" indicator next to the
	// commander tile. Sandbox: the engine doesn't enforce the
	// {2}-per-cast surcharge — players track mana themselves.
	// Omitted when empty.
	CommanderCasts map[string]int `json:"commander_casts,omitempty"`

	// Counters is the per-player named counter map (S13.2 — poison,
	// energy, experience, rad, plus homebrew). The legacy `poison`
	// and `energy` int fields above stay populated for backwards
	// compatibility with pre-S13.2 clients. Omitted when empty.
	Counters map[string]int `json:"counters,omitempty"`

	// MaxHandSize is the per-player cleanup-step hand-size cap
	// (S13.4, CR 402.2). DefaultMaxHandSize (7) on every freshly
	// seated player; -1 sentinel disables the cap (Reliquary Tower
	// / Thought Vessel). Surfaced on the wire so clients can
	// render "8 / ∞" or "3 / 2" next to the hand-count badge.
	//
	// This is the EFFECTIVE cap, not the raw Player.MaxHandSize:
	// a controlled permanent with Spec.NoMaxHandSize reports -1
	// here without the underlying field being written. Otherwise the
	// badge would keep saying "10 / 7" for a player the cleanup step
	// is (correctly) never going to prompt (#338).
	MaxHandSize int `json:"max_hand_size"`

	// LandDropsPerTurn / LandsPlayedThisTurn are the two halves of
	// the land-drop badge (#500, CR 305.2): how many lands this seat
	// may play this turn, and how many it already has. The engine
	// REFUSES a land play past the allowance since #500, so the
	// client needs both numbers to grey the hand's lands out before
	// the player clicks rather than only explain the error after.
	//
	// LandDropsPerTurn is the EFFECTIVE allowance, not the raw
	// Player.LandDropsPerTurn: a controlled Exploration or a
	// one-turn grant is already summed in, exactly as MaxHandSize
	// above reports the effective cap. Normally 1 / 0.
	LandDropsPerTurn    int `json:"land_drops_per_turn"`
	LandsPlayedThisTurn int `json:"lands_played_this_turn"`

	// ManaPool is the player's current mana pool projection — one
	// entry per floating mana token, in insertion order. Entries
	// are uppercase single-character mana letters ("W", "U", "B",
	// "R", "G", "C"). Empty / nil when the pool is empty (CR 106.4
	// empties at every step boundary, so this is the common case
	// outside a cast). Drives the S15 ManaPoolPips UI. Added in
	// S15 sub-PR 2.
	ManaPool []string `json:"mana_pool,omitempty"`

	// Emblems are the emblems this player has (CR 114), in creation
	// order. Absent for every seat that has none, which is nearly
	// every seat in nearly every game.
	//
	// Not a ZoneView: a ZoneView carries CardViews, and an emblem has
	// no characteristics at all (CR 114.1), so a CardView of one is a
	// row of empty strings that the hover-zoom, the card-image route
	// and the targeting layer would each have to learn to skip. The
	// label and the text are the whole thing a client needs.
	//
	// PUBLIC and unredacted. An emblem sits face up in the command
	// zone and any player may read it, so FilterViewFor leaves this
	// field alone - the same posture as delayed_triggers and
	// life_history. Added in S40 (#623, ADR 0064).
	Emblems []EmblemView `json:"emblems,omitempty"`

	// Keywords are the abilities this PLAYER has right now
	// (CR 702.11d, CR 702.16i) — engine tokens, "hexproof" or
	// "protection from everything". Derived grants from a controlled
	// permanent (Leyline of Sanctity, Aegis of the Gods) first, then
	// the ones granted for a duration (Teferi's Protection, The One
	// Ring). Absent for every seat that has none, which is nearly
	// every seat in nearly every game.
	//
	// EFFECTIVE, like max_hand_size and land_drops_per_turn above:
	// the derived half is not written anywhere on the engine's
	// Player, so this is computed on every projection.
	//
	// PUBLIC and unredacted. Protection and hexproof are facts about
	// the board that every player at the table can see, and the
	// targeting rule they drive is already visible through
	// legal_targets — a viewer who could see the refusal but not its
	// reason is strictly worse off. Same posture as emblems.
	//
	// The client has no badge for this yet (#1197 names the follow-up);
	// the field ships with the rule so the badge is a client-only
	// change when it comes. Added in S40 (#1197, ADR 0072).
	Keywords []string `json:"keywords,omitempty"`

	// LifeTotalLocked is "your life total can't change" on this seat
	// (CR 119.7, CR 119.8) — Platinum Emperion's printed static, or a
	// grant that lasts until this player's next turn (Teferi's
	// Protection, Teferi's Reproach). Absent for every seat that has
	// no lock, which is nearly every seat in nearly every game.
	//
	// EFFECTIVE, like keywords above: the derived half is not written
	// anywhere on the engine's Player, so this is computed on every
	// projection.
	//
	// PUBLIC and unredacted, same posture as keywords and emblems.
	// The lock changes what every player at the table may do — a Bolt
	// that moves no life, a drain that drains nobody, a Phyrexian
	// symbol this seat may not claim — so a viewer who could see the
	// refusal but not its reason is strictly worse off.
	//
	// NOT folded into keywords. That field carries the bare ability
	// TOKENS a seat has, which the client parses "protection from
	// <quality>" out of; a life-total lock is not an ability the
	// player has and would be a badge built on a grammar it does not
	// belong to. See ADR 0085 Decision 7 and game/life_lock.go.
	// Added in S39 (#1200, ADR 0085).
	LifeTotalLocked bool `json:"life_total_locked,omitempty"`
}

// EmblemView is one emblem on the wire (CR 114). Label is what the
// board calls it ("Elspeth, Sun's Champion emblem") and Text is its
// printed ability, for the chip's hover.
//
// Both are read from the catalog on every projection rather than
// stored on the object, so a wording fix in a card file reaches a
// game that is already in progress.
type EmblemView struct {
	InstanceID string `json:"instance_id"`
	Label      string `json:"label"`
	Text       string `json:"text"`
}

// LifeChangeView is the wire representation of a single life-change
// log entry. Always emitted as part of PlayerView; the Player's
// canonical history is bounded server-side at MaxLifeHistoryEntries
// (S08), so the wire payload stays small without per-snapshot
// pruning here.
//
// Seq is game.LifeChange.Seq: a per-player counter that starts at 1
// and only goes up, so the client can key the life-change popup on
// something that survives the cap. Neither the array index nor At
// can do that job — the log stops growing at the cap, and At is
// RFC3339 SECONDS, so two changes in one second collide (#703).
type LifeChangeView struct {
	Delta    int    `json:"delta"`
	NewTotal int    `json:"new_total"`
	At       string `json:"at"` // RFC3339
	Seq      uint64 `json:"seq"`
}

// ZoneView is the wire representation of a Zone. Count is sent
// explicitly so that future visibility-filtered views (e.g. "opponent
// library count but not contents") have a stable field to populate.
type ZoneView struct {
	Kind  string     `json:"kind"`
	Owner string     `json:"owner,omitempty"`
	Count int        `json:"count"`
	Cards []CardView `json:"cards"`
}

// CardView is the wire representation of a Card. Carries an
// unexported `knowers` map for the S13.5 per-viewer redaction in
// FilterViewFor — encoding/json skips unexported fields so the map
// never reaches the wire. FilterViewFor walks the map per viewer,
// clears it on the output, and stamps `KnownByYou` from the lookup.
type CardView struct {
	InstanceID string `json:"instance_id"`
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	Controller string `json:"controller"`
	ScryfallID string `json:"scryfall_id,omitempty"`
	// TypeLine is Scryfall's printed type line ("Legendary Creature
	// — Human Wizard"). Carried so the client can filter "creatures
	// only" UIs (the combat panel) without a Scryfall round-trip.
	// Omitted for placeholder demo cards that have no resolved type.
	// Added in S08.
	TypeLine string `json:"type_line,omitempty"`
	// Colors is the effective color list (W/U/B/R/G), including layer-5
	// changes. Absent means colorless; mana cost is not a color fallback.
	Colors []string `json:"colors,omitempty"`
	// Protection is this permanent's CR 702.16 protections, PARSED
	// server-side (#662). The raw tokens are in Abilities like every
	// other keyword; this is the same list with the quality pulled
	// out, because protection is the only keyword whose ability
	// carries a parameter and the grammar has exactly one owner
	// (game/protection.go).
	//
	// Two consumers and the same reason for both. The client's badge
	// row renders the quality rather than the raw token, and the bot
	// may not import internal/game at all (ADR 0033 §3) — so the
	// engine hands both the parse rather than letting either
	// re-implement it.
	Protection []ProtectionView `json:"protection,omitempty"`
	// Power and Toughness are the parsed printed stats. Zero for
	// non-creatures and for any card with non-numeric printed stats
	// ("*", "1+*"). The client uses Power to label combat-panel
	// creature rows; ResolveCombatDamage uses CurrentPower (base +
	// counter modifiers) on the server side. Both omitempty for
	// non-creatures. Added in S08.
	Power int `json:"power,omitempty"`
	// NegativePower preserves signed power for comparisons such as skulk.
	// Present only below zero; Power retains its combat-damage zero clamp.
	NegativePower int  `json:"negative_power,omitempty"`
	Toughness     int  `json:"toughness,omitempty"`
	Tapped        bool `json:"tapped,omitempty"`
	// NoUntap describes an untap-step restriction or one-shot marker on
	// this battlefield permanent. Static is hidden for face-down cards;
	// Next is public state and survives the face-down identity redaction.
	NoUntap     *NoUntapView   `json:"no_untap,omitempty"`
	Counters    map[string]int `json:"counters,omitempty"`
	IsCommander bool           `json:"is_commander,omitempty"`
	// DamageMarked is the damage currently noted on this creature
	// (S13.1 — feeds the lethal-damage SBA). Cleared by the
	// cleanup-step turn-based action and on zone exit. Only
	// meaningful for creatures on the battlefield; omitted when
	// zero. Added in S13.1.
	DamageMarked int `json:"damage_marked,omitempty"`
	// RegenerationShields is how many regeneration shields (CR 701.19a)
	// this permanent is carrying — the number of times the next
	// destructions this turn will tap it, remove all damage from it and
	// take it out of combat instead of killing it. Public, like the
	// damage above and for the same reason: everyone at the table needs
	// it to decide whether removal is worth casting. Cleared at the
	// cleanup step and on zone exit; omitted when zero. Added in #667.
	RegenerationShields int `json:"regeneration_shields,omitempty"`
	// FaceDown reflects Card.FaceDown — a card flipped face-down
	// by morph / manifest / mutate-bottom (CR 708). Distinct from
	// KnownByYou: a face-down creature is face-down to everyone
	// visually, but the morph caster (and anyone else who saw it
	// face-up or via reveal) still has KnownByYou=true so their
	// client can hover-reveal the printed characteristics. Added
	// in S13.5.
	FaceDown bool `json:"face_down,omitempty"`
	// FaceDownKind is WHY the card is face down (ADR 0069) — one of
	// "exiled", "foretold", "manifested", "morphed", "disguised",
	// "cloaked", "turned" (an effect, not a keyword, put it face down
	// — #1209, #1270), or absent for a face-up card. PUBLIC: every player
	// can see that a permanent is a morph and that an exiled card is
	// foretold, so it survives the non-knower redaction. The client
	// uses it to label the card back.
	FaceDownKind string `json:"face_down_kind,omitempty"`
	// FaceVisible is whether THIS viewer may look at the face of a
	// face-down object — the controller of a CR 708.5 permanent, the
	// owner of a foretold card (CR 702.143d), nobody for a plain
	// face-down exile (CR 406.3). Stamped per-viewer by
	// FilterViewFor beside KnownByYou and never trusted from the
	// input; it is exactly `face_down && known_by_you`, on the wire
	// rather than derived client-side because it is the rules
	// permission and one place should own it. The client renders the
	// real face plus a face-down badge when it is true and a card
	// back when it is not. Added by ADR 0069.
	FaceVisible bool `json:"face_visible,omitempty"`
	// PhasedOut marks a card in the `phased_out` zone (CR 702.26,
	// #1199, ADR 0084). Always true there and absent everywhere else,
	// so the bit is redundant with the zone it arrived in — carried
	// anyway so that a client rendering phased-out permanents IN PLACE
	// one day has it without another wire change, and so that a card
	// pulled out of the zone into a list still says what it is.
	//
	// PUBLIC, like face_down_kind: everyone at the table can see that
	// a permanent phased out, and everyone needs to, because the board
	// simply stops showing it otherwise. Survives the non-knower
	// redaction.
	PhasedOut bool `json:"phased_out,omitempty"`
	// KnownByYou reports whether the viewer is currently a knower
	// of this card's identity (S13.5). Computed per-viewer at
	// FilterViewFor time. When false, printed characteristics
	// (name, type_line, scryfall_id, power, toughness, counters)
	// are redacted to their zero values; the instance ID, zone,
	// controller, tapped state, and battlefield position remain
	// visible so the engine state stays observable. Added in S13.5.
	KnownByYou bool `json:"known_by_you,omitempty"`

	// knowers is the set of player UUID strings that currently
	// know this card. Populated by viewOfCard from
	// game.Card.KnownBy; consumed by FilterViewFor which clears it
	// on the output. Unexported so encoding/json drops it from
	// the wire — the only consumer is the in-process per-viewer
	// projection. Added in S13.5.
	knowers map[string]bool

	// libraryTop marks the card the CR 401.5 standing visibility rule
	// currently applies to — the top of its owner's library, under a
	// "you may look at the top card" or "play with the top card
	// revealed" permanent. Set by stampLibraryTop, read by
	// keepKnownTopInLibraryZone.
	//
	// It exists because being KNOWN and being VISIBLE IN THE ZONE are
	// different facts. A one-shot "reveal the top two cards of your
	// library" makes those cards known to everyone — that is what the
	// reveal frame carries — but it does not make them visible sitting
	// in the library afterwards. Keying the zone projection on
	// KnownByYou alone would have leaked exactly that. Unexported, so
	// encoding/json never puts it on the wire.
	libraryTop bool

	// oracleID is the card's catalog key, kept off the wire (the
	// client keys art on scryfall_id) so the post-projection
	// legal-target stamp can look the TargetSpec up. Added in S20.
	oracleID string

	// castOffers holds the announce-time cast stamps this card carries
	// for a seat whose answer is NOT the public one, keyed by seat
	// UUID string: legal_targets, clauses, modes, additional_cost,
	// optional_costs, tap_cost, alternative_costs,
	// alternative_cost_required, target_cost_notes, phyrexian_symbols,
	// cant_cast, castable_here and the grant behind them.
	//
	// It exists because a legal target set is not a view of the board:
	// hexproof, shroud and "target opponent" all narrow it by who is
	// asking, so seat A's answer is not seat B's to read. The exported
	// fields carry the answer that is PUBLIC — the zone owner's own
	// cast out of their own pile, which is every printed flashback
	// card in every graveyard — and everything else lives here until
	// FilterViewFor promotes ONE seat's entry into those fields and
	// drops the map (applyCastStampsFor).
	//
	// A MAP since #1037, where it was one seat and one set of stamps
	// (`castOffersFor`, #978). A CardView is one struct, so a card two
	// seats may both cast — a Snapcaster flashback under somebody
	// else's Wrexial grant, two impulse grants over one exiled card —
	// could only carry one of the two answers, and the second holder
	// got the public zone and no picker for a cast the engine would
	// accept. The wire shape is unchanged, because the frame is built
	// per viewer already: this is the in-process projection catching
	// up with that.
	//
	// Unexported, so encoding/json never puts it on the wire, and
	// cleared on the way out like `knowers` so repeated FilterViewFor
	// calls stay stable.
	castOffers map[string]castStamps
	// BattleX, BattleY are the normalised battlefield position in
	// [0, 1] stamped by the `set_battlefield_position` action.
	//
	// Pointers, and set by viewOfZone for the BATTLEFIELD ZONE ONLY
	// (#29). Absent therefore means exactly one thing — "this card is
	// not on the battlefield" — and never "this card is at the
	// origin". Previously both were plain float64 with omitempty,
	// which dropped the pair whenever it was (0, 0) and so conflated
	// a card stamped at the origin with one that had never been
	// positioned. That origin is not an exotic input:
	// SetBattlefieldPosition CLAMPS rather than rejects, so every
	// negative and NaN coordinate a client sends lands on exactly 0,
	// and x=0 is "first in the row" for the client's within-row sort —
	// the single most useful slot to be able to name.
	//
	// Scoping the emission to the battlefield rather than simply
	// dropping omitempty is what keeps the frame small. Off the
	// battlefield the pair is meaningless by construction (zone exit
	// clears it, see game/zone.go), and those zones are most of the
	// frame: a redacted library card is four keys, so always emitting
	// would have grown the bulk of every broadcast by ~20% to say
	// "(0, 0)" about cards that have no position at all.
	//
	// Note this makes the wire honest, not more expressive. game.Card
	// has no "positioned" flag, so "on the battlefield but never
	// positioned" is not a state the server can distinguish from
	// (0, 0) in the first place — every battlefield card carries the
	// pair and unpositioned ones read (0, 0). Giving that its own
	// representation means adding real state to game.Card, which
	// waits on the within-row reorder UX being designed (#30, #35).
	BattleX *float64 `json:"battle_x,omitempty"`
	BattleY *float64 `json:"battle_y,omitempty"`
	// AttackingTarget is the ID this card is currently declared to
	// attack, or omitted if not declared. Cleared on zone exit and by
	// clear_combat. Added in S08.
	//
	// S27: this is no longer always a player ID. An attacker may be
	// declared against a player, a planeswalker or a battle
	// (CR 508.1d), and the id is a seat id or an instance id
	// accordingly. AttackingTargetKind says which, so the client
	// never has to guess by trying one lookup and falling back to the
	// other.
	AttackingTarget string `json:"attacking_target,omitempty"`
	// AttackingTargetKind is "player", "planeswalker" or "battle" —
	// what AttackingTarget names. Omitted when nothing is declared.
	// Added in S27.
	AttackingTargetKind string `json:"attacking_target_kind,omitempty"`
	// DefendingPlayer is the seat defending against this attack — the
	// only seat whose creatures may block it (CR 802.4a, #1339): the
	// player attacked, the controller of the planeswalker attacked, or
	// the PROTECTOR of the battle attacked. Computed by the server so
	// the client's block picker never re-derives who defends a battle.
	// Omitted when nothing is declared, and for an attacker whose
	// planeswalker or battle has left the battlefield (CR 506.4c —
	// nobody can block it).
	DefendingPlayer string `json:"defending_player,omitempty"`
	// ProtectorPlayer is the seat protecting this battle (CR 310.9a),
	// or omitted for every other card type and for a battle whose
	// protector prompt has not been answered yet. Public information:
	// the whole table needs to know who is defending, because it
	// decides who may attack it. Added in S27.
	ProtectorPlayer string `json:"protector_player,omitempty"`
	// Defense is the battle's current defense counter total, lifted
	// out of the counters map for the same reason loyalty is rendered
	// specially: it is the card's life total, not one pip among
	// several. Omitted for non-battles. Added in S27.
	Defense int `json:"defense,omitempty"`
	// BlockingTarget is the attacker instance ID this card is
	// currently declared to block, or omitted if not declared.
	// Cleared on zone exit and by clear_combat. Added in S08.
	BlockingTarget string `json:"blocking_target,omitempty"`
	// GoadedBy is the player ID who goaded this creature, or empty
	// when not goaded. Cleared on zone exit. Sandbox marker — the
	// must-attack-not-the-goader rule is not enforced. Added in S10.
	GoadedBy string `json:"goaded_by,omitempty"`
	// AttachedTo is the CR 301.5c / CR 303.4 attachment relation for
	// an Equipment or an Aura: the permanent or player this card is
	// attached to. Omitted for the overwhelming majority of cards,
	// which are attached to nothing.
	//
	// One direction only. The reverse list ("what is attached to
	// this creature") is DERIVED on the client by partitioning the
	// battlefield, the same shape PlayerPanel already uses for its
	// type buckets — two wire representations of one relation can
	// disagree, and one cannot.
	//
	// Not redacted: attachment is public battlefield state exactly
	// like attacking_target. Added in S24 (ADR 0036 decision 13).
	AttachedTo *TargetRefView `json:"attached_to,omitempty"`
	// Auto signals that this card is in the S14 effect catalog —
	// when it resolves (or ETBs), a registered effect fires
	// automatically rather than relying on manual sandbox clicks.
	// Omitted when false so non-catalog cards (the majority) don't
	// carry the field on the wire. Added in S14 sub-PR 3.
	Auto bool `json:"auto,omitempty"`
	// Unimplemented is the honest inverse of Auto, and it is
	// deliberately NOT !Auto. Most cards in a real deck have no
	// catalog Spec and do not need one: printed keywords are
	// enforced for every card in the dump since #317 / #319 / #320,
	// a vanilla creature is complete, a basic land taps off its type
	// line. This bit is set only when the card prints rules the
	// engine will not run — the case behind reports #321, #324,
	// #325, #332 and #333, where five uncatalogued cards resolved
	// into silence and the player had no way to tell that was by
	// design. See game.Unimplemented.
	//
	// Omitted when false. The client surfaces it where a player
	// forms an expectation and nowhere else — the hover/inspect
	// panel and the stack, not as a permanent board badge, which on
	// a real battlefield would be most of the cards on the table.
	Unimplemented bool `json:"unimplemented,omitempty"`
	// CastSurfaceView is the announce surface of the card AS THE FACE
	// IT IS SHOWING. Embedded rather than listed, so `target_mode`,
	// `legal_targets` and the rest keep the names and the JSON keys
	// they have always had while CardFaceView carries the identical
	// block for a face that is NOT up (#992).
	CastSurfaceView
	// ExilePlay is the S21 sub-PR 6 impulse-exile grant. Present
	// only while the card is in exile with a live permission;
	// absent — which is nearly always — the card is inert exile.
	ExilePlay *ExilePlayView `json:"exile_play,omitempty"`
	// ActivatedAbilities are the CR 602 activated abilities this
	// permanent offers, from its controller's point of view (public
	// information, so present on every viewer's copy). Absent off
	// the battlefield and for cards with none. Added in S21 sub-PR 2.
	ActivatedAbilities []ActivatedAbilityView `json:"activated_abilities,omitempty"`
	// ZoneAbilities are the CR 602 activated abilities this card
	// offers from the NON-BATTLEFIELD zone it is sitting in — the
	// hand (cycling and typecycling, CR 702.29a/e), a graveyard
	// (unearth CR 702.82a, scavenge CR 702.96a, embalm CR 702.128a),
	// exile, the command zone. Whatever `ActivatedAbilityShape.Zones`
	// names. Absent for every card that prints none, which is nearly
	// all of them.
	//
	// It was `hand_abilities` through #660, when the hand was the only
	// zone the field existed for; #1221 opened the other three and
	// renamed it rather than adding `graveyard_abilities` beside it.
	// One field is the point: the engine has ONE activation path with
	// a zone dimension (ADR 0062 Decision 1), so the wire has one row
	// list and the client has one reader.
	//
	// A separate field from ActivatedAbilities rather than a reuse,
	// and ADR 0062 Decision 5 says why: the two lists are disjoint by
	// construction, but they are read by different UI (a permanent's
	// menu versus a hand card's popover or the zone browser's row),
	// and a client that has not been taught about this field shows
	// nothing rather than showing a cycling row on a battlefield
	// permanent.
	//
	// UNLIKE ActivatedAbilities, this is NOT public, and since #1221
	// that is enforced rather than inherited. Through #660 it rode
	// the hand's own secrecy — FilterViewFor strips other seats'
	// hands wholesale. A GRAVEYARD is public, so the same trick would
	// have shipped one seat's answer to the whole table: the rows
	// carry `legal_targets` and `clauses`, which hexproof, shroud and
	// "target opponent" narrow BY WHO IS ASKING (#1055). So the list
	// rides castOffers like every other per-seat announce answer and
	// reaches exactly the seat whose card it is.
	//
	// `index` is the ability's index in the card's FULL ability
	// list, so the client sends the same activate_ability payload it
	// sends for a permanent. Added for #660, widened by #1221.
	ZoneAbilities []ActivatedAbilityView `json:"zone_abilities,omitempty"`
	// ZoneManaAbilities are the CR 605 MANA abilities this card offers
	// from the non-battlefield zone it is sitting in (#1228) — a hand,
	// today and only a hand (game.supportedManaAbilityZones). The
	// Spirit Guides' "Exile this card from your hand: Add {R}" is the
	// whole printed family.
	//
	// `zone_abilities`' twin, one ability kind over, and separate from
	// it for the reason `mana_abilities` is separate from
	// `activated_abilities`: the two payloads differ
	// (`activate_mana_ability` against `activate_ability`) and a
	// client sends the row it read.
	//
	// Not public, and it rides castOffers for the same reason
	// `zone_abilities` does — an activation row answers "what may YOU
	// announce about this card, from here". A hand is already hidden
	// wholesale, so today that is belt-and-braces; it is written this
	// way because the graveyard is one entry in
	// supportedManaAbilityZones away and a public pile would ship one
	// seat's answer to the table (#1055, #1167).
	//
	// `index` is the ability's index in the card's FULL mana-ability
	// list, so the client sends the same activate_mana_ability payload
	// it sends for a permanent.
	ZoneManaAbilities []ManaAbilityView `json:"zone_mana_abilities,omitempty"`
	// SpecialActions are the CR 116.2 special actions this card
	// offers while it is IN HAND — "Foretell {2}" (CR 702.143a),
	// "Suspend 1—{R}" (CR 702.62a). Each entry is a menu row the
	// client fires directly as a `special_action` action; neither
	// kind's cost needs a choice, so there is no picker between the
	// row and the wire.
	//
	// `available` is the per-kind TIMING answer for right now, from
	// the engine's own table — so the client greys the row instead
	// of re-deriving "is it my turn" and "could I cast this" and
	// getting split second backwards. ADR 0062 Decision 4.
	//
	// Not public, and stripped by the same redaction that strips
	// zone_abilities: "Suspend 4—{U}" names Ancestral Vision.
	SpecialActions []SpecialActionView `json:"special_actions,omitempty"`
	// SummoningSick reports CR 302.6 sickness: the permanent is a
	// creature, it entered this turn and it has no haste, so it
	// can't attack or pay a {T} cost. Battlefield creatures only —
	// a Treasure, a mana rock or a fetchland that arrived this turn
	// is NOT sick and must not be greyed (#530, and the reports it
	// caused: #365, #368). Added in S21 sub-PR 2 for the
	// activated-ability menu's affordance; the server does the real
	// check.
	SummoningSick bool `json:"summoning_sick,omitempty"`
	// LoyaltyActivated reports CR 606.3: this planeswalker has
	// already had a loyalty ability activated this turn, so every
	// entry in ActivatedAbilities carrying a LoyaltyCost is greyed
	// until the turn cursor moves on. Battlefield planeswalkers
	// only. Before S27 this lived only in the server's
	// Game.LoyaltyActivatedThisTurn map, which is why the client's
	// canActivateLoyalty had to guess. Added in S27 (#329, #334).
	LoyaltyActivated bool `json:"loyalty_activated,omitempty"`
	// ClassLevel is a Class permanent's CR 716.2 level designation —
	// 1 for a Class nobody has levelled, up from there (ADR 0071).
	// Present only for a Class on the battlefield, absent for every
	// other card, so the client can render the badge on presence
	// rather than having to parse the type line.
	//
	// Public: a level is visible to everyone in paper, and it is what
	// says which of the card's printed lines are live. An
	// UNCATALOGUED Class still carries it — the levels are engine
	// state, not catalog state — exactly as an uncatalogued Saga
	// still shows its lore counters.
	ClassLevel int `json:"class_level,omitempty"`
	// Solved is a Case permanent's CR 719.3 solved designation
	// (ADR 0071). Public for the same reason, and absent — not
	// `false` — for every card that is not a solved Case.
	//
	// There is no wire field for "which printed abilities are active":
	// an inactive ACTIVATED ability is already absent from
	// activated_abilities, because the gate lives in the one accessor
	// that list is built from, and an inactive static or trigger has
	// no per-ability representation on the wire to grey out.
	Solved bool `json:"solved,omitempty"`
	// Prepared is a permanent's CR 722.3a prepared designation
	// (ADR 0090): while it is set, its controller may cast the copy of
	// its prepare spell that sits in exile — which the wire already
	// shows as an exile card with an `exile_play` stamp. Public, as a
	// level and a solved Case are, and absent rather than `false` for
	// every permanent that is not prepared.
	Prepared bool `json:"prepared,omitempty"`
	// ChosenColor and NamedTribe are the answers a player gave to this
	// permanent's "as this enters, choose a color" (CR 105.4) and "as
	// this enters, choose a creature type" (CR 614.12) instructions —
	// one uppercase colour letter (W/U/B/R/G) and one canonical
	// creature type ("Elf"). Absent when the permanent asks no such
	// question, and absent in the window between it entering and its
	// controller answering.
	//
	// PUBLIC (#781). A choice made as a card enters is announced at
	// the table and hidden from nobody, and CR 607.2d makes it the
	// only way to read the card's OTHER abilities: "creatures you
	// control of the chosen color" names a set nobody can compute
	// without the answer. Before this, an opponent could not tell
	// which creatures a Heraldic Banner was pumping and the
	// controller had to remember what they named.
	//
	// Projected once, here in viewOfCard, straight off the engine
	// fields — there is no per-card special case anywhere above this
	// line and there must not be one. Cleared by the non-knower
	// redaction with the other type-derived bits: a named tribe says
	// "Cavern of Souls" as loudly as loyalty says "planeswalker", and
	// CR 708.2 gives a face-down permanent no such choice to report.
	ChosenColor string `json:"chosen_color,omitempty"`
	NamedTribe  string `json:"named_tribe,omitempty"`
	// ChosenName is the same family's third answer (#1210): the CARD
	// NAME this permanent's "as this enters, choose a card name"
	// instruction was answered with — Pithing Needle, Phyrexian
	// Revoker, Sorcerous Spyglass. Absent when the permanent asks no
	// such question and in the window before it is answered.
	//
	// Public and cleared by the non-knower redaction for exactly the
	// reasons above, and with one more that makes it the loudest of
	// the three: a Needle that did not say what it named would leave
	// the table guessing why an ability is greyed out, and the greyed
	// row's `cant_activate` clause is printed text that says "the
	// chosen name" without saying which.
	ChosenName string `json:"chosen_name,omitempty"`
	// ManaCost is the printed casting cost as Scryfall returns it —
	// "{1}{R}", "{W/U}", "{X}{B}{B}", etc. Empty for lands and for
	// placeholder / demo-seed cards. Rendered by the client as a
	// read-only chip on hand-zone cards; S15 sub-PR 3 will parse
	// this into a ParsedCost at cast time for the strict-mode
	// validator. Added in S15 sub-PR 1.
	ManaCost string `json:"mana_cost,omitempty"`

	// ManaAbilities lists the card's activated mana abilities, one
	// entry per tap-or-cost-for-mana slot on the battlefield. The
	// client's right-click menu (ManaAbilityMenu.svelte) reads this
	// to build the activation buttons. Ability indices 0..N-1 match
	// the array ordering and are what activate_mana_ability's
	// payload carries. Empty / absent for non-producers. Added in
	// S15 sub-PR 2.
	ManaAbilities []ManaAbilityView `json:"mana_abilities,omitempty"`

	// Abilities is the card's effective keyword list — strings like
	// "flying", "first strike", "trample". Layered effects (Lord of
	// Atlantis grants flying to other Merfolk) populate this in
	// S16 sub-PR 4 onward. S18 reads this list to render keyword
	// badges and gate combat behaviour. Empty in S16 sub-PR 1
	// (engine ships, no card declares a static ability yet).
	// Added in S16 sub-PR 1.
	Abilities []string `json:"abilities,omitempty"`

	// TokenText is a TOKEN's printed ability text, verbatim (ADR
	// 0083) — "When this token dies, you gain 1 life." Empty for
	// every printed card and for a vanilla token.
	//
	// It exists because a token has no printing behind it: there is
	// no scryfall_id for the client to resolve oracle text from (ADR
	// 0078's art is still Proposed), and a Dragon Egg whose
	// dies-trigger the player cannot read is a surprise rather than a
	// play. The emblem's `text` is the same field for the same
	// reason.
	//
	// Newlines separate printed lines, as on the card. It is public —
	// a token's text is public information — and survives the
	// face-down redaction path only insofar as a token is never face
	// down.
	TokenText string `json:"token_text,omitempty"`

	// Restrictions is the S24 restriction set as stable snake_case
	// tokens — "cant_attack", "cant_block", "cant_be_blocked",
	// "cant_activate", "cant_activate_mana". Empty for the permanent
	// nothing is restricting, which is almost all of them.
	//
	// It is deliberately NOT folded into Abilities. A restriction is
	// not a keyword the permanent has, it is an effect something else
	// has (game/restrictions.go argues that at length), and the
	// client renders the two differently: a keyword gets a badge, a
	// restriction gets a control disabled with a reason.
	//
	// The client reads this INSTEAD of deriving the rule. The server
	// decides who can attack; the wire says so; attackAll and the
	// ability menu render it. A second derivation in TypeScript is
	// the thing #429 spent a PR deleting.
	Restrictions []string `json:"restrictions,omitempty"`

	// Layout is Scryfall's printing layout ("modal_dfc",
	// "transform", "adventure", …), omitted for the ordinary
	// single-faced card. The client reads it to decide whether
	// playing this card needs a face prompt at all. Added by
	// ADR 0034.
	Layout string `json:"layout,omitempty"`

	// Faces is every printed face of a multi-face card, front
	// first, and it is PURELY ADDITIVE: Name, TypeLine, ManaCost,
	// Power and Toughness above continue to mean "the ACTIVE
	// face's", which is what keeps the client change small. All
	// twenty-odd client-side type checks — cardTypes.ts,
	// Card.svelte's regexes, timing.ts's cast gate, the mana-source
	// estimator — keep working with zero edits, because they now
	// receive one clean type line instead of a concatenation.
	//
	// What this feeds is the face picker and the hover overlay's
	// back-face panel. Absent for single-faced cards.
	Faces []CardFaceView `json:"faces,omitempty"`

	// ActiveFace indexes Faces. Omitted when zero, which is the
	// front face and every single-faced card.
	ActiveFace int `json:"active_face,omitempty"`
}

// NoUntapView is the public projection of a permanent's untap-step
// restrictions. A controller-keyed marker is represented by the current
// controller's player ID in Next; duplicate and eliminated players are
// omitted by stampNoUntap.
//
// Static means the permanent does not untap during its controller's
// untap step RIGHT NOW: an UntapStepRestriction applies to it, or
// (#1313) a live "for as long as" hold does. Next lists one-shot
// next-untap-step markers only; a hold is not a "next step" statement.
type NoUntapView struct {
	Static bool     `json:"static,omitempty"`
	Next   []string `json:"next,omitempty"`
}

// CastSurfaceView is the announce-time cast surface of ONE CASTABLE
// OBJECT: everything the cast chain asks about before cast_spell goes
// out, for one specific half of one specific card out of one specific
// zone.
//
// It exists because a card can be TWO castable objects (ADR 0034). A
// modal DFC's faces are independently playable (CR 712.12a) and an
// adventure card's are too (CR 715.3), and the two halves have
// different catalog entries, different costs and different target
// clauses — Bonecrusher Giant targets nothing and Stomp deals 2 damage
// to any target. Until #992 the wire carried exactly one of these
// blocks, for the face that happened to be UP, and the client's
// `cardAsFace` had to CLEAR it when the player picked the other half
// because the front's answers are actively wrong attached to the back.
// Clearing it is why a targeted Adventure half reached cast_spell with
// no target picker ever opening.
//
// One type, embedded in both CardView (the face that is up) and
// CardFaceView (every castable face, including that one), so the wire
// key is `legal_targets` in both places and the client reads one
// shape. One writer too: castStamps embeds it, which is what keeps
// "the answer for a face" and "the answer for the card" from being two
// computations that can drift.
//
// The PER-VIEWER / PUBLIC split is ADR 0066's (#1055, #1167) and it
// applies per face exactly as it does per card: `castable_here`,
// `legal_targets` and `clauses` answer "what may YOU announce" and
// ride castStamps into one seat's frame, and everything else is a fact
// about the card in this zone that every viewer may read.
type CastSurfaceView struct {
	// TargetMode tells the client what kind of target to prompt
	// for at cast time. See effects.Spec.TargetMode for the enum.
	// Empty when the card takes no announce-time targets (either
	// not a catalog card, or a catalog card with no targeting
	// prompt — e.g. Pyroclasm, Wrath of God). Added in S14 sub-PR 4.
	TargetMode string `json:"target_mode,omitempty"`
	// LegalTargets is the S20 structured-targeting answer for a card
	// its controller could cast right now: the players and card IDs
	// the card's TargetSpec accepts against the current board. Only
	// present on cards in the viewer's own hand / command zone (an
	// opponent's hand card never carries it). Absent for cards
	// without a TargetSpec — those keep the S13.1 free-form picker
	// keyed off target_mode alone. An empty-but-present entry means
	// "no legal target right now", which the client treats as
	// uncastable. Added in S20 sub-PR 1.
	LegalTargets *LegalTargetsView `json:"legal_targets,omitempty"`
	// Clauses is every clause of a MULTI-clause target statement, in
	// printed order, each with its own legal set, count and printed
	// wording — Bite Down's "target creature you control" then
	// "target creature or planeswalker you don't control". Absent for
	// the single-clause card that is nearly every card, where
	// legal_targets alone says everything. The client walks the
	// clauses one prompt at a time. Added by #764 (ADR 0065 §7).
	Clauses []LegalTargetsView `json:"clauses,omitempty"`
	// Modes is the S20 sub-PR 4 modal-spell clause for a card in the
	// viewer's own hand / command zone: the prompt, how many options
	// to pick, and each option's label plus — for targeted options —
	// its target mode and legal set right now. Absent for non-modal
	// cards and stripped from opponents' hands. The client shows a
	// mode picker between the X prompt and targeting.
	//
	// The field is PUBLIC on a public pile — "choose two of three" is
	// printed on the card, and a bystander reads it off a graveyard in
	// paper — but the legal sets INSIDE it are not (#1172). See
	// ModeOptionView.LegalTargets and castStamps.publicIn.
	Modes *ModeSpecView `json:"modes,omitempty"`
	// AdditionalCost is the S21 sub-PR 5 "As an additional cost to
	// cast this spell, …" clause for a card in the viewer's own
	// hand / command zone. Absent for the overwhelming majority of
	// cards, which have none. The client must collect the payment
	// before firing cast_spell — the server rejects a cast that
	// arrives without it.
	AdditionalCost *AdditionalCostView `json:"additional_cost,omitempty"`
	// AlternativeCosts are the S22 "you may cast this spell for its
	// overload / evoke / cleave cost" offers for a card in the
	// viewer's own hand / command zone. Absent for the overwhelming
	// majority of cards. Optional, unlike AdditionalCost: the client
	// offers them alongside "pay the printed cost", and a cast that
	// names none is the ordinary case.
	AlternativeCosts []AlternativeCostView `json:"alternative_costs,omitempty"`
	// AlternativeCostRequired says the PRINTED mana cost is NOT one
	// of the prices this cast may claim out of the zone the card is
	// in right now, so the caster must name one of
	// `alternative_costs` — a Faithless Looting in the graveyard is
	// castable at its flashback cost and at nothing else (CR 702.34b,
	// rule 3 of validateCastPathLocked), and a card a permission
	// PRICES is the same shape (rule 4).
	//
	// Absent — every hand cast, every command-zone cast, and a
	// Gravecrawler whose graveyard permission carries no price —
	// means the printed cost is on the menu as usual.
	//
	// Shipped as a flag rather than as a nil-keyed entry in
	// `alternative_costs` (#1012) because the entry would break every
	// client that walks that list, and because the two questions are
	// genuinely different: the list is what you may pay INSTEAD, this
	// is whether the cost in the corner is still on the table. The
	// engine answers both out of one list — game.CastOffersForLocked,
	// where a nil entry IS the printed cost — so the picker and the
	// bot's move list cannot disagree.
	//
	// Stamped and stripped with `alternative_costs`, which it is only
	// meaningful beside.
	AlternativeCostRequired bool `json:"alternative_cost_required,omitempty"`
	// TapCost is the S22 convoke / waterbend clause for a card in the
	// viewer's own hand / command zone: which of your untapped
	// permanents may be tapped to help pay, and how many. Absent for
	// the overwhelming majority of cards. Optional like the
	// alternative costs — tapping nothing is always a legal cast.
	TapCost *TapCostView `json:"tap_cost,omitempty"`
	// PhyrexianSymbols is how many symbols in the cost this card is
	// being offered at carry CR 107.4's "or 2 life" option — {U/P}
	// on Gitaxian Probe, {B/P}{B/P} on Dismember, {G/W/P} on a
	// compleated planeswalker. It is the CEILING on the
	// `phyrexian_life` the client may send with the cast, and the
	// only reason the client knows to offer the stepper at all
	// (#916).
	//
	// Shipped as a COUNT rather than left to the client, for the
	// reason `demands_x` is: a client that re-derived it would be a
	// second parser of the mana-cost syntax, and the two would drift
	// the day a new symbol family is printed — which is exactly what
	// #787 was. It counts the effective cast cost, so a commander
	// tax or a cost modifier cannot change it and does not.
	// `alternative_costs[i].phyrexian_symbols` answers the same
	// question for an offer that replaces the printed cost.
	//
	// Stamped with the other cast clauses on cards in the viewer's
	// own hand, command zone and castable graveyard, and stripped
	// with them. Absent — which is nearly every card — means there
	// is no life half to offer.
	PhyrexianSymbols int `json:"phyrexian_symbols,omitempty"`
	// TargetCostNotes are the printed clauses of this card's own cost
	// modifiers whose price depends on its targets — Fireball's "This
	// spell costs {1} more to cast for each target beyond the first",
	// strive — for a card in the viewer's own hand / command zone /
	// castable graveyard. The X picker opens before targeting and its
	// readout is priced at one target, so it shows these clauses
	// under the readout instead of a surcharge it cannot know yet
	// (ADR 0048 addendum, open question 2). Absent for nearly every
	// card. Added for #746.
	TargetCostNotes []string `json:"target_cost_notes,omitempty"`
	// CastableHere is the S29 "this card can be cast from the zone
	// you are looking at it in" bit, for the zones where that is not
	// already implied by the surface: the graveyard, today. Hand and
	// command-zone cards never carry it — every card in a hand is a
	// cast candidate, and the command zone has its own button.
	//
	// It is the flag the zone browser keys its cast button off, the
	// way exile keys its impulse button off `exile_play`. The cost
	// to pay rides `alternative_costs`, already filtered to the
	// offers claimable from this zone — so a Faithless Looting in
	// the graveyard carries flashback and nothing else, while the
	// same card in hand carries neither.
	//
	// DERIVED, since #1015, from exactly two things and in exactly
	// one place (castStampsFor): the prices this cast may claim out
	// of this zone (game.CastOffersForLocked) and the ADR 0073 §7
	// cast gate. An empty price list is a real "no" — a card whose
	// only path out of the graveyard is an escape cost the caster
	// cannot pay is not a cast surface, and marking one rendered a
	// button the announce path refused with ErrCastCostRequired.
	//
	// PER VIEWER, and it means "YOU may cast this from here" (#1055).
	// Not "the pile's owner may": it is stamped only on the frame of
	// a seat that may actually make the cast — the pile's owner for a
	// printed flashback or escape, the holder of a CastPermission over
	// the card for a granted one, both of them on their own frames
	// when both are true — and is absent for everybody else,
	// spectators included.
	//
	// It was public until #1055, and that is the surface the S48 #891
	// pass is about: the field's NAME is a statement about the viewer
	// and its VALUE was a statement about somebody else, so both
	// client readers had to pair it with `exile_play` to work out
	// which of the two they were holding. They read the bit alone now.
	//
	// What stays public is the half that is a fact about the CARD IN
	// THIS ZONE rather than about a player — `alternative_costs` and
	// `alternative_cost_required`, `modes`, `additional_cost`,
	// `optional_costs`, `tap_cost`, `target_cost_notes`,
	// `phyrexian_symbols` and `cant_cast` — because a card in a
	// graveyard is a card every player may pick up and read
	// (castStamps.applyPublicTo).
	//
	// Two of those carry per-viewer lists one level down and those go
	// with this bit rather than with their parents (#1172):
	// `modes[i].legal_targets`, `modes[i].clauses` and
	// `alternative_costs[i].legal_targets` / `.pay_options`.
	CastableHere bool `json:"castable_here,omitempty"`
	// OptionalCosts are the "you may pay an additional cost" offers
	// this card makes (CR 601.2b, ADR 0073) — kicker, multikicker,
	// buyback. Absent for nearly every card. Stamped alongside
	// `alternative_costs`, because the client asks both questions in
	// one modal.
	OptionalCosts []OptionalCostView `json:"optional_costs,omitempty"`
	// CantCast is the printed clause that stops this card being cast
	// from the zone it is in right now (CR 101.2, ADR 0073 §7) —
	// "Each player can't cast more than one spell each turn", "Cast
	// this spell only if you control a legendary creature or
	// planeswalker". Absent, which is nearly always, means nothing
	// refuses the cast.
	//
	// The STAMP of the one gate function CastSpell and the bot
	// enumerator both call, so the client can grey the card and say
	// why from server data rather than from a rule it reimplemented.
	// A card carrying it is never castable: the client must not
	// dispatch cast_spell for it, and the server would refuse.
	//
	// Public, like `castable_here`: a Rule of Law on the battlefield
	// is visible to everyone, so the fact that it is stopping a cast
	// is not hidden information. Cleared with the other announce
	// hints on the non-knower redaction.
	CantCast string `json:"cant_cast,omitempty"`
}

// CardFaceView is one printed face on the wire (ADR 0034). Enough
// to render a picker row and a hover panel: what it is called, what
// it costs, what it is, and where its art lives.
type CardFaceView struct {
	Name     string `json:"name"`
	TypeLine string `json:"type_line,omitempty"`
	ManaCost string `json:"mana_cost,omitempty"`
	// OracleText is the face's rules text, so the picker can show
	// what each half actually does without a round-trip to
	// GET /cards/{id}.
	OracleText string `json:"oracle_text,omitempty"`
	// Power and Toughness for a creature face. Both omitted when
	// zero, as on CardView.
	Power     int `json:"power,omitempty"`
	Toughness int `json:"toughness,omitempty"`
	// Image is the path to this face's art:
	// "/cards/{scryfall_id}/image?face=N". Built server-side so the
	// client never has to know that the face index is a query
	// parameter, and empty for a card with no Scryfall ID
	// (fixtures, the demo seed).
	Image string `json:"image,omitempty"`

	// CastSurfaceView is the announce surface of a cast of THIS face
	// out of the zone the card is sitting in (#992) — the same field
	// names, the same JSON keys and the same per-viewer split as the
	// block CardView carries for the face that is up.
	//
	// Present only on a face the card actually offers a cast of:
	// game.CastableFaces, narrowed by a grant that names faces
	// (CR 715.4's Adventure permission opens the creature and no
	// other). So it is two blocks for a modal DFC and an adventure
	// card, one for a transform card's front, and nothing at all for
	// the ~33,000 single-faced oracle IDs, which carry no `faces` on
	// the wire to hang it off.
	//
	// The client's face picker swaps this in when the player chooses
	// the half — `cardAsFace` — instead of clearing what the front
	// published, which is what left a targeted Adventure half with no
	// target picker.
	CastSurfaceView

	// castOffers is this FACE's per-viewer answers, keyed by seat,
	// exactly as CardView.castOffers is for the card. Unexported and
	// never serialised: applyCastStampsFor promotes the viewer's own
	// entry into the exported block above and drops the rest, which
	// is what keeps one seat's legal target set off another seat's
	// frame for a face as well as for a card.
	castOffers map[string]castStamps
}

// ManaAbilityView is the wire shape of one activated mana ability
// on a permanent. Mirrors effects.ManaAbility via the
// game.ManaAbilityShape projection, so the client can render the
// button labels without reaching into the catalog directly. Added
// in S15 sub-PR 2.
// ExilePlayView is the impulse-exile grant on a card in exile —
// "exile the top card of that player's library, you may play it
// this turn" (S21 sub-PR 6). Public information: the trigger that
// created it resolved in the open, so every viewer sees who may
// play the card. The client offers the action only to `player`.

// SpecialActionView is one CR 116.2 special action offered on a card
// in the viewer's own hand — a menu row and nothing more. There are
// no targets, no modes and no cost picker, so the client fires the
// `special_action` verb straight from the row. ADR 0062 Decision 4;
// #658 / #659.
type SpecialActionView struct {
	// Kind is the wire kind the action payload carries: "foretell",
	// "suspend".
	Kind string `json:"kind"`
	// Label is the row's text, as the card prints the keyword:
	// "Foretell {2}", "Suspend 1—{R}".
	Label string `json:"label"`
	// Cost is the PRINTED mana cost of TAKING the action, in Scryfall
	// brace notation — always "{2}" for foretell (CR 702.143a), the
	// printed suspend cost for suspend. Empty is a free action —
	// Lotus Bloom suspends for nothing.
	Cost string `json:"cost,omitempty"`
	// ChargedCost is what the engine actually charges for Cost right
	// now, after every CR 601.2f cost modifier on the battlefield
	// (#1319) — Ranar the Ever-Watchful's "The first card you
	// foretell each turn costs {0} to foretell" turns a printed `{2}`
	// into a charged `""`. Rendered from the same ParsedCost
	// SpecialActionManaCostForEffect charges (ParsedCost.String()),
	// so the row and the payment can never disagree the way Cost
	// alone could once a discount applied. EQUAL to Cost when no
	// modifier reaches this action.
	//
	// A POINTER for the reason ActivatedAbilityView.ChargedManaCost
	// is one: a discount that empties the cost out completely renders
	// "", which is a real answer ("this action now costs nothing")
	// that `omitempty` on a plain string would erase, indistinguishable
	// from Cost being empty to begin with. Nil is "not priced" — an
	// unparseable printed string, or a modifier that itself errors —
	// in which case the row falls back to Cost exactly as a
	// pre-#1319 client would. Absent whenever Cost is empty; a cost
	// with no mana component is not made of mana, so there is
	// nothing for this field to price.
	ChargedCost *string `json:"charged_cost,omitempty"`
	// Available is the engine's own per-kind timing answer for this
	// moment: foretell only during its owner's turn (and legal under
	// split second), suspend only when the card could begin to be
	// cast (and not under split second). False means grey the row,
	// never hide the keyword — a player has to be able to see that
	// the card has it.
	Available bool `json:"available,omitempty"`
}

type ExilePlayView struct {
	// Player is the instance ID of the player who may play it —
	// usually not the card's owner.
	Player string `json:"player"`
	// CastOnly marks a grant that permits casting but not playing a
	// land (Ragavan). The client labels the action accordingly.
	CastOnly bool `json:"cast_only,omitempty"`
	// AnyColor marks "you may spend mana as though it were mana of
	// any color" (Breeches).
	AnyColor bool `json:"any_color,omitempty"`
	// CostOverride is the mana cost the holder pays INSTEAD of the
	// card's printed one — airbend's "{2} rather than its mana
	// cost". Empty for impulse exile, which charges the printed
	// cost. The card's `mana_cost` field still carries the printed
	// value, so a client that ignores this shows the wrong price.
	CostOverride string `json:"cost_override,omitempty"`
	// NotBeforeTurn is the earliest turn number the grant is live on
	// — warp's "you may cast it from exile ON A LATER TURN" (S29).
	// Absent for every grant that is live as soon as it is made,
	// which is all of impulse exile and airbend. The client compares
	// it against `turn.number` and withholds the button until then;
	// the server rejects an early cast regardless.
	NotBeforeTurn int `json:"not_before_turn,omitempty"`
	// Faces are the printed faces this grant opens, when it names any
	// (S32). Absent — every impulse, airbend, warp and cascade grant
	// — means the grant does not speak about faces and the card's own
	// layout decides, which is what `faces` and `layout` already tell
	// the client.
	//
	// Present means the grant opens THOSE FACES AND NO OTHER. Two
	// grants name one each, in opposite directions: a defeated
	// Siege's "cast it transformed" names the BACK face, where the
	// card in the exile pile is still showing the battle and the
	// thing the button will actually cast is `faces[1]`; and CR
	// 715.4's Adventure grant names the CREATURE face, face 0, which
	// is why this is a list rather than the bare integer it was until
	// #719 — absent and "face 0" are different facts and one integer
	// could not tell them apart.
	//
	// A client that ignores this labels the button with the wrong
	// card name; it does not cast the wrong thing, because the server
	// settles the face from the grant rather than from the request.
	Faces []int `json:"faces,omitempty"`

	// XLockedAtZero is CR 107.3b for a cast taken under this grant:
	// the card prints an {X} in its mana cost and `cost_override`
	// does not, so the only legal X is 0 and the client must not open
	// its X picker. This is what a cascade hit carries. Absent — the
	// overwhelming majority, including every impulse grant that
	// charges the printed cost — means X is announced as usual.
	//
	// Server-computed (game.CastCost.LocksXAtZero), the same field
	// and the same rule an alternative-cost offer carries. Added for
	// #831.
	XLockedAtZero bool `json:"x_locked_at_zero,omitempty"`
}

// ActivatedAbilityView is one CR 602 activated ability on a
// battlefield permanent, from its controller's point of view. The
// cost flags tell the client what to collect before firing
// activate_ability with `ability_index`: a sacrifice pick from
// SacrificeOptions, a target from LegalTargets. Only stamped for
// the controller — an opponent's menu isn't theirs to open.
// Added in S21 sub-PR 2.
type ActivatedAbilityView struct {
	Index int    `json:"index"`
	Label string `json:"label,omitempty"`
	// Cost components. TapCost greys the entry when the source is
	// tapped or summoning-sick; ManaCost / LifeCost are advisory
	// (the server does the real check).
	TapCost       bool   `json:"tap_cost,omitempty"`
	SacrificeSelf bool   `json:"sacrifice_self,omitempty"`
	ManaCost      string `json:"mana_cost,omitempty"`
	// ChargedManaCost is what the engine actually charges for
	// ManaCost's mana component right now, after every CR 601.2f cost
	// modifier on the battlefield (#1190) — Boom Scholar's "Exhaust
	// abilities of other permanents you control cost {2} less to
	// activate" turns a printed `{3}{R}` into a charged `{1}{R}`.
	// Rendered from the same ParsedCost AbilityManaCostForEffect
	// charges (ParsedCost.String()), so the row and the payment can
	// never disagree the way ManaCost alone could once a discount
	// applied. EQUAL to ManaCost when no modifier reaches this
	// ability — the client always prefers this field and shows
	// ManaCost as a tooltip only when the two differ.
	//
	// A POINTER for the reason LoyaltyCost below is one: a discount
	// that empties the component out completely renders "", which is
	// a real answer ("this ability now costs nothing") that
	// `omitempty` on a plain string would erase, indistinguishable
	// from the printed cost could not be priced at all. Nil is that
	// second case — an unparseable printed string, or a modifier that
	// itself errors — in which case the row falls back to ManaCost
	// exactly as a pre-#1190 client would. Absent whenever ManaCost
	// is empty; a cost with no mana component is not made of mana, so
	// there is nothing for this field to price.
	ChargedManaCost *string `json:"charged_mana_cost,omitempty"`
	LifeCost        int     `json:"life_cost,omitempty"`
	SorcerySpeed    bool    `json:"sorcery_speed,omitempty"`
	// ConditionUnmet is true when the ability carries an activation
	// condition (CR 602.1b — "Activate only if an opponent controls
	// four or more lands", "Activate only during your turn") and that
	// condition is false right now. Absent when there is no condition
	// or it holds. Negative so `omitempty` keeps the common case off
	// the wire. Evaluated once, with the permanent's controller as
	// "you", and sent to every viewer: a condition reads only public
	// information. The client greys the row the way it greys
	// SorcerySpeed; the server refuses the activation with
	// ErrConditionNotMet either way. ADR 0020's #743 addendum.
	ConditionUnmet bool `json:"condition_unmet,omitempty"`
	// Exhausted is true when this is an exhaust ability ("Activate
	// each exhaust ability only once") that this object has already
	// activated, so the engine will refuse it for the rest of the
	// game (#1181). Its own flag rather than ConditionUnmet because
	// the two recover differently and the client says so: a condition
	// may be true again next turn, an exhaust never is until the
	// permanent becomes a new object (CR 400.7 — a flicker, not an
	// untap). The client greys the row the same way; the server
	// refuses with ErrAbilityExhausted either way.
	Exhausted bool `json:"exhausted,omitempty"`
	// CantActivate is the printed clause of a board-wide "can't be
	// activated" static that refuses THIS ability right now (CR
	// 602.5a, #1210) — "Activated abilities of creatures can't be
	// activated" (Cursed Totem), "…of sources with the chosen name …
	// unless they're mana abilities" (Pithing Needle). Absent, which
	// is nearly always, means nothing refuses it.
	//
	// The STAMP of the one gate function ActivateCatalogAbility and
	// the bot enumerator both call, so the client greys the row and
	// says why from server data rather than from a rule it
	// reimplemented — `cant_cast`'s shape, one level down, because a
	// restriction on activating is a fact about an ability and not
	// about the card.
	//
	// Distinct from CardView.Restrictions' `cant_activate` token,
	// which is the per-PERMANENT Arrest bit and says nothing about
	// which ability or why. Distinct from ConditionUnmet and
	// Exhausted for the reason those two are distinct from each
	// other: the three recover differently and the client says so —
	// a condition may hold again next turn, an exhaust never does
	// until the object is new, and a restriction ends when somebody
	// kills the artifact.
	CantActivate string `json:"cant_activate,omitempty"`
	// TimingClosed says the engine will refuse this activation RIGHT
	// NOW for timing (CR 602.5d, CR 606.3) — the stamp of
	// game.ActivationTimingOpenLocked, the one read the activation
	// path and the bot enumerator also ask (#1208).
	//
	// It is the row's own verdict, where SorcerySpeed above is only
	// the ability's printed clause. The two differ exactly where a
	// per-player statement speaks: The Wandering Emperor's loyalty
	// rows carry `sorcery_speed` and NOT this on the turn she
	// entered, and a Leonin Shikari's controller's equip rows carry
	// `sorcery_speed` and not this at any time.
	//
	// NEGATIVE, so `omitempty` keeps it off every row the engine has
	// no objection to — which is every instant-speed ability, always.
	// Present on a sorcery-speed or loyalty row whenever the window
	// is shut, which is the common state of such a row and the one
	// the client already greys.
	//
	// Evaluated once with the ability's CONTROLLER as "you" and sent
	// to every viewer, exactly as ConditionUnmet is: whose turn it
	// is, what is on the stack and what the battlefield says are all
	// public.
	TimingClosed bool `json:"timing_closed,omitempty"`
	// LoyaltyCost is the loyalty component of a planeswalker's
	// loyalty ability: +N / 0 / −N (CR 606.4). A POINTER because [0]
	// is a real printed cost and `omitempty` would erase it — the
	// client needs "no loyalty component" and "costs zero loyalty"
	// to stay different, since only the first leaves the ability
	// activatable more than once a turn. Added in S27 (#329, #334).
	LoyaltyCost *int `json:"loyalty_cost,omitempty"`
	// SacrificeLabel / SacrificeOptions describe a "Sacrifice a
	// creature"-style cost: the clause and the permanents the
	// controller may pay with right now. Absent when the cost has
	// no sacrifice component. SacrificeOptions' min / max are the
	// count ("Sacrifice two artifacts" ships 2 / 2, #747) and its
	// cards come in payment order — see sacrificeCostOptions.
	SacrificeLabel   string            `json:"sacrifice_label,omitempty"`
	SacrificeOptions *LegalTargetsView `json:"sacrifice_options,omitempty"`
	// CrewCost is the crew number of a Vehicle's crew ability
	// (CR 702.122a) — "Crew 3" ships 3. Zero and absent for every
	// ability that is not a crew ability.
	//
	// CrewOptions is the set of creatures that could pay it right
	// now: untapped creatures the controller controls, summoning
	// sickness deliberately NOT filtered out, because tapping to
	// crew is not paying a {T} cost and a creature cast this turn
	// may crew. The client collects a subset whose total power
	// reaches CrewCost and sends them as `crew_ids`; the server
	// re-checks. Each option's power is already on the CardView the
	// client holds, so the running total is computable client-side
	// without a second round trip. Added in S27.
	CrewCost    int               `json:"crew_cost,omitempty"`
	CrewOptions *LegalTargetsView `json:"crew_options,omitempty"`
	// CounterCostView is the counter half of the cost — embedded, so
	// its fields sit at the top level of the JSON exactly as they did
	// before #789 split them out, and so a mana ability can carry the
	// same ones without a second declaration to keep in step.
	CounterCostView
	// DiscardSelf is cycling's "Discard this card" cost component
	// (CR 702.29a, #660). Advisory: the client renders the cost
	// chip, and there is nothing to collect — the source IS the
	// payment, so no `discard_ids` is sent for it.
	DiscardSelf bool `json:"discard_self,omitempty"`
	// ExileSelf is scavenge's and embalm's "Exile this card from your
	// graveyard" cost component (CR 702.96a / CR 702.128a, #1221).
	// DiscardSelf's sibling one zone over, and advisory for exactly
	// the same reason: the source IS the payment, so there is nothing
	// to collect and nothing on the payload. The keyword's own label
	// spells the clause out, so like `discard_self` this has no
	// renderer of its own today — it is here so a client that wants
	// to mark the row does not have to parse the label for it.
	ExileSelf bool `json:"exile_self,omitempty"`
	// DiscardCostN / Label / Options describe a "Discard N cards"
	// cost component (#660): Fauna Shaman's "Discard a creature
	// card", Cryptbreaker's "Discard a card". DiscardCostN is the
	// count and its presence marks the component; zero and absent
	// for every ability without one.
	//
	//   - DiscardCostLabel is the clause as printed, without the
	//     verb ("a creature card").
	//   - DiscardCostOptions is what could pay right now: the cards
	//     in the activator's hand that match the clause, in hand
	//     order, with the source excluded (an ability activated from
	//     hand cannot pay for itself). Absent when nothing can pay.
	//
	// The client sends the chosen cards as `discard_ids` and skips
	// its picker entirely when the options number exactly
	// DiscardCostN — a modal with one possible answer is a worse
	// version of no modal.
	DiscardCostN       int      `json:"discard_cost_n,omitempty"`
	DiscardCostLabel   string   `json:"discard_cost_label,omitempty"`
	DiscardCostOptions []string `json:"discard_cost_options,omitempty"`
	// ReturnLabel / ReturnOptions describe a "Return a permanent you
	// control to its owner's hand" cost component (#1213) — Quirion
	// Ranger's Forest, Master Transmuter's artifact, Meloku's land.
	// Absent when the cost has no return component.
	//
	// ReturnOptions is a LegalTargetsView whose min and max are both
	// the clause's count, exactly as SacrificeOptions' are, and whose
	// cards are the permanents that could pay right now in payment
	// order — so the client reuses the sacrifice picker and its
	// "Choose for me" button fills with what a bot would have paid.
	// The chosen permanents go back as `return_ids`.
	//
	// It is NOT a target list: returning a permanent to pay a cost
	// does not target it (CR 601.2h), so a hexproof permanent you
	// control is on this list.
	ReturnLabel   string            `json:"return_label,omitempty"`
	ReturnOptions *LegalTargetsView `json:"return_options,omitempty"`
	// Waterbend is the CR 701.67 clause of a "Waterbend {N}:" cost
	// (#1310) — Aang, Swift Savior, Katara, Water Tribe's Hope — in
	// the SAME TapCostView shape a hand card's convoke / waterbend
	// ships as `tap_cost`, so the client's one picker serves both.
	// Absent for every ability without the clause.
	//
	//   - Options are the untapped artifacts and creatures the
	//     activator could tap right now, from the walk the engine
	//     validates against (game.WaterbendOptionsForEffect). The
	//     source is among them unless its cost also prints {T}.
	//   - Max is how many may be tapped: the waterbend's generic, capped
	//     by what the priced cost still charges. 0 with DemandsX set
	//     means "size it from the X you are about to announce"
	//     (Katara's "Waterbend {X}").
	//
	// OPTIONAL in both directions: tapping none pays the whole cost
	// with mana. The picks ride activate_ability as `waterbend_ids`.
	Waterbend *TapCostView `json:"waterbend,omitempty"`
	// TapOthersLabel / TapOthersOptions describe a "Tap another
	// untapped creature you control" cost component (#758's
	// TapOthersCost, wired to the wire by #759 for station, CR
	// 702.184a). Absent when the cost has no such component.
	//
	// TapOthersOptions is ReturnOptions one verb over: a
	// LegalTargetsView whose min and max are both the clause's count,
	// whose cards are the untapped permanents that could pay right
	// now, from the same walk the engine validates against
	// (game.TapOthersOptionsForEffect). The client reuses the
	// sacrifice picker with the verb "Tap" and sends the choice as
	// `tap_ids`.
	//
	// NOT a target list and NOT the {T} symbol: a hexproof creature
	// you control is on it (CR 601.2h), and so is one that arrived
	// this turn (CR 302.6). The source is left off when the clause
	// says "another" or when the ability also taps its source
	// (CR 118.3 — it will already be tapped).
	TapOthersLabel   string            `json:"tap_others_label,omitempty"`
	TapOthersOptions *LegalTargetsView `json:"tap_others_options,omitempty"`
	// DemandsX marks an ability whose mana component contains {X}
	// (Helm of Obedience, Treasure Vault, Soothsaying). The client
	// opens its X picker before the targeting step and sends the
	// answer as `x_value` on the activate_ability payload; the
	// server validates it at announce and locks it onto the stack
	// item (CR 602.2b).
	//
	// Derived from ManaCost rather than declared, so the two can
	// never disagree — but shipped explicitly all the same, because
	// a client that had to re-parse the cost string to find out
	// would be a second parser of the same syntax.
	//
	// MinX is the floor the printed text puts on the announcement
	// ("X can't be 0" ships 1). Absent means the ordinary floor of
	// zero. XSlots is how many {X} tokens the cost carries — 2 for
	// Treasure Vault's "{X}{X}" — so the picker can show what a
	// given X actually costs without parsing.
	DemandsX bool `json:"demands_x,omitempty"`
	MinX     int  `json:"min_x,omitempty"`
	XSlots   int  `json:"x_slots,omitempty"`
	// PhyrexianSymbols is how many symbols in this ability's mana
	// component carry CR 107.4's "or 2 life" option — 1 for Birthing
	// Pod's "{1}{G/P}", 2 for Solphim's "{1}{R/P}{R/P}". The client
	// offers a stepper bounded by it and by the activator's life
	// total, and sends the answer as `phyrexian_life` (#917, #916).
	//
	// Derived from ManaCost rather than declared, exactly as
	// DemandsX is, and shipped explicitly for the same reason: a
	// client that re-parsed the cost string to find out would be a
	// second parser of the same syntax.
	PhyrexianSymbols int `json:"phyrexian_symbols,omitempty"`
	// TargetMode / LegalTargets mirror the cast-time targeting
	// fields for an ability that targets.
	TargetMode   string            `json:"target_mode,omitempty"`
	LegalTargets *LegalTargetsView `json:"legal_targets,omitempty"`
	// Clauses is every clause of the ability's statement when it has
	// more than one; Modes is its CR 700.2 mode clause when it is
	// modal. Both absent for the ordinary ability. Added by #764.
	Clauses []LegalTargetsView `json:"clauses,omitempty"`
	Modes   *ModeSpecView      `json:"modes,omitempty"`
}

// CounterCostView describes the counter components of an ability's
// activation cost. ONE view for both ability kinds (#789): an
// activated ability and a mana ability embed it, because they carry
// the same game.CounterRemovalCost and the client renders them with
// the same picker.
//
// CounterCostN is the number removed and its presence marks the
// removal component; zero and absent for every ability without one.
//
//   - CounterCostKind is the printed kind ("loyalty", "charge");
//     empty means "a counter" of ANY kind, and the client asks for
//     the kind as well as the permanent. Empty WITH CounterCostAmong
//     is #943's last shape (Tekuthal): the kind is asked once per
//     permanent, which the option list below already answers —
//     nothing new is projected for it, because a row is a
//     (permanent, kind) pair and always was.
//   - CounterCostSelf is the "from this" form: the counters come off
//     the source, and no permanent is sent.
//   - CounterCostLabel is the "from" clause of the other and among
//     forms ("a planeswalker you control", "artifacts you control");
//     empty for the self form.
//   - CounterCostAmong is "from AMONG …" (#789): the N counters may
//     be split across any number of the listed permanents, and the
//     client sends a count per permanent in `counter_counts`. With
//     an empty CounterCostKind the kinds may differ too, and the
//     client sends `counter_kinds` beside them (#943).
//   - CounterCostVariable is "Remove X counters" / "any number"
//     (#789): the count is announced, CounterCostN is the FLOOR
//     rather than the amount, and CounterCostMax is the most the
//     payer could name right now — so the picker can bound its
//     stepper without a second round trip.
//   - CounterCostOptions is what could pay right now: permanents the
//     controller controls that match the clause (the source alone for
//     the self form) and hold enough counters, each with the kinds
//     that could pay — most counters first. Built from the
//     NON-targeting candidate walk, so a hexproof or shrouded
//     permanent of yours is offered: a cost does not target. Absent
//     when nothing can pay.
//   - CounterCostAdd / CounterCostAddKind are a cost that PUTS
//     counters on the source (Devoted Druid's -1/-1). Nothing is
//     chosen, so there is no option list; CounterAddBlocked is CR
//     118.3 saying the permanent can't have them, which is the one
//     reason such a cost is unpayable.
//
// The client sends the chosen permanents as `counter_source_ids`
// (omitted for the self form), the split as `counter_counts` (only
// for the among and variable forms) and, for the any-kind form, the
// chosen kind as `counter_kind` — or, when an any-kind among payment
// mixes kinds, one kind per permanent as `counter_kinds` (#943).
type CounterCostView struct {
	CounterCostN        int                     `json:"counter_cost_n,omitempty"`
	CounterCostKind     string                  `json:"counter_cost_kind,omitempty"`
	CounterCostSelf     bool                    `json:"counter_cost_self,omitempty"`
	CounterCostLabel    string                  `json:"counter_cost_label,omitempty"`
	CounterCostAmong    bool                    `json:"counter_cost_among,omitempty"`
	CounterCostVariable bool                    `json:"counter_cost_variable,omitempty"`
	CounterCostMax      int                     `json:"counter_cost_max,omitempty"`
	CounterCostOptions  []CounterCostOptionView `json:"counter_cost_options,omitempty"`
	CounterCostAdd      int                     `json:"counter_cost_add,omitempty"`
	CounterCostAddKind  string                  `json:"counter_cost_add_kind,omitempty"`
	CounterAddBlocked   bool                    `json:"counter_add_blocked,omitempty"`
}

// CounterCostOptionView is one permanent that could pay a "remove N
// counters" cost, and the counter kinds on it that could (#625).
type CounterCostOptionView struct {
	CardID string             `json:"card_id"`
	Kinds  []CounterKindCount `json:"kinds"`
}

// CounterKindCount is a counter kind and how many the permanent holds.
type CounterKindCount struct {
	Kind  string `json:"kind"`
	Count int    `json:"count"`
}

type ManaAbilityView struct {
	// Index is the 0-based position in the card's ability list;
	// what the activate_mana_ability payload carries.
	Index int `json:"index"`
	// Label is the human-readable menu entry ("Add {C}{C}",
	// "Add one mana of any color"). Empty falls back to the raw
	// Produced string on the client side.
	Label string `json:"label,omitempty"`
	// TapCost reflects the "{T}:" portion of the ability. Drives
	// the client's greying of the menu entry when the source is
	// tapped — or, for a creature source like Birds of Paradise,
	// summoning-sick (CardView.SummoningSick carries that).
	// SacrificeCost is a sacrifice-SELF cost (Treasure, Lotus
	// Petal); both went live in S21 sub-PR 1.
	TapCost       bool `json:"tap_cost,omitempty"`
	SacrificeCost bool `json:"sacrifice_cost,omitempty"`
	// ExileSelf is the "Exile this card from your hand" component of
	// a mana ability that functions from a hand (#1228) — the Spirit
	// Guides, and the whole printed family. The same wire name
	// ActivatedAbilityView carries it under (#1221), so a client that
	// renders a cost chip renders one chip for both ability kinds.
	//
	// Advisory, like every other cost flag on this view: the server
	// validates the zone and pays the exile.
	ExileSelf bool `json:"exile_self,omitempty"`
	// SacrificeLabel / SacrificeOptions describe a sacrifice-ANOTHER
	// cost — Ashnod's Altar's "Sacrifice a creature" — exactly as
	// ActivatedAbilityView carries them, so the client reuses one
	// picker for both ability kinds. Absent when the cost has no
	// such component. Stamped by stampActivatedAbilities, which is
	// the pass that has the game handle to compute a legal set.
	SacrificeLabel   string            `json:"sacrifice_label,omitempty"`
	SacrificeOptions *LegalTargetsView `json:"sacrifice_options,omitempty"`
	// LifeCost is a "Pay N life" component of the activation cost —
	// Mana Confluence's "{T}, Pay 1 life:". Advisory, exactly like
	// ActivatedAbilityView.LifeCost: the client renders the cost
	// chip, the server does the real CR 119.4 check. A damage RIDER
	// ("This land deals 1 damage to you") is not a cost and does not
	// appear here — it's part of the ability's Label.
	// Added in the S22 mana-ability-rider pass.
	LifeCost int `json:"life_cost,omitempty"`
	// ManaCost is a mana component of the activation cost — the
	// Signet cycle's "{1}, {T}", Cabal Coffers' "{2}, {T}".
	// Advisory, like LifeCost: the client renders the cost chip so
	// the player knows to float the mana first, and the server does
	// the real check. The engine deliberately does NOT auto-tap
	// into a mana ability, so an ability with this set can only be
	// fired against mana the player has already produced.
	// Added in the S32 mana-pipeline pass (#352).
	ManaCost string `json:"mana_cost,omitempty"`
	// ChargedManaCost is ActivatedAbilityView.ChargedManaCost for a
	// mana ability (#1191, #1190): CR 605.1a makes a mana ability an
	// activated ability, so Boom Scholar's discount reaches Loot, the
	// Pathfinder's "{G}, {T}" exactly as it reaches a CR 602 ability,
	// and the row says so through the same field name, the same
	// pointer-for-a-real-empty-answer shape, and the same rule: equal
	// to ManaCost absent a modifier, what the client shows in place of
	// ManaCost with ManaCost itself as the tooltip when they differ,
	// and nil (never a bare "") when the printed cost could not be
	// priced at all. Stamped by stampManaChargedCost, the pass with a
	// game handle.
	ChargedManaCost *string `json:"charged_mana_cost,omitempty"`
	// CounterCostView is the counter half of the activation cost —
	// Vivid Creek's charge counter, Ramos's five +1/+1 counters,
	// Mage-Ring Network's "any number of storage counters" (#789).
	// The SAME embedded view an activated ability carries, so the
	// client's picker, its greyed-row reason and its payload builder
	// are each written once.
	CounterCostView
	// DiscardCostN / Label / Options describe a "Discard N cards"
	// cost component on a MANA ability (#1213) — Skirge Familiar's
	// "Discard a card: Add {B}". Exactly the three fields and exactly
	// the wire names ActivatedAbilityView carries them under, so the
	// client's discard picker is one component for both ability
	// kinds, and the chosen cards go back as `discard_ids` either
	// way. Absent for every other mana ability, which is all of them.
	DiscardCostN       int      `json:"discard_cost_n,omitempty"`
	DiscardCostLabel   string   `json:"discard_cost_label,omitempty"`
	DiscardCostOptions []string `json:"discard_cost_options,omitempty"`
	// ExileCostN / Label / Options describe an "Exile N cards from your
	// hand" cost component (#1283) — Cadaverous Bloom's "Exile a card
	// from your hand: Add {B}{B} or {G}{G}". The discard triple's shape
	// under its own names, because it is its own component: an exiled
	// card is not discarded (no discard event, nothing for madness to
	// see). The chosen cards go back as `exile_ids`, and the client
	// skips its picker when the options are exactly ExileCostN.
	// Absent for every other mana ability.
	ExileCostN       int      `json:"exile_cost_n,omitempty"`
	ExileCostLabel   string   `json:"exile_cost_label,omitempty"`
	ExileCostOptions []string `json:"exile_cost_options,omitempty"`
	// ConditionUnmet is ActivatedAbilityView.ConditionUnmet for a
	// mana ability: true while the ability's "Activate only if …"
	// condition is false — Temple of the False God with four lands,
	// Mox Opal without metalcraft. Absent otherwise. Same closure the
	// engine gates on, evaluated for the controller and stamped by
	// stampActivatedAbilities (the pass with a game handle). Added
	// with #743 on the owner's decision to grey both kinds of row.
	ConditionUnmet bool `json:"condition_unmet,omitempty"`
	// Exhausted is ActivatedAbilityView.Exhausted for a mana ability
	// (#1183): an exhaust ability ("Activate each exhaust ability
	// only once") this object has already activated, so the engine
	// refuses it for the rest of this object's life. Loot, the
	// Pathfinder's "Exhaust — {G}, {T}: Add three mana of any one
	// color" is the one printed card.
	//
	// The SAME key and the same wire name the activated view uses, so
	// the client's row logic is one structural predicate for both
	// ability kinds and its ABILITY_EXHAUSTED string needed no
	// sibling. Absent for every other mana ability, which is all of
	// them.
	//
	// It is also the one greyed-row reason on a mana ability that the
	// AUTO-TAPPER honours: a spent exhaust source is not planned and
	// not counted as producible (CR 106.7), so the row being greyed
	// and the cast being priced agree.
	Exhausted bool `json:"exhausted,omitempty"`
	// AddsNoMana is CR 903.4f (#844): this ability's printed text
	// says "any color in your commander's color identity" and the
	// controller has no commander, or a commander whose colour
	// identity is colourless (Kozilek, Karn). The quality is
	// undefined or empty, so the ability adds no mana at all —
	// Command Tower taps for nothing. The client greys the row with
	// its own reason, the same way it greys ConditionUnmet; the
	// server does not refuse the activation (the ability exists, it
	// just does nothing), it simply stops offering it. Absent for
	// every other ability, which is all but four cards. Stamped by
	// stampManaIdentity, the pass with a game handle.
	AddsNoMana bool `json:"adds_no_mana,omitempty"`
	// CantActivate is ActivatedAbilityView.CantActivate for a mana
	// ability (#1210): the printed clause of a board-wide "can't be
	// activated" static that refuses this one. Cursed Totem's
	// "activated abilities of creatures can't be activated" does not
	// exempt mana abilities and reaches this row; Pithing Needle's
	// does exempt them and never will. Absent for every other mana
	// ability, which is all of them.
	//
	// The SAME key and the same wire name the activated view uses, so
	// the client's greyed-row logic is one structural predicate for
	// both ability kinds. Stamped by stampManaConditions, the pass
	// with a game handle.
	CantActivate string `json:"cant_activate,omitempty"`
	// Restrictions are the "spend this mana only on …" tags the
	// produced tokens will carry — Ancient Ziggurat, Eldrazi
	// Temple, the coloured half of Delighted Halfling. Present so
	// the client can warn before a player floats mana they cannot
	// spend on what they were about to cast. Purely informational;
	// enforcement is server-side, in the pool solver.
	// Added in the S32 mana-pipeline pass (#352).
	Restrictions []string `json:"restrictions,omitempty"`
	// Produced is the raw production string ("{C}{C}",
	// "{W|U|B|R|G}"). Lets the client render the produced-mana
	// pills alongside the activation button even when Label is
	// empty. EMPTY for a derived or scaled ability (Exotic Orchard,
	// Cabal Coffers), whose output only exists once computed at
	// activation — those carry the description in Label instead.
	Produced string `json:"produced,omitempty"`
}

// TurnView is the wire representation of the turn cursor. PriorityHolder
// is the seat index that currently holds priority within the step
// (S07+); it equals ActiveSeat at every step boundary and rotates on
// pass_priority.
type TurnView struct {
	Number         int    `json:"number"`
	ActiveSeat     int    `json:"active_seat"`
	PriorityHolder int    `json:"priority_holder"`
	Phase          string `json:"phase"`
	Step           string `json:"step"`
	// BlockDecisionSeats lists the seat indices that owe a
	// declare-blockers decision right now — under attack, with at
	// least one creature that could legally block one of the
	// attackers (CR 509.1a / 509.1b). Empty and omitted outside the
	// declare_blockers step.
	//
	// #328: blocking is a turn-based action, not a response, so the
	// client's "does this player have a legal response?" auto-pass
	// predicate could never see it and happily passed the defending
	// player's one chance to block. This field is what the client
	// consults to refuse to auto-pass the window. Public
	// information — attackers and untapped creatures are both on the
	// board — so it survives per-viewer filtering unredacted.
	BlockDecisionSeats []int `json:"block_decision_seats,omitempty"`
	// AttackTargets is the set of things the ACTIVE player's
	// creatures may be declared against right now (CR 506.2,
	// 508.1d) — the other seated players, the planeswalkers they do
	// not control, and the battles they do not protect. Present only
	// during the declare_attackers step, and only then because that
	// is the only step where it means anything.
	//
	// On the turn cursor rather than on each attacking creature: the
	// set is a property of the ATTACKING PLAYER, not of the creature
	// (summoning sickness, defender and tapped state gate the
	// creature separately), so hanging it off every card would be the
	// same list repeated once per attacker.
	//
	// Public information — every seat and every planeswalker on the
	// board is already visible — so it survives the per-viewer filter
	// unredacted. Added in S27.
	AttackTargets []AttackTargetView `json:"attack_targets,omitempty"`
}

// AttackTargetView is one legal attack target: what it is, and the id
// to send in declare_attacker's `target`. Kind is "player",
// "planeswalker" or "battle"; ID is a seat id for a player and an
// instance id for a permanent. Added in S27.
type AttackTargetView struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`

	// Tax is what ONE creature pays to attack this target under
	// CR 508.1a — "{2}" against a seat with Propaganda out, "{2}{2}"
	// against one with Propaganda and Ghostly Prison, "" when
	// attacking it is free. ADR 0080 (#1063).
	//
	// Priced by the engine for the ACTIVE seat, per creature, so the
	// client labels the control without re-deriving a rule — #429's
	// line, the one ADR 0045 §6 repeats: nothing in the client
	// re-derives who can attack, and nothing here re-derives what it
	// costs. A declaration's real price is this once per attacking
	// creature, concatenated.
	//
	// It is a per-target flat rate, which is exactly what every
	// printed card in the family charges. A hypothetical tax that
	// priced one creature differently from another would make this
	// field a lie, and the enumerator's per-move MoveCost.Mana — which
	// IS per creature — is the honest reading for a client that needs
	// one. The field would go then, rather than grow a caveat.
	Tax string `json:"tax,omitempty"`
}

// ViewOfGameFor builds a per-viewer wire snapshot. Same shape as
// ViewOfGame, plus per-card S13.5 KnownBy redaction:
//   - Every card in every zone has its KnownByYou flag set per the
//     viewer's KnownBy membership.
//   - Cards the viewer doesn't know have their printed
//     characteristics (name / type_line / scryfall_id / power /
//     toughness / counters / commander flag) zeroed out so the
//     wire doesn't leak identity. The instance ID, zone, owner,
//     controller, tapped state, position, face-down flag, and
//     damage_marked stay intact so the engine state is still
//     observable.
//   - Opponent hand + library are still rendered as backs by the
//     S04 zone-default heuristic (HandHasFilteredContents) — but
//     with KnownBy in place, individual hand cards revealed via
//     Thoughtseize-style effects will surface their characteristics
//     to the knowing viewer once the hand-zone filter is taught
//     to honour KnownBy. For S13.5 sub-PR 1, library + hand
//     redaction stays zone-wide; per-card knowledge of opponent
//     hand cards (the Thoughtseize case) lands as a follow-up.
//
// `viewerID` is the player UUID string; pass empty string for an
// admin / spectator session that sees everything (KnownByYou = true
// on every card). Added in S13.5.
func ViewOfGameFor(g *game.Game, viewerID string) GameView {
	view := ViewOfGame(g)
	return FilterViewFor(view, viewerID)
}

// ViewOfGame builds a wire snapshot from a game.Game. It acquires a
// read lock on the game via ReadSnapshot and reads every field it
// needs in one consistent pass. Safe to call from any goroutine.
func ViewOfGame(g *game.Game) GameView {
	var view GameView
	g.ReadSnapshot(func() {
		view = GameView{
			ID:          g.ID.String(),
			State:       string(g.State),
			Seats:       viewOfSeats(g, g.Seats),
			Battlefield: viewOfZone(g.Battlefield),
			Stack:       viewOfZone(g.Stack),
			Exile:       viewOfZone(g.Exile),
			PhasedOut:   viewOfZone(g.PhasedOut),
			Turn: TurnView{
				Number:         g.Turn.Number,
				ActiveSeat:     g.Turn.ActiveSeat,
				PriorityHolder: g.Turn.PriorityHolder,
				Phase:          string(g.Turn.Phase),
				Step:           string(g.Turn.Step),
				// #328: who still owes a block declaration. Read
				// surface takes the lock we already hold and the
				// layers ReadSnapshot just refreshed.
				BlockDecisionSeats: g.SeatsOwingBlockDecisionLocked(),
			},
			MulligansOpen:     g.MulligansOpen,
			Monarch:           uuidStringOrEmpty(g.Monarch),
			Initiative:        uuidStringOrEmpty(g.Initiative),
			Promises:          viewOfPromises(g.Promises),
			Vote:              viewOfVote(g.Vote),
			UndoLimit:         g.Settings.UndoLimit,
			Settings:          viewOfTableSettings(g.Settings),
			StartingSeat:      g.StartingSeat,
			StackItems:        viewOfStackItemsInStackOrder(g),
			PendingTriggers:   viewOfStackItemSlice(g.PendingTriggers),
			DelayedTriggers:   viewOfDelayedTriggers(g.DelayedTriggers),
			SplitSecondActive: g.SplitSecondActive,
			DiscardPending:    viewOfDiscardPending(g.DiscardPending),
			PendingChoices:    viewOfPendingChoices(g),
			LoopNotice:        viewOfLoopNotice(g.LoopNotice),
		}
		// ADR 0066. Asked ONCE per frame and threaded down, not per
		// seat and not per card: the answer is "nobody may play this"
		// in almost every game, it costs a walk of the battlefield,
		// and a bot's decision loop pays for every frame it builds.
		anyGrant := g.AnyCastPermissionsForEffect()
		stampLegalTargets(g, view.Seats, anyGrant)
		if anyGrant {
			// Exile has no owner, so no seat's answer is the public
			// one and every holder is filed per seat (uuid.Nil skips
			// nobody).
			stampGrantedPermissions(g, view.Seats, &view.Exile, g.Exile, uuid.Nil, 0)
			for si := range view.Seats {
				id := mustParseSeatID(view.Seats[si].ID)
				seat := g.PlayerByIDForEffect(id)
				if seat == nil {
					continue
				}
				// The pile's OWNER is skipped in both: their own cast
				// out of their own graveyard or library is the public
				// answer stampLegalTargets has already stamped.
				stampGrantedPermissions(g, view.Seats, &view.Seats[si].Graveyard, seat.Graveyard, id, 0)
				// CR 401.5: a library is a cast surface for exactly
				// one card, its top, which is the LAST element.
				stampGrantedPermissions(g, view.Seats, &view.Seats[si].Library, seat.Library,
					id, len(view.Seats[si].Library.Cards)-1)
			}
		}
		stampLibraryTop(g, view.Seats)
		stampActivatedAbilities(g, &view.Battlefield)
		stampZoneAbilities(g, view.Seats, &view.Exile)
		stampCombatTargets(g, &view)
		stampNoUntap(g, &view.Battlefield)
		view.legalBySeat = enumerateLegalMoves(g)
		// S31 sub-PR 0: the public log resolves card names and knower
		// sets out of the view that was just assembled, so it must run
		// last — and inside the same read lock, so the log and the
		// board it describes come from one consistent read.
		view.Log = publicLogOf(g, &view)
		// S22: the reveal window resolves printed identity out of the
		// same assembled view, for the same reason and under the same
		// lock. Unlike the log it needs no knower sets — see
		// reveal_frame.go.
		view.Reveals = publicRevealsOf(g, &view)
	})
	return view
}

// enumerateLegalMoves runs the legal-move enumerator once per seat
// and returns the result keyed by player UUID string. Runs under the
// read lock ViewOfGame already holds, hence EnumerateLocked.
//
// Every seat, not just the one holding priority, and deliberately so:
// "who owes a decision right now" is a question the enumerator
// already answers — a seat with neither priority nor a pending choice
// nor a combat declaration gets an empty list and costs two map
// lookups to find that out. Re-deriving the predicate here would be
// a second copy of the rule, and the one case it would get wrong is
// the expensive one: a defender declaring blockers holds no priority
// (the active player still does), and blocking is the window a
// client must never be left guessing about (#328).
//
// A nil map is fine — FilterViewFor reads it with a comma-less index
// and gets nil back for every seat.
func enumerateLegalMoves(g *game.Game) map[string][]LegalMoveView {
	var out map[string][]LegalMoveView
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		moves := capLegalMoves(legal.EnumerateLocked(g, p.ID, legal.Options{}))
		if len(moves) == 0 {
			continue
		}
		if out == nil {
			out = make(map[string][]LegalMoveView, len(g.Seats))
		}
		out[p.ID.String()] = moves
	}
	return out
}

// legalMovesWireCap bounds how many moves one seat's list may put on
// the wire before it degrades.
//
// The enumerator's own cap is PER SOURCE (legal.Options
// MaxExpansionPerSource, 12), which bounds one card and not the
// board. A seat with a Lightning Bolt, two sacrifice outlets and a
// modal spell out is four crossed products, and the measured cost of
// a single expanded source on a four-player table is ~5 KB — so
// "capped" without a global bound is not actually capped. This is
// that bound.
//
// 48 lands the worst realistic frame just under the 24 KiB the
// protocol tests budget for the field. The bot is unaffected either
// way: internal/aiseat enumerates in-process at full fidelity and
// never reads this projection.
const legalMovesWireCap = 48

// capLegalMoves degrades an over-long move list instead of truncating
// it, because truncation would be a correctness bug rather than a
// size one: the client's question is "does this card have a move?",
// and dropping the tail would grey out a card that is perfectly
// playable — the exact class of "greyed in the UI, accepted by the
// server" bug this whole sub-PR exists to kill.
//
// So the degraded list keeps the FIRST move of every (source, kind,
// targets_stack) tuple and drops only the alternatives. Every card
// that had a move still has one; what is lost is the choice between
// its twelve targets, which no client consumes today (targeting is
// driven by CardView.legal_targets, and targeting.ts stays the
// presentation layer for it). TargetsStack rides along in the key
// because it is NOT an alternative target of the same shape: a modal
// card with a counter mode and a burn mode is two different moves
// that happen to share a source and a kind, and keeping only the
// burn one would silently delete the counterspell response smart
// autopass needs to see. docs/protocol.md states this as part of the
// field's contract.
func capLegalMoves(moves []LegalMoveView) []LegalMoveView {
	if len(moves) <= legalMovesWireCap {
		return moves
	}
	type key struct {
		source       uuid.UUID
		kind         legal.Kind
		targetsStack bool
	}
	seen := make(map[key]bool, len(moves))
	out := make([]LegalMoveView, 0, legalMovesWireCap)
	for _, m := range moves {
		k := key{m.Source, m.Kind, m.TargetsStack}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, m)
	}
	return out
}

// stampLegalTargets walks the PER-SEAT cast surfaces — hand, the
// command zone, the graveyard and the top of the library — and calls
// castStampsFor for each card the seat that owns the zone could
// cast from it. Runs under the read lock ViewOfGame already holds;
// the per-viewer filter strips the fields from opponents' hands.
//
// Which cards are surfaces differs per zone, and that is all this
// function decides. Hand and the command zone are surfaces for every
// card that sits in them, so they are walked unconditionally (S29).
// The graveyard is a surface only for the cards whose own text says
// so (flashback, escape, Gravecrawler) or that a permission opens
// (ADR 0066), so it is gated per card and stamps CastableHere on the
// ones that pass. The library is a surface for exactly one card, its
// top, and only under a permission (S42, CR 401.5).
//
// THE SEAT THIS PASS ASKS ABOUT IS THE SEAT THAT OWNS THE PILE, and
// since #1055 its answer is split rather than published whole. The
// PUBLIC half — the prices claimable out of this zone, the modes, the
// printed clause that refuses the cast — is a fact about a card in a
// public zone and goes on every viewer's copy. The half that answers
// "may YOU cast this" (`castable_here`) and "what may YOU target"
// (`legal_targets`, `clauses`) is filed under the owner's own seat and
// reaches their frame alone, exactly as a foreign holder's does. A
// permission somebody ELSE holds over a card sitting here is stamped
// by stampGrantedPermissions under their seat — including when the
// owner may cast the card too, which is the case a single holder
// could not express.
//
// EXILE IS NOT HERE, and could not be: it is a shared top-level zone
// with no seat to hang a per-seat walk off, so every one of its
// answers is private and stampGrantedPermissions files them all
// (#978). All of them call the same castStampsFor, so the answer
// cannot differ by zone.
func stampLegalTargets(g *game.Game, seats []PlayerView, anyGrant bool) {
	for si := range seats {
		seat := &seats[si]
		caster, err := uuid.Parse(seat.ID)
		if err != nil {
			continue
		}
		live := g.PlayerByIDForEffect(caster)
		zones := []struct {
			view *ZoneView
			kind game.ZoneKind
			live *game.Zone
		}{
			{&seat.Hand, game.ZoneHand, nil},
			{&seat.Command, game.ZoneCommand, nil},
			{&seat.Graveyard, game.ZoneGraveyard, zoneOf(live, game.ZoneGraveyard)},
			// S42 / CR 401.5: the library is a cast surface for its
			// top card when a permission opens it. Walked for every
			// seat and gated per card below, exactly as the graveyard
			// is — the per-viewer filter drops the whole zone for a
			// library the viewer may not see into.
			{&seat.Library, game.ZoneLibrary, zoneOf(live, game.ZoneLibrary)},
		}
		for _, zone := range zones {
			// CR 401.5: a library is a cast surface for exactly one
			// card, the top one, and the top is the LAST element.
			// Walking the rest would be both wrong and the most
			// expensive loop in the view.
			first := 0
			if zone.kind == game.ZoneLibrary {
				if !anyGrant || len(zone.view.Cards) == 0 {
					continue
				}
				first = len(zone.view.Cards) - 1
			}
			for ci := first; ci < len(zone.view.Cards); ci++ {
				c := &zone.view.Cards[ci]
				if zone.kind == game.ZoneGraveyard || zone.kind == game.ZoneLibrary {
					// The card's own text, or a granted permission
					// (ADR 0066) — the same merge the cast path makes,
					// so the client can never render a button
					// CastSpell would refuse.
					//
					// #1012: the gate is the PERMISSION, not the offer it
					// prices. A grant that charges a flat price — Bolas's
					// Citadel's life, an impulse grant that charges the
					// printed cost — synthesises no claimable offer, and
					// reading the offer instead of the permission skipped the
					// card here and left stampGrantedPermissions to paint
					// `castable_here` back on with no clauses behind it. One
					// question, asked once: may this seat cast this card out
					// of this zone.
					//
					// THE ZONE'S OWNER and nobody else: this pass
					// computes the PUBLIC answer, which is the one the
					// pile's owner gets out of their own pile. A
					// permission somebody ELSE holds over a card sitting
					// here is that seat's private answer and is stamped
					// by stampGrantedPermissions, which already walks
					// these piles asking whose permission covers each
					// card (#1022, #1037).
					//
					// Both questions are asked of the ENGINE's card
					// rather than of the projection (#1171). viewOfZone
					// builds a pile card-for-card and in order — the
					// alignment stampGrantedPermissions already relies
					// on — so the live card is the one beside it, and
					// reading it is what lets the gate below ask about
					// a FACE the pile is not showing.
					var live *game.Card
					if zone.live != nil && ci < len(zone.live.Cards) {
						live = &zone.live.Cards[ci]
					}
					var grant *game.CastPermission
					if anyGrant && live != nil {
						grant = grantedCast(g, caster, *live, zone.kind)
					}
					// #1171: EVERY face a cast may choose, through the
					// one predicate the bot enumerator reads
					// (game.CardCastableFromAnyFace). This used to ask
					// `c.oracleID`, which is the BARE oracle ID and so
					// resolves to face 0's catalog entry whatever the
					// card's other halves declare — while
					// legal/cast.go asked every face. A card whose BACK
					// face prints flashback was therefore a legal move
					// for a bot and a card with no announce stamps at
					// all on the wire: no `castable_here`, no price
					// list, no target clause, and no cast button behind
					// a cast CastSpell would have accepted.
					if grant == nil && (live == nil || !game.CardCastableFromAnyFace(*live, zone.kind)) {
						continue
					}
					// `castable_here` is NOT set here (#1015):
					// castStampsFor derives it from the price list and
					// the cast gate, in the one place both are known.
					//
					// #1055: and it is filed for the OWNER rather than
					// written to the public fields. The owner's answer
					// to "may you cast this" is still computed here and
					// is still the only one this pass knows, but it is
					// a statement about a PLAYER, so it reaches that
					// player's frame alone (applyCastStampsFor). What
					// stays public is the half that is a statement
					// about the CARD IN THIS ZONE — its price list, its
					// modes, the clause refusing the cast — which every
					// player may read off a card in a public zone.
					s := castStampsFor(g, caster, c, activeFace(c), zone.kind, grant)
					s.applyPublicTo(c, zone.kind)
					c.stampsFor(caster, s)
					stampCastableFaces(g, caster, c, zone.kind, grant, true)
					continue
				}
				// Hand and the command zone, by the same rule
				// (#1166). They took the pre-#1055 path until now —
				// one call that wrote the ZONE OWNER's whole announce
				// surface, `legal_targets` included, straight into the
				// exported fields — and that was invisible for almost
				// every card, because an opponent's hand is full of
				// cards they are not a knower of and
				// redactCardForViewer clears the lot.
				//
				// It stops being invisible for a REVEALED one. A
				// Thoughtseize or a Telepathy makes the viewer a
				// knower, keepKnownInHandZone keeps the card, and
				// their copy then carried the hand owner's legal
				// target set for a spell only that seat may cast — the
				// #891 shape one zone over from the graveyard #1055
				// fixed. A target set is narrowed by hexproof, shroud,
				// protection and "target opponent", so seat A's list
				// is not seat B's to read, wherever the card is
				// sitting.
				//
				// `castable_here` was never part of this: castStampsFor
				// only ever sets it for a graveyard or a library top,
				// so a hand card has never carried one.
				s := castStampsFor(g, caster, c, activeFace(c), zone.kind, nil)
				s.applyPublicTo(c, zone.kind)
				c.stampsFor(caster, s)
				// #992: and the halves this hand card is NOT showing.
				// A hand is where an adventure card and a modal DFC
				// are actually cast from, so this is the walk the face
				// picker reads — and it is one extra face for those
				// two layouts and zero for every other card in the
				// game.
				stampCastableFaces(g, caster, c, zone.kind, nil, true)
			}
		}
	}
}

// castHolder is one seat that may cast a card out of the zone it is
// sitting in, and the permission that says so.
type castHolder struct {
	seat  uuid.UUID
	grant *game.CastPermission
}

// castHoldersOf answers "who may cast this card out of this zone",
// and with which permission — EVERY seat that may, not the first
// (#1037).
//
// ADR 0066 makes a CastPermission a statement about an OBJECT — "you
// may cast that card" — and nothing in it says the object has to be in
// the holder's own zone. Wrexial's "you may cast target instant or
// sorcery card from that player's graveyard" and Xanathar's "you may
// play the top card of their library" are the printed shapes. Until
// #1022 the view asked `grantedCast` for the ZONE's owner and no other
// seat, so such a permission reached the wire as nothing at all; until
// #1037 it asked the other seats only until one said yes, so a card
// two seats may both cast carried one of the two answers and the other
// holder got the public zone and no picker.
//
// `skip` is the seat whose answer is already the PUBLIC one — the
// pile's owner for a graveyard or a library, nobody (uuid.Nil) for
// exile, which has no owner to be public on behalf of.
//
// In SEAT ORDER, so the frame is deterministic; nothing depends on the
// order any more, because each holder's answer is filed under their
// own key.
//
// Caller must hold g.mu.
func castHoldersOf(g *game.Game, seats []PlayerView, skip uuid.UUID, card game.Card, kind game.ZoneKind) []castHolder {
	var out []castHolder
	for i := range seats {
		id, err := uuid.Parse(seats[i].ID)
		if err != nil || (skip != uuid.Nil && id == skip) {
			continue
		}
		if grant := grantedCast(g, id, card, kind); grant != nil {
			out = append(out, castHolder{seat: id, grant: grant})
		}
	}
	return out
}

// castStamps is the announce-time cast surface ONE seat is offered on
// ONE card out of ONE zone — every field of a CardView whose value
// depends on WHO is asking, in one value that can be computed for
// several seats and handed to the right one later (#1037).
//
// The field list is the same one stripCastOffersNotFor used to clear
// by hand, in one place now: a stamp added to castStampsFor without a
// field here would be a stamp that never reached the holder, and a
// field here that castStampsFor does not fill is cleared on the way
// in, which is what makes applyTo safe to run over a card that was
// stamped publicly first.
type castStamps struct {
	// CastSurfaceView is THE field list, and it is the wire's own
	// (#992). It used to be spelled out here a second time, which was
	// the drift this type existed to end one level down: a stamp
	// added to castStampsFor without a field here was a stamp that
	// never reached the holder. Embedding the wire block instead
	// means applyTo and applyToFace are each one assignment and
	// neither can forget a field.
	CastSurfaceView

	// ExilePlay is the grant these stamps were computed under, for a
	// card in a zone a permission opens. `exile_play` is PUBLIC —
	// the trigger that created the permission resolved in the open —
	// and stampGrantedPermissions stamps one for everybody; this is
	// the holder's own, which is the one their client must read,
	// because the public one names whichever live permission came
	// first and two seats may hold two.
	ExilePlay *ExilePlayView

	// ZoneAbilities is the seat's CR 602 activation rows for this
	// card in this zone (#1221) — cycling out of a hand, unearth out
	// of a graveyard, whatever ActivatedAbilityShape.Zones names.
	//
	// It rides this per-seat carrier rather than the exported field
	// for the reason `castable_here` and `legal_targets` do, and it
	// is the same sentence one surface over: an ability row answers
	// "what may YOU announce about this card, from here", and its
	// target sets are narrowed by hexproof, shroud and "target
	// opponent" — one seat's list is not another's to read (#1055).
	// Through #660 the field was written publicly and the HAND's own
	// secrecy hid it; a graveyard is public, so the hand's rows moved
	// here with the graveyard's rather than leaving two lifecycles.
	//
	// Never in publicIn's half: a bystander reads the card's printed
	// text off the pile in paper and gets the ability from there.
	// Written by stampZoneAbilities, which merges into whatever
	// castStampsFor already filed for this seat rather than replacing
	// it — the two passes answer different questions about the same
	// card.
	ZoneAbilities []ActivatedAbilityView

	// ZoneManaAbilities is the seat's CR 605 MANA activation rows for
	// this card in this zone (#1228) — a Spirit Guide's "Exile this
	// card from your hand: Add {R}". It rides the per-seat carrier
	// beside ZoneAbilities and for the same sentence: an activation
	// row answers "what may YOU announce about this card, from here".
	//
	// Written by the same pass, through the same merge.
	ZoneManaAbilities []ManaAbilityView
}

// applyTo writes one seat's answer onto the card they will receive.
// Wholesale, including the zero values: these fields have exactly one
// writer per card per zone, so "the holder has no modes" must clear a
// public answer that did.
func (s castStamps) applyTo(c *CardView) {
	c.CastSurfaceView = s.CastSurfaceView
	if s.ExilePlay != nil {
		c.ExilePlay = s.ExilePlay
	}
	// #1221: wholesale, zero value included, for the reason the
	// surface above is — this seat's answer is the whole answer, and
	// "you have no rows here" has to clear anything a public pass
	// wrote. Nothing writes it publicly today; assigning rather than
	// guarding on nil is what keeps that true if something ever does.
	c.ZoneAbilities = s.ZoneAbilities
	// #1228: the mana half, the same way and for the same reason.
	c.ZoneManaAbilities = s.ZoneManaAbilities
}

// applyToFace is applyTo for ONE PRINTED FACE of the card (#992) —
// the same answer written to the same field names, one level down.
//
// No ExilePlay half: a grant is a permission over an OBJECT, not over
// a face of one, so `exile_play` stays on the card where both client
// readers already look for it. The face a grant NAMES is expressed by
// which faces carry a block at all.
func (s castStamps) applyToFace(f *CardFaceView) {
	f.CastSurfaceView = s.CastSurfaceView
}

// applyPublicTo writes the half of one seat's answer that EVERY viewer
// legitimately sees, and leaves the rest to applyCastStampsFor (#1055).
//
// The line is "is this a fact about the card in this zone, or about a
// player". A card in a graveyard or on top of a revealed library is a
// card every player may read in paper, so what it prints — the prices
// claimable out of this zone, the modes it chooses among, its
// additional and optional costs, its convoke clause, its Phyrexian
// symbols, and the printed clause that refuses the cast (a Rule of Law
// on the battlefield is on the battlefield) — is public, and the pile
// owner is simply the seat the view computes it for.
//
// Three fields are not:
//
//   - CastableHere answers "may YOU cast this from here". It has as
//     many answers as there are seats — the owner's printed flashback,
//     a Wrexial holder's grant over the same card, and "no" for
//     everyone else — and shipping one seat's as a public bit is the
//     surface #1055 is about: the name says "castable here" and the
//     value meant "castable here by somebody else", so both client
//     readers had to pair it with `exile_play` to find out which.
//   - LegalTargets and Clauses are the same kind of answer one step
//     further in: hexproof, shroud, protection and "target opponent"
//     all narrow a target set by WHO IS ASKING, so seat A's list is
//     not seat B's to read. applyCastStampsFor has said so about a
//     spectator since #978; this is the same sentence applied to the
//     bystander a public stamp used to reach.
//
// And the same sentence again inside the two public fields that carry
// legal sets of their own — `modes[i]` and `alternative_costs[i]`
// (#1172). The field stays, the lists inside it do not; publicIn
// spells out which is which.
//
// A seat with an answer of its own gets all three back wholesale when
// FilterViewFor promotes their entry.
//
// A HAND's public half is narrower than a graveyard's, and `kind` is
// how this function knows (#1169) — see publicIn.
func (s castStamps) applyPublicTo(c *CardView, kind game.ZoneKind) {
	s.publicIn(kind).applyTo(c)
}

// applyPublicToFace is applyPublicTo one level down (#992): the same
// split, applied to a face's block. A face's legal target set is one
// seat's answer for exactly the reason the card's is, and a face of a
// card in a HAND is as unreadable as the card is (#1169).
func (s castStamps) applyPublicToFace(f *CardFaceView, kind game.ZoneKind) {
	s.publicIn(kind).applyToFace(f)
}

// publicIn is the half of one seat's answer that EVERY viewer of a
// card in `kind` legitimately sees. One place, so the card's split and
// the face's cannot disagree about which fields those are.
//
// THREE FIELDS ARE NEVER PUBLIC, in any zone: `castable_here`,
// `legal_targets` and `clauses` each answer "what may YOU announce"
// (#1055).
//
// A HAND IS NOT A PUBLIC ZONE the way a graveyard is, and that is the
// zone-publicity half (#1169). A revealed card — Thoughtseize,
// Telepathy — is ONE card the viewer has been shown, not a pile they
// may read, and the cost-shaped announce fields are documented as
// "the viewer's own hand". One of them is also a live leak of the
// rest of that hand: an offer with a pitch cost (Force of Will's
// "exile a blue card from your hand") carries `pay_options`, which is
// a list of instance IDs out of the hand the viewer was shown exactly
// one card of.
//
// keepKnownInHandZone used to clear these from a hand-rolled list of
// field names in the per-viewer filter, which is the second list in a
// second place that castStamps exists to end — and being a list of
// what to REMOVE it had gone stale twice over: it never covered
// `optional_costs` (ADR 0073), and since #992 it never covered the
// per-face blocks at all, so a revealed adventure card handed a
// knower its owner's whole per-face price list. handPublicCastSurface
// is an ALLOWLIST instead, so a field added to CastSurfaceView
// tomorrow is private in a hand until somebody says otherwise.
//
// THE SAME LINE RUNS ONE LEVEL DOWN (#1172). `modes` and
// `alternative_costs` stay public on a pile because they are what the
// CARD PRINTS — "choose two of three", "Overload {4}{R}" — and a
// bystander at a paper table reads them off a graveyard card. Each of
// them also carries board-derived lists computed for the seat the
// stamp was built for, and those are one seat's answer for exactly the
// reason the top-level `legal_targets` is: hexproof, shroud,
// protection and "target opponent" narrow a target set by who is
// asking, and "a blue card in YOUR hand" names one seat's cards. So
// the nested per-viewer fields ride the SAME per-seat castStamps /
// castOffers split as the top-level ones — the public projection keeps
// the mode's and the offer's static facts, the asking seat's frame
// gets the lists — and the rule that decides which is which is "does
// this field come off the printed card, or off the board for one
// player".
func (s castStamps) publicIn(kind game.ZoneKind) castStamps {
	s.CastableHere = false
	s.LegalTargets = nil
	s.Clauses = nil
	// #1221: an activation row is the same kind of answer as
	// `castable_here` — "what may YOU announce from here" — and it
	// carries legal sets of its own. None of it is public.
	s.ZoneAbilities = nil
	// #1228: and the mana rows beside them. A mana ability names no
	// targets, but the row still answers "what may YOU announce", and
	// the day the graveyard joins supportedManaAbilityZones is the
	// day a public pile would otherwise start shipping it.
	s.ZoneManaAbilities = nil
	s.Modes = publicModeSpec(s.Modes)
	s.AlternativeCosts = publicAlternativeCosts(s.AlternativeCosts)
	if kind == game.ZoneHand {
		s.CastSurfaceView = handPublicCastSurface(s.CastSurfaceView)
	}
	return s
}

// publicModeSpec is the half of a modal card's clause every viewer of
// a public pile may read (#1172): the prompt, how many options to
// pick, whether they repeat, and each option's printed label and
// target-prompt shape. Each option's `legal_targets` and `clauses` are
// left to the per-seat stamp, which is where the same two fields live
// one level up.
//
// It COPIES rather than blanking in place, and that is the whole
// reason it is a function rather than two lines in publicIn. A
// castStamps is passed by value, but `Modes` is a POINTER and
// `Options` a slice header, both shared with the stamp stampsFor files
// for the asking seat — the public projection and that seat's answer
// come from ONE castStampsFor call (stampLegalTargets). Blanking
// through the pointer would take the nested sets off the owner's own
// frame as well as off the bystander's, which is the bug
// applyFaceCastStampsFor copies the face slice to avoid, arriving one
// level further in.
//
// One allocation per modal card per public stamp, and none at all for
// the overwhelming majority of cards, which are not modal.
func publicModeSpec(ms *ModeSpecView) *ModeSpecView {
	if ms == nil {
		return nil
	}
	out := *ms
	out.Options = make([]ModeOptionView, len(ms.Options))
	for i, o := range ms.Options {
		o.LegalTargets = nil
		o.Clauses = nil
		out.Options[i] = o
	}
	return &out
}

// publicAlternativeCosts is publicModeSpec for the offer list (#1172):
// every price the cast may claim out of this zone, with each offer's
// printed half — key, label, mana cost, life, the pay clause's copy,
// CR 107.3b's X lock and CR 107.4's Phyrexian count — and without the
// two lists that are one seat's.
//
//   - `legal_targets` is the target clause the spell has WHEN THIS
//     COST IS PAID, resolved against the board through
//     LegalTargetsForEffect. Cleave's wider clause narrowed by
//     hexproof is the asking seat's answer, exactly as the card's own
//     clause is.
//   - `pay_options` is the card-shaped half of the cost: the blue
//     cards in the CASTER's hand for a pitch cost, the Islands THEY
//     control for Daze, the cards in THEIR graveyard for escape. Not
//     a target list — a cost does not target (CR 601.2h) — but a list
//     of instance IDs picked out by "you control" / "your hand" /
//     "your graveyard" predicates, so it answers "what may YOU pay"
//     and belongs on the asking seat's frame. #1169 already said so
//     about a hand, where the whole offer list is dropped; this is
//     the same sentence on a public pile, where it is not.
//
// Copies for the reason publicModeSpec does: the slice header is
// shared with the seat's own stamp.
func publicAlternativeCosts(offers []AlternativeCostView) []AlternativeCostView {
	if len(offers) == 0 {
		return nil
	}
	out := make([]AlternativeCostView, len(offers))
	for i, o := range offers {
		o.LegalTargets = nil
		o.PayOptions = nil
		out[i] = o
	}
	return out
}

// handPublicCastSurface is the announce surface of a card in a HAND
// that a viewer who is not its owner may read (#1169).
//
// Written as an allowlist — a fresh block carrying the fields that
// stay, rather than a list of the fields that go — because the
// question a new announce field has to answer is "may somebody who
// was SHOWN this card read it", and the safe default for a field
// nobody has thought about is no. The four that stay are the three
// the line has always kept plus `target_mode`, and they are the ones
// that are not cost-shaped:
//
//   - `target_mode` is the shape of the card's target prompt, which
//     is its printed text — and it is already public on a revealed
//     card's FACES (#992), so clearing it here would have made the
//     card and its faces disagree.
//   - `additional_cost` and `optional_costs` are printed clauses
//     whose pickers read the battlefield, which is public.
//   - `cant_cast` is a Rule of Law on the battlefield, visible to
//     everybody in exactly the same words.
//
// The list of Go field names is pinned by the reflection guard in
// face_down_view_test.go, which fails on a field of CastSurfaceView
// this function has not placed.
func handPublicCastSurface(s CastSurfaceView) CastSurfaceView {
	return CastSurfaceView{
		TargetMode:     s.TargetMode,
		AdditionalCost: s.AdditionalCost,
		OptionalCosts:  s.OptionalCosts,
		CantCast:       s.CantCast,
	}
}

// stampsFor files one seat's answer on the card for FilterViewFor to
// hand out. Never the exported fields: those are the public answer,
// and a per-viewer answer shipped publicly is the thing #1015 took off
// this surface.
func (c *CardView) stampsFor(seat uuid.UUID, s castStamps) {
	if c.castOffers == nil {
		c.castOffers = make(map[string]castStamps, 1)
	}
	c.castOffers[seat.String()] = s
}

// stampZoneAbilitiesFor files one seat's CR 602 activation rows for
// this card in this zone (#1221), MERGING into whatever answer that
// seat already has rather than replacing it.
//
// Merging is the whole reason this is not a stampsFor call. The two
// passes that write a seat's entry answer different questions about
// one card — stampLegalTargets and stampGrantedPermissions say what
// the seat may CAST out of this pile, this says what it may ACTIVATE
// there — and they run in that order over the same map. A plain
// stampsFor here would hand a flashback card's own announce surface
// back as a zero value the moment the card also printed a graveyard
// ability, which is the #544 failure mode arriving through the view
// instead of the enumerator.
//
// Empty rows file nothing: the overwhelming majority of cards in a
// graveyard print no ability that functions there, and an entry per
// card per seat would be one map allocation per graveyard card per
// frame for an answer that is always nil.
func (c *CardView) stampZoneAbilitiesFor(seat uuid.UUID, rows []ActivatedAbilityView) {
	if len(rows) == 0 {
		return
	}
	if c.castOffers == nil {
		c.castOffers = make(map[string]castStamps, 1)
	}
	key := seat.String()
	s := c.castOffers[key]
	s.ZoneAbilities = rows
	c.castOffers[key] = s
}

// stampZoneManaAbilitiesFor is stampZoneAbilitiesFor for the CR 605
// half (#1228), with the same merge and the same empty-rows
// short-circuit — and the short-circuit matters more here, because
// almost every card in a hand has a mana ability list that is empty
// once the zone predicate has run over it.
func (c *CardView) stampZoneManaAbilitiesFor(seat uuid.UUID, rows []ManaAbilityView) {
	if len(rows) == 0 {
		return
	}
	if c.castOffers == nil {
		c.castOffers = make(map[string]castStamps, 1)
	}
	key := seat.String()
	s := c.castOffers[key]
	s.ZoneManaAbilities = rows
	c.castOffers[key] = s
}

// castFace names the printed half ONE cast surface is computed for
// (#992): its index on the card, the catalog key that half registers
// under (ADR 0034's "<oracle_id>#N") and the mana cost it prints.
//
// It is one parameter rather than three because the three always
// travel together and getting two of them from one face and the third
// from another is the bug it exists to prevent: Stomp's target clause
// priced against Bonecrusher Giant's {2}{R} would count the wrong
// Phyrexian symbols and size the wrong convoke budget.
type castFace struct {
	index    int
	key      string
	manaCost string
}

// activeFace is the castFace for the half the view is SHOWING — every
// stamp that existed before #992. A pile's cards are front-up (CR
// 712.8, MoveCard), so for all of them this is face 0 and the key is
// the bare oracle ID, exactly as the pre-#992 call sites passed.
func activeFace(c *CardView) castFace {
	return castFace{
		index:    c.ActiveFace,
		key:      game.CatalogKeyForFace(c.oracleID, c.ActiveFace),
		manaCost: c.ManaCost,
	}
}

// castFaceOf is activeFace for a face the card is NOT showing, read
// off the projected face list.
func castFaceOf(c *CardView, i int) castFace {
	return castFace{
		index:    i,
		key:      game.CatalogKeyForFace(c.oracleID, i),
		manaCost: c.Faces[i].ManaCost,
	}
}

// stampCastableFaces files one caster's answer for every face of a
// multi-face card that a cast may actually choose, and returns
// nothing for the ~33,000 single-faced oracle IDs, which carry no
// `faces` on the wire at all (#992).
//
// `public` also writes the half every viewer sees. True for the piles
// stampLegalTargets walks on their OWNER's behalf — a graveyard, a
// library top, a hand, a command zone — and false for a foreign
// holder's stamps, where nothing about the answer is public because
// the seat it was computed for is not the seat the pile belongs to.
//
// The face list is game.CastableFacesUnder, which is the same answer
// the bot enumerator walks (legal/cast.go): both halves of a modal
// DFC and of an adventure card, the front alone of a transform card,
// and exactly the faces a grant names when one does. A view that
// offered a face the enumerator does not would be a picker row
// CastSpell refuses with ErrInvalidFace.
//
// Caller must hold g.mu.
func stampCastableFaces(g *game.Game, caster uuid.UUID, c *CardView, kind game.ZoneKind, grant *game.CastPermission, public bool) {
	if len(c.Faces) < 2 {
		return
	}
	live, ok := liveCardForView(g, c)
	if !ok {
		return
	}
	for _, i := range game.CastableFacesUnder(live, grant, caster) {
		if i < 0 || i >= len(c.Faces) {
			continue
		}
		s := castStampsFor(g, caster, c, castFaceOf(c, i), kind, grant)
		if public {
			s.applyPublicToFace(&c.Faces[i], kind)
		}
		c.Faces[i].stampsFor(caster, s)
	}
}

// stampsFor is CardView.stampsFor for one printed face (#992).
func (f *CardFaceView) stampsFor(seat uuid.UUID, s castStamps) {
	if f.castOffers == nil {
		f.castOffers = make(map[string]castStamps, 1)
	}
	f.castOffers[seat.String()] = s
}

// castStampsFor fills the announce-time clauses ONE card offers ONE
// caster out of ONE zone: the target mode, the modes, the additional
// cost, the tap cost, the X notes, the alternative costs and the legal
// target set. It RETURNS the answer rather than writing it, so the
// same computation serves the public stamp, the zone owner's own and
// each foreign holder's.
//
// The one body, called once per zone (#978). It used to be inlined in
// stampLegalTargets' innermost loop, which walked hand, command, the
// graveyard and the library top — and so an exiled card that a
// CastPermission let its owner cast carried `target_mode` and
// `mana_cost` and nothing else, and the client's cast chain had to
// guess. Exile is a SHARED zone, so it could not simply be added to
// that per-seat loop; stampGrantedPermissions calls this instead, for
// the holder the engine named.
//
// `key` is the catalog key to read the card's clauses off, rather
// than the card's bare oracle ID, because a permission can name a
// FACE (ADR 0034): a defeated Siege's back face is a different entry
// with different modes and a different target clause, and the exile
// pile is showing the front.
//
// `grant` is the CastPermission that opens this zone for this caster,
// or nil. It is NOT a price: since #1012 the price question has one
// answer for the whole engine — game.CastOffersForLocked — and this
// function reads it rather than building a second list. The view used
// to stamp the grant's offer first and let a printed offer for the
// same zone overwrite the WHOLE slice, so a Gravecrawler in the
// graveyard under an Underworld Breach showed only what the card
// prints and the Breach's escape offer was gone from the picker even
// though resolveAlternativeCostLocked would have accepted it.
//
// Caller must hold g.mu.
func castStampsFor(g *game.Game, caster uuid.UUID, c *CardView, f castFace, kind game.ZoneKind, grant *game.CastPermission) castStamps {
	key := f.key
	// #662 / #979: the source of a SPELL is the spell itself, so every
	// legal-target stamp below names the card rather than only its
	// caster — a creature with protection from red is off the picker
	// for a red spell and on it for a white one. Built here, once per
	// card, so every zone builds it the same way; a card cast off
	// somebody else's grant is still its own source.
	var out castStamps
	src := castSourceOf(g, caster, c.InstanceID)
	// The engine's own card, faced the way the cast path would face
	// it (ADR 0034): a permission that names a face opens that face
	// and no other, so the gate below and the price list further down
	// both judge the half being cast rather than the one the pile
	// happens to be showing. `f` names that half — its index, its
	// catalog key and its printed cost — which the caller has already
	// resolved.
	//
	// SetFace before grantedFace, and both: #992 prices a face the
	// card is NOT showing, which is the other direction from a grant
	// naming one, and a grant that names a face still wins. SetFace
	// to the face already up is a no-op, so this is the same call the
	// pre-#992 line made for every stamp that has ever existed.
	//
	// A card the lookup cannot find is in no zone at all, which a
	// zone walk cannot produce; it stamps no gate and no offers
	// rather than guessing at either.
	live, haveLive := liveCardForView(g, c)
	if haveLive {
		live.SetFace(f.index)
		live = grantedFace(live, grant)
	}
	// S14: the announce-time target kind, read off the catalog entry
	// of the half being cast. Stamped HERE since #992 rather than
	// only in viewOfCard, because viewOfCard knows one face — the one
	// that is up — and a face picker needs the answer for the other.
	out.TargetMode = game.TargetModeFor(key)
	if ms := game.ModeSpecFor(key); ms != nil {
		out.Modes = viewOfModeSpec(g, src, ms)
	}
	if ac := game.AdditionalCostFor(key); !ac.Empty() {
		out.AdditionalCost = &AdditionalCostView{
			DiscardCards: ac.DiscardCards,
			DemandsX:     ac.PayLifeX,
			Label:        ac.Label,
		}
		if ac.Sacrifice != nil {
			// SpecCandidatesForEffect, not LegalTargetsForEffect: an
			// additional sacrifice cost doesn't target, so the
			// hexproof / shroud gate must not narrow the list the
			// client offers. The same list, count and order the
			// abilities ship (#747).
			out.AdditionalCost.SacrificeOptions = sacrificeCostOptions(g, caster, ac.Sacrifice, uuid.Nil, false)
		}
	}
	// ADR 0073: the optional costs this card OFFERS. Stamped next to
	// the mandatory one and read by the same modal the alternative
	// costs open, because CR 601.2b announces all of them together.
	out.OptionalCosts = viewOfOptionalCosts(g, caster, src, key)
	// S22: convoke / waterbend. Stamped before the target clause
	// because the caster pays it first, and the count of a "X target
	// creatures" clause depends on what they paid.
	if tc := game.TapPermanentsCostFor(key); !tc.Empty() {
		out.TapCost = viewOfTapCost(g, caster, f.manaCost, tc)
	}
	// #746: the printed clauses of a per-target price, for the X
	// picker's note.
	out.TargetCostNotes = game.TargetPricedCostClauses(key)
	// #760, ADR 0073 §7: the one cast gate's view stamp. The SAME
	// function CastSpell and the bot enumerator call, so the client
	// can never render a cast button the server would refuse.
	//
	// The zone the card sits in is the zone a cast would come out of,
	// which is what makes Grafdigger's Cage answerable here at all —
	// and, since #978, answerable for exile and the library top as
	// well, because they reach this function too.
	if haveLive {
		if err := g.CastGateLocked(caster, live, kind, game.CastSpellParams{}); err != nil {
			// A card the gate refuses is not a cast surface, whatever
			// opened the zone — but the bit itself is derived below,
			// once, out of this refusal and the price list together.
			out.CantCast = cantCastReason(err)
		}
	}
	// #916: the ceiling on the cast's `phyrexian_life`. Read off the
	// EFFECTIVE cost — the commander tax is generic and cost
	// modifiers add generic, so the two agree today, and reading the
	// effective one keeps them agreeing if that ever stops being
	// true.
	out.PhyrexianSymbols = phyrexianSymbolsIn(f.manaCost)
	spec := game.TargetSpecFor(key)
	// #1012: THE list of CR 118.9 prices this cast may claim out of
	// this zone, from the one function that answers the question for
	// the engine, the bot enumerator and the view. A nil entry is the
	// printed mana cost, present only when a cast that claims nothing
	// would be accepted from here.
	//
	// The view used to answer it twice and get three different
	// answers: AlternativeCostsOfferedFromZone for the printed
	// offers, a wholesale overwrite of the grant's offer, and
	// silence about whether the printed cost was on the menu at all.
	// Stamped before the early-out below, because a card can offer a
	// price without having any target clause of its own.
	var offers []*game.AlternativeCost
	if haveLive {
		offers = g.CastOffersForLocked(caster, live, kind, grant)
	}
	out.AlternativeCosts = viewOfAlternativeCosts(g, caster, src, c.InstanceID, f.manaCost, spec, offers)
	// #1012: and the wire says so when the printed cost is not one of
	// them. `castable_here` is one bit and means "you may cast this
	// from here", never "you may cast this from here for the cost in
	// the corner" — a Faithless Looting in the graveyard is castable
	// at its flashback cost and at no other, and the client had to
	// infer that from the shape of the offer list.
	out.AlternativeCostRequired = len(offers) > 0 && !printedCostAmong(offers)
	// #1015: and THE cast-surface bit, derived here because this is
	// the one place that knows both halves of it — the prices this
	// cast may claim, and the ADR 0073 §7 gate. An empty price list
	// is a real answer: a card whose only path out of this zone is an
	// offer the caster cannot pay (an escape card in a graveyard too
	// small to pay for it) is not a cast surface, and marking one was
	// a button the announce path refused with ErrCastCostRequired.
	//
	// Hand and the command zone never carry the bit: every card in
	// them is a cast candidate and an always-true flag would be noise
	// the client had to ignore. Exile keys its button off `exile_play`
	// instead, whose per-viewer stamps this bit — public since S29 —
	// is not part of.
	//
	// #1195: and the THIRD input, game.CastTimingOpenLocked — the one
	// CR 307.1 read CastSpell and the bot enumerator also call. Until
	// it was added, a flashback SORCERY in a graveyard was marked a
	// cast surface in an opponent's end step and the announce path
	// refused it with ErrSorcerySpeedRequired; a Vedalken Orrery on
	// the board was invisible to this bit in the other direction. The
	// predicate folds the card's own timing, the permission's ADR 0066
	// override, the per-player grants and the per-player restrictions
	// in CR 101.2's order, so the client cannot render a cast button
	// out of a rule it reimplemented.
	switch kind {
	case game.ZoneGraveyard, game.ZoneLibrary:
		out.CastableHere = out.CantCast == "" && len(offers) > 0 &&
			haveLive && g.CastTimingOpenLocked(caster, live, kind, grant)
	}
	if spec == nil {
		return out
	}
	out.LegalTargets = viewOfLegalTargets(g.LegalTargetsForEffect(src, spec), spec)
	out.Clauses = viewOfClauses(g, src, spec)
	return out
}

// phyrexianSymbolsIn counts CR 107.4's "or 2 life" symbols in a cost
// string — the ceiling on the `phyrexian_life` an announcement may
// claim (#916). One parse, three readers: a castable card, an
// alternative cost offer and an activated ability, so the client is
// never handed two different answers to the same question.
//
// An unparseable cost counts zero rather than erroring: the cast or
// activation is refused with ErrUnparseableCost long before any
// payment is announced (#289), so there is no life half to offer and
// nothing for the view to say about it.
func phyrexianSymbolsIn(costStr string) int {
	if costStr == "" {
		return 0
	}
	cost, err := game.ParseCost(costStr)
	if err != nil {
		return 0
	}
	return cost.PhyrexianSymbols()
}

// ProtectionView is one "protection from <quality>" on a permanent,
// parsed by game.ProtectionQualities so nothing downstream has to
// read the token. #662, CR 702.16.
type ProtectionView struct {
	// Printed is the quality as the CARD prints it — "red",
	// "Demons", "artifacts", "everything". The badge tooltip's text,
	// which is why the engine's token keeps the card's own spelling
	// and plural instead of a canonical singular.
	Printed string `json:"printed"`
	// Kind is which characteristic of a source the quality is
	// compared against: "color", "card_type", "subtype",
	// "everything" or "player". Stable tokens; see
	// game.ProtectionQualityKind.
	Kind string `json:"kind"`
	// Value is what the rules actually compare — the wire colour
	// ("R"), the lowercase card type ("artifact"), the canonical
	// singular subtype ("Demon"). Empty for "everything", which
	// compares nothing.
	//
	// For "player" (CR 702.16k, #980) it is the chosen seat's id, and
	// the comparison is against the SOURCE'S CONTROLLER rather than
	// against any characteristic of it. An id and not a name: this is
	// the rules value, the same shape CardView.controller carries, and
	// the client resolves seats to names the way it already does.
	// `printed` stays "the chosen player" — the display string never
	// holds a UUID.
	//
	// Empty for a player quality whose permanent has not been answered
	// yet, which reads correctly as "protected from nobody".
	Value string `json:"value,omitempty"`
}

// viewOfProtection projects a card's parsed protections. Nil for the
// overwhelming majority of cards, which have none.
func viewOfProtection(c *game.Card) []ProtectionView {
	qs := game.ProtectionQualities(c)
	if len(qs) == 0 {
		return nil
	}
	out := make([]ProtectionView, 0, len(qs))
	for _, q := range qs {
		v := ProtectionView{
			Printed: q.Printed,
			Kind:    q.Kind.String(),
			Value:   q.Value,
		}
		// CR 702.16k: the player quality's rules value is a seat, and
		// it lives on the quality rather than in the token because the
		// token names no seat. ProtectionQualities has already
		// resolved it off the permanent (#980).
		if q.Player != uuid.Nil {
			v.Value = q.Player.String()
		}
		out = append(out, v)
	}
	return out
}

func viewOfLegalTargets(lt game.LegalTargets, spec *game.TargetSpec) *LegalTargetsView {
	view := &LegalTargetsView{Min: spec.Min, Max: spec.Max, CountFromX: spec.CountFromX, Distinct: spec.Distinct}
	for _, id := range lt.Players {
		view.Players = append(view.Players, id.String())
	}
	for _, id := range lt.Cards {
		view.Cards = append(view.Cards, id.String())
	}
	return view
}

// viewOfTapCost projects a card's convoke / waterbend clause: the
// untapped permanents the caster may tap and the cap on how many.
//
// The cap is computed from the card's printed cost rather than from
// the cost the cast will actually owe, because the alternative cost
// and the commander tax are not chosen yet when a hand snapshot is
// built. It is an affordance, not the gate — the server re-derives
// the real budget at announce and rejects an over-tap there.
//
// Caller must hold g.mu.
func viewOfTapCost(g *game.Game, caster uuid.UUID, manaCost string, tc *game.TapPermanentsCost) *TapCostView {
	v := &TapCostView{
		Key:         tc.Key,
		Label:       tc.Label,
		ColorClause: tc.ColorClause,
		DemandsX:    tc.DemandsX(),
	}
	// SpecCandidatesForEffect, not LegalTargetsForEffect: tapping a
	// permanent for convoke or waterbend doesn't target it, so a
	// shrouded creature you control is still a legal tapper. The
	// engine's own check (tap_cost.go) already works this way.
	opts := viewOfLegalTargets(g.SpecCandidatesForEffect(caster, tc.Spec), tc.Spec)
	opts.Cards = filterToController(g, opts.Cards, caster)
	opts.Players = nil
	opts.Min, opts.Max, opts.CountFromX = 0, 0, false
	v.Options = opts
	v.Max = game.TapPermanentsBudgetFor(tc, manaCost, 0)
	return v
}

// viewOfWaterbend projects a waterbend clause that is not a spell's —
// an activated ability's (#1310) or a pay-unless prompt's (#1311) —
// as the TapCostView a hand card's clause already uses. The options
// come from game.WaterbendOptionsForEffect, the walk the engine
// validates against, so the picker never offers a permanent the
// server refuses; `exclude` is a permanent another component of the
// same cost already spends.
//
// Caller must hold g.mu.
func viewOfWaterbend(g *game.Game, player, exclude uuid.UUID, wb *game.TapPermanentsCost, budget int) *TapCostView {
	return &TapCostView{
		Key:      wb.Key,
		Label:    wb.Label,
		DemandsX: wb.DemandsX(),
		Max:      budget,
		Options:  &LegalTargetsView{Cards: cardIDStrings(g.WaterbendOptionsForEffect(player, exclude, wb))},
	}
}

// viewOfAlternativeCosts projects a card's "you may cast this for
// its overload / evoke / cleave cost" offers, each already resolved
// against the target clause it would leave the spell with (S22).
// `base` is the card's own clause, or nil when it has none.
//
// Resolving the rewrite here rather than on the client is what lets
// the picker be dumb: an entry with no target_mode casts straight
// away, an entry with one enters targeting on the legal set it
// carries, and neither case needs the client to know what the word
// "overload" means. Caller must hold g.mu.
// `self` is the instance ID of the card the offers belong to, which
// escape needs and nothing else does: "exile five OTHER cards from
// your graveyard" is paid out of the same zone the spell is being
// cast from, so the spell itself is sitting in the candidate list
// until it is filtered out here. The server rejects it anyway (CR
// 601.2a moved it to the stack before the cost is paid), but a
// picker that offers a card the cast will be rejected for choosing
// is a trap rather than an affordance.
//
// `offers` is game.CastOffersForLocked's answer verbatim (#1012), so
// this function no longer decides WHICH offers exist — only how each
// one looks on the wire. The nil entry in that list is the printed
// mana cost, which is not an alternative cost and is projected as
// `alternative_cost_required` instead.
func viewOfAlternativeCosts(g *game.Game, caster uuid.UUID, src game.TargetSource, self, printedCost string, base *game.TargetSpec, offers []*game.AlternativeCost) []AlternativeCostView {
	// A nil result rather than a present-and-empty one: `absent`
	// is what the field means for a card with no offers, and every
	// card in every cast surface reaches this function since #1012.
	var out []AlternativeCostView
	for i := range offers {
		if offers[i] == nil {
			// The printed mana cost. Not an alternative cost, and
			// `alternative_cost_required` is where the wire says
			// whether it is claimable from this zone.
			continue
		}
		ac := *offers[i]
		// S28 asked the offer's Condition here and #695 widened it to
		// the whole of CR 601.2b — "a greyed-out button the server
		// would reject is worse than no button". Neither test lives
		// here any more: CastOffersForLocked applies both, through
		// the same AlternativeCostPayableLocked the announce
		// validator and the bot enumerator read, and a second copy
		// of the filter is exactly the drift #1012 was about.
		v := AlternativeCostView{
			Key: ac.Key, Label: ac.Label, ManaCost: ac.ManaCost,
			Life: ac.Life, PayLabel: ac.PayLabel,
			// CR 107.3b (#831). An offer is a cost; the rule asks
			// only whether the printed {X} survives into it, so the
			// pair of strings IS the whole question here. A grant
			// that charges a FLAT price rather than making an offer —
			// cascade's {0}, airbend's {2} — never reaches this
			// function; `exile_play.x_locked_at_zero` is where the
			// same predicate answers for those.
			XLockedAtZero: game.CastCost{Printed: printedCost, Paid: ac.ManaCost}.LocksXAtZero(),
			// #916: the offer replaces the mana cost, so it replaces
			// the "or 2 life" count the client's stepper is bounded
			// by.
			PhyrexianSymbols: phyrexianSymbolsIn(ac.ManaCost),
		}
		if spec := game.TargetSpecUnderAlternativeCost(base, &ac); spec != nil {
			v.TargetMode = spec.Mode
			v.LegalTargets = viewOfLegalTargets(g.LegalTargetsForEffect(src, spec), spec)
		}
		// The card-shaped half. SpecCandidatesForEffect, not
		// LegalTargetsForEffect, for the same reason the additional
		// cost's sacrifice picker uses it: a cost does not target, so
		// the hexproof / shroud gate must not narrow what the client
		// offers.
		if paySpec := ac.ExileFromHand; paySpec != nil {
			opts := viewOfLegalTargets(g.SpecCandidatesForEffect(caster, paySpec), paySpec)
			opts.Players = nil
			v.PayOptions = opts
		} else if paySpec := ac.ReturnToHand; paySpec != nil {
			opts := viewOfLegalTargets(g.SpecCandidatesForEffect(caster, paySpec), paySpec)
			opts.Cards = filterToController(g, opts.Cards, caster)
			opts.Players = nil
			v.PayOptions = opts
		} else if paySpec := ac.ExileFromGraveyard; paySpec != nil {
			// The spec's own predicate already narrows this to the
			// caster's graveyard; what it cannot know is which card
			// is being cast, hence the `self` filter. Min / Max ride
			// through from the spec, so the client's picker sizes
			// itself to "exile five" without knowing the keyword.
			opts := viewOfLegalTargets(g.SpecCandidatesForEffect(caster, paySpec), paySpec)
			opts.Cards = withoutID(opts.Cards, self)
			opts.Players = nil
			v.PayOptions = opts
		}
		out = append(out, v)
	}
	return out
}

// printedCostAmong reports whether the printed mana cost is one of
// the prices this cast may claim — the nil entry game.
// CastOffersForLocked puts at the head of its list when a cast that
// claims nothing would be accepted from the zone (#1012).
//
// A named function rather than a loop at the call site because the
// nil-means-printed convention is the one thing a reader of the offer
// list has to know, and it should be spelled out once.
func printedCostAmong(offers []*game.AlternativeCost) bool {
	for _, offer := range offers {
		if offer == nil {
			return true
		}
	}
	return false
}

// viewOfOptionalCosts projects a card's "you may pay an additional
// cost" offers (CR 601.2b, ADR 0073) for one caster. Nil for nearly
// every card.
//
// Next to viewOfAlternativeCosts because the client asks both
// questions in one modal: an alternative cost REPLACES the mana cost
// and an optional one ADDS to it, so the two compose and CR 601.2b
// announces them together.
//
// The sacrifice pool goes through sacrificeCostOptions, the same
// helper the mandatory cost uses, so a non-mana kicker's picker is
// the picker every other sacrifice cost opens — filtered to the
// caster's own permanents (CR 701.21a), in payment order. A
// present-and-empty list is how the client knows the offer cannot be
// taken right now, exactly as it is for the mandatory cost.
//
// ADR 0089: a gift offer also carries who may receive it, and any
// optional cost that rewrites the target clause carries the rewritten
// clause's legal set — `src` is the spell as its own source (#662),
// exactly as the alternative-cost stamp reads it.
func viewOfOptionalCosts(g *game.Game, caster uuid.UUID, src game.TargetSource, key string) []OptionalCostView {
	costs := game.OptionalCostsFor(key)
	if len(costs) == 0 {
		return nil
	}
	out := make([]OptionalCostView, 0, len(costs))
	for i, oc := range costs {
		v := OptionalCostView{
			Index:        i,
			Key:          oc.Key,
			Label:        oc.Label,
			ManaCost:     oc.ManaCost,
			MaxTimes:     oc.MaxPayments(),
			DiscardCards: oc.DiscardCards,
		}
		if oc.Sacrifice != nil {
			v.SacrificeOptions = sacrificeCostOptions(g, caster, oc.Sacrifice, uuid.Nil, false)
		}
		if oc.ChoosesOpponent {
			v.ChoosesOpponent = true
			// Nobody left to choose serialises as an absent list
			// beside chooses_opponent, which the client greys.
			for _, id := range g.GiftOpponentsLocked(caster) {
				v.OpponentOptions = append(v.OpponentOptions, id.String())
			}
		}
		if spec := oc.Targets; spec != nil {
			v.TargetMode = spec.Mode
			v.LegalTargets = viewOfLegalTargets(g.LegalTargetsForEffect(src, spec), spec)
			v.Clauses = viewOfClauses(g, src, spec)
		}
		out = append(out, v)
	}
	return out
}

// viewOfModeSpec projects a modal card's options with each targeted
// option's legal set from the caster's point of view. Caller must
// hold g.mu.
func viewOfModeSpec(g *game.Game, src game.TargetSource, ms *game.ModeSpec) *ModeSpecView {
	out := &ModeSpecView{
		Prompt:     ms.Prompt,
		Min:        ms.Min,
		Max:        ms.Max,
		Repeatable: ms.Repeatable,
		Options:    make([]ModeOptionView, 0, len(ms.Options)),
	}
	for _, o := range ms.Options {
		ov := ModeOptionView{Label: o.Label, Cost: o.Cost}
		if o.Targets != nil {
			ov.TargetMode = o.Targets.Mode
			ov.LegalTargets = viewOfLegalTargets(g.LegalTargetsForEffect(src, o.Targets), o.Targets)
			ov.Clauses = viewOfClauses(g, src, o.Targets)
		}
		out.Options = append(out.Options, ov)
	}
	return out
}

// viewOfClauses projects a multi-clause statement's clauses, each
// with its own legal set and bounds. Nil for a single-clause
// statement, where `legal_targets` alone is the whole answer and
// every pre-#764 client path keeps working. Caller must hold g.mu.
func viewOfClauses(g *game.Game, src game.TargetSource, spec *game.TargetSpec) []LegalTargetsView {
	if spec.ClauseCount() < 2 {
		return nil
	}
	out := make([]LegalTargetsView, 0, spec.ClauseCount())
	for i := 0; i < spec.ClauseCount(); i++ {
		c := spec.Clause(i)
		v := viewOfLegalTargets(g.LegalTargetsForEffect(src, c), c)
		v.Label = c.Label
		out = append(out, *v)
	}
	return out
}

// stampActivatedAbilities fills CardView.ActivatedAbilities for
// every battlefield permanent, computed from its CONTROLLER's point
// of view (they're the only player who can activate it). A
// permanent's abilities are public information in Magic, so unlike
// hand-card legal targets these are not stripped per viewer — the
// client simply doesn't offer the menu on permanents you don't
// control. Runs under the read lock ViewOfGame already holds.
func stampActivatedAbilities(g *game.Game, bf *ZoneView) {
	for i := range bf.Cards {
		c := &bf.Cards[i]
		// #521: this guard used to be `c.oracleID == ""`, which
		// dropped EVERY token's activated abilities on the way to the
		// client — Food, Clue, Blood and the Lander reached the table
		// with no way to crack them, because they carried their
		// ability on the instance for want of an oracle ID and this
		// line read the want of one as "nothing to project". A token
		// has a catalog key of its own now, so the loop simply looks
		// the engine card up and asks the ability accessor, which
		// answers for a token and a printed card in the same words;
		// a permanent with no abilities at all gets a nil list from
		// it, which is what the wire wants anyway.
		//
		// Treasure escaped the old guard only by accident: its
		// ability is a MANA ability, and viewOfManaAbilities is
		// stamped unconditionally in viewOfCard. The four mana
		// stampers below were skipped for it all the same.
		controller, err := uuid.Parse(c.Controller)
		if err != nil {
			continue
		}
		instanceID, err := uuid.Parse(c.InstanceID)
		if err != nil {
			continue
		}
		card, ok := g.LookupCardForEffect(instanceID)
		if !ok {
			continue
		}
		c.ActivatedAbilities = viewOfActivatedAbilities(g, card, controller, game.ZoneBattlefield)
		// ADR 0082 decision 9: a FACE-DOWN permanent carries its
		// CR 116.2g "turn face up" row, from the same projection the
		// hand's foretell and suspend rows come from and with the
		// same server-computed `available`.
		//
		// The "you" is the CONTROLLER (CR 708.6), and it leaks
		// nothing: the row quotes the morph cost, which names the
		// card, and redactCardForViewer clears special_actions for
		// every non-knower — of which a face-down permanent has all
		// but one (CR 708.5).
		//
		// Every other permanent on the board gets a nil list: the
		// offer is derived from the face-down kind, so
		// SpecialActionsOfferedByCard answers nothing for a face-up
		// permanent, and the declared kinds are hand keywords.
		c.SpecialActions = viewOfSpecialActions(g, card, controller)
		c.LoyaltyActivated = g.LoyaltyActivatedThisTurn[instanceID]
		stampManaSacrificeOptions(g, card, controller, c.ManaAbilities)
		stampManaConditions(g, card, controller, c.ManaAbilities)
		stampManaIdentity(g, card, controller, c.ManaAbilities)
		stampManaCounterCosts(g, card, controller, c.ManaAbilities)
		stampManaChargedCost(g, card, controller, c.ManaAbilities)
	}
}

// stampZoneAbilities fills CardView.ZoneAbilities for every card in
// every NON-BATTLEFIELD zone an ability can function from (CR 113.6):
// a seat's own hand (cycling and typecycling, CR 702.29), their own
// graveyard (unearth CR 702.82a, scavenge CR 702.96a, embalm
// CR 702.128a), their command zone, and their cards in exile.
// #660 / ADR 0062 Decision 5, widened by #1221.
//
// The view's half of the same walk the legal enumerator makes
// (legal.enumerator.abilityZones) and the activation path accepts, so
// the rows a client can see, the moves a bot is offered and the
// activations the engine allows are one answer computed three times
// from one predicate rather than three answers (#544). The LIBRARY is
// missing from all three for the same reason: it is hidden, and no
// printed ability functions from one.
//
// Mirrors stampActivatedAbilities and is split from viewOfCard for
// the same reason: the cost projections (which cards in hand could
// pay a discard clause, which targets are legal) need a game handle,
// and viewOfCard has one card.
//
// The "you" is the card's OWNER, not a controller: a card outside the
// battlefield and the stack has no controller (CR 108.4), which is
// the same rule ActivateCatalogAbility reads off Card.Owner. For the
// per-seat piles that is the seat; for exile, the one shared pile
// holding every seat's cards, it is read off the card.
//
// The rows go to that seat ALONE, through castOffers (#1055): a
// graveyard is public but an ability row is not, because its target
// sets are narrowed by who is asking. Through #660 they were written
// to the exported field and the hand's own secrecy did the hiding;
// that stops working the moment the pile is one everybody can read.
// SpecialActions stays on the exported field and stays hand-only —
// foretell and suspend are announced out of a hand and nowhere else,
// so the zone redaction that has always hidden them still does.
//
// Runs under the read lock ViewOfGame already holds.
func stampZoneAbilities(g *game.Game, seats []PlayerView, exile *ZoneView) {
	for si := range seats {
		seat := &seats[si]
		owner, err := uuid.Parse(seat.ID)
		if err != nil {
			continue
		}
		for ci := range seat.Hand.Cards {
			c := &seat.Hand.Cards[ci]
			card, ok := liveCardForAbilityRows(g, c)
			if !ok {
				continue
			}
			c.stampZoneAbilitiesFor(owner, viewOfActivatedAbilities(g, card, owner, game.ZoneHand))
			// #1228: the CR 605 rows beside the CR 602 ones. The same
			// pass, because it is the same question about the same
			// card — "what may you activate here" — and a second walk
			// of every hand would be a second answer to keep in step.
			c.stampZoneManaAbilitiesFor(owner, viewOfManaAbilitiesFromZone(card, game.ZoneHand))
			c.SpecialActions = viewOfSpecialActions(g, card, owner)
		}
		stampZoneAbilitiesInPile(g, &seat.Graveyard, game.ZoneGraveyard, owner)
		stampZoneAbilitiesInPile(g, &seat.Command, game.ZoneCommand, owner)
	}
	// Exile last, and with no seat of its own: it is one shared pile
	// and each card's owner is its "you" (CR 108.4), so the walk reads
	// the owner off the card rather than off the loop.
	stampZoneAbilitiesInPile(g, exile, game.ZoneExile, uuid.Nil)
}

// stampZoneAbilitiesInPile is stampZoneAbilities' per-pile body.
//
// `owner` is the seat every card in the pile belongs to, or uuid.Nil
// for a shared pile whose cards belong to different seats — exile is
// the one such pile, and passing Nil is what makes the walk read
// CardView.Owner instead of assuming one.
//
// Runs under the read lock ViewOfGame already holds.
func stampZoneAbilitiesInPile(g *game.Game, zone *ZoneView, kind game.ZoneKind, owner uuid.UUID) {
	if zone == nil {
		return
	}
	for ci := range zone.Cards {
		c := &zone.Cards[ci]
		you := owner
		if you == uuid.Nil {
			parsed, err := uuid.Parse(c.Owner)
			if err != nil {
				continue
			}
			you = parsed
		}
		card, ok := liveCardForAbilityRows(g, c)
		if !ok {
			continue
		}
		c.stampZoneAbilitiesFor(you, viewOfActivatedAbilities(g, card, you, kind))
		// #1228: nil for every pile today — game.supportedManaAbilityZones
		// is the hand alone, so the zone predicate answers no here —
		// and the call is made anyway so that adding a zone to that
		// list is adding it in one place rather than two.
		c.stampZoneManaAbilitiesFor(you, viewOfManaAbilitiesFromZone(card, kind))
	}
}

// liveCardForAbilityRows resolves a projected card back to the
// engine's own Card so the ability rows can be computed off it, and
// declines the ones that could not have any.
//
// The empty oracle ID is the fast negative and the correctness one at
// once: a card with no catalog key has no catalog abilities to
// project, and since ADR 0069 an object CR 708.2a has silenced has
// the empty key too — so a face-down card in exile offers no rows
// without this function knowing what face-down means.
func liveCardForAbilityRows(g *game.Game, c *CardView) (game.Card, bool) {
	if c.oracleID == "" {
		return game.Card{}, false
	}
	instanceID, err := uuid.Parse(c.InstanceID)
	if err != nil {
		return game.Card{}, false
	}
	return g.LookupCardForEffect(instanceID)
}

// viewOfSpecialActions projects the CR 116.2 special actions a card
// offers, with the engine's own per-kind timing answer stamped on
// each (ADR 0062 Decision 4).
//
// Two callers and two "you": a card in its owner's HAND, where the
// foretell and suspend rows live, and a face-down BATTLEFIELD
// permanent, where the turn-face-up row does and the "you" is its
// controller (CR 708.6, ADR 0082 decision 9). One projection for
// both, off one engine accessor, so a row the client can see is a row
// the engine would accept.
//
// A kind the engine cannot carry out is not projected at all: the
// client must never show a row the server would refuse.
//
// Runs under the read lock ViewOfGame already holds.
func viewOfSpecialActions(g *game.Game, card game.Card, owner uuid.UUID) []SpecialActionView {
	var out []SpecialActionView
	for _, sa := range game.SpecialActionsOfferedByCard(card) {
		if !game.SpecialActionKindBuilt(sa.Kind) {
			continue
		}
		view := SpecialActionView{
			Kind:      string(sa.Kind),
			Label:     sa.Label,
			Cost:      sa.Cost,
			Available: g.SpecialActionTimingOKLocked(owner, card, sa.Kind),
		}
		// #1319: the charged price, after every CR 601.2f cost
		// modifier on the battlefield — Ranar the Ever-Watchful's
		// foretell discount. Left nil on a pricing error, exactly as
		// ActivatedAbilityView.ChargedManaCost does, so the client
		// falls back to Cost rather than showing a stale charge.
		if sa.Cost != "" {
			if charged, err := g.SpecialActionManaCostForEffect(owner, card, sa.Kind, sa); err == nil {
				s := charged.String()
				view.ChargedCost = &s
			}
		}
		out = append(out, view)
	}
	return out
}

// stampManaConditions sets ManaAbilityView.ConditionUnmet for every
// mana ability whose activation condition is false right now (#743):
// the closure ActivateManaAbility gates on, with the controller as
// "you". Split from viewOfManaAbilities for the reason the sacrifice
// options are — that projection has no game handle. Caller must hold
// g's read lock.
func stampManaConditions(g *game.Game, card game.Card, controller uuid.UUID, views []ManaAbilityView) {
	raw := game.ManaAbilitiesForCard(card)
	// #1210: the board-wide "can't be activated" fast negative, taken
	// once per card rather than once per row.
	restricted := g.AnyActivationRestrictionsForEffect()
	for i := range views {
		if i >= len(raw) {
			continue
		}
		// #1210, CR 602.5a: the board-wide gate's reason, from the
		// one function the engine, the enumerator and the auto-tapper
		// all call. Mana: true, and the RESTRICTION decides what that
		// means — Cursed Totem reaches a Birds of Paradise's {G},
		// Pithing Needle never does.
		if restricted {
			views[i].CantActivate = g.CantActivateReasonLocked(controller, card, game.ZoneBattlefield,
				game.ActivationAbility{Label: raw[i].Label, Mana: true})
		}
		// #1183: the exhaust flag, stamped beside the condition
		// because the client reads them off the same row and greys
		// with the same code — and separately from it, because the
		// two recover differently (a condition may hold again next
		// turn; an exhaust only if the permanent becomes a new
		// object). The one reader the engine, the enumerator and the
		// auto-tapper all use.
		views[i].Exhausted = g.ManaAbilityExhausted(controller, card.InstanceID, raw[i])
		if raw[i].Condition == nil {
			continue
		}
		views[i].ConditionUnmet = !raw[i].Condition(g, controller, card.InstanceID)
	}
}

// stampManaIdentity sets ManaAbilityView.AddsNoMana for every mana
// ability that CR 903.4f leaves with nothing to add (#844): Command
// Tower, Arcane Signet, Commander's Sphere or Path of Ancestry under a
// controller with no commander or a colourless one. Same question the
// engine answers when the ability fires, asked here so the client can
// grey the row instead of letting a player tap a land for no mana.
// Split from viewOfManaAbilities for the reason the conditions are:
// that projection has no game handle. Caller must hold g's read lock.
func stampManaIdentity(g *game.Game, card game.Card, controller uuid.UUID, views []ManaAbilityView) {
	raw := game.ManaAbilitiesForCard(card)
	for i := range views {
		if i >= len(raw) {
			continue
		}
		views[i].AddsNoMana = game.ManaAbilityAddsNoMana(g, controller, card.InstanceID, raw[i])
	}
}

// stampManaChargedCost sets ManaAbilityView.ChargedManaCost for every
// mana ability whose own activation cost has a mana component (#1191,
// #1190): CR 605.1a makes a mana ability an activated ability, so the
// CR 601.2f pass — and the row that shows what it charges — reaches
// the Signet cycle's "{1}, {T}" and Loot, the Pathfinder's exhaust
// "{G}, {T}" exactly as it reaches a CR 602 ability. Split from
// viewOfManaAbilities for the reason the conditions and the identity
// are — that projection has no game handle. Left nil on a pricing
// error, exactly as the activated-ability row does, so the client
// falls back to ManaCost rather than showing a stale charge; a
// discount that empties the component out completely still stamps a
// pointer to "", which is a real answer and not the same as nil (see
// ManaAbilityView.ChargedManaCost). Caller must hold g's read lock.
func stampManaChargedCost(g *game.Game, card game.Card, controller uuid.UUID, views []ManaAbilityView) {
	raw := game.ManaAbilitiesForCard(card)
	for i := range views {
		if i >= len(raw) || raw[i].ManaCost == "" {
			continue
		}
		if charged, err := g.ManaAbilityManaCostForEffect(controller, card, raw[i]); err == nil {
			s := charged.String()
			views[i].ChargedManaCost = &s
		}
	}
}

// stampManaCounterCosts fills ManaAbilityView's counter half (#789):
// Vivid Creek's charge counter, Ramos's five +1/+1 counters, Mage-Ring
// Network's "any number of storage counters". Split from
// viewOfManaAbilities for the reason the conditions and the identity
// are — that projection has no game handle, and the option list is a
// walk of the battlefield. Caller must hold g's read lock.
func stampManaCounterCosts(g *game.Game, card game.Card, controller uuid.UUID, views []ManaAbilityView) {
	raw := game.ManaAbilitiesForCard(card)
	for i := range views {
		if i >= len(raw) {
			continue
		}
		views[i].CounterCostView = counterCostView(g, controller, card.InstanceID, raw[i].RemoveCounters, raw[i].AddCounter)
	}
}

// stampCombatTargets fills the two S27 combat-target projections that
// need a game handle: the KIND of each attacker's declared target,
// and the active player's legal attack-target set.
//
// Split out of viewOfCard for the reason stampActivatedAbilities is:
// classifying an id as a seat, a planeswalker or a battle means
// looking at the rest of the game, and viewOfCard has one card.
//
// Caller must hold g's read lock (ReadSnapshot already does).
func stampCombatTargets(g *game.Game, view *GameView) {
	for i := range view.Battlefield.Cards {
		c := &view.Battlefield.Cards[i]
		if c.AttackingTarget == "" {
			continue
		}
		id, err := uuid.Parse(c.AttackingTarget)
		if err != nil {
			continue
		}
		if kind := g.ClassifyAttackTargetForEffect(id); kind != "" {
			c.AttackingTargetKind = string(kind)
		}
		if d := g.DefendingPlayerForAttackForEffect(id); d != uuid.Nil {
			c.DefendingPlayer = d.String()
		}
	}
	if g.Turn.Step != game.StepDeclareAttackers {
		return
	}
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return
	}
	active := g.Seats[g.Turn.ActiveSeat]
	if active == nil {
		return
	}
	// ADR 0080: the per-creature attack tax, priced against a
	// REPRESENTATIVE attacker rather than a real one, because the
	// field is about the target and the printed family charges the
	// same for every creature. Any creature the active seat controls
	// answers the question; the first one is used so the answer is
	// stable frame to frame.
	probe := attackTaxProbe(g, active.ID)
	for _, t := range g.AttackTargetsForEffect(active.ID) {
		row := AttackTargetView{
			Kind: string(t.Kind),
			ID:   t.ID.String(),
		}
		if probe != uuid.Nil {
			row.Tax = g.PriceAttackDeclarationForEffect([]game.AttackDeclaration{{
				Attacker: probe,
				Target:   t.ID,
			}}).Cost
		}
		view.Turn.AttackTargets = append(view.Turn.AttackTargets, row)
	}
}

// attackTaxProbe picks the creature the attack-tax preview is priced
// against: the active seat's first battlefield creature, or uuid.Nil
// when it controls none (in which case there is nothing to declare and
// no price to show).
//
// A probe rather than a per-creature matrix because AttackTargetView
// is a per-TARGET row and every printed attack tax charges the same
// for every creature. The enumerator's MoveCost.Mana is the per-move
// answer for a consumer that needs one.
//
// Caller holds g's read lock.
func attackTaxProbe(g *game.Game, seat uuid.UUID) uuid.UUID {
	if g.Battlefield == nil {
		return uuid.Nil
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == seat && c.IsCreature() {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// stampNoUntap projects the untap-step state that needs the game handle.
// Caller holds g's read lock. The view and battlefield slices are kept in
// the same order by viewOfZone, so this pass can use the card index without
// exposing the internal marker representation.
func stampNoUntap(g *game.Game, view *ZoneView) {
	if g == nil || g.Battlefield == nil || view == nil {
		return
	}
	for i := range view.Cards {
		if i >= len(g.Battlefield.Cards) {
			break
		}
		card := &g.Battlefield.Cards[i]
		// A face-down permanent has no abilities, so its own static
		// restriction reads false — but a hold (#1313) comes from
		// ANOTHER object's resolved ability and applies face-down or
		// not, like a next-step marker.
		static := (!card.FaceDown && g.UntapStepRestrictedLocked(card)) || g.UntapHeldLocked(card)
		next := projectedUntapSkipPlayers(g, card)
		if !static && len(next) == 0 {
			continue
		}
		view.Cards[i].NoUntap = &NoUntapView{Static: static, Next: next}
	}
}

func projectedUntapSkipPlayers(g *game.Game, card *game.Card) []string {
	if card == nil || len(card.NextUntapSkips) == 0 {
		return nil
	}
	seen := make(map[uuid.UUID]struct{}, len(card.NextUntapSkips))
	players := make([]string, 0, len(card.NextUntapSkips))
	for _, skip := range card.NextUntapSkips {
		if skip.While != nil {
			// A hold is projected as Static by stampNoUntap.
			continue
		}
		playerID := skip.Player
		if playerID == uuid.Nil {
			playerID = card.Controller
		}
		if playerID == uuid.Nil || !livePlayer(g, playerID) {
			continue
		}
		if _, ok := seen[playerID]; ok {
			continue
		}
		seen[playerID] = struct{}{}
		players = append(players, playerID.String())
	}
	return players
}

func livePlayer(g *game.Game, id uuid.UUID) bool {
	for _, p := range g.Seats {
		if p != nil && p.ID == id {
			return !p.Eliminated
		}
	}
	return false
}

// stampManaSacrificeOptions fills the card-shaped cost clauses on a
// permanent's MANA abilities: the sacrifice clause (Ashnod's Altar,
// Phyrexian Altar), since #1213 the discard clause (Skirge
// Familiar), and since #1283 the exile-a-card clause (Cadaverous Bloom).
//
// Split from viewOfManaAbilities because that runs while building the
// base card view, which has no game handle — computing a legal set
// needs one. Same division the activated abilities already use, and
// the same two lists: sacrificeCostOptions and
// DiscardCostOptionsForEffect.
func stampManaSacrificeOptions(g *game.Game, card game.Card, controller uuid.UUID, views []ManaAbilityView) {
	raw := game.ManaAbilitiesForCard(card)
	for i := range views {
		if i >= len(raw) {
			continue
		}
		if raw[i].SacrificeOther != nil {
			views[i].SacrificeLabel = raw[i].SacrificeOther.Label
			views[i].SacrificeOptions = sacrificeCostOptions(g, controller, raw[i].SacrificeOther, card.InstanceID, raw[i].SacrificeCost)
		}
		// #1213: the same three fields the activated view carries,
		// off the same walk, so the client's picker is one component
		// for both ability kinds.
		if dc := raw[i].DiscardCards; dc != nil && dc.N > 0 {
			views[i].DiscardCostN = dc.N
			views[i].DiscardCostLabel = dc.Label
			views[i].DiscardCostOptions = cardIDStrings(g.DiscardCostOptionsForEffect(controller, card.InstanceID, dc))
		}
		// #1283: the exile-a-card sibling, off its own walk.
		if ec := raw[i].ExileCards; ec != nil && ec.N > 0 {
			views[i].ExileCostN = ec.N
			views[i].ExileCostLabel = ec.Label
			views[i].ExileCostOptions = cardIDStrings(g.ExileCostOptionsForEffect(controller, card.InstanceID, ec))
		}
	}
}

// viewOfPendingChoices materialises the PendingChoices queue,
// inlining the candidate cards from the live game state so the
// per-viewer filter can redact them uniformly. Caller must hold
// g's read lock (ReadSnapshot already does).
func viewOfPendingChoices(g *game.Game) []PendingChoiceView {
	if len(g.PendingChoices) == 0 {
		return nil
	}
	out := make([]PendingChoiceView, 0, len(g.PendingChoices))
	for _, c := range g.PendingChoices {
		if c == nil {
			continue
		}
		v := PendingChoiceView{
			ID:            c.ID.String(),
			Kind:          string(c.Kind),
			Chooser:       c.Chooser.String(),
			FromPlayer:    c.FromPlayer.String(),
			Count:         c.Count,
			Source:        uuidStringOrEmpty(c.Source),
			Reason:        c.Reason,
			NoLegalTarget: c.NoLegalTarget,
			PayCost:       c.PayCost,
		}
		// #1311: the waterbend half of a pay-unless, sized against the
		// prompt's own cost — the budget ResolvePayUnlessWithTaps
		// validates against.
		if tc := c.PayTapCost(); !tc.Empty() {
			budget := 0
			if parsed, err := game.ParseCost(c.PayCost); err == nil {
				budget = game.WaterbendBudget(tc, parsed, 0)
			}
			v.TapCost = viewOfWaterbend(g, c.Chooser, uuid.Nil, tc, budget)
		}
		if c.Kind == game.PendingChoiceTriggerPrompt || c.Kind == game.PendingChoicePickTarget {
			doubledBy, doubledByName := c.TriggerDoubler()
			if doubledBy != uuid.Nil {
				v.DoubledBy = doubledBy.String()
				v.DoubledByName = doubledByName
			}
		}
		// For discard_from_hand, inline the source player's hand
		// as Options. Per-viewer redaction in FilterViewFor
		// projects the cards through isKnower — the Thoughtseize
		// caster (chooser) sees face-up because QueueDiscardFromRevealedHand
		// marked them knower; everyone else sees backs.
		if c.Kind == game.PendingChoiceDiscardFromHand {
			fromP := g.PlayerByIDForEffect(c.FromPlayer)
			if fromP != nil && fromP.Hand != nil {
				v.Options = make([]CardView, len(fromP.Hand.Cards))
				for i, card := range fromP.Hand.Cards {
					v.Options[i] = viewOfCard(card)
				}
			}
		}
		// PendingChoiceSacrifice — "each player sacrifices a
		// creature". Inline the chooser's own candidate permanents as
		// Options so the picker renders card faces rather than UUIDs.
		// Battlefield cards are public, so no redaction concern.
		if c.Kind == game.PendingChoiceSacrifice && len(c.SacrificeOptions) > 0 {
			v.Options = make([]CardView, 0, len(c.SacrificeOptions))
			for _, id := range c.SacrificeOptions {
				if card, ok := g.LookupCardForEffect(id); ok {
					v.Options = append(v.Options, viewOfCard(card))
				}
			}
		}
		// The scry family — scry, surveil, plain "look at the top
		// N, put them back in any order", and ADR 0088's
		// put_in_library — projects the cards, top-first, as Options. All three are "look at", not
		// "reveal": only the chooser was marked a knower, so
		// FilterViewFor redacts these to backs for every other seat
		// and the top of the library stays private. The COUNT is
		// public, which is correct — "scry 2" is a printed number.
		if c.Kind == game.PendingChoicePutInLibrary {
			v.Placement = string(c.LibraryPlacement)
		}
		if game.IsLookAtTopKind(c.Kind) && len(c.ScryCards) > 0 {
			v.Options = make([]CardView, 0, len(c.ScryCards))
			for _, id := range c.ScryCards {
				if card, ok := g.LookupCardForEffect(id); ok {
					v.Options = append(v.Options, viewOfCard(card))
				}
			}
		}
		// PendingChoiceSearchLibrary — the matching library cards the
		// searcher may take. Only the chooser was marked a knower, so
		// FilterViewFor redacts these for every other seat — and then
		// drops the list outright for non-choosers, because the
		// COUNT of matches is itself information about a hidden zone
		// that nobody else is entitled to.
		if c.Kind == game.PendingChoiceSearchLibrary && len(c.SearchCards) > 0 {
			v.SearchMax = c.SearchMax
			v.Options = make([]CardView, 0, len(c.SearchCards))
			for _, id := range c.SearchCards {
				if card, ok := g.LookupCardForEffect(id); ok {
					v.Options = append(v.Options, viewOfCard(card))
				}
			}
		}
		// PendingChoiceConfirm — the chained-choice two-way prompt.
		// Only the branch labels travel; everything else about the
		// question is already in Reason.
		if c.Kind == game.PendingChoiceCoinCall {
			v.AllowStop = c.CoinAllowStop
			v.Coins = c.CoinCount
			v.MaxUsefulWins = c.CoinMaxUsefulWins
			v.Wins = c.CoinWins
		}
		if c.Kind == game.PendingChoiceConfirm {
			v.AcceptLabel = c.AcceptLabel
			v.DeclineLabel = c.DeclineLabel
			v.LifeCost = c.LifeCost
		}
		// PendingChoiceLoopShortcut — the CR 726 proposal (#804). The
		// count is the N in "has resolved N times this turn" and the
		// max is the ceiling on the client's number field; Reason
		// already carries "<card> — <ability>". Public, like the
		// loop_notice beside it: a loop is something the whole table
		// can see running, and everyone can see whose question it is.
		if c.Kind == game.PendingChoiceLoopShortcut {
			v.LoopCount = c.LoopShortcutCount
			v.LoopMaxIterations = game.MaxLoopShortcutIterations
		}
		// PendingChoiceChooseCards — the chained-choice card-set pick.
		// The candidates are frequently cards in a hand, so they go
		// through the ordinary per-viewer knower redaction in
		// FilterViewFor like every other Options list: the chooser was
		// made a knower by whatever effect queued the prompt, and a
		// seat that is not a knower sees nothing at all rather than a
		// count.
		//
		// #826's untap_choice (CR 502.3, "choose which of these
		// untap") carries the same payload and projects the same way.
		// What it does NOT share is the redaction below: its
		// candidates are tapped permanents on the battlefield, which
		// every seat can already see.
		//
		// #1198's entry_reveal_from_hand (CR 614.1c) carries it too,
		// and IS redacted below with choose_cards: its candidates are
		// cards in the revealer's own hand, and which of them match
		// "an Island or Swamp card" is the hidden information the
		// prompt is about. The other seats learn what was shown after
		// the answer, through the reveal's own log line, which is the
		// order CR 701.20 puts them in.
		//
		// #1214's three resolution-time picks (reveal_pick,
		// their_permanents, own_permanents) carry the same payload and
		// project the same way. What they do NOT share is
		// filterPendingChoices' bound-stripping: a reveal_pick's
		// candidates were REVEALED, and the two permanent picks are
		// battlefield cards, so "choose 2 of these 5" is public for all
		// three and hiding it would be hiding a fact the table watched
		// happen.
		if game.IsCardSetPickKind(c.Kind) {
			v.ChooseMin = c.ChooseMin
			v.ChooseMax = c.ChooseMax
			v.Options = make([]CardView, 0, len(c.ChooseCards))
			for _, id := range c.ChooseCards {
				if card, ok := g.LookupCardForEffect(id); ok {
					v.Options = append(v.Options, viewOfCard(card))
				}
			}
		}
		// PendingChoiceOptionPick — #568's "choose one of the
		// following", addressed to any seat. The labels are the
		// card's own words; an option's cards (a pile) are inlined so
		// the client renders faces rather than quoting names into a
		// sentence, and they go through the same per-viewer redaction
		// as every other Options list in filterPendingChoices.
		if c.Kind == game.PendingChoiceOptionPick && len(c.PickOptions) > 0 {
			v.PickOptions = make([]PickOptionView, 0, len(c.PickOptions))
			for _, opt := range c.PickOptions {
				out := PickOptionView{Label: opt.Label, LifeCost: opt.LifeCost}
				if opt.Player != uuid.Nil {
					out.Player = opt.Player.String()
				}
				for _, id := range opt.Cards {
					if card, ok := g.LookupCardForEffect(id); ok {
						out.Cards = append(out.Cards, viewOfCard(card))
					}
				}
				v.PickOptions = append(v.PickOptions, out)
			}
		}
		// PendingChoiceModePick — #764, CR 603.3c: a modal trigger's
		// bullets, chosen as the ability is put on the stack. Public
		// information the moment it is asked (the card's text is
		// public), so nothing here is redacted; the labels are oracle
		// text and the indexes are what the answer names.
		if c.Kind == game.PendingChoiceModePick {
			v.ModeOptions = append([]string(nil), c.ModeOptionLabel...)
			v.ModeIndexes = append([]int(nil), c.ModeOptionIndex...)
			v.ModeMin = c.ModeMin
			v.ModeMax = c.ModeMax
			v.ModeRepeatable = c.ModeRepeatable
		}
		// PendingChoiceMana carries a color-option list server-
		// filtered against the chooser's commander identity (see
		// ActivateManaAbility). Clone the slice so post-wire
		// mutations on the engine's copy don't leak onto the view.
		if c.Kind == game.PendingChoiceMana && len(c.ColorOptions) > 0 {
			v.ColorOptions = append([]string(nil), c.ColorOptions...)
			if len(c.ManaAmounts) > 0 {
				v.ColorAmounts = make(map[string]int, len(c.ManaAmounts))
				for k, n := range c.ManaAmounts {
					v.ColorAmounts[k] = n
				}
			}
		}
		// PendingChoiceColor — #742 "choose a color" (CR 105.4). The
		// legal colours ride in the same field a mana pick uses, so
		// the client's colour buttons render both.
		//
		// #986: ordered by the prompt's own purpose, by the SAME
		// function that orders a bot seat's answers
		// (legal.OrderColorOptionsLocked) — so the button under the
		// cursor and the first move in the bot's list are the same
		// colour, and neither client nor policy needs a ranking rule
		// of its own. The ordering is battlefield counts, which is
		// public, so it is the chooser's order for every viewer. We
		// are already inside ViewOfGame's ReadSnapshot, which is what
		// the Locked suffix means.
		if c.Kind == game.PendingChoiceColor && len(c.ColorOptions) > 0 {
			v.ColorOptions = legal.OrderColorOptionsLocked(g, c.Chooser, c.ColorOptions, c.ColorPurpose)
			v.ColorPurpose = string(c.ColorPurpose)
		}
		// PendingChoiceCreatureType — S26. The option set is the
		// whole CR 205.3m vocabulary; the clone keeps the engine's
		// package-level slice off the wire path, where a marshaller
		// has no business holding a reference to it.
		if c.Kind == game.PendingChoiceCreatureType {
			v.TypeOptions = append([]string(nil), game.AllCreatureTypes...)
		}
		// PendingChoiceCardName — #1210. There is no vocabulary to
		// send (CR 201.2 admits any card name), so what goes on the
		// wire is a SUGGESTION list rebuilt from the public zones.
		// We are already inside ViewOfGame's ReadSnapshot, which is
		// what the Locked suffix means.
		if c.Kind == game.PendingChoiceCardName {
			v.NameOptions = g.PublicCardNamesLocked()
		}
		// PendingChoiceReplacementOrder — S17 sub-PR 2. Emit the
		// ordered list of replacement-effect IDs with a human-
		// readable label + source-card ID (empty for engine
		// built-ins). The client renders a drag-reorder list and
		// returns the IDs in the chosen order.
		if c.Kind == game.PendingChoiceReplacementOrder && len(c.ReplacementEffectIDs) > 0 {
			v.ReplacementOptions = make([]ReplacementOptionView, 0, len(c.ReplacementEffectIDs))
			for _, id := range c.ReplacementEffectIDs {
				label, srcID := g.ReplacementOptionMetaForEffect(id)
				opt := ReplacementOptionView{
					ID:    game.ReplacementEffectIDToString(id),
					Label: label,
				}
				if srcID != (uuid.UUID{}) {
					opt.SourceCardID = srcID.String()
				}
				v.ReplacementOptions = append(v.ReplacementOptions, opt)
			}
		}
		// PendingChoicePickTarget — S20 sub-PR 2: the frozen legal set.
		// S27: the legend rule asks the same question shape — pick one
		// from a server-computed set — and rides the same projection
		// so the client's existing highlight flow answers it. It is
		// NOT targeting (a state-based action chooses nothing on the
		// stack); the kind is what keeps the two distinguishable on
		// the way back.
		// #1196: the CR 115.7 retarget prompt is a fourth. Same
		// question shape again — pick one from a server-computed set
		// — so it rides the same projection and the client's existing
		// highlight flow answers it; the KIND is what routes the
		// answer to ResolveRetarget rather than to ResolvePickTarget.
		if c.Kind == game.PendingChoicePickTarget || c.Kind == game.PendingChoiceLegendRule ||
			c.Kind == game.PendingChoiceChooseProtector || c.Kind == game.PendingChoiceRetarget {
			pt := &LegalTargetsView{Min: c.PickTargetMin, Max: c.PickTargetMax}
			for _, id := range c.PickTargetPlayers {
				pt.Players = append(pt.Players, id.String())
			}
			for _, id := range c.PickTargetCards {
				pt.Cards = append(pt.Cards, id.String())
			}
			v.PickTarget = pt
		}
		// PendingChoiceTriggerOrder — S19 sub-PR 8. Resolve each
		// pending-trigger ID to its label + source so the reorder
		// list reads as card names, not UUIDs.
		if c.Kind == game.PendingChoiceTriggerOrder && len(c.TriggerOrderIDs) > 0 {
			v.TriggerOptions = make([]ReplacementOptionView, 0, len(c.TriggerOrderIDs))
			for _, id := range c.TriggerOrderIDs {
				opt := ReplacementOptionView{ID: id.String()}
				for _, t := range g.PendingTriggers {
					if t != nil && t.ID == id {
						opt.Label = t.Label
						opt.SourceCardID = t.SourceCardID.String()
						break
					}
				}
				v.TriggerOptions = append(v.TriggerOptions, opt)
			}
		}
		// PendingChoiceDamageAssignment — S18 sub-PR 3. Emit the
		// attacker + blocker IDs + attacker power + trample flag so
		// the client can render the per-blocker damage picker.
		if c.Kind == game.PendingChoiceDamageAssignment && c.DamageAssignment != nil {
			frame := c.DamageAssignment
			blockers := make([]string, len(frame.BlockerIDs))
			for i, id := range frame.BlockerIDs {
				blockers[i] = id.String()
			}
			v.DamageAssignment = &DamageAssignmentView{
				AttackerCardID: frame.AttackerID.String(),
				BlockerCardIDs: blockers,
				AttackerPower:  frame.AttackerPower,
				AllowTrample:   frame.AllowTrample,
				HasDeathtouch:  frame.HasDeathtouch,
			}
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// viewOfDiscardPending mirrors Game.DiscardPending to the wire
// shape (string-keyed UUIDs). Returns nil for an empty map so
// json.Marshal's omitempty drops the field.
func viewOfDiscardPending(in map[uuid.UUID]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k.String()] = v
	}
	return out
}

// viewOfStackItemsInStackOrder projects every stack item bottom..top
// in resolution order — the inverse of the LIFO the engine uses
// (CR 608.1): items are sorted by insertion Seq, so an ability
// triggered in response to a spell sits above it. Spell items are
// matched to cards via InstanceID and, for equal Seq (legacy
// zero-Seq snapshots), keep Game.Stack zone order ahead of ability
// items — the pre-Seq behaviour. Returns nil for an empty stack so
// json omitempty drops the field.
func viewOfStackItemsInStackOrder(g *game.Game) []StackItemView {
	if len(g.StackMeta) == 0 {
		return nil
	}
	type entry struct {
		seq  uint64
		view StackItemView
	}
	entries := make([]entry, 0, len(g.StackMeta))
	seen := make(map[uuid.UUID]bool, len(g.StackMeta))
	if g.Stack != nil {
		for _, c := range g.Stack.Cards {
			if item, ok := g.StackMeta[c.InstanceID]; ok && item != nil {
				entries = append(entries, entry{item.Seq, viewOfStackItem(item)})
				seen[item.ID] = true
			}
		}
	}
	// Ability items have no card on Game.Stack; collect them by Seq
	// so the projection is deterministic regardless of map order.
	abilities := make([]*game.StackItem, 0, len(g.StackMeta))
	for id, item := range g.StackMeta {
		if seen[id] || item == nil {
			continue
		}
		abilities = append(abilities, item)
	}
	sort.SliceStable(abilities, func(i, j int) bool { return abilities[i].Seq < abilities[j].Seq })
	for _, item := range abilities {
		entries = append(entries, entry{item.Seq, viewOfStackItem(item)})
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].seq < entries[j].seq })
	if len(entries) == 0 {
		return nil
	}
	out := make([]StackItemView, len(entries))
	for i, e := range entries {
		out[i] = e.view
	}
	return out
}

// viewOfStackItemSlice projects a slice of *StackItem (the pending-
// triggers queue) to the wire shape, preserving order. Returns nil
// for an empty input.
func viewOfStackItemSlice(items []*game.StackItem) []StackItemView {
	if len(items) == 0 {
		return nil
	}
	out := make([]StackItemView, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		out = append(out, viewOfStackItem(it))
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// viewOfDelayedTriggers mirrors the queue of CR 603.7 delayed
// triggered abilities. Nil for an empty queue so json.Marshal's
// omitempty drops the field. Added in S22.
func viewOfDelayedTriggers(queue []*game.DelayedTrigger) []DelayedTriggerView {
	if len(queue) == 0 {
		return nil
	}
	out := make([]DelayedTriggerView, 0, len(queue))
	for _, dt := range queue {
		if dt == nil {
			continue
		}
		view := DelayedTriggerView{
			ID:          dt.ID.String(),
			Controller:  uuidStringOrEmpty(dt.Controller),
			Source:      uuidStringOrEmpty(dt.SourceCardID),
			Label:       dt.Label,
			At:          string(dt.At),
			CreatedTurn: dt.CreatedTurn,
		}
		for _, kind := range dt.On {
			view.On = append(view.On, string(kind))
		}
		for _, cardID := range dt.Cards {
			view.Cards = append(view.Cards, cardID.String())
		}
		out = append(out, view)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// viewOfStackItem mirrors a single game.StackItem to its wire shape.
// Distribution and Targets are reallocated; scalar fields are
// stringified UUIDs.
func viewOfStackItem(it *game.StackItem) StackItemView {
	view := StackItemView{
		ID:           it.ID.String(),
		Kind:         string(it.Kind),
		Controller:   it.Controller.String(),
		Owner:        it.Owner.String(),
		SourceCardID: it.SourceCardID.String(),
		Label:        it.Label,
		Modes:        append([]int(nil), it.Modes...),
		ModeLabels:   it.ModeLabels(),
		XValue:       it.XValue,
		HoldPriority: it.HoldPriority,
		SplitSecond:  it.SplitSecond,
		AltCost:      it.AltCost,
		IsCopy:       it.IsCopy,
		// #761: what paid for it. Mana is spent face up, so this is
		// public, and a responder to a converge spell needs it.
		ManaSpent:        it.Paid.ManaSpentCount(),
		ColorsSpent:      it.Paid.ColorsSpent(),
		ManaSpentUnknown: !it.Paid.Known(),
	}
	if it.Paid.GiftPromised() {
		view.GiftTo = it.Paid.GiftOpponent.String()
	}
	if it.DoubledBy != uuid.Nil {
		view.DoubledBy = it.DoubledBy.String()
		view.DoubledByName = it.DoubledByName
	}
	if len(it.Targets) > 0 {
		view.Targets = make([]TargetRefView, len(it.Targets))
		for i, t := range it.Targets {
			view.Targets[i] = TargetRefView{
				Kind: string(t.Kind),
				ID:   uuidStringOrEmpty(t.ID),
				Slot: t.Slot,
				Mode: t.Mode,
			}
		}
	}
	if len(it.Distribution) > 0 {
		view.Distribution = make(map[string]int, len(it.Distribution))
		for k, v := range it.Distribution {
			view.Distribution[k.String()] = v
		}
	}
	return view
}

func viewOfSeats(g *game.Game, seats []*game.Player) []PlayerView {
	out := make([]PlayerView, len(seats))
	for i, p := range seats {
		out[i] = viewOfPlayer(g, p)
	}
	return out
}

func viewOfPlayer(g *game.Game, p *game.Player) PlayerView {
	cmdrDamage := make(map[string]int, len(p.CommanderDamage))
	for k, v := range p.CommanderDamage {
		cmdrDamage[k.String()] = v
	}
	var cmdrCasts map[string]int
	if len(p.CommanderCasts) > 0 {
		cmdrCasts = make(map[string]int, len(p.CommanderCasts))
		for k, v := range p.CommanderCasts {
			cmdrCasts[k.String()] = v
		}
	}
	history := make([]LifeChangeView, len(p.LifeHistory))
	for i, c := range p.LifeHistory {
		history[i] = LifeChangeView{
			Delta:    c.Delta,
			NewTotal: c.NewTotal,
			At:       c.At.UTC().Format(time.RFC3339),
			Seq:      c.Seq,
		}
	}
	var manaPool []string
	if len(p.ManaPool) > 0 {
		manaPool = make([]string, len(p.ManaPool))
		for i, t := range p.ManaPool {
			manaPool[i] = t.Color
		}
	}
	// #623 / CR 114. The label and text come from the catalog, so a
	// game restored on a binary whose card file reworded the emblem
	// shows the new wording.
	var emblems []EmblemView
	for _, e := range g.EmblemsForPlayer(p.ID) {
		emblems = append(emblems, EmblemView{
			InstanceID: e.InstanceID.String(),
			Label:      e.Label,
			Text:       e.Text,
		})
	}
	return PlayerView{
		ID:                p.ID.String(),
		Name:              p.Name,
		Seat:              p.Seat,
		Life:              p.Life,
		Poison:            p.Poison,
		Energy:            p.Energy,
		Library:           viewOfZone(p.Library),
		Hand:              viewOfZone(p.Hand),
		Graveyard:         viewOfZone(p.Graveyard),
		Command:           viewOfZone(p.Command),
		CommanderDamage:   cmdrDamage,
		LifeHistory:       history,
		Eliminated:        p.Eliminated,
		HandKept:          p.HandKept,
		MulligansTaken:    p.MulligansTaken,
		DeckImported:      p.DeckImported,
		UndosRemaining:    p.UndosRemaining,
		DiscordID:         p.DiscordID,
		DiscordAvatarHash: p.DiscordAvatarHash,
		DisplayName:       p.DisplayName,
		IsBot:             p.IsBot,
		BotTier:           p.BotTier,
		BotDeck:           p.BotDeck,
		CommanderCasts:    cmdrCasts,
		Counters:          cloneStringIntMap(p.Counters),
		MaxHandSize:       g.EffectiveMaxHandSizeLocked(p),
		// Locked variants: this builder already runs under the
		// game's read lock (see legal.EnumerateFor's note), and the
		// public accessors would take it a second time.
		LandDropsPerTurn:    g.EffectiveLandDropsLocked(p),
		LandsPlayedThisTurn: g.LandsPlayedThisTurnFor(p.ID),
		ManaPool:            manaPool,
		Emblems:             emblems,
		Keywords:            g.PlayerAbilitiesForEffect(p),
		LifeTotalLocked:     g.PlayerLifeTotalCantChangeLocked(p),
	}
}

// cloneStringIntMap returns nil for an empty input so json.Marshal's
// omitempty drops the field. Otherwise allocates a fresh map so the
// caller doesn't observe future mutations on the engine's copy.
func cloneStringIntMap(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// viewOfPromises projects the per-pair promise map to wire format.
// Keys are encoded as "{from}->{to}" strings; entries with count <= 0
// are dropped so the wire payload stays sparse.
func viewOfPromises(in map[game.PromiseKey]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		if v <= 0 {
			continue
		}
		out[k.From.String()+"->"+k.To.String()] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// viewOfVote projects an open vote to wire format. Returns nil for a
// nil vote so json's omitempty drops the field entirely.
func viewOfVote(v *game.Vote) *VoteView {
	if v == nil {
		return nil
	}
	ballots := make(map[string]int, len(v.Ballots))
	for k, opt := range v.Ballots {
		ballots[k.String()] = opt
	}
	return &VoteView{
		ID:        v.ID.String(),
		Topic:     v.Topic,
		Options:   append([]string(nil), v.Options...),
		Initiator: v.Initiator.String(),
		Ballots:   ballots,
	}
}

// uuidStringOrEmpty returns u.String() unless u is the zero UUID, in
// which case it returns "" so json's omitempty drops the field. Used
// for nullable scalar IDs like Game.Monarch / Game.Initiative.
func uuidStringOrEmpty(u uuid.UUID) string {
	if u == uuid.Nil {
		return ""
	}
	return u.String()
}

func viewOfZone(z *game.Zone) ZoneView {
	if z == nil {
		return ZoneView{}
	}
	cards := make([]CardView, len(z.Cards))
	onBattlefield := z.Kind == game.ZoneBattlefield
	phasedOut := z.Kind == game.ZonePhasedOut
	for i, c := range z.Cards {
		cards[i] = viewOfCard(c)
		// #1199: read off the zone rather than off the card, because
		// the card carries no flag — ADR 0084 keeps phasing as a
		// membership rather than as a bit several subsystems have to
		// agree about, and this is the one place the wire needs it
		// spelled out.
		cards[i].PhasedOut = phasedOut
		// #29: the position pair is battlefield-only, and on the
		// battlefield it is ALWAYS sent — including (0, 0), which is
		// both a legitimate stamp (the clamp lands every negative and
		// NaN input there) and the default for a permanent nobody has
		// positioned. Taking the address of the loop-local copy is
		// safe under Go 1.22+ per-iteration scoping, and each CardView
		// gets its own pair rather than aliasing its neighbours'.
		if onBattlefield {
			x, y := c.BattleX, c.BattleY
			cards[i].BattleX, cards[i].BattleY = &x, &y
		}
	}
	owner := ""
	if !z.IsShared() {
		owner = z.Owner.String()
	}
	return ZoneView{
		Kind:  string(z.Kind),
		Owner: owner,
		Count: len(z.Cards),
		Cards: cards,
	}
}

// FilterViewFor returns a copy of v with zones hidden from the given
// viewer zeroed out. The input is not mutated; only the copy's seat
// entries for non-viewer players get new (empty) card slices.
//
// Visibility rules at S04:
//   - Own seat: full fidelity (hand + library cards visible).
//   - Opponent seat: Hand.Cards and Library.Cards replaced with empty
//     slices; the Count field is preserved so the UI can render a
//     hidden stack. Graveyard and Command zones stay visible
//     (graveyard is public in MTG; command is public because
//     commanders are public).
//   - Shared zones (battlefield, stack, exile, phased_out):
//     unchanged but for the per-card redaction every zone gets.
//
// viewerID is the player UUID string; pass the empty string to get a
// "spectator" view where every opponent hand and library is hidden
// (i.e. no seat is treated as "own"). An observer without a claimed
// seat ends up here.
func FilterViewFor(v GameView, viewerID string) GameView {
	// S13.5: build a per-card "is the viewer a knower" closure that
	// every zone projection consults. Empty viewerID = admin /
	// spectator → knows everything.
	isKnower := func(c CardView) bool {
		if viewerID == "" {
			return true
		}
		return c.knowers[viewerID]
	}

	seats := make([]PlayerView, len(v.Seats))
	for i, p := range v.Seats {
		out := p
		// S13.5: redact every visible card based on KnownBy.
		// Hand + library still get their wholesale-hide (S04
		// zone-default heuristic) for opponents, but the per-card
		// pass below covers all visible zones uniformly.
		out.Library = applyCastStampsFor(redactZone(p.Library, isKnower), viewerID)
		// #1166: and the hand, for the reason the graveyard is
		// promoted below. A hand card's announce surface is the HAND
		// OWNER's answer, and a card another seat is a knower of —
		// Thoughtseize, Telepathy, a reveal — reaches that seat's
		// frame with keepKnownInHandZone. Until this line it reached
		// them carrying the owner's `legal_targets`, which hexproof,
		// shroud, protection and "target opponent" all narrow by who
		// is asking.
		out.Hand = applyCastStampsFor(redactZone(p.Hand, isKnower), viewerID)
		// #1022: a graveyard is public, but a cast permission over one
		// of its cards held by ANOTHER seat is that seat's answer, so
		// the stamps it produced come off for everybody else — the
		// pass exile has ridden since #978, now on the one per-seat
		// zone that can carry a foreign holder's stamps.
		out.Graveyard = applyCastStampsFor(redactZone(p.Graveyard, isKnower), viewerID)
		// #1166: the command zone is public and CR 903.4's permission
		// is its owner's alone, so the same split applies — the
		// commander's price list is a fact about the card, and the
		// legal targets of a cast only that seat may make are not.
		out.Command = applyCastStampsFor(redactZone(p.Command, isKnower), viewerID)
		// Opponent hand: strip only UNREVEALED cards (seated-viewer
		// path — Thoughtseize-style reveals survive via KnownBy);
		// wholesale-hide for spectator / admin (viewerID empty)
		// where the isKnower short-circuit to true would otherwise
		// leak every hand on the wire. Opponent library: always
		// wholesale-hidden — no per-card reveal paths for library
		// identity leak to the opponent client today. Viewer's own
		// hand + library: redactZone already handled visibility.
		if p.ID == "" || p.ID != viewerID {
			if viewerID == "" {
				out.Hand = hideZoneContents(out.Hand)
				// Spectator / admin: isKnower short-circuits to true,
				// so the library has to be hidden wholesale here for
				// the same reason the hand is — otherwise every
				// library's top card would ride out on the wire.
				out.Library = hideZoneContents(out.Library)
			} else {
				out.Hand = keepKnownInHandZone(out.Hand)
				// S42 / CR 401.5. An opponent's library used to be
				// wholesale-hidden, because no path revealed a card in
				// one. "Play with the top card of your library
				// revealed" (Oracle of Mul Daya, Courser of Kruphix)
				// is that path, so the projection keeps the TOP CARD
				// and only when this viewer knows it. Narrow on
				// purpose: a tucked card in the middle of a library
				// that still carries a stale knower is not exposed,
				// because only the last element is even considered.
				out.Library = keepKnownTopInLibraryZone(out.Library)
			}
		}
		seats[i] = out
	}
	return GameView{
		ID:          v.ID,
		State:       v.State,
		Seats:       seats,
		Battlefield: redactZone(v.Battlefield, isKnower),
		Stack:       redactZone(v.Stack, isKnower),
		// #978: exile is the one SHARED zone that carries
		// announce-time cast stamps, and they were computed for one
		// seat — the holder of the CastPermission that opened the
		// card. Everyone else gets the card without them.
		Exile: applyCastStampsFor(redactZone(v.Exile, isKnower), viewerID),
		// #1199: shared and public like the battlefield, and redacted
		// the same way — a permanent can phase out face down, and the
		// card under it is no more knowable for having phased.
		PhasedOut:         redactZone(v.PhasedOut, isKnower),
		Turn:              v.Turn,
		MulligansOpen:     v.MulligansOpen,
		Monarch:           v.Monarch,
		Initiative:        v.Initiative,
		Promises:          v.Promises,
		Vote:              v.Vote,
		UndoLimit:         v.UndoLimit,
		Settings:          v.Settings,
		StartingSeat:      v.StartingSeat,
		StackItems:        v.StackItems,
		PendingTriggers:   v.PendingTriggers,
		DelayedTriggers:   v.DelayedTriggers,
		SplitSecondActive: v.SplitSecondActive,
		DiscardPending:    v.DiscardPending,
		PendingChoices:    filterPendingChoices(v.PendingChoices, isKnower, viewerID),
		LegalMoves:        legalMovesFor(v.legalBySeat, viewerID),
		// S31 sub-PR 0: the public log rides the same isKnower closure
		// as every zone above it. Not a parallel visibility model —
		// literally the same predicate, applied to the card each entry
		// names.
		Log: redactLogForViewer(v.Log, isKnower),
		// S22: the reveal window is the one field here that is NOT
		// projected through isKnower, and the omission is the feature.
		// A reveal is public by construction — every seat saw the same
		// cards at the same moment — so there is no per-viewer answer
		// to give. There is also nothing left to redact: the entries
		// carry printed identity and no instance ID, so the handle
		// that filterPendingChoices has to drop and redactLogForViewer
		// has to reason about does not exist on this type.
		Reveals: v.Reveals,
		// #628: public, and identical for every seat — see the field
		// comment. Nothing in it names a card in a hidden zone.
		LoopNotice: v.LoopNotice,
	}
}

// legalMovesFor picks the viewer's own move list out of the per-seat
// enumeration and drops every other seat's.
//
// The empty viewerID — spectator, admin, replay reader — gets
// nothing, which is the one place this parts company with the rest of
// FilterViewFor's "empty means see everything" convention. A move
// list is not a view of the board, it is a view of what a specific
// player is holding: "cast Lightning Bolt targeting Kess" names a
// card in a hand. There is no seat whose moves an unseated viewer is
// entitled to, so there is nothing to hand back.
func legalMovesFor(bySeat map[string][]LegalMoveView, viewerID string) []LegalMoveView {
	if viewerID == "" || len(bySeat) == 0 {
		return nil
	}
	return bySeat[viewerID]
}

// filterPendingChoices projects each choice's Options through the
// viewer's knower filter. Unknown cards come back redacted (backs)
// so an opponent browsing the wire can't peek at a Thoughtseize-
// revealed hand.
//
// A "search_library" prompt goes further and drops its Options for
// anyone but the chooser. Redaction alone would still ship one card
// back per match, which tells the table how many Islands are left in
// a library mid-fetch — a number CR 400.2 does not entitle them to.
// Spectators and admins (empty viewerID) are held to the same rule
// here rather than being handed the whole match set.
//
// Every OTHER kind takes the same medicine one card at a time: an
// option the viewer does not know is dropped from a non-chooser's
// copy rather than redacted into an anonymous back. Redaction zeroes
// the printed characteristics and keeps the instance ID, which is the
// right trade for a card sitting in a public zone — the ID is already
// on that viewer's wire, so withholding it would be theatre — and the
// wrong one here, because these options come out of HIDDEN zones that
// FilterViewFor strips for exactly this viewer a few lines above.
//
// The concrete leak this closes: scry, surveil and "look at the top
// N" inline the chooser's top library cards as Options, and the zone
// projection had already removed every trace of that library from an
// opponent's frame. Shipping the same cards back as N stable UUIDs
// hands the table a handle on specific cards in a hidden zone that it
// can correlate the moment one of them is cast — the same "the
// instance ID alone is the leak" argument S31 sub-PR 0 settled for
// the public log, pointed the other way. Thoughtseize is the same
// shape: the caster is entitled to the hand it revealed, the two
// seats watching are not.
//
// The chooser always keeps their whole list, known or not. They are
// the one being asked to pick, and a coercive discard that revealed
// nothing is still answered by choosing one of the backs.
func filterPendingChoices(src []PendingChoiceView, isKnower func(CardView) bool, viewerID string) []PendingChoiceView {
	if len(src) == 0 {
		return nil
	}
	out := make([]PendingChoiceView, len(src))
	for i, c := range src {
		out[i] = c
		if c.Kind == string(game.PendingChoiceSearchLibrary) && c.Chooser != viewerID {
			out[i].Options = nil
			out[i].SearchMax = 0
			continue
		}
		// choose_cards candidates are usually cards in a hand. The
		// knower filter below already strips the cards themselves, but
		// the BOUNDS would still say "2 of 3" about a hidden zone, so
		// they go too — a non-chooser learns that a choice is open and
		// who owes it, and nothing about its contents.
		// #1198's entry_reveal_from_hand is the same pool and the
		// same rule: the candidates are the revealer's own hand, and
		// the COUNT of the ones that match the land's clause is
		// itself information about it. Every seat sees that the
		// prompt is open and whose it is; what was actually revealed
		// reaches them afterwards as an EventRevealCards run.
		if (c.Kind == string(game.PendingChoiceChooseCards) ||
			c.Kind == string(game.PendingChoiceEntryRevealFromHand)) && c.Chooser != viewerID {
			out[i].Options = nil
			out[i].ChooseMin = 0
			out[i].ChooseMax = 0
			continue
		}
		if len(c.Options) > 0 {
			out[i].Options = redactChoiceCards(c, c.Options, isKnower, viewerID)
		}
		// #568: an option's pile rides the SAME pass. A second card
		// list on a prompt that the filter did not know about would
		// be raw identity on every seat's wire, which is the leak PR
		// #513 fixed pointed at a new field.
		if len(c.PickOptions) > 0 {
			opts := make([]PickOptionView, len(c.PickOptions))
			for j, opt := range c.PickOptions {
				opts[j] = opt
				opts[j].Cards = redactChoiceCards(c, opt.Cards, isKnower, viewerID)
			}
			out[i].PickOptions = opts
		}
	}
	return out
}

// redactChoiceCards is THE answer to "which of a prompt's cards may
// this viewer see", and the only place it is answered: every card
// list on a PendingChoiceView — Options and each option's pile —
// goes through it.
//
// Two rules, and the second is what #568 added.
//
//  1. A viewer who is not a knower of a card does not get it. Not a
//     back either: the card is DROPPED, because these lists come out
//     of hidden zones that FilterViewFor strips for this same viewer
//     a few lines up, and shipping N stable instance IDs back hands
//     the table a handle it can correlate the moment one of them is
//     cast. Redaction-to-a-back is the right trade for a card in a
//     public zone and the wrong one here (PR #513).
//
//  2. The CHOOSER keeps their whole list — unknown cards included, as
//     answerable backs — when the pool is their OWN material, or when
//     it is another player's HAND. A hand's SIZE is public (CR 400.2),
//     so handing the chooser N anonymous backs out of one tells them
//     nothing the rest of the table does not already have, and a
//     coercive discard that revealed nothing still has to be
//     answerable by picking one of them (#513's own conclusion, and
//     TestChooserKeepsUnknownOptions).
//
//     A pool that is neither is different, and that difference is
//     #568. Fact or Fiction asks an opponent to separate five cards
//     off the top of your LIBRARY, whose depth and contents nobody
//     else is entitled to; the only thing that lets that opponent look
//     is that the card revealed them first (#549). Cards they were not
//     made a knower of are dropped from their copy too — no backs, so
//     not even the instance IDs, which are the correlation handle PR
//     #513 was about.
//
// Rule 2 is therefore also a contract on the effect: a prompt
// addressed to a seat over another seat's non-hand cards must reveal
// them, or its chooser is handed an empty list. That is the right
// failure — an engine that leaked the library instead would be the
// bug, and the cards a card like this asks about are public by
// construction.
//
// discard_from_hand is named explicitly because its pool IS a hand,
// definitionally: the kind means "pick from FromPlayer's hand", and
// viewOfPendingChoices inlines exactly that zone. Thoughtseize is
// unaffected for a second reason as well — QueueDiscardFromRevealedHand
// marks the caster a knower of every card first, so nothing is a back
// to begin with.
func redactChoiceCards(c PendingChoiceView, cards []CardView, isKnower func(CardView) bool, viewerID string) []CardView {
	if len(cards) == 0 {
		return nil
	}
	// The pool is the chooser's own when no other seat is named, or
	// when the named seat IS the chooser.
	ownPool := c.FromPlayer == "" || c.FromPlayer == c.Chooser
	handPool := c.Kind == string(game.PendingChoiceDiscardFromHand)
	keepUnknown := c.Chooser == viewerID && (ownPool || handPool)
	out := make([]CardView, 0, len(cards))
	for _, card := range cards {
		known := isKnower(card)
		if !known && !keepUnknown {
			continue
		}
		out = append(out, redactCardForViewer(card, known))
	}
	return out
}

// redactZone returns a copy of `z` with every card projected through
// redactCardForViewer — known cards keep their characteristics with
// known_by_you=true; unknown cards get redacted characteristics with
// known_by_you=false. Always allocates a fresh non-nil slice so the
// JSON shape stays `[]` (not `null`) regardless of input.
func redactZone(z ZoneView, isKnower func(CardView) bool) ZoneView {
	out := ZoneView{
		Kind:  z.Kind,
		Owner: z.Owner,
		Count: z.Count,
		Cards: make([]CardView, len(z.Cards)),
	}
	for i, c := range z.Cards {
		out.Cards[i] = redactCardForViewer(c, isKnower(c))
	}
	return out
}

// applyCastStampsFor promotes the announce-time cast stamps this
// viewer's own seat was offered into the exported fields, and drops
// everybody else's (#978, #1022, #1037).
//
// The exported fields arrive carrying the PUBLIC half of the zone
// owner's answer — the prices a cast out of this pile may claim, the
// modes, the clause that refuses it, all of it printed on a card in a
// public zone (castStamps.applyPublicTo) — and a seat with an answer
// of its own overwrites the lot wholesale. A seat with none keeps the
// public half, which is what a bystander is entitled to see: the card,
// its printed price list, the public `exile_play`, and no cast
// surface.
//
// `castable_here`, `legal_targets` and `clauses` are never in that
// public half (#1055). All three answer "what may YOU announce", the
// first one in its very name, and one seat's answer shipped to the
// whole table is the surface #891 is about. Nor are the legal sets
// nested inside `modes` and `alternative_costs`, whose parents DO stay
// public (#1172) — the promotion below is one assignment of the whole
// CastSurfaceView, so a seat with an answer of its own gets the nested
// lists back with the top-level ones and nobody else gets either.
//
// The empty viewerID — spectator, admin, replay reader — gets nothing
// private, for the reason legalMovesFor gives: a legal target set is
// not a view of the board, it is a view of what one specific player
// may announce, and there is no seat whose answer an unseated viewer
// is entitled to.
//
// Only for a KNOWER. The stamps name the card as loudly as its mana
// cost does — "this spell chooses two of three modes" — and
// redactCardForViewer has just cleared them off a card this viewer
// cannot read; promoting a private set over the top would hand them
// straight back. A holder who cannot see the card has no cast to
// announce anyway: the engine's own predicates (CR 401.5's look, a
// foretold card's owner) are what put them in the map.
//
// The map is cleared on the way out exactly as `knowers` is, so
// repeated FilterViewFor calls stay stable.
func applyCastStampsFor(z ZoneView, viewerID string) ZoneView {
	for i := range z.Cards {
		c := &z.Cards[i]
		mine := viewerID != "" && c.KnownByYou
		if c.castOffers != nil {
			stamps, ok := c.castOffers[viewerID]
			c.castOffers = nil
			if ok && mine {
				stamps.applyTo(c)
			}
		}
		applyFaceCastStampsFor(c, viewerID, mine)
	}
	return z
}

// applyFaceCastStampsFor is applyCastStampsFor's per-face half (#992):
// the same promotion, over the blocks CardFaceView carries for the
// halves the card is not showing.
//
// It COPIES the face slice before touching it, and that is the whole
// reason this is a function rather than four lines inside the loop
// above. redactCardForViewer copies the CardView by value, so nilling
// the card's own `castOffers` map pointer is local to this viewer's
// frame — but `Faces` is a slice header, and every viewer's copy
// shares one backing array with the unfiltered GameView. Writing one
// seat's answer into it in place would hand that answer to every
// later FilterViewFor call, which is the bug this whole pass is
// about, arriving through the back door.
//
// One allocation per multi-face card per viewer, and none at all for
// a single-faced card or one whose faces carry no stamps.
func applyFaceCastStampsFor(c *CardView, viewerID string, mine bool) {
	stamped := false
	for i := range c.Faces {
		if c.Faces[i].castOffers != nil {
			stamped = true
			break
		}
	}
	if !stamped {
		return
	}
	faces := make([]CardFaceView, len(c.Faces))
	copy(faces, c.Faces)
	for i := range faces {
		f := &faces[i]
		stamps, ok := f.castOffers[viewerID]
		f.castOffers = nil
		if ok && mine {
			stamps.applyToFace(f)
		}
	}
	c.Faces = faces
}

// redactCardForViewer applies the S13.5 visibility rule to a single
// card. When `known` is true the card keeps every printed
// characteristic; when false, every field derived from the card's
// identity zeroes out — name, type line, costs, faces, abilities,
// catalog flags — so the wire doesn't leak identity.
// face_down_view_test.go pins the survivors as an allowlist. The unexported `knowers` map is always
// cleared on the output so repeated FilterViewFor calls stay
// idempotent.
func redactCardForViewer(c CardView, known bool) CardView {
	out := c
	out.knowers = nil
	out.KnownByYou = known
	// `castOffers` deliberately SURVIVES this function, unlike
	// `knowers`: applyCastStampsFor runs immediately after the zone
	// projection and needs the map to hand this viewer their own cast
	// stamps. It reads the `KnownByYou` set just above before it
	// promotes anything, so a non-knower gets the redaction below and
	// keeps it; and it clears the map itself, so nothing unexported
	// leaves FilterViewFor either way.
	// ADR 0069 decision 6: "may this viewer look at the face" is the
	// rules permission, computed here beside known_by_you and never
	// carried in from the input. A face-down object's knowers are set
	// to the CR 406.3a / CR 702.143d / CR 708.5 answer as it enters
	// the state, so being a knower of a face-down card IS being
	// allowed to look at it.
	out.FaceVisible = c.FaceDown && known
	if known {
		return out
	}
	out.Name = ""
	out.TypeLine = ""
	out.Colors = nil
	out.ScryfallID = ""
	out.Power = 0
	out.NegativePower = 0
	out.Toughness = 0
	out.Counters = nil
	out.IsCommander = false
	out.ManaCost = ""
	out.Abilities = nil
	// #662: the parsed half of Abilities. "Protection from Demons"
	// names a card as loudly as the raw token does, and clearing one
	// without the other would put the leak back.
	out.Protection = nil
	// S22: an alternative cost names the card as loudly as its mana
	// cost does — "Overload {6}{U}" on a face-down card would give
	// away the Cyclonic Rift.
	out.AlternativeCosts = nil
	// #1012: and "this one cannot be cast for the cost in its corner"
	// says the card prints a zone-bound price, which is the same leak
	// one step removed.
	out.AlternativeCostRequired = false
	// ADR 0073: "Kicker {4}" names the card as loudly as an overload
	// cost does, and CR 708.2 leaves a face-down object with no text
	// to offer it from.
	out.OptionalCosts = nil
	// #978: the rest of the announce-time cast surface, for the same
	// reason. Before the exile pass these only ever landed on a card
	// in a zone this filter drops wholesale, so nothing cleared them;
	// a face-down foretold card in exile is a card in a SHARED zone
	// that carries them, and "this spell chooses two of three modes"
	// names a card.
	out.Modes = nil
	out.AdditionalCost = nil
	out.LegalTargets = nil
	out.Clauses = nil
	out.CantCast = ""
	out.TapCost = nil
	// #916: derived from the mana cost, which is cleared above, so
	// it goes with it — "two Phyrexian symbols" on a face-down card
	// would name Dismember out loud.
	out.PhyrexianSymbols = 0
	// #746: a quoted cost clause names the card like its mana cost.
	out.TargetCostNotes = nil
	// S29: "castable from where it sits" is only ever set on cards
	// whose text grants an extra cast zone, so it partitions the
	// card the same weak way `unimplemented` does. Cleared with the
	// rest of the cost surface.
	out.CastableHere = false
	// ADR 0073 §7: "Cast this spell only if you control a legendary
	// creature or planeswalker" says the card is a legendary sorcery,
	// which is more than its mana cost gives away. CR 708.2 also
	// leaves a face-down object with no text for such a clause to be
	// printed in.
	out.CantCast = ""
	// Weak evidence of identity, but evidence: it partitions the
	// card into "prints rules we don't run" or not. Cleared for the
	// same reason as the cost fields above rather than because
	// anyone could read much from it.
	out.Unimplemented = false
	// ADR 0034: the face list names the card twice over — both
	// halves, their costs and their type lines. A face-down Sea Gate
	// Restoration that still shipped "Sea Gate, Reborn // Land" in
	// its faces array would be the loudest leak on the wire.
	out.Faces = nil
	out.Layout = ""
	out.ActiveFace = 0
	// #95: everything below is read off the card's own text or type
	// line — the catalog entry, its abilities, its target prompt — and
	// so names it as surely as the fields above. A face-down Forest
	// that still shipped "Add {G}" in mana_abilities, or a face-down
	// catalog card that still shipped auto=true and target_mode, was
	// the leak: Necropotence's face-down exiles carried both, and so
	// did every card in the owner's own library.
	//
	// What survives is the game state around the card rather than the
	// card: instance id, owner, controller, tapped, damage, combat
	// declarations, goad, attachment and battlefield position.
	out.Auto = false
	out.TargetMode = ""
	out.ManaAbilities = nil
	out.ActivatedAbilities = nil
	// #660: a hand ability quotes the card's text as loudly as its
	// mana cost — "Cycling {3}" on an opponent's face-down hand card
	// would name the Triome. It is also the only ability list on the
	// wire that is not public information in the first place.
	out.ZoneAbilities = nil
	// #1228: and the mana rows beside them. "Exile this card from
	// your hand: Add {R}" names Simian Spirit Guide as surely as
	// "Cycling {3}" names a Triome.
	out.ZoneManaAbilities = nil
	// #658 / #659: "Foretell {1}{U}" and "Suspend 4—{U}" quote the
	// card exactly as a hand ability does, and are hidden for the
	// same reason.
	out.SpecialActions = nil
	out.Restrictions = nil
	out.ExilePlay = nil
	// The hand / command / graveyard stamps are only meaningful to a
	// player who can read the card, and each one quotes it: a mode
	// prompt, an additional-cost label, a target clause's bounds.
	out.LegalTargets = nil
	// #764: the per-clause legal sets name the card as loudly as the
	// single one does — "target creature you control, then target
	// creature or planeswalker you don't control" is the card.
	out.Clauses = nil
	out.Modes = nil
	out.AdditionalCost = nil
	// Type-derived bits. Sickness says "creature without haste",
	// loyalty says "planeswalker", defense and a protector say
	// "battle". Defense is also just the defense counter, and the
	// counters map is already gone.
	out.SummoningSick = false
	out.LoyaltyActivated = false
	out.Defense = 0
	out.ProtectorPlayer = ""
	// ADR 0071: a level says "Class" and a solved flag says "Case",
	// each as loudly as loyalty says "planeswalker" — and CR 708.2
	// gives a face-down permanent no subtypes at all, so it is
	// neither. Cleared with the rest of the type-derived bits.
	out.ClassLevel = 0
	out.Solved = false
	// ADR 0090: a face-down permanent has no prepare spell (CR 708.2)
	// and cannot be prepared, but the field is cleared with the other
	// designations rather than trusted to be false.
	out.Prepared = false
	// #781: a chosen colour and a named tribe are PUBLIC on a card the
	// viewer can see — that is the whole point of the fields — but
	// they are read off the card's own text, so they name it exactly
	// as loudly as the ability lists above. "Elf" on a face-down
	// permanent is Cavern of Souls or Adaptive Automaton; "G" is
	// Coldsteel Heart. CR 708.2 also leaves a face-down permanent with
	// no such ability to have asked the question, so there is nothing
	// true left to say.
	out.ChosenColor = ""
	out.NamedTribe = ""
	out.ChosenName = ""
	// ADR 0083: a token's printed text is public on a token the
	// viewer can see, and a token is always known to every seat
	// (mintTokenLocked adds every seat as a knower), so in practice
	// this line never fires. It is here because the field is read off
	// the object's catalog entry exactly as mana_abilities is, and
	// "when this token dies, you gain 1 life" would name the object
	// as loudly as any of them if a hidden object ever carried one.
	out.TokenText = ""
	return stampFaceDownPublicBody(out, c)
}

// stampFaceDownPublicBody puts the CR 708.2 body back onto a card the
// redaction has just stripped, when that card is a FACE-DOWN PERMANENT
// (ADR 0069 decision 6).
//
// The redaction above clears name, type line, colours and P/T because
// on an ordinary hidden card those fields name it. On a face-down
// permanent they are not the card's — they are the projection the
// engine put in at layer 0, a 2/2 colourless creature with no name
// that identifies nothing and that CR 708.2 makes PUBLIC. Opponents
// have to see the 2/2 to block it, target it and count it.
//
// It re-stamps rather than widening `redactedCardKeys`, so the #646
// zone × viewer table stays exactly as strict as it is for every card
// that is not a face-down permanent, and this one case gets its own
// cells. Everything that names the card — scryfall_id (the art),
// mana_cost, faces, layout, oracle_id, auto, target_mode and the
// ability lists — is still gone, and CatalogKey suppression means most
// of it was never stamped.
//
// The restored fields are taken from `orig` — the view viewOfCard
// built — rather than re-derived, because viewOfCard already read them
// off the layer-0 projection and added the counter deltas the pip
// needs. game.FaceDownBody is consulted only for "is this a CR 708.2
// object", so there is still one definition of the 2/2 for the engine
// and the wire both.
//
// A LISTED body (CR 708.2, #1270) needs nothing here for the same
// reason: the listing is read at layer 0 too, so `orig` already
// carries "Artifact Creature — Cyberman" or "Land — Forest", and a
// listed body is exactly as public as the default one. The kind check
// is still the right gate — a listing only ever rides a permanent
// state (Card.SetFaceDownListed).
// IsFaceDownPermanent reports whether this view is a CR 708.2 object —
// a face-down permanent, which every viewer sees as a 2/2 colourless
// creature with no name (ADR 0069 decision 3), as opposed to a
// face-down card in exile, which has no characteristics at all.
//
// Exported for readers on this side of the wire — the bot's board
// evaluator and prompt renderer — so "is this the public 2/2" has one
// definition there too rather than a re-derivation per reader.
func (c CardView) IsFaceDownPermanent() bool {
	if !c.FaceDown {
		return false
	}
	_, ok := game.FaceDownBody(game.FaceDownKind(c.FaceDownKind))
	return ok
}

func stampFaceDownPublicBody(out, orig CardView) CardView {
	if !out.IsFaceDownPermanent() {
		// A face-down card in EXILE has no characteristics at all
		// (CR 406.3a) and gets nothing back — it is not a permanent,
		// and its real characteristics are exactly what must not
		// reach a non-knower.
		return out
	}
	// Name is "" on the projection, so restoring it is a no-op; it is
	// here to say that the empty name is the OBJECT's name (CR 708.2:
	// no name) and not a redaction.
	out.Name = orig.Name
	out.TypeLine = orig.TypeLine
	out.Colors = orig.Colors
	out.Power = orig.Power
	out.NegativePower = orig.NegativePower
	out.Toughness = orig.Toughness
	// nil today; disguise and cloak's ward is #95's (ADR 0069 §3).
	out.Abilities = orig.Abilities
	// Counters on a face-down permanent are public — they are what
	// makes the restored P/T add up, and a +1/+1 counter on a morph
	// is visible across the table.
	out.Counters = orig.Counters
	// "Creature without haste" is public and combat-relevant: an
	// opponent has to know whether the face-down 2/2 can attack this
	// turn. It says "creature", which is already on the type line.
	out.SummoningSick = orig.SummoningSick
	return out
}

// hideZoneContents returns a copy of z with Cards replaced by an empty
// (but non-nil) slice. Non-nil matters: json.Marshal of a nil []CardView
// is `null`, but every other zone on the wire is `[]`; clients would
// see a shape inconsistency across seats.
func hideZoneContents(z ZoneView) ZoneView {
	return ZoneView{
		Kind:  z.Kind,
		Owner: z.Owner,
		Count: z.Count,
		Cards: []CardView{},
	}
}

// keepKnownTopInLibraryZone hides an opponent's library except for
// its top card, and even that only when the viewer is a knower of it
// (CR 401.5's "play with the top card of your library revealed").
//
// The narrower sibling of keepKnownInHandZone: that one keeps every
// revealed card in the zone, this one considers only the last element,
// which is the top. Count is preserved either way, so the client still
// renders the right number of backs.
//
// Input z is expected to have been projected through redactZone — we
// key off KnownByYou, which redactZone set.
func keepKnownTopInLibraryZone(z ZoneView) ZoneView {
	out := ZoneView{
		Kind:  z.Kind,
		Owner: z.Owner,
		Count: z.Count,
		Cards: []CardView{},
	}
	if n := len(z.Cards); n > 0 && z.Cards[n-1].libraryTop && z.Cards[n-1].KnownByYou {
		out.Cards = append(out.Cards, z.Cards[n-1])
	}
	return out
}

// keepKnownInHandZone drops cards the viewer isn't a knower of
// from an opponent's hand zone. Revealed cards (Thoughtseize
// reveal; scry-to-hand in future) survive; others are stripped
// from the Cards slice. Count is preserved so the client can
// still fan out the right number of face-down backs for the
// unrevealed remainder.
//
// Input z is expected to have already been projected through
// redactZone — we key off KnownByYou, which redactZone set.
func keepKnownInHandZone(z ZoneView) ZoneView {
	out := ZoneView{
		Kind:  z.Kind,
		Owner: z.Owner,
		Count: z.Count,
		Cards: []CardView{},
	}
	// NO CAST-SURFACE FIELD LIST HERE, and that is the point (#1169).
	// This function drops the cards this viewer is not a knower of and
	// nothing else.
	//
	// It used to clear six announce fields by name — the narrowing
	// that keeps "the viewer's own hand" true for a revealed card,
	// because a hand is not a public zone the way a graveyard is —
	// and that was a second list of cast-surface fields in a second
	// place, which is precisely the drift castStamps was created to
	// end. The narrowing itself was right and is unchanged; it is
	// stamped now rather than un-stamped here, by
	// castStamps.publicIn, which is the one function that knows the
	// field list. A field added to CastSurfaceView tomorrow cannot
	// reach a knower's copy of somebody else's hand card by being
	// forgotten here, because there is nothing here to forget.
	//
	// (`legal_targets` and `clauses` went the same way one issue
	// earlier: since #1166 the hand is routed through
	// applyCastStampsFor like every other cast surface, so the
	// owner's answer to "what may YOU target" reaches the owner's
	// frame alone and is already gone by the time a knower's copy
	// gets here.)
	for _, c := range z.Cards {
		if c.KnownByYou {
			out.Cards = append(out.Cards, c)
		}
	}
	return out
}

func viewOfCard(c game.Card) CardView {
	var counters map[string]int
	if len(c.Counters) > 0 {
		counters = make(map[string]int, len(c.Counters))
		for k, v := range c.Counters {
			counters[k] = v
		}
	}
	var knowers map[string]bool
	if len(c.KnownBy) > 0 {
		knowers = make(map[string]bool, len(c.KnownBy))
		for k := range c.KnownBy {
			knowers[k.String()] = true
		}
	}
	// S16 sub-PR 1: read post-layer characteristics off Card.Effective()
	// rather than the raw printed fields. Sub-PR 4 makes TypeLine
	// effective-aware too — Mycosynth Lattice's Layer-4 type-add
	// inserts into Effective().Types and the wire string is rebuilt
	// from the layered type fields. Cards with no static type
	// changes (the vast majority) reproduce the printed type-line
	// byte-for-byte via the parser round-trip.
	eff := c.Effective()
	view := CardView{
		InstanceID: c.InstanceID.String(),
		Name:       eff.Name,
		Owner:      c.Owner.String(),
		Controller: c.Controller.String(),
		ScryfallID: c.ScryfallID,
		TypeLine:   effectiveTypeLine(c, eff),
		Colors:     append([]string(nil), eff.Colors...),
		Protection: viewOfProtection(&c),
		// S16 sub-PR 1 + hotfix: CardView.power / .toughness is the
		// COMBAT-RELEVANT value — effective P/T from the layer engine
		// PLUS the +1/+1 / -1/-1 counter delta. S13.2's CurrentPower /
		// CurrentToughness helpers encode this math so the SBA loop
		// and combat-damage path use the same value the client pip
		// renders. Prior code sent eff.Power / eff.Toughness only,
		// which missed counter deltas — the on-card P/T pip would
		// stay at printed even after +1/+1 counters landed.
		Power:         c.CurrentPower(),
		NegativePower: min(0, c.PowerForComparison()),
		Toughness:     c.CurrentToughness(),
		Tapped:        c.Tapped,
		Counters:      counters,
		IsCommander:   c.IsCommander,
		// BattleX / BattleY are deliberately NOT stamped here: they
		// are battlefield-only, and viewOfCard has no idea which zone
		// it is projecting. viewOfZone fills them in for the
		// battlefield and leaves them nil everywhere else (#29).
		DamageMarked:        c.DamageMarked,
		RegenerationShields: c.RegenerationShields,
		FaceDown:            c.FaceDown,
		FaceDownKind:        string(c.FaceDownKind),
		Auto:                game.IsAutoCard(game.CatalogKey(c)),
		Unimplemented:       game.Unimplemented(c),
		// #992: `target_mode` is the one announce field this
		// function stamps, because it is a pure catalog read and
		// every zone wants it — a spell on the stack renders its
		// target clause too. The cast-surface zones overwrite it
		// through castStamps a moment later with the same answer for
		// the same face; the difference is that they can answer for
		// a face the card is NOT showing and this cannot.
		CastSurfaceView: CastSurfaceView{
			TargetMode: game.TargetModeFor(game.CatalogKey(c)),
		},
		oracleID:      c.OracleID,
		ManaCost:      c.ManaCost,
		ManaAbilities: viewOfManaAbilities(c),
		SummoningSick: game.HasSummoningSickness(&c),
		Abilities:     viewOfAbilityBadges(eff),
		Restrictions:  eff.Restrictions.Names(),
		// #781. Straight off the card, with no zone gate and no
		// catalog lookup: both are cleared on every battlefield exit
		// (game/zone.go, game/entry_tail.go), so "non-empty" already
		// means "a permanent on the battlefield whose controller has
		// answered". This is the one place either is projected.
		ChosenColor: c.ChosenColor,
		NamedTribe:  c.NamedTribe,
		ChosenName:  c.ChosenName,
		knowers:     knowers,
		Layout:      c.Layout,
		Faces:       viewOfFaces(c),
		ActiveFace:  c.ActiveFace,
		// ADR 0083. A token has no printing behind it, so there is no
		// oracle text for the client to fetch by scryfall_id and a
		// token that prints an ability would otherwise reach the board
		// as a bare name. Derived on every read like the emblem's
		// text, so fixing a token's wording reaches a game already in
		// progress; empty for every printed card and every vanilla
		// token, which is what keeps the field additive.
		TokenText: game.TokenTextForCard(c),
	}
	if c.AttackingTarget != uuid.Nil {
		view.AttackingTarget = c.AttackingTarget.String()
	}
	// S27 battles. Both read straight off the card, so they need no
	// game handle and land here rather than in a stamping pass; the
	// attack-target KIND does need one and is stamped in
	// stampCombatTargets.
	if c.ProtectorPlayerID != uuid.Nil {
		view.ProtectorPlayer = c.ProtectorPlayerID.String()
	}
	if c.IsBattle() {
		view.Defense = c.Counters[game.CounterDefense]
	}
	// ADR 0071. Both read straight off the card, like the battle pair
	// above. EnteredBattlefieldAt is the battlefield test viewOfCard
	// has — it is stamped on entry and cleared on exit — and it is
	// what keeps a Class in a hand or a graveyard from claiming a
	// level it does not have (CR 400.7 cleared the designation on the
	// way out; this stops the derived "level 1" from showing up in
	// its place).
	//
	// Both are gated on the permanent still BEING what the badge
	// names, because that is what the client renders: a permanent an
	// effect turned into a creature keeps whatever designation it
	// had, but a level badge on it would be describing a card that
	// is not there any more.
	if c.EnteredBattlefieldAt != 0 {
		if game.IsClass(c) {
			view.ClassLevel = game.ClassLevelOf(c)
		}
		if game.IsCase(c) {
			view.Solved = c.Solved
		}
		view.Prepared = c.Prepared
	}
	if c.BlockingTarget != uuid.Nil {
		view.BlockingTarget = c.BlockingTarget.String()
	}
	if c.GoadedBy != uuid.Nil {
		view.GoadedBy = c.GoadedBy.String()
	}
	if c.IsAttached() {
		id := ""
		if c.AttachedTo.ID != uuid.Nil {
			id = c.AttachedTo.ID.String()
		}
		view.AttachedTo = &TargetRefView{
			Kind: string(c.AttachedTo.Kind),
			ID:   id,
		}
	}
	return view
}

// stampGrantedPermissions fills the public `exile_play` field on every
// card in a zone a granted permission can cover, and files the
// announce-time cast stamps of every seat that may cast one of them.
//
// A stamping pass rather than a line in viewOfCard, because ADR 0066
// moved the permission off the card and onto the player: answering
// "may anyone play this" now needs the game, which viewOfCard does not
// have. The grant itself is public information either way — the
// trigger that created it resolved in the open — so one pass over the
// zone is all it takes, and the per-viewer filter strips the field
// from a card the viewer cannot read.
//
// The STAMPS are not public, and that is the other half of this pass
// (#978, #1022, #1037). A legal target set is narrowed by hexproof and
// by who is asking, so each holder's answer is filed under their own
// seat and FilterViewFor hands out exactly one. Two seats may hold two
// permissions over one card — a Snapcaster flashback under somebody
// else's Wrexial grant, two impulse grants over one exiled card — and
// both now get a picker.
//
// `skip` is the seat whose answer is already the PUBLIC one:
//
//   - a graveyard or a library is a per-seat pile, and its owner's own
//     cast out of it is the same answer for every viewer (a printed
//     flashback cost is public). stampLegalTargets has already stamped
//     it in the exported fields, so this pass files everybody ELSE;
//   - exile has no owner, so nothing about it is public but the grant,
//     and `skip` is uuid.Nil: every holder is filed, including the one
//     the engine names in `exile_play`.
//
// `first` is where to start in the zone. Zero everywhere but a
// LIBRARY, where CR 401.5 makes exactly one card — the top, the last
// element — a cast surface, and walking the rest would be both wrong
// and the most expensive loop in the view.
func stampGrantedPermissions(g *game.Game, seats []PlayerView, zone *ZoneView, live *game.Zone, skip uuid.UUID, first int) {
	if live == nil || len(zone.Cards) != len(live.Cards) {
		// viewOfZone projects a zone card-for-card and in order, so a
		// length mismatch means something moved between the two and
		// the indices no longer line up. Skipping is the safe half of
		// that race: a frame without the stamp, never a stamp on the
		// wrong card.
		return
	}
	if first < 0 {
		// An empty library, whose "top card" index is -1. Nothing to
		// stamp, and the loop below must not start before the slice.
		return
	}
	kind := live.Kind
	for i := first; i < len(zone.Cards); i++ {
		v := &zone.Cards[i]
		card := live.Cards[i]
		perm := g.CastPermissionOnCardForEffect(card, kind)
		if perm == nil {
			// THE fast negative for the per-holder walk below, and it
			// is exact rather than approximate: the only permission
			// that can reach a pile its holder does not own is a
			// STORED one, because a derived standing permission is
			// scoped to its holder's own graveyard or library and
			// carries no ZoneOwner (cast_permission.go). This function
			// answers "is there a stored permission over this card for
			// anyone", so a nil here rules every foreign holder out
			// without deriving a single seat's standing set.
			continue
		}
		v.ExilePlay = exilePlayViewOf(card, perm)
		// Each holder's own answer, including their own grant: the
		// public `exile_play` above names whichever live permission
		// came first, and a client whose seat holds the second one
		// must read theirs or it will render the wrong cost and the
		// wrong face (ADR 0034).
		for _, h := range castHoldersOf(g, seats, skip, card, kind) {
			granted := grantedFace(card, h.grant)
			stamps := castStampsFor(g, h.seat, v, castFace{
				index:    granted.ActiveFace,
				key:      game.CatalogKey(granted),
				manaCost: granted.ManaCost,
			}, kind, h.grant)
			stamps.ExilePlay = exilePlayViewOf(card, h.grant)
			v.stampsFor(h.seat, stamps)
			// #992: and per face, for the grant that leaves the
			// choice open. An adventure card impulse-exiled by
			// Ragavan is the shape — CastableFacesUnder returns both
			// halves, because the grant names none — and without this
			// the client's picker would open on the pile's front face
			// with the back's announce data missing. Not public: a
			// foreign holder's answer is theirs alone, exactly as the
			// card-level stamps above are.
			stampCastableFaces(g, h.seat, v, kind, h.grant, false)
		}
	}
}

// exilePlayViewOf projects one permission over one card into the
// `exile_play` wire shape. One constructor, because the field is
// stamped twice now: publicly, for whichever live permission the
// engine names first, and privately for each holder's own (#1037).
func exilePlayViewOf(card game.Card, perm *game.CastPermission) *ExilePlayView {
	return &ExilePlayView{
		Player:        perm.Player.String(),
		CastOnly:      perm.CastOnly,
		AnyColor:      perm.AnyColor,
		CostOverride:  perm.Cost,
		NotBeforeTurn: perm.NotBeforeTurn,
		Faces:         append([]int(nil), perm.Faces...),
		// CR 107.3b (#831): a cascade hit is granted at {0}, so
		// its printed {X} is not being paid and the only legal
		// announcement is 0. Through CastCostFor rather than a
		// read of the price, because an unpriced permission pays
		// the PRINTED cost and locks nothing. Against the face the
		// permission opens, because that is the cost the cast path
		// will read (ADR 0034).
		XLockedAtZero: game.CastCostFor(grantedFace(card, perm), perm.AlternativeCostFor(card), perm).LocksXAtZero(),
	}
}

// liveCardForView resolves a CardView back to the engine's own Card,
// which the cast gate needs because its predicates read a type line,
// a controller and a zone rather than a projection.
//
// By instance ID through the game's own lookup rather than by index
// into the live zone: a per-viewer projection may be redacted or
// reordered relative to the zone it came from, and an off-by-one here
// would grey the wrong card.
func liveCardForView(g *game.Game, c *CardView) (game.Card, bool) {
	id, err := uuid.Parse(c.InstanceID)
	if err != nil {
		return game.Card{}, false
	}
	return g.LookupCardForEffect(id)
}

// cantCastReason is the printed clause behind a refused cast, for the
// client's grey-out tooltip. The generic fallback should be
// unreachable — CastGateLocked only ever returns a *CantCastError and
// Register refuses a restriction with no label — but a view that
// rendered an empty tooltip would look like a bug in the client.
func cantCastReason(err error) string {
	var cant *game.CantCastError
	if errors.As(err, &cant) && cant.Reason != "" {
		return cant.Reason
	}
	return "An effect prevents casting this spell."
}

// grantedCast answers "may this viewer cast this card out of this
// zone": the permission, or nil.
//
// It used to answer the PRICE half too, and #1012 took that away —
// game.CastOffersForLocked is the one list of CR 118.9 prices a cast
// may claim, it derives the granted offer from this permission
// itself, and a second caller synthesising one beside it is how the
// view and the enumerator came to disagree.
//
// Straight through game.CastPermissionForLocked, which is the same
// function CastSpell validates with and the same one legal/cast.go's
// enumerator asks. That is the whole point: the client cannot be
// shown a cast the engine would refuse, and it cannot be shown the
// wrong price for one it would allow.
//
// The FACE a permission names (ADR 0034) is applied by the caller and
// by castStampsFor, not here: the permission is found by instance,
// and which half of the card it opens is a question about the price
// and the clauses rather than about whether a cast is allowed.
//
// No second liveness check: CastPermissionForLocked has already asked
// CastPermissionActiveForEffect, which is the one function that reads
// a permission's CR 611.2 duration (#945).
//
// Caller must hold g.mu.
func grantedCast(g *game.Game, caster uuid.UUID, card game.Card, kind game.ZoneKind) *game.CastPermission {
	return g.CastPermissionForLocked(caster, card, kind)
}

// stampLibraryTop applies CR 401.5's visibility to the card currently
// on top of each seat's library: "you may look at the top card of your
// library any time" makes it known to its owner, "play with the top
// card of your library revealed" makes it known to everyone.
//
// It writes into the projected card's `knowers` set rather than into
// game state, for two reasons. The view runs under a READ lock, so it
// may not mutate Card.KnownBy; and the top of a library is a POSITION
// rather than a card, so a stamp that persisted would have to be
// invalidated on every draw, mill, shuffle and scry. Re-deriving it
// per frame is both cheaper and exactly CR 401.6 — a top card that
// stops being revealed and is revealed again is a new object, and
// nothing here remembers the old one.
func stampLibraryTop(g *game.Game, seats []PlayerView) {
	if game.CatalogLibraryTopVisible == nil {
		return
	}
	for si := range seats {
		seat := &seats[si]
		owner, err := uuid.Parse(seat.ID)
		if err != nil || len(seat.Library.Cards) == 0 {
			continue
		}
		knowers, topID := g.LibraryTopKnowersLocked(owner)
		if len(knowers) == 0 {
			continue
		}
		top := &seat.Library.Cards[len(seat.Library.Cards)-1]
		if top.InstanceID != topID.String() {
			continue
		}
		top.libraryTop = true
		if top.knowers == nil {
			top.knowers = make(map[string]bool, len(knowers))
		}
		for _, k := range knowers {
			top.knowers[k.String()] = true
		}
	}
}

// grantedFace materialises the face a permission opens on a copy of
// the card, the way the cast path does before it prices anything
// (ADR 0034). A permission that names no face — every impulse,
// airbend, warp and cascade grant — gets the card back untouched, and
// so does one naming several, which is a choice the caster has not
// made yet.
//
// Through NamedFace rather than a read of the slice, so a grant that
// names face 0 (CR 715.4's Adventure creature) is a real answer here
// and not the "no opinion" a bare integer made of it. SetFace(0) is a
// no-op on a card already showing its front, which is every card in
// exile (CR 712.8, MoveCard) — the call is what makes the code say
// the rule rather than rely on the coincidence.
func grantedFace(c game.Card, perm *game.CastPermission) game.Card {
	face, ok := perm.NamedFace()
	if !ok {
		return c
	}
	c.SetFace(face)
	return c
}

// effectiveTypeLine renders the wire `type_line` string from the
// post-layer characteristic. Two short paths:
//
//  1. The card has no static type/sub/supertype changes — eff
//     matches the parsed printed type-line — return c.TypeLine
//     verbatim. This preserves the printed string byte-for-byte
//     for the common case (no Mycosynth Lattice in play, no
//     creature-typing aura, etc.) so the wire output is identical
//     to pre-S16 for the ~99% of cards with no static type
//     changes.
//
//  2. Layer 4 has mutated the type fields — rebuild the canonical
//     "Supertypes Types — Subtypes" string from eff. Mycosynth
//     Lattice + a Forest produces "Basic Land Artifact — Forest".
//
// Empty type fields return "" (matches placeholder demo cards
// pre-S16).
//
// Added in S16 sub-PR 4.
func effectiveTypeLine(c game.Card, eff game.Characteristic) string {
	if len(eff.Types) == 0 && len(eff.Subtypes) == 0 && len(eff.Supertypes) == 0 {
		return c.TypeLine
	}
	// Round-trip parse the printed line; if eff matches printed (no
	// layer mutation), return printed verbatim to avoid drift like
	// double spaces.
	pSuper, pTypes, pSubs := game.ParseTypeLine(c.TypeLine)
	if equalStrings(pSuper, eff.Supertypes) && equalStrings(pTypes, eff.Types) && equalStrings(pSubs, eff.Subtypes) {
		return c.TypeLine
	}
	// Rebuild from eff. Format mirrors Scryfall: supertypes + types
	// joined with spaces; em-dash and subtypes appended only when
	// subtypes exist.
	out := joinSpace(eff.Supertypes)
	if out != "" && len(eff.Types) > 0 {
		out += " "
	}
	out += joinSpace(eff.Types)
	if len(eff.Subtypes) > 0 {
		if out != "" {
			out += " — "
		}
		out += joinSpace(eff.Subtypes)
	}
	return out
}

// joinSpace concatenates ss with single spaces. Empty input → "".
// Cheap; the type-fields slices are usually 1-3 entries.
func joinSpace(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}

// equalStrings reports whether two []string slices contain the
// same elements in the same order. Used to decide when the
// printed type-line and the post-layer types match (the no-op
// fast path in effectiveTypeLine).
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// viewOfManaAbilities projects a card's catalog + synthetic mana
// abilities to wire format. Empty for non-producers (the vast
// majority of cards); populated for Sol Ring / Arcane Signet /
// Birds of Paradise + every basic land (via the synthetic path).
// Cheap — ManaAbilitiesForCard is a single catalog lookup + a
// TypeLine substring scan. Added in S15 sub-PR 2.
// viewOfActivatedAbilities projects a battlefield permanent's
// catalog activated abilities (S21 sub-PR 2). The client renders
// them in the same right-click menu the mana abilities use; the
// cost flags tell it which extra picks to collect before firing
// activate_ability (a sacrifice choice, a target).
// `zone` is where the card is sitting, and abilities that do not
// function there are left out (CR 113.6, #660): a Ketria Triome on
// the battlefield offers no cycling row, and the same card in hand
// offers nothing but one. The index is the ability's index in the
// card's FULL list either way, which is what the engine validates
// against — so a filtered list never renumbers.
func viewOfActivatedAbilities(g *game.Game, c game.Card, caster uuid.UUID, zone game.ZoneKind) []ActivatedAbilityView {
	raw := game.ActivatedAbilitiesForCard(c)
	if len(raw) == 0 {
		return nil
	}
	// #662: every target clause below is chosen BY this object — the
	// permanent, or since #660 the hand card whose cycling ability
	// this is — so the protection check has the source it needs.
	abilitySrc := game.SourceObject(caster, &c)
	// #1210: the board-wide "can't be activated" fast negative, taken
	// once per card rather than once per ability row.
	restricted := g.AnyActivationRestrictionsForEffect()
	var out []ActivatedAbilityView
	for i, a := range raw {
		if !game.AbilityFunctionsFromZone(a, zone) {
			continue
		}
		v := ActivatedAbilityView{
			Index:         i,
			Label:         a.Label,
			TapCost:       a.Cost.Tap,
			SacrificeSelf: a.Cost.SacrificeSelf,
			DiscardSelf:   a.Cost.DiscardSelf,
			ExileSelf:     a.Cost.ExileSelf,
			ManaCost:      a.Cost.Mana,
			LifeCost:      a.Cost.Life,
			SorcerySpeed:  a.SorcerySpeed,
			LoyaltyCost:   a.Cost.Loyalty,
		}
		// CR 606.3 is carried by the loyalty component itself, so a
		// catalog entry doesn't have to remember to set SorcerySpeed
		// — but the client greys on this flag, so stamp it.
		if a.Cost.Loyalty != nil {
			v.SorcerySpeed = true
		}
		// #1208, CR 602.5d / CR 606.3: the row's own timing verdict,
		// from the one read ActivateCatalogAbility and internal/legal
		// ask. Not derived from SorcerySpeed above — a per-player
		// statement can open a sorcery-speed row (The Wandering
		// Emperor, Leonin Shikari) or shut an instant-speed one, and
		// the printed clause says neither.
		if !g.ActivationTimingOpenLocked(caster, c, zone, game.ActivationAbilityOf(a)) {
			v.TimingClosed = true
		}
		// #743: the same closure ActivateCatalogAbility gates on,
		// with the controller as "you" — `caster` is the permanent's
		// controller here (stampActivatedAbilities).
		if a.Condition != nil && !a.Condition(g, caster, c.InstanceID) {
			v.ConditionUnmet = true
		}
		// #1181: the same reader ActivateCatalogAbility refuses on and
		// internal/legal drops the move for.
		if g.AbilityExhausted(caster, c.InstanceID, a) {
			v.Exhausted = true
		}
		// #1210, CR 602.5a: the board-wide "can't be activated"
		// gate's reason, from the one function the engine and the
		// enumerator call. Behind the fast negative taken once for
		// the whole card, because almost no board restricts anything.
		if restricted {
			v.CantActivate = g.CantActivateReasonLocked(caster, c, zone,
				game.ActivationAbility{Label: a.Label})
		}
		if a.Cost.SacrificeOther != nil {
			v.SacrificeLabel = a.Cost.SacrificeOther.Label
			v.SacrificeOptions = sacrificeCostOptions(g, caster, a.Cost.SacrificeOther, c.InstanceID, a.Cost.SacrificeSelf)
		}
		if a.Cost.Crew > 0 {
			v.CrewCost = a.Cost.Crew
			v.CrewOptions = crewOptions(g, caster)
		}
		v.CounterCostView = counterCostView(g, caster, c.InstanceID, a.Cost.RemoveCounters, a.Cost.AddCounter)
		if a.Cost.DemandsX() {
			v.DemandsX = true
			v.MinX = a.Cost.FloorX()
			v.XSlots = a.Cost.XSlots()
		}
		// #917 / #916: the "or 2 life" half of the announcement.
		v.PhyrexianSymbols = phyrexianSymbolsIn(a.Cost.Mana)
		// #1190: the row shows what AbilityManaCostForEffect actually
		// charges, not the printed string alone — `caster` is the
		// permanent's controller here, the same activator the
		// activation path prices for. Left nil (falls back to
		// ManaCost) on a pricing error rather than guessing; a card
		// with no cost modifier reaching it renders byte-identical to
		// ManaCost, which is nearly every ability in the catalog.
		if a.Cost.Mana != "" {
			if charged, err := g.AbilityManaCostForEffect(caster, c, zone, a); err == nil {
				s := charged.String()
				v.ChargedManaCost = &s
			}
		}
		if dc := a.Cost.DiscardCards; dc != nil && dc.N > 0 {
			v.DiscardCostN = dc.N
			v.DiscardCostLabel = dc.Label
			v.DiscardCostOptions = cardIDStrings(g.DiscardCostOptionsForEffect(caster, c.InstanceID, dc))
		}
		// #1213: the return-to-hand component, stamped from the same
		// walk the engine validates against.
		if rc := a.Cost.ReturnToHand; !rc.Empty() {
			v.ReturnLabel = rc.Label
			v.ReturnOptions = returnCostOptions(g, caster, c.InstanceID, rc)
		}
		// #1310: the waterbend clause, sized against the same priced
		// cost the activation path charges. X is not announced yet,
		// so a Waterbend {X} ships Max 0 and DemandsX and the client
		// sizes the picker from the X it collects first.
		if wb := a.Cost.Waterbend; !wb.Empty() {
			budget := 0
			if priced, err := g.AbilityManaCostForEffect(caster, c, zone, a); err == nil {
				budget = game.WaterbendBudget(wb, priced, 0)
			}
			v.Waterbend = viewOfWaterbend(g, caster, game.AbilityWaterbendExclusion(c.InstanceID, a.Cost), wb, budget)
		}
		// #759: the tap-another component, from the same walk.
		if tc := a.Cost.TapOthers; !tc.Empty() {
			v.TapOthersLabel = tc.Label
			v.TapOthersOptions = tapOthersCostOptions(g, caster, c.InstanceID, tc, a.Cost.Tap)
		}
		if a.Targets != nil {
			v.TargetMode = a.Targets.Mode
			// #662: an activated ability's source is the permanent
			// that has it, so the picker hides a creature with
			// protection from that permanent's colour or type.
			v.LegalTargets = abilityLegalTargets(g, abilitySrc, a.Targets)
			v.Clauses = viewOfClauses(g, abilitySrc, a.Targets)
		}
		// #764: a modal activated ability announces its modes with
		// its targets, so the menu needs the same picker a modal
		// spell's hand card gets.
		if a.Modes != nil {
			v.Modes = viewOfModeSpec(g, abilitySrc, a.Modes)
		}
		out = append(out, v)
	}
	return out
}

// cardIDStrings renders a list of instance IDs for the wire.
func cardIDStrings(ids []uuid.UUID) []string {
	if len(ids) == 0 {
		return nil
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

// crewOptions is the set of creatures that can pay a crew cost right
// now: untapped creatures the activator controls (CR 702.122a).
//
// Not built through abilityLegalTargets, and that is the point: crew
// does not TARGET. Routing it through the targeting machinery would
// apply the CR 702 keyword gate, and a hexproof creature you control
// can crew your Vehicle exactly as a hexproof creature you control
// can be sacrificed to a cost. The same reasoning keeps sacrifice
// costs off targetLegalLocked in the engine.
//
// Summoning-sick creatures are included deliberately — tapping to
// crew is not paying a {T} cost (CR 702.122b). Caller must hold g.mu.
func crewOptions(g *game.Game, caster uuid.UUID) *LegalTargetsView {
	out := &LegalTargetsView{Min: 1, Max: 0}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != caster || !c.IsCreature() || c.Tapped {
			continue
		}
		out.Cards = append(out.Cards, c.InstanceID.String())
	}
	return out
}

// counterCostOptions projects game.CounterCostOptionsForEffect — the
// one candidate walk the engine validates against and the move
// enumerator expands — so the client's picker, the bots and the
// engine cannot disagree about which permanent pays (#625). NOT
// abilityLegalTargets: that is the targeting walk, and a cost does
// not target (CR 601.2h), so it would hide a shrouded planeswalker
// the engine accepts — the same rule sacrificeCostOptions follows.
// Its own shape rather than sacrificeCostOptions' LegalTargetsView
// because each option carries the counter kinds that could pay.
// Caller must hold g.mu.
// counterCostView projects the counter components of one ability's
// cost. ONE function for both ability kinds (#789), for the same
// reason there is one view struct and one validator: an activated
// ability's charge counter and a mana ability's are the same
// component, and two projections would be two chances to disagree
// about what the client may offer. Caller must hold g.mu.
func counterCostView(g *game.Game, caster, sourceID uuid.UUID, rc *game.CounterRemovalCost, ac *game.CounterAddCost) CounterCostView {
	var v CounterCostView
	if rc != nil && (rc.N > 0 || rc.Variable) {
		v.CounterCostN = rc.N
		v.CounterCostKind = rc.Counter
		v.CounterCostSelf = rc.From == nil
		v.CounterCostAmong = rc.Among
		v.CounterCostVariable = rc.Variable
		if rc.From != nil {
			v.CounterCostLabel = rc.From.Label
		}
		v.CounterCostOptions = counterCostOptions(g, caster, sourceID, rc)
		if rc.Variable {
			// The ceiling the picker's stepper needs, from the same
			// walk the engine validates against — so a player can
			// never name a number the server then refuses.
			v.CounterCostMax = g.MaxCounterPaymentForEffect(caster, sourceID, rc)
		}
	}
	if ac != nil && ac.N > 0 {
		v.CounterCostAdd = ac.N
		v.CounterCostAddKind = ac.Counter
		// CR 118.3: the only way an add-a-counter cost is unpayable.
		v.CounterAddBlocked = !g.CanPlaceCounterForEffect(caster, sourceID, ac)
	}
	return v
}

func counterCostOptions(g *game.Game, caster, sourceID uuid.UUID, rc *game.CounterRemovalCost) []CounterCostOptionView {
	opts := g.CounterCostOptionsForEffect(caster, sourceID, rc)
	if len(opts) == 0 {
		return nil
	}
	out := make([]CounterCostOptionView, 0, len(opts))
	for _, o := range opts {
		ov := CounterCostOptionView{CardID: o.CardID.String()}
		for _, k := range o.Kinds {
			ov.Kinds = append(ov.Kinds, CounterKindCount{Kind: k.Kind, Count: k.Count})
		}
		out = append(out, ov)
	}
	return out
}

// abilityLegalTargets is the legal set for an ability's TARGET
// clause. It applies the targeting gate (hexproof, shroud), so it is
// only for clauses whose text says "target". A cost clause goes
// through sacrificeCostOptions instead. Caller must hold g.mu.
func abilityLegalTargets(g *game.Game, src game.TargetSource, spec *game.TargetSpec) *LegalTargetsView {
	return abilityClauseView(g.LegalTargetsForEffect(src, spec), spec)
}

// castSourceOf is the TargetSource for a card a player is about to
// cast: the card itself, looked up by the wire instance ID the view
// carries. Falls back to the chooser alone when the ID does not
// resolve, which is the same posture every other lookup in this file
// takes for a card that has moved under the projection.
func castSourceOf(g *game.Game, caster uuid.UUID, instanceID string) game.TargetSource {
	id, err := uuid.Parse(instanceID)
	if err != nil {
		return game.SourceChooser(caster)
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return game.SourceChooser(caster)
	}
	return game.SourceObject(caster, &c)
}

// sacrificeCostOptions is the set of permanents that can pay an
// ability's "sacrifice another" cost, for both activated and mana
// abilities.
//
// It uses SpecCandidatesForEffect, not the targeting gate: paying a
// cost doesn't target (CR 601.2h), so a shrouded creature you control
// can still be sacrificed. The engine validates the payment the same
// way (activated.go). A sacrifice cost can only be paid with
// permanents you control (CR 701.21a), and the spec walk doesn't know
// that, so this filters to the controller. Caller must hold g.mu.
//
// #747: Min / Max are the clause's count (game.SacrificeCostCount) —
// "Sacrifice three Foods" ships 3 / 3 — and the cards come in
// game.SacrificePaymentOrderForEffect's order (tokens first, then
// lower mana value, then the source last, then board order), which is
// the order the legal enumerator takes its payment from. The client's
// "Choose for me" button takes the first N of this list, so it picks
// what a bot would. `sourceID` is the ability's source (uuid.Nil for a
// spell's additional cost); `selfToo` drops the source from the list
// when the cost also sacrifices it, because the engine refuses paying
// one permanent twice.
func sacrificeCostOptions(g *game.Game, controller uuid.UUID, spec *game.TargetSpec, sourceID uuid.UUID, selfToo bool) *LegalTargetsView {
	lt := g.SpecCandidatesForEffect(controller, spec)
	var ids []uuid.UUID
	for _, id := range lt.Cards {
		if selfToo && id == sourceID {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && c.Controller == controller {
			ids = append(ids, id)
		}
	}
	// #1213: the bounds come from the one function the announce path
	// validates against and the enumerator pays from, so the picker
	// cannot enforce a count the engine refuses (#544). A VARIABLE
	// clause ships its real bounds — "one or more" is min 1 with max
	// 0 (LegalTargetsView's "unbounded"), and "Sacrifice X" ships
	// CountFromX so the client knows the count is the X it is about
	// to announce rather than a number it picks.
	lo, hi := game.SacrificeCostBounds(spec, 0)
	out := &LegalTargetsView{Min: lo, Max: hi, CountFromX: spec != nil && spec.CountFromX}
	for _, id := range g.SacrificePaymentOrderForEffect(ids, sourceID) {
		out.Cards = append(out.Cards, id.String())
	}
	return out
}

// returnCostOptions is sacrificeCostOptions one verb over (#1213):
// the permanents that could pay a return-to-hand cost right now, in
// the same payment order, with the same min / max convention — so the
// client reuses one picker and its "Choose for me" button fills with
// what the legal enumerator would have paid.
//
// The walk is the engine's own (ReturnToHandOptionsForEffect), so an
// option offered here is one validateReturnToHandCostLocked accepts.
//
// Caller must hold g.mu.
func returnCostOptions(g *game.Game, controller, sourceID uuid.UUID, rc *game.ReturnToHandCost) *LegalTargetsView {
	if rc.Empty() {
		return nil
	}
	ids := g.ReturnToHandOptionsForEffect(controller, sourceID, rc)
	out := &LegalTargetsView{Min: rc.Count, Max: rc.Count}
	for _, id := range g.SacrificePaymentOrderForEffect(ids, sourceID) {
		out.Cards = append(out.Cards, id.String())
	}
	return out
}

// tapOthersCostOptions is returnCostOptions for a TapOthers component
// (#759): the untapped permanents that could be tapped to pay it, in
// the same payment order, with min and max both the clause's count.
//
// `alsoTapsSource` is the CR 118.3 rule the engine enforces where the
// {T} symbol and this clause meet on one ability: the source is
// tapped by its own half of the cost and cannot be named for the
// other. Filtering it here keeps the picker from offering a choice
// the server refuses.
func tapOthersCostOptions(g *game.Game, controller, sourceID uuid.UUID, tc *game.TapOthersCost, alsoTapsSource bool) *LegalTargetsView {
	if tc.Empty() {
		return nil
	}
	ids := g.TapOthersOptionsForEffect(controller, sourceID, tc)
	out := &LegalTargetsView{Min: tc.Count, Max: tc.Count}
	for _, id := range g.SacrificePaymentOrderForEffect(ids, sourceID) {
		if alsoTapsSource && id == sourceID {
			continue
		}
		out.Cards = append(out.Cards, id.String())
	}
	return out
}

// abilityClauseView converts a candidate set for one ability clause
// into its wire form.
func abilityClauseView(lt game.LegalTargets, spec *game.TargetSpec) *LegalTargetsView {
	// Min / Max come off the spec, exactly as the cast-time path
	// stamps them (viewOfLegalTargets). Omitting them shipped every
	// ability clause to the client as min 0 / max 0 — "unbounded,
	// confirm with nothing picked" — which is right for no clause
	// the catalog actually declares. Teferi's "up to one target"
	// is the first clause whose Min is genuinely 0, so the
	// difference became visible.
	out := &LegalTargetsView{Min: spec.Min, Max: spec.Max}
	for _, id := range lt.Players {
		out.Players = append(out.Players, id.String())
	}
	for _, id := range lt.Cards {
		out.Cards = append(out.Cards, id.String())
	}
	return out
}

// filterToController drops card IDs not controlled by the given
// player. Caller must hold g.mu.
func filterToController(g *game.Game, ids []string, controller uuid.UUID) []string {
	out := ids[:0]
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		if c, ok := g.LookupCardForEffect(parsed); ok && c.Controller == controller {
			out = append(out, id)
		}
	}
	return out
}

// withoutID drops one instance ID from a candidate list. Escape's
// "N OTHER cards from your graveyard" is the only caller: every
// other cost component is paid out of a zone the spell being cast
// has already left.
func withoutID(ids []string, drop string) []string {
	if drop == "" {
		return ids
	}
	out := ids[:0]
	for _, id := range ids {
		if id != drop {
			out = append(out, id)
		}
	}
	return out
}

// viewOfFaces projects a multi-face card's printed faces onto the
// wire. Returns nil for single-faced cards, which is every one of
// the ~33,000 ordinary oracle IDs — the field is omitempty, so their
// CardView is byte-identical to what it was before ADR 0034.
func viewOfFaces(c game.Card) []CardFaceView {
	if len(c.Faces) < 2 {
		return nil
	}
	out := make([]CardFaceView, 0, len(c.Faces))
	for i, f := range c.Faces {
		v := CardFaceView{
			Name:       f.Name,
			TypeLine:   f.TypeLine,
			ManaCost:   f.ManaCost,
			OracleText: f.OracleText,
			Power:      f.Power,
			Toughness:  f.Toughness,
		}
		if c.ScryfallID != "" {
			v.Image = fmt.Sprintf("/cards/%s/image?face=%d", c.ScryfallID, i)
		}
		out = append(out, v)
	}
	return out
}

// viewOfAbilityBadges is the keyword row the client renders: the
// effective ability list, plus `changeling` whenever the permanent is
// every creature type and does not already carry the keyword.
//
// Since #670 "is every creature type" is a layer-4 fact on the
// Characteristic and not a keyword, so a Maskwood Nexus grant no
// longer writes anything into Abilities. The badge is worth keeping —
// a player looking at a Bear that is suddenly a legal Goblin
// Chieftain target deserves to be told why — so the wire PROJECTS it
// from the fact rather than the engine storing it twice.
func viewOfAbilityBadges(eff game.Characteristic) []string {
	if !eff.AllCreatureTypes {
		return eff.Abilities
	}
	for _, a := range eff.Abilities {
		if a == game.KeywordChangeling {
			return eff.Abilities
		}
	}
	return append(append([]string(nil), eff.Abilities...), game.KeywordChangeling)
}

func viewOfManaAbilities(c game.Card) []ManaAbilityView {
	return viewOfManaAbilitiesFromZone(c, game.ZoneBattlefield)
}

// viewOfManaAbilitiesFromZone is viewOfManaAbilities for a card in a
// named zone (#1228, CR 113.6): the abilities that function THERE and
// no others.
//
// The exported `mana_abilities` field goes through it with the
// BATTLEFIELD, which is what that field has always meant — "what does
// this permanent do" — and what keeps a Simian Spirit Guide's hand
// ability off a public projection of the card. A Forest in hand is
// unaffected: its "{T}: Add {G}" declares nothing and so functions
// from the battlefield, which is where the row it publishes is about.
//
// The INDEX is the ability's index in the card's FULL list either
// way, which is what the engine validates against — so a filtered
// list never renumbers, exactly as viewOfActivatedAbilities' does
// not.
func viewOfManaAbilitiesFromZone(c game.Card, zone game.ZoneKind) []ManaAbilityView {
	raw := game.ManaAbilitiesForCard(c)
	if len(raw) == 0 {
		return nil
	}
	var out []ManaAbilityView
	for i, a := range raw {
		if !game.ManaAbilityFunctionsFromZone(a, zone) {
			continue
		}
		out = append(out, ManaAbilityView{
			Index:         i,
			Label:         a.Label,
			TapCost:       a.TapCost,
			SacrificeCost: a.SacrificeCost,
			ExileSelf:     a.ExileSelf,
			LifeCost:      a.LifeCost,
			ManaCost:      a.ManaCost,
			Restrictions:  a.Restrictions,
			Produced:      a.Produced,
		})
	}
	return out
}

// mustParseSeatID parses a seat's UUID string, answering uuid.Nil for
// anything it cannot read. Seat IDs are minted by the engine and are
// always valid; the fallback exists so a projection never panics on a
// hand-built fixture.
func mustParseSeatID(id string) uuid.UUID {
	out, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}
	return out
}

// zoneOf picks one of a player's own zones by kind, nil-safe on both
// the player and the kind. The view's stamping passes index a live
// zone alongside the projected one, and a nil here simply skips the
// stamp rather than reaching for a lookup by ID — which is a scan of
// every zone in the game, per card, per frame.
func zoneOf(p *game.Player, kind game.ZoneKind) *game.Zone {
	if p == nil {
		return nil
	}
	switch kind {
	case game.ZoneGraveyard:
		return p.Graveyard
	case game.ZoneLibrary:
		return p.Library
	case game.ZoneHand:
		return p.Hand
	case game.ZoneCommand:
		return p.Command
	}
	return nil
}

// TableSettingsView is game.TableSettings on the wire (ADR 0075). The
// keys match game.SettingsPatch's, so a client can send back exactly
// the fields it read.
type TableSettingsView struct {
	// UndoLimit is the per-player per-turn undo budget; -1 is
	// unlimited, 0 is no undos.
	UndoLimit int `json:"undo_limit"`
	// UndoScope is "own" or "host_any".
	UndoScope string `json:"undo_scope"`
	// StartingLife is fixed once the game is active.
	StartingLife int `json:"starting_life"`
	// CommanderDamage is the damage from one commander that loses.
	CommanderDamage int `json:"commander_damage"`
	// BotPace is "fast", "normal" or "slow".
	BotPace string `json:"bot_pace"`
	// AllowSpawn is the production spawn switch.
	AllowSpawn bool `json:"allow_spawn"`
}

// ViewOfTableSettings projects one table's settings for a caller
// outside this package — today the lobby's PATCH
// /games/{id}/settings, which answers with the settings it just wrote
// so the client learns whether its patch landed without waiting for
// the WebSocket broadcast. One projection, so the HTTP answer and the
// `settings` object on the game view can never disagree about a key.
func ViewOfTableSettings(s game.TableSettings) TableSettingsView {
	return *viewOfTableSettings(s)
}

func viewOfTableSettings(s game.TableSettings) *TableSettingsView {
	return &TableSettingsView{
		UndoLimit:       s.UndoLimit,
		UndoScope:       string(s.UndoScope),
		StartingLife:    s.StartingLife,
		CommanderDamage: s.CommanderDamage,
		BotPace:         string(s.BotPace),
		AllowSpawn:      s.AllowSpawn,
	}
}
