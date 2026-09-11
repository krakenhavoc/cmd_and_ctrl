package protocol

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
	Turn        TurnView     `json:"turn"`
	// MulligansOpen reflects Game.MulligansOpen — true between Start
	// and the moment all seated players have committed to their
	// opening hand via the keep_hand action. Clients render the
	// keep / mulligan dialog while open. Added in S08.
	MulligansOpen bool `json:"mulligans_open"`
	// Monarch is the player ID currently designated as the monarch,
	// or empty string if no monarch is set. Sandbox marker; the must-
	// attack-when-able and combat-damage-transfer rules are not
	// enforced. Added in S10.
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
	UndoLimit int `json:"undo_limit,omitempty"`
	// StartingSeat is the seat index that took the first turn. Used
	// by the server to enforce the CR 103.7c turn-1 skip-draw rule;
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
	// while any item with split second is on the stack (CR 702.79).
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
}

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
	// "C"). Absent for non-mana choices. Server-side filtered
	// against commander identity before the wire leaves the engine.
	// Added in S15 sub-PR 2.
	ColorOptions []string `json:"color_options,omitempty"`

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

	// SearchMax populates the S22 "search_library" kind: how many of
	// Options the searcher may take. The minimum is always zero —
	// CR 701.19c permits failing to find — so the client's submit
	// button is live from the first render. Absent for other kinds.
	SearchMax int `json:"search_max,omitempty"`
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
}

// ModeSpecView / ModeOptionView are the wire shape of game.ModeSpec
// for the owner's hand cards. Added in S20 sub-PR 4.
type ModeSpecView struct {
	Prompt  string           `json:"prompt"`
	Min     int              `json:"min"`
	Max     int              `json:"max"`
	Options []ModeOptionView `json:"options"`
}

type ModeOptionView struct {
	Label string `json:"label"`
	// TargetMode / LegalTargets mirror CardView.target_mode /
	// legal_targets for the option's own target clause; both absent
	// for untargeted options.
	TargetMode   string            `json:"target_mode,omitempty"`
	LegalTargets *LegalTargetsView `json:"legal_targets,omitempty"`
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
	// caster's own board (CR 701.17b). The picked instance ID rides
	// back on cast_spell's sacrifice_ids. Absent when the cost has
	// no sacrifice component; present-and-empty means the cost is
	// unpayable, which makes the spell uncastable.
	SacrificeOptions *LegalTargetsView `json:"sacrifice_options,omitempty"`
	// Label is the clause as printed ("Discard a card"), shown
	// above the picker.
	Label string `json:"label,omitempty"`
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
	TargetMode   string            `json:"target_mode,omitempty"`
	LegalTargets *LegalTargetsView `json:"legal_targets,omitempty"`
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
	XValue       int             `json:"x_value,omitempty"`
	Distribution map[string]int  `json:"distribution,omitempty"`
	HoldPriority bool            `json:"hold_priority,omitempty"`
	SplitSecond  bool            `json:"split_second,omitempty"`
	// AltCost is the key of the alternative cost this spell was cast
	// for — "overload", "evoke", "cleave" — empty for an ordinary
	// cast (S22). Public information the moment it is announced, and
	// load-bearing for the table: an overloaded Cyclonic Rift is a
	// one-sided board wipe and a hard-cast one bounces a single
	// permanent, so a responder needs to see which is on the stack.
	AltCost string `json:"alt_cost,omitempty"`
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
}

// TargetRefView is the wire shape of a single announce-time target
// slot. Kind is one of "player", "card", "self", "none"; ID is the
// referenced UUID (zero string for "self" / "none"). See
// server/internal/game/stack.go for the canonical taxonomy.
type TargetRefView struct {
	Kind string `json:"kind"`
	ID   string `json:"id,omitempty"`
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
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Seat            int              `json:"seat"`
	Life            int              `json:"life"`
	Poison          int              `json:"poison,omitempty"`
	Energy          int              `json:"energy,omitempty"`
	Library         ZoneView         `json:"library"`
	Hand            ZoneView         `json:"hand"`
	Graveyard       ZoneView         `json:"graveyard"`
	Command         ZoneView         `json:"command"`
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

	// ManaPool is the player's current mana pool projection — one
	// entry per floating mana token, in insertion order. Entries
	// are uppercase single-character mana letters ("W", "U", "B",
	// "R", "G", "C"). Empty / nil when the pool is empty (CR 106.4
	// empties at every step boundary, so this is the common case
	// outside a cast). Drives the S15 ManaPoolPips UI. Added in
	// S15 sub-PR 2.
	ManaPool []string `json:"mana_pool,omitempty"`
}

// LifeChangeView is the wire representation of a single life-change
// log entry. Always emitted as part of PlayerView; the Player's
// canonical history is bounded server-side at MaxLifeHistoryEntries
// (S08), so the wire payload stays small without per-snapshot
// pruning here.
type LifeChangeView struct {
	Delta    int    `json:"delta"`
	NewTotal int    `json:"new_total"`
	At       string `json:"at"` // RFC3339
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
	// Power and Toughness are the parsed printed stats. Zero for
	// non-creatures and for any card with non-numeric printed stats
	// ("*", "1+*"). The client uses Power to label combat-panel
	// creature rows; ResolveCombatDamage uses CurrentPower (base +
	// counter modifiers) on the server side. Both omitempty for
	// non-creatures. Added in S08.
	Power       int            `json:"power,omitempty"`
	Toughness   int            `json:"toughness,omitempty"`
	Tapped      bool           `json:"tapped,omitempty"`
	Counters    map[string]int `json:"counters,omitempty"`
	IsCommander bool           `json:"is_commander,omitempty"`
	// DamageMarked is the damage currently noted on this creature
	// (S13.1 — feeds the lethal-damage SBA). Cleared by the
	// cleanup-step turn-based action and on zone exit. Only
	// meaningful for creatures on the battlefield; omitted when
	// zero. Added in S13.1.
	DamageMarked int `json:"damage_marked,omitempty"`
	// FaceDown reflects Card.FaceDown — a card flipped face-down
	// by morph / manifest / mutate-bottom (CR 708). Distinct from
	// KnownByYou: a face-down creature is face-down to everyone
	// visually, but the morph caster (and anyone else who saw it
	// face-up or via reveal) still has KnownByYou=true so their
	// client can hover-reveal the printed characteristics. Added
	// in S13.5.
	FaceDown bool `json:"face_down,omitempty"`
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

	// oracleID is the card's catalog key, kept off the wire (the
	// client keys art on scryfall_id) so the post-projection
	// legal-target stamp can look the TargetSpec up. Added in S20.
	oracleID string
	// BattleX, BattleY are the normalised battlefield position in
	// [0, 1]. Emitted only for cards on the battlefield (other zones
	// clear them to zero on exit); clients should ignore these fields
	// outside the battlefield zone. omitempty drops them for cards that
	// have never been positioned (e.g. just-played cards waiting for a
	// drag-release).
	BattleX float64 `json:"battle_x,omitempty"`
	BattleY float64 `json:"battle_y,omitempty"`
	// AttackingTarget is the player ID this card is currently
	// declared to attack, or omitted if not declared. Cleared on
	// zone exit and by clear_combat. Added in S08.
	AttackingTarget string `json:"attacking_target,omitempty"`
	// BlockingTarget is the attacker instance ID this card is
	// currently declared to block, or omitted if not declared.
	// Cleared on zone exit and by clear_combat. Added in S08.
	BlockingTarget string `json:"blocking_target,omitempty"`
	// GoadedBy is the player ID who goaded this creature, or empty
	// when not goaded. Cleared on zone exit. Sandbox marker — the
	// must-attack-not-the-goader rule is not enforced. Added in S10.
	GoadedBy string `json:"goaded_by,omitempty"`
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
	// Modes is the S20 sub-PR 4 modal-spell clause for a card in the
	// viewer's own hand / command zone: the prompt, how many options
	// to pick, and each option's label plus — for targeted options —
	// its target mode and legal set right now. Absent for non-modal
	// cards and stripped from opponents' hands. The client shows a
	// mode picker between the X prompt and targeting.
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
	// TapCost is the S22 convoke / waterbend clause for a card in the
	// viewer's own hand / command zone: which of your untapped
	// permanents may be tapped to help pay, and how many. Absent for
	// the overwhelming majority of cards. Optional like the
	// alternative costs — tapping nothing is always a legal cast.
	TapCost *TapCostView `json:"tap_cost,omitempty"`
	// ExilePlay is the S21 sub-PR 6 impulse-exile grant. Present
	// only while the card is in exile with a live permission;
	// absent — which is nearly always — the card is inert exile.
	ExilePlay *ExilePlayView `json:"exile_play,omitempty"`
	// ActivatedAbilities are the CR 602 activated abilities this
	// permanent offers, from its controller's point of view (public
	// information, so present on every viewer's copy). Absent off
	// the battlefield and for cards with none. Added in S21 sub-PR 2.
	ActivatedAbilities []ActivatedAbilityView `json:"activated_abilities,omitempty"`
	// SummoningSick reports CR 302.1 sickness: the creature entered
	// this turn and has no haste, so it can't attack or pay a {T}
	// cost. Battlefield creatures only. Added in S21 sub-PR 2 for
	// the activated-ability menu's affordance; the server does the
	// real check.
	SummoningSick bool `json:"summoning_sick,omitempty"`
	// LoyaltyActivated reports CR 606.5: this planeswalker has
	// already had a loyalty ability activated this turn, so every
	// entry in ActivatedAbilities carrying a LoyaltyCost is greyed
	// until the turn cursor moves on. Battlefield planeswalkers
	// only. Before S27 this lived only in the server's
	// Game.LoyaltyActivatedThisTurn map, which is why the client's
	// canActivateLoyalty had to guess. Added in S27 (#329, #334).
	LoyaltyActivated bool `json:"loyalty_activated,omitempty"`
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
	LifeCost      int    `json:"life_cost,omitempty"`
	SorcerySpeed  bool   `json:"sorcery_speed,omitempty"`
	// LoyaltyCost is the loyalty component of a planeswalker's
	// loyalty ability: +N / 0 / −N (CR 606.1). A POINTER because [0]
	// is a real printed cost and `omitempty` would erase it — the
	// client needs "no loyalty component" and "costs zero loyalty"
	// to stay different, since only the first leaves the ability
	// activatable more than once a turn. Added in S27 (#329, #334).
	LoyaltyCost *int `json:"loyalty_cost,omitempty"`
	// SacrificeLabel / SacrificeOptions describe a "Sacrifice a
	// creature"-style cost: the clause and the permanents the
	// controller may pay with right now. Absent when the cost has
	// no sacrifice component.
	SacrificeLabel   string            `json:"sacrifice_label,omitempty"`
	SacrificeOptions *LegalTargetsView `json:"sacrifice_options,omitempty"`
	// TargetMode / LegalTargets mirror the cast-time targeting
	// fields for an ability that targets.
	TargetMode   string            `json:"target_mode,omitempty"`
	LegalTargets *LegalTargetsView `json:"legal_targets,omitempty"`
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
	// chip, the server does the real CR 118.8 check. A damage RIDER
	// ("This land deals 1 damage to you") is not a cost and does not
	// appear here — it's part of the ability's Label.
	// Added in the S22 mana-ability-rider pass.
	LifeCost int `json:"life_cost,omitempty"`
	// Produced is the raw production string ("{C}{C}",
	// "{W|U|B|R|G}"). Lets the client render the produced-mana
	// pills alongside the activation button even when Label is
	// empty.
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
			UndoLimit:         g.UndoLimit,
			StartingSeat:      g.StartingSeat,
			StackItems:        viewOfStackItemsInStackOrder(g),
			PendingTriggers:   viewOfStackItemSlice(g.PendingTriggers),
			DelayedTriggers:   viewOfDelayedTriggers(g.DelayedTriggers),
			SplitSecondActive: g.SplitSecondActive,
			DiscardPending:    viewOfDiscardPending(g.DiscardPending),
			PendingChoices:    viewOfPendingChoices(g),
		}
		stampLegalTargets(g, view.Seats)
		stampActivatedAbilities(g, &view.Battlefield)
	})
	return view
}

// stampLegalTargets fills CardView.LegalTargets for every hand and
// command-zone card that has a TargetSpec, from its owner's point
// of view. Runs under the read lock ViewOfGame already holds; the
// per-viewer filter strips the field from opponents' hands.
func stampLegalTargets(g *game.Game, seats []PlayerView) {
	for si := range seats {
		seat := &seats[si]
		caster, err := uuid.Parse(seat.ID)
		if err != nil {
			continue
		}
		for _, zone := range []*ZoneView{&seat.Hand, &seat.Command} {
			for ci := range zone.Cards {
				c := &zone.Cards[ci]
				if ms := game.ModeSpecFor(c.oracleID); ms != nil {
					c.Modes = viewOfModeSpec(g, caster, ms)
				}
				if ac := game.AdditionalCostFor(c.oracleID); !ac.Empty() {
					c.AdditionalCost = &AdditionalCostView{
						DiscardCards: ac.DiscardCards,
						Label:        ac.Label,
					}
					if ac.Sacrifice != nil {
						opts := viewOfLegalTargets(g.LegalTargetsForEffect(caster, ac.Sacrifice), ac.Sacrifice)
						opts.Cards = filterToController(g, opts.Cards, caster)
						opts.Players = nil
						c.AdditionalCost.SacrificeOptions = opts
					}
				}
				// S22: convoke / waterbend. Stamped before the target
				// clause because the caster pays it first, and the
				// count of a "X target creatures" clause depends on
				// what they paid.
				if tc := game.TapPermanentsCostFor(c.oracleID); !tc.Empty() {
					c.TapCost = viewOfTapCost(g, caster, c, tc)
				}
				spec := game.TargetSpecFor(c.oracleID)
				// S22: the alternative costs are stamped before the
				// early-out below, because a card can offer one
				// without having any target clause of its own.
				if alts := game.AlternativeCostsFor(c.oracleID); len(alts) > 0 {
					c.AlternativeCosts = viewOfAlternativeCosts(g, caster, spec, alts)
				}
				if spec == nil {
					continue
				}
				c.LegalTargets = viewOfLegalTargets(g.LegalTargetsForEffect(caster, spec), spec)
			}
		}
	}
}

func viewOfLegalTargets(lt game.LegalTargets, spec *game.TargetSpec) *LegalTargetsView {
	view := &LegalTargetsView{Min: spec.Min, Max: spec.Max, CountFromX: spec.CountFromX}
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
func viewOfTapCost(g *game.Game, caster uuid.UUID, c *CardView, tc *game.TapPermanentsCost) *TapCostView {
	v := &TapCostView{
		Key:         tc.Key,
		Label:       tc.Label,
		ColorClause: tc.ColorClause,
		DemandsX:    tc.DemandsX(),
	}
	opts := viewOfLegalTargets(g.LegalTargetsForEffect(caster, tc.Spec), tc.Spec)
	opts.Cards = filterToController(g, opts.Cards, caster)
	opts.Players = nil
	opts.Min, opts.Max, opts.CountFromX = 0, 0, false
	v.Options = opts
	v.Max = game.TapPermanentsBudgetFor(tc, c.ManaCost, 0)
	return v
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
func viewOfAlternativeCosts(g *game.Game, caster uuid.UUID, base *game.TargetSpec, alts []game.AlternativeCost) []AlternativeCostView {
	out := make([]AlternativeCostView, 0, len(alts))
	for i := range alts {
		ac := alts[i]
		v := AlternativeCostView{Key: ac.Key, Label: ac.Label, ManaCost: ac.ManaCost}
		if spec := game.TargetSpecUnderAlternativeCost(base, &ac); spec != nil {
			v.TargetMode = spec.Mode
			v.LegalTargets = viewOfLegalTargets(g.LegalTargetsForEffect(caster, spec), spec)
		}
		out = append(out, v)
	}
	return out
}

// viewOfModeSpec projects a modal card's options with each targeted
// option's legal set from the caster's point of view. Caller must
// hold g.mu.
func viewOfModeSpec(g *game.Game, caster uuid.UUID, ms *game.ModeSpec) *ModeSpecView {
	out := &ModeSpecView{Prompt: ms.Prompt, Min: ms.Min, Max: ms.Max, Options: make([]ModeOptionView, 0, len(ms.Options))}
	for _, o := range ms.Options {
		ov := ModeOptionView{Label: o.Label}
		if o.Targets != nil {
			ov.TargetMode = o.Targets.Mode
			ov.LegalTargets = viewOfLegalTargets(g.LegalTargetsForEffect(caster, o.Targets), o.Targets)
		}
		out.Options = append(out.Options, ov)
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
		if c.oracleID == "" {
			continue
		}
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
		c.ActivatedAbilities = viewOfActivatedAbilities(g, card, controller)
		c.LoyaltyActivated = g.LoyaltyActivatedThisTurn[instanceID]
		stampManaSacrificeOptions(g, card, controller, c.ManaAbilities)
	}
}

// stampManaSacrificeOptions fills the sacrifice clause on a
// permanent's MANA abilities (Ashnod's Altar, Phyrexian Altar).
//
// Split from viewOfManaAbilities because that runs while building the
// base card view, which has no game handle — computing a legal set
// needs one. Same division the activated abilities already use, and
// the same CR 701.17b filter: a sacrifice cost may only be paid with
// permanents you control, which the generic legal-target walk doesn't
// know.
func stampManaSacrificeOptions(g *game.Game, card game.Card, controller uuid.UUID, views []ManaAbilityView) {
	raw := game.ManaAbilitiesForCard(card)
	for i := range views {
		if i >= len(raw) || raw[i].SacrificeOther == nil {
			continue
		}
		views[i].SacrificeLabel = raw[i].SacrificeOther.Label
		opts := abilityLegalTargets(g, controller, raw[i].SacrificeOther)
		if opts == nil {
			continue
		}
		opts.Cards = filterToController(g, opts.Cards, controller)
		opts.Players = nil
		views[i].SacrificeOptions = opts
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
		// PendingChoiceScry — the looked-at cards, top-first, as
		// Options. Scry is "look at", not "reveal": only the chooser
		// was marked a knower, so FilterViewFor redacts these to backs
		// for every other seat and the top of the library stays
		// private.
		if c.Kind == game.PendingChoiceScry && len(c.ScryCards) > 0 {
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
		// PendingChoiceMana carries a color-option list server-
		// filtered against the chooser's commander identity (see
		// ActivateManaAbility). Clone the slice so post-wire
		// mutations on the engine's copy don't leak onto the view.
		if c.Kind == game.PendingChoiceMana && len(c.ColorOptions) > 0 {
			v.ColorOptions = append([]string(nil), c.ColorOptions...)
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
		if c.Kind == game.PendingChoicePickTarget {
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
		XValue:       it.XValue,
		HoldPriority: it.HoldPriority,
		SplitSecond:  it.SplitSecond,
		AltCost:      it.AltCost,
	}
	if len(it.Targets) > 0 {
		view.Targets = make([]TargetRefView, len(it.Targets))
		for i, t := range it.Targets {
			view.Targets[i] = TargetRefView{
				Kind: string(t.Kind),
				ID:   uuidStringOrEmpty(t.ID),
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
		}
	}
	var manaPool []string
	if len(p.ManaPool) > 0 {
		manaPool = make([]string, len(p.ManaPool))
		for i, t := range p.ManaPool {
			manaPool[i] = t.Color
		}
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
		CommanderCasts:    cmdrCasts,
		Counters:          cloneStringIntMap(p.Counters),
		MaxHandSize:       g.EffectiveMaxHandSizeLocked(p),
		ManaPool:          manaPool,
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
	for i, c := range z.Cards {
		cards[i] = viewOfCard(c)
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
//   - Shared zones (battlefield, stack, exile): unchanged.
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
		out.Library = redactZone(p.Library, isKnower)
		out.Hand = redactZone(p.Hand, isKnower)
		out.Graveyard = redactZone(p.Graveyard, isKnower)
		out.Command = redactZone(p.Command, isKnower)
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
			} else {
				out.Hand = keepKnownInHandZone(out.Hand)
			}
			out.Library = hideZoneContents(out.Library)
		}
		seats[i] = out
	}
	return GameView{
		ID:                v.ID,
		State:             v.State,
		Seats:             seats,
		Battlefield:       redactZone(v.Battlefield, isKnower),
		Stack:             redactZone(v.Stack, isKnower),
		Exile:             redactZone(v.Exile, isKnower),
		Turn:              v.Turn,
		MulligansOpen:     v.MulligansOpen,
		Monarch:           v.Monarch,
		Initiative:        v.Initiative,
		Promises:          v.Promises,
		Vote:              v.Vote,
		UndoLimit:         v.UndoLimit,
		StartingSeat:      v.StartingSeat,
		StackItems:        v.StackItems,
		PendingTriggers:   v.PendingTriggers,
		DelayedTriggers:   v.DelayedTriggers,
		SplitSecondActive: v.SplitSecondActive,
		DiscardPending:    v.DiscardPending,
		PendingChoices:    filterPendingChoices(v.PendingChoices, isKnower, viewerID),
	}
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
		if len(c.Options) > 0 {
			opts := make([]CardView, len(c.Options))
			for j, card := range c.Options {
				opts[j] = redactCardForViewer(card, isKnower(card))
			}
			out[i].Options = opts
		}
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

// redactCardForViewer applies the S13.5 visibility rule to a single
// card. When `known` is true the card keeps every printed
// characteristic; when false, name / type_line / scryfall_id /
// power / toughness / counters / is_commander zero out so the wire
// doesn't leak identity. The unexported `knowers` map is always
// cleared on the output so repeated FilterViewFor calls stay
// idempotent.
func redactCardForViewer(c CardView, known bool) CardView {
	out := c
	out.knowers = nil
	out.KnownByYou = known
	if known {
		return out
	}
	out.Name = ""
	out.TypeLine = ""
	out.ScryfallID = ""
	out.Power = 0
	out.Toughness = 0
	out.Counters = nil
	out.IsCommander = false
	out.ManaCost = ""
	out.Abilities = nil
	// S22: an alternative cost names the card as loudly as its mana
	// cost does — "Overload {6}{U}" on a face-down card would give
	// away the Cyclonic Rift.
	out.AlternativeCosts = nil
	out.TapCost = nil
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
	for _, c := range z.Cards {
		if c.KnownByYou {
			// S20: legal targets are computed from the OWNER's point
			// of view and only meaningful to them. S22 folds the
			// alternative-cost offers in for the same reason — they
			// carry their own legal sets.
			c.LegalTargets = nil
			c.Modes = nil
			c.AlternativeCosts = nil
			c.TapCost = nil
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
		// S16 sub-PR 1 + hotfix: CardView.power / .toughness is the
		// COMBAT-RELEVANT value — effective P/T from the layer engine
		// PLUS the +1/+1 / -1/-1 counter delta. S13.2's CurrentPower /
		// CurrentToughness helpers encode this math so the SBA loop
		// and combat-damage path use the same value the client pip
		// renders. Prior code sent eff.Power / eff.Toughness only,
		// which missed counter deltas — the on-card P/T pip would
		// stay at printed even after +1/+1 counters landed.
		Power:         c.CurrentPower(),
		Toughness:     c.CurrentToughness(),
		Tapped:        c.Tapped,
		Counters:      counters,
		IsCommander:   c.IsCommander,
		BattleX:       c.BattleX,
		BattleY:       c.BattleY,
		DamageMarked:  c.DamageMarked,
		FaceDown:      c.FaceDown,
		Auto:          game.IsAutoCard(game.CatalogKey(c)),
		Unimplemented: game.Unimplemented(c),
		TargetMode:    game.TargetModeFor(game.CatalogKey(c)),
		oracleID:      c.OracleID,
		ManaCost:      c.ManaCost,
		ManaAbilities: viewOfManaAbilities(c),
		SummoningSick: game.HasSummoningSickness(&c),
		Abilities:     eff.Abilities,
		knowers:       knowers,
		Layout:        c.Layout,
		Faces:         viewOfFaces(c),
		ActiveFace:    c.ActiveFace,
	}
	if c.AttackingTarget != uuid.Nil {
		view.AttackingTarget = c.AttackingTarget.String()
	}
	if c.BlockingTarget != uuid.Nil {
		view.BlockingTarget = c.BlockingTarget.String()
	}
	if c.GoadedBy != uuid.Nil {
		view.GoadedBy = c.GoadedBy.String()
	}
	// S21 sub-PR 6: the impulse-exile grant rides the card itself,
	// so no per-viewer stamping pass is needed — and it's zeroed as
	// the card leaves exile, so this can't linger on a permanent.
	if c.ExilePlay.Granted() {
		view.ExilePlay = &ExilePlayView{
			Player:       c.ExilePlay.Player.String(),
			CastOnly:     c.ExilePlay.CastOnly,
			AnyColor:     c.ExilePlay.AnyColor,
			CostOverride: c.ExilePlay.CostOverride,
		}
	}
	return view
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
func viewOfActivatedAbilities(g *game.Game, c game.Card, caster uuid.UUID) []ActivatedAbilityView {
	raw := game.ActivatedAbilitiesForCard(c)
	if len(raw) == 0 {
		return nil
	}
	out := make([]ActivatedAbilityView, len(raw))
	for i, a := range raw {
		v := ActivatedAbilityView{
			Index:         i,
			Label:         a.Label,
			TapCost:       a.Cost.Tap,
			SacrificeSelf: a.Cost.SacrificeSelf,
			ManaCost:      a.Cost.Mana,
			LifeCost:      a.Cost.Life,
			SorcerySpeed:  a.SorcerySpeed,
			LoyaltyCost:   a.Cost.Loyalty,
		}
		// CR 606.5 is carried by the loyalty component itself, so a
		// catalog entry doesn't have to remember to set SorcerySpeed
		// — but the client greys on this flag, so stamp it.
		if a.Cost.Loyalty != nil {
			v.SorcerySpeed = true
		}
		if a.Cost.SacrificeOther != nil {
			v.SacrificeLabel = a.Cost.SacrificeOther.Label
			v.SacrificeOptions = abilityLegalTargets(g, caster, a.Cost.SacrificeOther)
			// A sacrifice cost can only be paid with your own
			// permanents (CR 701.17b); the legal-target walk doesn't
			// know that, so filter here.
			v.SacrificeOptions.Cards = filterToController(g, v.SacrificeOptions.Cards, caster)
			v.SacrificeOptions.Players = nil
		}
		if a.Targets != nil {
			v.TargetMode = a.Targets.Mode
			v.LegalTargets = abilityLegalTargets(g, caster, a.Targets)
		}
		out[i] = v
	}
	return out
}

// abilityLegalTargets is the legal set for one of an ability's
// clauses (its target clause, or the permanents that can pay its
// sacrifice cost). Caller must hold g.mu.
func abilityLegalTargets(g *game.Game, caster uuid.UUID, spec *game.TargetSpec) *LegalTargetsView {
	lt := g.LegalTargetsForEffect(caster, spec)
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

func viewOfManaAbilities(c game.Card) []ManaAbilityView {
	raw := game.ManaAbilitiesForCard(c)
	if len(raw) == 0 {
		return nil
	}
	out := make([]ManaAbilityView, len(raw))
	for i, a := range raw {
		out[i] = ManaAbilityView{
			Index:         i,
			Label:         a.Label,
			TapCost:       a.TapCost,
			SacrificeCost: a.SacrificeCost,
			LifeCost:      a.LifeCost,
			Produced:      a.Produced,
		}
	}
	return out
}
