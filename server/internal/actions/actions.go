// Package actions is the glue between the wire protocol and the
// authoritative game state. It takes a decoded ActionPayload from the
// protocol layer, validates its shape, and calls the appropriate
// mutation method on a *game.Game. Errors from Dispatch are returned
// verbatim to the caller, which the ws hub turns into error frames.
//
// actions is the only package that imports both protocol and game.
// That's deliberate: the protocol and game packages stay independent
// of each other (modulo the narrow view.go projection), and actions
// is the one place where wire format meets state mutation.
package actions

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Type is the discriminator for an action payload.
type Type string

const (
	TypeDrawCard               Type = "draw_card"
	TypePlayCard               Type = "play_card"
	TypeMoveCard               Type = "move_card"
	TypeTap                    Type = "tap"
	TypeUntap                  Type = "untap"
	TypeUntapAll               Type = "untap_all"
	TypePassPriority           Type = "pass_priority"
	TypePassTurn               Type = "pass_turn"
	TypeMulligan               Type = "mulligan"
	TypeShuffleLibrary         Type = "shuffle_library"
	TypeChangeLife             Type = "change_life"
	TypeAddCounter             Type = "add_counter"
	TypeSetCommanderDamage     Type = "set_commander_damage"
	TypeSetBattlefieldPosition Type = "set_battlefield_position"
	TypeConcede                Type = "concede"
	TypeKeepHand               Type = "keep_hand"
	TypeDeclareAttacker        Type = "declare_attacker"
	// TypeDeclareAttackers (plural) declares a whole attacking set in
	// one mutation — the wire verb behind the client's "attack with
	// all" cluster (#318). Params carry
	// `{attackers: [{attacker, target}, ...]}`; ineligible entries are
	// skipped silently by game.DeclareAttackers. Deliberately NOT a
	// loop over TypeDeclareAttacker: each room.Apply pushes an undo
	// clone and broadcasts a snapshot, so N creatures would cost N of
	// both and N undo presses to take back.
	TypeDeclareAttackers Type = "declare_attackers"
	TypeDeclareBlocker   Type = "declare_blocker"
	// TypeDeclareBlockers (plural) declares a whole block in one
	// mutation. Params carry `{blocks: [{blocker, attacker}, ...]}`.
	//
	// Not an optimisation like its attacking twin — a RULES need
	// (#750, ADR 0045 addendum Decision 13). A block COUNT (menace's
	// minimum of two, Hungering Hydra's maximum of one) is a property
	// of a whole declaration, so a two-creature menace block is legal
	// only as a pair and cannot be sent as two actions: the first
	// would be refused. Unlike declare_attackers, which skips
	// ineligible entries, this is all-or-nothing — game.DeclareBlockers
	// refuses the set whole and stores none of it.
	TypeDeclareBlockers Type = "declare_blockers"
	// TypeFinishBlocks completes the caller's CR 509.1 block
	// declaration — the "done blocking" / "no blocks" button (#1279,
	// ADR 0045 Decision 38). Whatever the seat has staged is its
	// declaration. Player-scoped and NOT priority-gated: the
	// declaration is a turn-based action a defender takes while the
	// active player still holds priority.
	TypeFinishBlocks Type = "finish_blocks"
	TypeClearCombat  Type = "clear_combat"
	TypeAdvanceStep  Type = "advance_step"
	// S10 Commander UX additions.
	TypeSetMonarch    Type = "set_monarch"
	TypeSetInitiative Type = "set_initiative"
	TypeSetGoaded     Type = "set_goaded"
	TypeSetPoison     Type = "set_poison"
	TypeSetEnergy     Type = "set_energy"
	TypeSetPromise    Type = "set_promise"
	TypeStartVote     Type = "start_vote"
	TypeCastVote      Type = "cast_vote"
	TypeEndVote       Type = "end_vote"
	// S11.
	//
	// TypeSetUndoLimit is DEPRECATED since S35 (ADR 0075 §2.3): it is
	// a one-field alias for TypeSetTableSettings and, like it, is now
	// host-or-admin only. The gate lives at the WebSocket edge, not
	// here — see ws.Client.handleAction and Room.CanManageTable.
	TypeSetUndoLimit Type = "set_undo_limit"
	// S35 — table settings (ADR 0075 §2.3). TypeSetTableSettings
	// carries a partial game.SettingsPatch: only the fields present in
	// the JSON are applied. Host or admin only, gated at the
	// WebSocket edge, and NOT undoable — the ws layer routes it
	// through Room.ApplyExternal so lowering the undo limit cannot be
	// taken back with the undo it was meant to stop.
	//
	// Deliberately absent from internal/legal: a bot never hosts and
	// never changes a table's rules (ADR 0075 §3).
	TypeSetTableSettings Type = "set_table_settings"
	// S13.1 — stack actions. cast_spell is the canonical "play a card
	// from hand" verb (CR 601). Lands route to the battlefield;
	// every other type goes to the stack with a fresh StackMeta
	// entry and the caster retains priority. play_card is kept as
	// the sandbox / admin direct-drop verb.
	TypeCastSpell       Type = "cast_spell"
	TypeCounterSpell    Type = "counter_spell"
	TypeCounterAbility  Type = "counter_ability"
	TypeActivateAbility Type = "activate_ability"
	TypeActivateLoyalty Type = "activate_loyalty"
	TypeAnnounceTrigger Type = "announce_trigger"
	TypeMarkDamage      Type = "mark_damage"
	// S13.2 — counter mechanics. add_player_counter modifies a named
	// player-level counter (poison, energy, experience, rad, plus
	// homebrew names). Drives the SBA loop after the mutation so
	// poison ≥ 10 immediately applies.
	TypeAddPlayerCounter Type = "add_player_counter"
	// S13.4 — interactive cleanup discard + per-player hand-size cap.
	TypeDiscardSelection Type = "discard_selection"
	// TypeResolveChoice drains a PendingChoice entry by ID with
	// the chooser's picks. Used by Thoughtseize-style effects
	// where the chooser isn't the player who's being discarded
	// from (S14). For the simpler S13.4 cleanup case, callers
	// still use TypeDiscardSelection.
	TypeResolveChoice  Type = "resolve_choice"
	TypeSetMaxHandSize Type = "set_max_hand_size"
	// S15 sub-PR 2 — fires a mana ability on a battlefield permanent.
	// Params carry `{card_id, ability_index}`; `card_id` must be on
	// the battlefield controlled by `player`, `ability_index` is the
	// 0-based offset into the catalog's Spec.ManaAbilities (or 0 for
	// the synthetic basic-land ability derived from TypeLine).
	TypeActivateManaAbility Type = "activate_mana_ability"
	// TypeSpecialAction is a CR 116.2 special action: a game action a
	// player takes without using the stack and without passing
	// priority. Params carry `{card_id, kind}` plus the usual
	// `{strict, auto_tap}` payment pair; `kind` is one of `foretell`
	// (CR 702.143a) and `suspend` (CR 702.62a), with `turn_face_up`
	// (CR 116.2g) reserved for #95. An optional `cost` picks between two
	// offers of the same kind on one card (#1391: a plot card on top of
	// the library under Fblthp, Lost on the Range).
	//
	// ONE verb rather than one per keyword, because CR 116 groups
	// these actions precisely because their contract is identical —
	// no stack, no announce, no response window, no targets. The
	// per-kind differences are when it may be taken and what it
	// costs, which is a switch in game.PerformSpecialAction's timing
	// table and not a second verb. ADR 0062 Decision 4; #658 / #659.
	TypeSpecialAction Type = "special_action"
	// S21 sub-PR 1 — sacrifice a permanent you control (CR 701.21).
	// Params carry `{instance_id}`. Distinct from a manual
	// move_card to the graveyard: it emits EventSacrifice, so
	// "whenever you sacrifice" payoffs fire. Only the permanent's
	// controller may sacrifice it.
	TypeSacrificePermanent Type = "sacrifice_permanent"
	// ADR 0121 §3 — the opening roll. roll_opening rolls the caller's
	// d20 in the current round; host_roll_remaining rolls for every seat
	// in the round that has not (host or admin only, gated at the
	// WebSocket edge like set_table_settings); choose_starting_player
	// is the winner's choice of who takes the first turn, params
	// `{seat}`. None of the three mints an undo entry (MintsNoUndo).
	TypeRollOpening          Type = "roll_opening"
	TypeHostRollRemaining    Type = "host_roll_remaining"
	TypeChooseStartingPlayer Type = "choose_starting_player"
	// ADR 0121 §5 — "Roll a die": a d6, a d20 or a coin flip at the
	// table, for fun, params `{die: "d6" | "d20" | "coin"}`. Not a game
	// roll: nothing can trigger on it. Legal at any time while the game
	// is active, the opening roll and the mulligan included; mints no
	// undo entry (MintsNoUndo). The hub allows one per seat per 2 s.
	TypeRollTableDie Type = "roll_table_die"
	// #1530 — "Always ask me to order my triggers": a seat's own
	// preference, params `{always_ask: bool}`. A setting, not a play:
	// legal while the game is active (the opening roll included) and
	// mints no undo entry (MintsNoUndo); undo carries it forward. Never
	// a bot move — the enumerator does not offer it.
	TypeSetTriggerOrderPreference Type = "set_trigger_order_preference"
	// ADR 0127 §3 — a seat's standing answers to repeated prompts,
	// params `{rules: [{key, answer}]}` with answer "always" or "never";
	// the list replaces the seat's rules. A setting, not a play, like
	// set_trigger_order_preference: legal during the opening roll, mints
	// no undo entry, carried across an undo, never a bot move.
	TypeSetAutoAnswers Type = "set_auto_answers"
)

// openingRollActions are the only action types Dispatch accepts while
// the opening roll is open (ADR 0121 §1): the roll's own verbs, a
// concession, and the table settings. An ALLOWLIST, so every other
// type — and every type added after this one — is refused with
// game.ErrOpeningRollOpen until somebody decides otherwise here.
// ADR 0121 §5's roll_table_die is on it: a table roll is never a game
// action, and decision 4 makes it legal at any time.
var openingRollActions = map[Type]struct{}{
	TypeRollOpening:               {},
	TypeHostRollRemaining:         {},
	TypeChooseStartingPlayer:      {},
	TypeRollTableDie:              {},
	TypeSetTriggerOrderPreference: {},
	TypeSetAutoAnswers:            {},
	TypeConcede:                   {},
	TypeSetTableSettings:          {},
	TypeSetUndoLimit:              {},
}

// MintsNoUndo reports whether an action of type t is applied without an
// undo entry (ADR 0121 §3): the opening roll's verbs, a table roll
// (§5), the two table settings verbs, and every action while the
// opening roll is open (in practice a concession). The hub and the bot runner, the two callers
// of ws.Room.Apply, route such an action through Room.ApplyExternal.
//
// The opening roll's dice are simultaneous: an undo entry for each would
// bury every seat's real last action under other seats' rolls and make
// "undo your own most recent action" refuse. The choice of who goes
// first is final, as at a paper table. And no undo, the admin's
// included, may reach back into the roll: undoing a concession made
// mid-roll would silently take back every die rolled since.
//
// It reads the game outside the room lock. The window only ever closes,
// so the one stale answer it can give is "no undo" for an action that
// lands just after the choice: a concession that cannot be taken back.
func MintsNoUndo(g *game.Game, t Type) bool {
	switch t {
	case TypeRollOpening, TypeHostRollRemaining, TypeChooseStartingPlayer,
		TypeRollTableDie, TypeSetTableSettings, TypeSetUndoLimit,
		TypeSetTriggerOrderPreference, TypeSetAutoAnswers:
		return true
	}
	return g.OpeningRollOpen()
}

// ErrUnknownType is returned when Dispatch receives an action type it
// does not recognise. Clients sending unknown types get an error
// frame and the game state is untouched.
var ErrUnknownType = errors.New("actions: unknown action type")

// ErrInvalidPlayer is returned when an action that requires a player
// ID has an empty or malformed Player field.
var ErrInvalidPlayer = errors.New("actions: invalid or missing player ID")

// ErrMissingParams is returned when an action that requires a params
// object is dispatched with no payload. Kept separate from JSON parse
// errors so the client-facing message is clean.
var ErrMissingParams = errors.New("actions: missing required params")

// MaxBulkAttackers caps a single declare_attackers batch. No legal
// board comes close; the cap just bounds the parse cost of a hostile
// payload before it reaches the game lock.
const MaxBulkAttackers = 256

// attackTaxParams is the payment posture both declaration verbs carry
// for the CR 508.1a attack tax (ADR 0080, #1063) — the same trio
// cast_spell, activate_ability and special_action already send.
//
// Embedded in each verb's params struct rather than written twice, so
// the two actions cannot drift on what `auto_tap` means. Every field
// is absent from a payload built before the tax existed and inert at a
// table with no tax on it.
//
// There is deliberately no `strict`: an attack tax waived on paper is
// Propaganda as a blank. See game.payAttackTaxLocked.
type attackTaxParams struct {
	AutoTap       bool     `json:"auto_tap,omitempty"`
	LockedSources []string `json:"locked_sources,omitempty"`
	PhyrexianLife int      `json:"phyrexian_life,omitempty"`
}

// decode turns the wire form into the engine's params, naming the
// action in any parse error the way every other verb does.
func (p attackTaxParams) decode(action string) (game.DeclareAttackersParams, error) {
	out := game.DeclareAttackersParams{
		AutoTap:       p.AutoTap,
		PhyrexianLife: p.PhyrexianLife,
	}
	if len(p.LockedSources) > 0 {
		out.LockedSources = make([]uuid.UUID, 0, len(p.LockedSources))
		for i, raw := range p.LockedSources {
			id, err := uuid.Parse(raw)
			if err != nil {
				return game.DeclareAttackersParams{}, fmt.Errorf("%s locked_sources[%d]: %w", action, i, err)
			}
			out.LockedSources = append(out.LockedSources, id)
		}
	}
	return out, nil
}

// ErrEmptyAttackerSet is returned when declare_attackers arrives with
// an empty `attackers` list. "Attack with nobody" is the default
// state of the step, not an action — a player who wants it passes
// priority instead.
var ErrEmptyAttackerSet = errors.New("actions: declare_attackers needs at least one attacker")

// ErrTooManyAttackers is returned when a declare_attackers batch
// exceeds MaxBulkAttackers entries.
var ErrTooManyAttackers = errors.New("actions: declare_attackers batch is too large")

// MaxBulkBlockers caps a single declare_blockers batch, for the same
// reason MaxBulkAttackers caps its twin: bound the parse cost of a
// hostile payload before it reaches the game lock.
const MaxBulkBlockers = 256

// ErrEmptyBlockerSet is returned when declare_blockers arrives with an
// empty `blocks` list. "Block with nobody" is the default state of the
// step, not an action — a defender who wants it passes priority.
var ErrEmptyBlockerSet = errors.New("actions: declare_blockers needs at least one blocker")

// ErrTooManyBlockers is returned when a declare_blockers batch exceeds
// MaxBulkBlockers entries.
var ErrTooManyBlockers = errors.New("actions: declare_blockers batch is too large")

// ErrNotPriorityHolder is returned when pass_priority is dispatched by
// a seated player who does not currently hold priority. Admin /
// spectator callers (Caller == uuid.Nil) bypass this check.
var ErrNotPriorityHolder = errors.New("actions: caller does not hold priority")

// ErrNotActivePlayer is returned when pass_turn is dispatched by a
// seated player whose seat is not the active turn's seat. Admin /
// spectator callers (Caller == uuid.Nil) bypass this check.
var ErrNotActivePlayer = errors.New("actions: caller is not the active player")

// ErrPlayerCallerMismatch is returned when a seated player issues an
// action whose Player field targets a different seat (e.g. draw_card
// with someone else's player ID, change_life trying to drain an
// opponent). Admin / spectator callers bypass this check.
var ErrPlayerCallerMismatch = errors.New("actions: caller may not act on another player's behalf")

// Action is the decoded form of a wire ActionPayload, with the player
// ID already parsed as a uuid.UUID and Params still raw JSON for the
// individual handler to decode. Caller is the authenticated player ID
// of the connection that sent the frame; the hub stamps it after
// Decode and the gated action branches in Dispatch validate against
// it. uuid.Nil means "no seat bound" (admin / spectator) and bypasses
// per-seat checks — admins may need to act on behalf of any seat.
//
// Admin is set by the hub for a connection authenticated as an admin
// (ws.Binding.Admin: the shared token, or a signed-in person on the
// admin allowlist, ADR 0110 §3). An admin bound to a seat plays as that
// seat — Caller is the seat and every per-seat gate applies — except
// for the sandbox card overrides (overrideCaller), which are the admin
// context menu's whole point.
type Action struct {
	Type   Type
	Player uuid.UUID
	Caller uuid.UUID
	Admin  bool
	Params json.RawMessage
}

// Decode parses a protocol.ActionPayload-shaped map into an Action.
// The ActionPayload's Player field is parsed as a uuid.UUID if
// present; Params is passed through unchanged. Does NOT validate that
// Type is a known action — that's Dispatch's job.
func Decode(payloadType, payloadPlayer string, payloadParams json.RawMessage) (Action, error) {
	a := Action{
		Type:   Type(payloadType),
		Params: payloadParams,
	}
	if payloadPlayer != "" {
		id, err := uuid.Parse(payloadPlayer)
		if err != nil {
			return Action{}, fmt.Errorf("%w: %q", ErrInvalidPlayer, payloadPlayer)
		}
		a.Player = id
	}
	return a, nil
}

// requirePriorityHolder verifies that caller (an authenticated player
// ID, possibly uuid.Nil for admin / spectator) is the seated player
// whose seat currently holds priority. uuid.Nil bypasses the check —
// admins may need to advance the game on a player's behalf.
func requirePriorityHolder(g *game.Game, caller uuid.UUID) error {
	if caller == uuid.Nil {
		return nil
	}
	snap := g.Snapshot()
	if snap.State != game.StateActive {
		// Let the underlying mutation surface ErrGameNotActive with its
		// canonical message.
		return nil
	}
	if snap.Turn.PriorityHolder < 0 || snap.Turn.PriorityHolder >= len(snap.Seats) {
		// #1501 / CR 509.1: priority PARKED for a block declaration
		// is held by nobody, so nobody may cast, activate or pass
		// before the defenders have declared. The other NoPriority
		// steps (untap, cleanup) keep the old pass-through.
		if snap.Turn.PriorityHolder == game.NoPriority && snap.Turn.Step == game.StepDeclareBlockers {
			return ErrNotPriorityHolder
		}
		return nil
	}
	holder := snap.Seats[snap.Turn.PriorityHolder]
	if holder == nil || holder.ID != caller {
		return ErrNotPriorityHolder
	}
	return nil
}

// overrideCaller is the caller the sandbox card overrides check
// (move_card, tap and untap, add_counter, sacrifice_permanent,
// mark_damage, set_battlefield_position): uuid.Nil — "the admin" — for
// an admin connection, so an admin seated at the table can still drive
// any card from the admin context menu (ADR 0110 §3, owner requirement
// of 2026-10-02: an admin plays as their own seat and keeps the admin
// menu). Every other gate, combat declarations included, reads Caller.
func overrideCaller(a Action) uuid.UUID {
	if a.Admin {
		return uuid.Nil
	}
	return a.Caller
}

// requireCardController verifies that caller is the current controller
// of the card with the given instance ID. uuid.Nil (admin / spectator)
// bypasses the check. If the card is not present in any zone, the
// helper returns nil so the underlying mutation can surface the
// canonical ErrCardNotFound — keeping "I don't know about that card"
// distinct from "you don't control that card" on the wire.
func requireCardController(g *game.Game, caller uuid.UUID, instanceID uuid.UUID) error {
	if caller == uuid.Nil {
		return nil
	}
	controller, ok := g.ControllerOfCard(instanceID)
	if !ok {
		return nil
	}
	if controller != caller {
		return game.ErrCardCallerMismatch
	}
	return nil
}

// requireActivePlayer verifies that caller is the seated player whose
// seat is the active turn's seat. uuid.Nil (admin / spectator)
// bypasses the check.
func requireActivePlayer(g *game.Game, caller uuid.UUID) error {
	if caller == uuid.Nil {
		return nil
	}
	snap := g.Snapshot()
	if snap.State != game.StateActive {
		return nil
	}
	if snap.Turn.ActiveSeat < 0 || snap.Turn.ActiveSeat >= len(snap.Seats) {
		return nil
	}
	active := snap.Seats[snap.Turn.ActiveSeat]
	if active == nil || active.ID != caller {
		return ErrNotActivePlayer
	}
	return nil
}

// unmarshalParams is a small helper that returns a friendly
// "missing required params" error for nil or empty Params, and
// otherwise decodes into dest. Without this, every action that
// takes params and receives an empty payload leaks the raw
// "unexpected end of JSON input" message through to the wire.
func unmarshalParams(raw json.RawMessage, actionType Type, dest any) error {
	if len(raw) == 0 {
		return fmt.Errorf("%w: %s", ErrMissingParams, actionType)
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("%s params: %w", actionType, err)
	}
	return nil
}

// playerScopedActions are the action types that carry a Player field
// identifying which seat the action operates on. For these, a seated
// (non-admin) caller may only act on their own seat — the wire
// Player must match Caller. Card-instance-scoped actions (tap,
// move_card, add_counter, set_battlefield_position, declare_attacker,
// declare_blocker) are gated separately via requireCardController
// against the card's current Controller — see the case branches in
// Dispatch. clear_combat is intentionally still loose: it's the
// "combat is wedged" escape hatch and any seated player may invoke it.
var playerScopedActions = map[Type]struct{}{
	TypeDrawCard:       {},
	TypePlayCard:       {},
	TypeCastSpell:      {},
	TypeUntapAll:       {},
	TypeMulligan:       {},
	TypeShuffleLibrary: {},
	TypeChangeLife:     {},
	TypeConcede:        {},
	TypeKeepHand:       {},
	// ADR 0121 §3: a seat rolls its own opening die, and only the
	// winner chooses who goes first. The admin acts for any seat.
	TypeRollOpening:          {},
	TypeChooseStartingPlayer: {},
	// ADR 0121 §5: a seat rolls its own table die.
	TypeRollTableDie: {},
	// #1530: a seat sets only its own preference.
	TypeSetTriggerOrderPreference: {},
	// ADR 0127 §3: and only its own standing answers.
	TypeSetAutoAnswers: {},
	// Poison and energy follow change_life's posture: the affected
	// player adjusts their own counters in the sandbox. Monarch and
	// initiative are NOT player-scoped — any seated player may flip
	// the marker because card effects routinely make someone else
	// the monarch.
	TypeSetPoison:           {},
	TypeSetEnergy:           {},
	TypeAddPlayerCounter:    {},
	TypeDiscardSelection:    {},
	TypeResolveChoice:       {},
	TypeSetMaxHandSize:      {},
	TypeActivateManaAbility: {},
	// A special action is taken on a card in your OWN hand
	// (CR 702.143a, CR 702.62a), so a seat may never take one for
	// another seat.
	TypeSpecialAction: {},
	// Cast vote on behalf of self only — the seated voter is the
	// authoritative caller. start_vote is also self-driven (the
	// initiator is `Player`) but the "any caller may start a vote"
	// posture matches set_monarch / set_initiative — not scoped.
	TypeCastVote: {},
	// #1279: a seat finishes its OWN block declaration, never
	// another seat's — finishing a defender early would decide their
	// blocks for them.
	TypeFinishBlocks: {},
}

// Dispatch applies an action to a game. Returns nil on success, an
// action-specific error on failure. Dispatch itself is stateless; all
// state lives on the game.
//
// #1289, CR 704.3: an action can be the answer that finishes a paused
// resolution, and a finished resolution is owed state-based actions
// and the trigger drain before anyone acts again. Most answer paths
// run that boundary themselves. SettleResolution is the backstop for
// the ones that do not, and a no-op otherwise.
//
// #1501: SettleBlockDeclaration is the same backstop for a block
// declaration whose priority is parked with nobody left declaring — a
// sandbox verb that took the last attacker out of combat, say.
func Dispatch(g *game.Game, a Action) error {
	err := dispatch(g, a)
	g.SettleResolution()
	g.SettleBlockDeclaration()
	return err
}

func dispatch(g *game.Game, a Action) error {
	// ADR 0121 §1: while the opening roll is open, nothing but the
	// roll, a concession and the table settings happens.
	if _, ok := openingRollActions[a.Type]; !ok && g.OpeningRollOpen() {
		return game.ErrOpeningRollOpen
	}
	// Player-scoped guard: a seated player may not target a different
	// seat. Admin / spectator (Caller == uuid.Nil) bypasses so a
	// trusted moderator can advance any seat.
	if _, scoped := playerScopedActions[a.Type]; scoped {
		if a.Caller != uuid.Nil && a.Player != uuid.Nil && a.Caller != a.Player {
			return ErrPlayerCallerMismatch
		}
	}
	switch a.Type {
	case TypeDrawCard:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		return g.DrawCard(a.Player)

	case TypePlayCard:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			InstanceID string `json:"instance_id"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("play_card instance_id: %w", err)
		}
		return g.PlayCard(a.Player, instanceID)

	case TypeCastSpell:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		if err := requirePriorityHolder(g, a.Caller); err != nil {
			return err
		}
		var p struct {
			InstanceID   string           `json:"instance_id"`
			FromZone     string           `json:"from_zone,omitempty"`
			Targets      []castTargetWire `json:"targets,omitempty"`
			Modes        []int            `json:"modes,omitempty"`
			XValue       int              `json:"x_value,omitempty"`
			Distribution map[string]int   `json:"distribution,omitempty"`
			HoldPriority bool             `json:"hold_priority,omitempty"`
			SplitSecond  bool             `json:"split_second,omitempty"`
			// S15 sub-PR 3 — strict-mode flag + override hatch.
			// Strict comes from the client's gameplay.strictMana
			// setting; ForceCast is set by the override toast that
			// follows an insufficient_mana error frame.
			Strict    bool `json:"strict,omitempty"`
			ForceCast bool `json:"force_cast,omitempty"`
			// S15 sub-PR 5 — auto-tap-and-cast. AutoTap asks the
			// server to plan + execute a tap pass against the
			// caller's untapped permanents before the strict gate
			// runs. LockedSources is the lock-tap UI's exclusion
			// list (permanents the player has reserved).
			AutoTap       bool     `json:"auto_tap,omitempty"`
			LockedSources []string `json:"locked_sources,omitempty"`
			// S21 sub-PR 5 — cards paid to an additional cost of the
			// form "As an additional cost to cast this spell,
			// discard a card". Empty for every card without one.
			DiscardIDs []string `json:"discard_ids,omitempty"`
			// S21 sub-PR 6 — the permanent paid to a "sacrifice a
			// creature" additional cost (Village Rites).
			SacrificeIDs []string `json:"sacrifice_ids,omitempty"`
			// ADR 0073 (#664) — the optional additional costs being
			// paid (CR 601.2b): kicker, multikicker, buyback, as
			// POSITIONS in the card's OptionalCosts slice, repeated
			// once per payment for a multikicker. Absent is the
			// ordinary "decline them all" case, and absent on a card
			// that offers none is every cast the engine has ever had.
			OptionalCosts []int `json:"optional_costs,omitempty"`
			// ADR 0100 §2 — which branch of an either/or additional
			// cost is paid, as an index into the card's branches
			// ("sacrifice an artifact or discard a card"). Required on
			// a card with such a cost and refused on any other; a
			// pointer, so an absent field is not branch 0.
			CostBranch *int `json:"cost_branch,omitempty"`
			// ADR 0089 (#1267) — the opponent a gift is promised to
			// (CR 702.174a). Present exactly when optional_costs
			// names the card's gift cost.
			GiftOpponent string `json:"gift_opponent,omitempty"`
			// #1703 — the creatures tapped to pay an announced
			// teamwork cost (CR 702.194a), and the one creature an
			// announced blight cost puts its -1/-1 counters on
			// (CR 701.68a). Absent unless optional_costs names one.
			TeamworkIDs []string `json:"teamwork_ids,omitempty"`
			BlightIDs   []string `json:"blight_ids,omitempty"`
			// ADR 0100 amendment 2026-10-07 — the one card an
			// announced reveal / behold branch shows ("reveal an Elf
			// card from your hand or pay {3}"). Absent unless the
			// chosen cost_branch reveals.
			RevealIDs []string `json:"reveal_ids,omitempty"`
			// S22 — the untapped permanents tapped to help pay
			// (convoke, waterbend). Optional even on a card that
			// offers the cost: tapping nothing and paying the whole
			// cost with mana is always legal.
			TapIDs []string `json:"tap_ids,omitempty"`
			// S22 — the alternative cost being paid INSTEAD of the
			// mana cost: the key of one of the card's declared
			// alternative costs ("overload", "evoke", "cleave").
			// Empty is the ordinary "pay the printed cost" case.
			AlternativeCost string `json:"alternative_cost,omitempty"`
			// S28 — the card paid to the non-mana half of that
			// alternative cost: Force of Will's pitched blue card,
			// Daze's returned Island, Solitude's evoke pitch. Exactly
			// one entry when the claimed cost charges one, absent
			// otherwise.
			AltCostIDs []string `json:"alt_cost_ids,omitempty"`
			// ADR 0100 — delve_ids names the cards in the caster's
			// graveyard exiled to delve the spell (CR 702.66a), each
			// paying for {1} of the generic mana. Absent pays the
			// whole cost with mana.
			DelveIDs []string `json:"delve_ids,omitempty"`
			// ADR 0034 — which printed face of a multi-face card is
			// being cast or played. Absent (0) is the front face,
			// which is the right answer for every single-faced card
			// and for any client that predates the face picker.
			Face int `json:"face,omitempty"`
			// Fuse casts both halves of a split card with fuse from
			// hand (CR 702.102a, ADR 0103).
			Fuse bool `json:"fuse,omitempty"`
			// CR 107.4 / CR 601.2b (#787) — how many of the cost's
			// Phyrexian symbols are being paid with 2 life each
			// instead of mana. Absent (0) pays every symbol with its
			// coloured half, which is what every client that predates
			// hybrid Phyrexian mana sends.
			PhyrexianLife int `json:"phyrexian_life,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("cast_spell instance_id: %w", err)
		}
		params := game.CastSpellParams{
			FromZone:        p.FromZone,
			Modes:           append([]int(nil), p.Modes...),
			XValue:          p.XValue,
			HoldPriority:    p.HoldPriority,
			SplitSecond:     p.SplitSecond,
			Strict:          p.Strict,
			ForceCast:       p.ForceCast,
			AutoTap:         p.AutoTap,
			AlternativeCost: p.AlternativeCost,
			Face:            p.Face,
			Fuse:            p.Fuse,
			PhyrexianLife:   p.PhyrexianLife,
		}
		if len(p.DiscardIDs) > 0 {
			params.DiscardIDs = make([]uuid.UUID, 0, len(p.DiscardIDs))
			for i, raw := range p.DiscardIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("cast_spell discard_ids[%d]: %w", i, err)
				}
				params.DiscardIDs = append(params.DiscardIDs, id)
			}
		}
		if len(p.SacrificeIDs) > 0 {
			params.SacrificeIDs = make([]uuid.UUID, 0, len(p.SacrificeIDs))
			for i, raw := range p.SacrificeIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("cast_spell sacrifice_ids[%d]: %w", i, err)
				}
				params.SacrificeIDs = append(params.SacrificeIDs, id)
			}
		}
		// ADR 0073: indices, not IDs, so there is nothing to parse —
		// but they are copied rather than aliased, because the
		// decoded payload does not outlive the dispatch and the
		// engine stamps this slice onto a stack item that does.
		if len(p.OptionalCosts) > 0 {
			params.OptionalCosts = append([]int(nil), p.OptionalCosts...)
		}
		if p.CostBranch != nil {
			b := *p.CostBranch
			params.CostBranch = &b
		}
		if p.GiftOpponent != "" {
			id, err := uuid.Parse(p.GiftOpponent)
			if err != nil {
				return fmt.Errorf("cast_spell gift_opponent: %w", err)
			}
			params.GiftOpponent = id
		}
		if len(p.AltCostIDs) > 0 {
			params.AltCostIDs = make([]uuid.UUID, 0, len(p.AltCostIDs))
			for i, raw := range p.AltCostIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("cast_spell alt_cost_ids[%d]: %w", i, err)
				}
				params.AltCostIDs = append(params.AltCostIDs, id)
			}
		}
		if len(p.TapIDs) > 0 {
			params.TapIDs = make([]uuid.UUID, 0, len(p.TapIDs))
			for i, raw := range p.TapIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("cast_spell tap_ids[%d]: %w", i, err)
				}
				params.TapIDs = append(params.TapIDs, id)
			}
		}
		for _, l := range []struct {
			name string
			raw  []string
			dst  *[]uuid.UUID
		}{
			{"teamwork_ids", p.TeamworkIDs, &params.TeamworkIDs},
			{"blight_ids", p.BlightIDs, &params.BlightIDs},
			{"reveal_ids", p.RevealIDs, &params.RevealIDs},
			{"delve_ids", p.DelveIDs, &params.DelveIDs},
		} {
			for i, raw := range l.raw {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("cast_spell %s[%d]: %w", l.name, i, err)
				}
				*l.dst = append(*l.dst, id)
			}
		}
		if len(p.LockedSources) > 0 {
			params.LockedSources = make([]uuid.UUID, 0, len(p.LockedSources))
			for i, raw := range p.LockedSources {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("cast_spell locked_sources[%d]: %w", i, err)
				}
				params.LockedSources = append(params.LockedSources, id)
			}
		}
		if len(p.Targets) > 0 {
			params.Targets = make([]game.TargetRef, 0, len(p.Targets))
			for i, t := range p.Targets {
				ref, err := t.toRef()
				if err != nil {
					return fmt.Errorf("cast_spell targets[%d]: %w", i, err)
				}
				params.Targets = append(params.Targets, ref)
			}
		}
		dist, err := parseDistribution(p.Distribution)
		if err != nil {
			return fmt.Errorf("cast_spell %w", err)
		}
		params.Distribution = dist
		return g.CastSpell(a.Player, instanceID, params)

	case TypeMoveCard:
		var p struct {
			Src         zoneRefWire `json:"src"`
			Dst         zoneRefWire `json:"dst"`
			InstanceID  string      `json:"instance_id"`
			AsCommander bool        `json:"as_commander,omitempty"`
			ToBottom    bool        `json:"to_bottom,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("move_card instance_id: %w", err)
		}
		srcRef, err := p.Src.toRef()
		if err != nil {
			return fmt.Errorf("move_card src: %w", err)
		}
		dstRef, err := p.Dst.toRef()
		if err != nil {
			return fmt.Errorf("move_card dst: %w", err)
		}
		if err := requireCardController(g, overrideCaller(a), instanceID); err != nil {
			return err
		}
		// #170: `to_bottom` seats the card at the bottom of the
		// destination instead of the top. Only meaningful for a
		// library destination; the admin context menu's "Library
		// (bottom)" override is the only caller today.
		if p.ToBottom {
			return g.MoveCardByIDToBottom(srcRef, dstRef, instanceID, p.AsCommander)
		}
		return g.MoveCardByIDAsCommander(srcRef, dstRef, instanceID, p.AsCommander)

	case TypeTap, TypeUntap:
		var p struct {
			InstanceID string `json:"instance_id"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("tap instance_id: %w", err)
		}
		if err := requireCardController(g, overrideCaller(a), instanceID); err != nil {
			return err
		}
		return g.TapCard(instanceID, a.Type == TypeTap)

	case TypeUntapAll:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		return g.UntapAll(a.Player)

	case TypePassPriority:
		if err := requirePriorityHolder(g, a.Caller); err != nil {
			return err
		}
		return g.PassPriority()

	case TypePassTurn:
		if err := requireActivePlayer(g, a.Caller); err != nil {
			return err
		}
		return g.PassTurn()

	case TypeMulligan:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			HandSize int `json:"hand_size"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		return g.Mulligan(a.Player, p.HandSize)

	case TypeShuffleLibrary:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		return g.ShuffleLibrary(a.Player)

	case TypeChangeLife:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Delta int `json:"delta"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		_, err := g.ChangePlayerLife(a.Player, p.Delta)
		return err

	case TypeAddCounter:
		var p struct {
			InstanceID string `json:"instance_id"`
			Name       string `json:"name"`
			Delta      int    `json:"delta"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("add_counter instance_id: %w", err)
		}
		if err := requireCardController(g, overrideCaller(a), instanceID); err != nil {
			return err
		}
		return g.AddCounter(instanceID, p.Name, p.Delta)

	case TypeSetCommanderDamage:
		// `from` is a COMMANDER CARD instance ID since S25 (#77); it
		// was an opposing player ID before the CR 903.10a rekey.
		var p struct {
			From   string `json:"from"`
			To     string `json:"to"`
			Amount int    `json:"amount"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		from, err := uuid.Parse(p.From)
		if err != nil {
			return fmt.Errorf("set_commander_damage from: %w", err)
		}
		to, err := uuid.Parse(p.To)
		if err != nil {
			return fmt.Errorf("set_commander_damage to: %w", err)
		}
		return g.SetCommanderDamage(from, to, p.Amount)

	case TypeConcede:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		return g.Concede(a.Player)

	case TypeKeepHand:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		return g.KeepHand(a.Player)

	case TypeRollOpening:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		return g.RollOpening(a.Player)

	case TypeHostRollRemaining:
		// WHO may press it (the host or the admin) is decided at the
		// WebSocket edge, which knows the host (ADR 0121 §3). The
		// caller is recorded as the presser: uuid.Nil is the admin.
		return g.HostRollRemaining(a.Caller)

	case TypeChooseStartingPlayer:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Seat *int `json:"seat"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		if p.Seat == nil {
			return fmt.Errorf("%w: %s seat", ErrMissingParams, a.Type)
		}
		return g.ChooseStartingPlayer(a.Player, *p.Seat)

	case TypeSetTriggerOrderPreference:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			AlwaysAsk *bool `json:"always_ask"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		if p.AlwaysAsk == nil {
			return fmt.Errorf("%w: %s always_ask", ErrMissingParams, a.Type)
		}
		return g.SetTriggerOrderPreference(a.Player, *p.AlwaysAsk)

	case TypeSetAutoAnswers:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Rules *[]struct {
				Key    string `json:"key"`
				Answer string `json:"answer"`
			} `json:"rules"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		if p.Rules == nil {
			return fmt.Errorf("%w: %s rules", ErrMissingParams, a.Type)
		}
		if len(*p.Rules) > game.MaxAutoAnswerRules {
			return game.ErrTooManyAutoAnswers
		}
		rules := make(map[string]game.AutoAnswer, len(*p.Rules))
		for _, r := range *p.Rules {
			if _, dup := rules[r.Key]; dup {
				return fmt.Errorf("%w: %s names %q twice", game.ErrInvalidParam, a.Type, r.Key)
			}
			rules[r.Key] = game.AutoAnswer(r.Answer)
		}
		return g.SetAutoAnswers(a.Player, rules)

	case TypeRollTableDie:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Die string `json:"die"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		if p.Die == "" {
			return fmt.Errorf("%w: %s die", ErrMissingParams, a.Type)
		}
		_, err := g.RollTableDie(a.Player, p.Die)
		return err

	case TypeDeclareAttacker:
		var p struct {
			Attacker string `json:"attacker"`
			Target   string `json:"target"`
			// Exert is the choice to exert the attacker as it attacks
			// (CR 701.43d, ADR 0130 §5). Absent means no.
			Exert bool `json:"exert,omitempty"`
			attackTaxParams
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		attackerID, err := uuid.Parse(p.Attacker)
		if err != nil {
			return fmt.Errorf("declare_attacker attacker: %w", err)
		}
		targetID, err := uuid.Parse(p.Target)
		if err != nil {
			return fmt.Errorf("declare_attacker target: %w", err)
		}
		if err := requireCardController(g, a.Caller, attackerID); err != nil {
			return err
		}
		declParams, err := p.attackTaxParams.decode("declare_attacker")
		if err != nil {
			return err
		}
		return g.DeclareAttackerDeclWith(game.AttackDeclaration{Attacker: attackerID, Target: targetID, Exert: p.Exert}, declParams)

	case TypeDeclareAttackers:
		var p struct {
			Attackers []struct {
				Attacker string `json:"attacker"`
				Target   string `json:"target"`
				// Exert: ADR 0130 §5, per attacker.
				Exert bool `json:"exert,omitempty"`
			} `json:"attackers"`
			attackTaxParams
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		if len(p.Attackers) == 0 {
			return ErrEmptyAttackerSet
		}
		if len(p.Attackers) > MaxBulkAttackers {
			return ErrTooManyAttackers
		}
		decls := make([]game.AttackDeclaration, 0, len(p.Attackers))
		for _, e := range p.Attackers {
			attackerID, err := uuid.Parse(e.Attacker)
			if err != nil {
				return fmt.Errorf("declare_attackers attacker: %w", err)
			}
			targetID, err := uuid.Parse(e.Target)
			if err != nil {
				return fmt.Errorf("declare_attackers target: %w", err)
			}
			// Authorization is NOT part of the silent-skip contract. An
			// unattackable creature is an eligibility question and gets
			// dropped inside the mutation; a creature the caller does not
			// control is a permission question and rejects the whole
			// batch, exactly as the single-card verb does.
			if err := requireCardController(g, a.Caller, attackerID); err != nil {
				return err
			}
			decls = append(decls, game.AttackDeclaration{Attacker: attackerID, Target: targetID, Exert: e.Exert})
		}
		declParams, err := p.attackTaxParams.decode("declare_attackers")
		if err != nil {
			return err
		}
		_, err = g.DeclareAttackersWith(decls, declParams)
		return err

	case TypeDeclareBlocker:
		var p struct {
			Blocker  string `json:"blocker"`
			Attacker string `json:"attacker"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		blockerID, err := uuid.Parse(p.Blocker)
		if err != nil {
			return fmt.Errorf("declare_blocker blocker: %w", err)
		}
		attackerID, err := uuid.Parse(p.Attacker)
		if err != nil {
			return fmt.Errorf("declare_blocker attacker: %w", err)
		}
		if err := requireCardController(g, a.Caller, blockerID); err != nil {
			return err
		}
		return g.DeclareBlocker(blockerID, attackerID)

	case TypeDeclareBlockers:
		var p struct {
			Blocks []struct {
				Blocker  string `json:"blocker"`
				Attacker string `json:"attacker"`
			} `json:"blocks"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		if len(p.Blocks) == 0 {
			return ErrEmptyBlockerSet
		}
		if len(p.Blocks) > MaxBulkBlockers {
			return ErrTooManyBlockers
		}
		decls := make([]game.BlockDeclaration, 0, len(p.Blocks))
		for _, e := range p.Blocks {
			blockerID, err := uuid.Parse(e.Blocker)
			if err != nil {
				return fmt.Errorf("declare_blockers blocker: %w", err)
			}
			attackerID, err := uuid.Parse(e.Attacker)
			if err != nil {
				return fmt.Errorf("declare_blockers attacker: %w", err)
			}
			// Authorization is per entry and rejects the whole batch,
			// exactly as the single-card verb does. A defender may
			// only ever block with their own creatures.
			if err := requireCardController(g, a.Caller, blockerID); err != nil {
				return err
			}
			decls = append(decls, game.BlockDeclaration{Blocker: blockerID, Attacker: attackerID})
		}
		return g.DeclareBlockers(decls)

	case TypeFinishBlocks:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		return g.FinishBlocks(a.Player)

	case TypeClearCombat:
		return g.ClearCombat()

	case TypeAdvanceStep:
		// Sandbox affordance: advance the step cursor by one without
		// requiring both players to have passed priority. This is the
		// "done — next step" button. NOT caller-gated — any seated
		// player (or admin) can advance the cursor; in casual play
		// the active player is usually the one driving combat
		// progression. Auto-resolve combat damage still fires when
		// the cursor lands on combat_damage (handled inside
		// game.AdvanceStep).
		_, err := g.AdvanceStep()
		return err

	case TypeSetMonarch:
		// Sandbox affordance — any seated player can flip the marker
		// (no enforcement of "combat damage transfers monarchy"). Empty
		// player clears the marker.
		if a.Player == uuid.Nil {
			return g.SetMonarch(uuid.Nil)
		}
		return g.SetMonarch(a.Player)

	case TypeSetInitiative:
		if a.Player == uuid.Nil {
			return g.SetInitiative(uuid.Nil)
		}
		return g.SetInitiative(a.Player)

	case TypeSetGoaded:
		var p struct {
			InstanceID string `json:"instance_id"`
			By         string `json:"by"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("set_goaded instance_id: %w", err)
		}
		var by uuid.UUID
		if p.By != "" {
			by, err = uuid.Parse(p.By)
			if err != nil {
				return fmt.Errorf("set_goaded by: %w", err)
			}
		}
		return g.SetGoaded(instanceID, by)

	case TypeSetPoison:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Amount int `json:"amount"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		return g.SetPoison(a.Player, p.Amount)

	case TypeSetEnergy:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Amount int `json:"amount"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		return g.SetEnergy(a.Player, p.Amount)

	case TypeSetPromise:
		var p struct {
			From  string `json:"from"`
			To    string `json:"to"`
			Count int    `json:"count"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		from, err := uuid.Parse(p.From)
		if err != nil {
			return fmt.Errorf("set_promise from: %w", err)
		}
		to, err := uuid.Parse(p.To)
		if err != nil {
			return fmt.Errorf("set_promise to: %w", err)
		}
		// Sandbox: any seated player may flip a promise marker. The
		// authoritative thing here is the visual reminder; if it
		// became contentious the playgroup would self-police.
		return g.SetPromise(from, to, p.Count)

	case TypeStartVote:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Topic   string   `json:"topic"`
			Options []string `json:"options"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		_, err := g.StartVote(a.Player, p.Topic, p.Options)
		return err

	case TypeCastVote:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Option int `json:"option"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		return g.CastVote(a.Player, p.Option)

	case TypeEndVote:
		_, err := g.EndVote()
		return err

	case TypeSetUndoLimit:
		var p struct {
			Limit int `json:"limit"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		// Deprecated alias for set_table_settings (ADR 0075 §2.3).
		// Host or admin only since S35 — the gate is at the WebSocket
		// edge (ws.Client.handleAction), which is the only layer that
		// knows whether the connection is the admin's. A negative
		// limit clamps to 0; this legacy action cannot select
		// UndoUnlimited.
		return g.SetUndoLimit(a.Caller, p.Limit)

	case TypeSetTableSettings:
		// The patch IS the params object: {"undo_limit": 3,
		// "allow_spawn": true}. Absent fields stay nil and are left
		// alone; UpdateSettings validates the whole patch before it
		// applies any of it, so one bad field changes nothing.
		var p game.SettingsPatch
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		// a.Caller is uuid.Nil for the admin, which is exactly what
		// UpdateSettings records on the event. WHO may send this is
		// decided before Dispatch (ADR 0075 §2.1).
		return g.UpdateSettings(a.Caller, p)

	case TypeCounterSpell:
		if err := requirePriorityHolder(g, a.Caller); err != nil {
			return err
		}
		var p struct {
			InstanceID string       `json:"instance_id"`
			ToZone     *zoneRefWire `json:"to_zone,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("counter_spell instance_id: %w", err)
		}
		var dst *game.ZoneRef
		if p.ToZone != nil {
			ref, err := p.ToZone.toRef()
			if err != nil {
				return fmt.Errorf("counter_spell to_zone: %w", err)
			}
			dst = &ref
		}
		return g.CounterSpell(instanceID, dst)

	case TypeCounterAbility:
		if err := requirePriorityHolder(g, a.Caller); err != nil {
			return err
		}
		var p struct {
			InstanceID string `json:"instance_id"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("counter_ability instance_id: %w", err)
		}
		return g.CounterAbility(instanceID)

	case TypeActivateAbility:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		if err := requirePriorityHolder(g, a.Caller); err != nil {
			return err
		}
		var p struct {
			SourceCardID string           `json:"source_card_id"`
			Label        string           `json:"label,omitempty"`
			Targets      []castTargetWire `json:"targets,omitempty"`
			Modes        []int            `json:"modes,omitempty"`
			XValue       int              `json:"x_value,omitempty"`
			Distribution map[string]int   `json:"distribution,omitempty"`
			// S21 sub-PR 2 — catalog activated abilities. AbilityIndex
			// selects the entry in Spec.Activated; sacrifice_ids names
			// the permanents paid to a "Sacrifice a creature" cost.
			AbilityIndex *int `json:"ability_index,omitempty"`
			// ADR 0093 Decision 5 — the row's stable ref, from the
			// view's ActivatedAbilityView.Ref or the legal move's
			// params. Optional: a stale one is refused before anything
			// is paid (ErrStaleAbilityRef), an absent one is accepted.
			Ref          string   `json:"ref,omitempty"`
			SacrificeIDs []string `json:"sacrifice_ids,omitempty"`
			// #758 — the permanents paying a TapOthers component on this
			// CR 602 activation (station, The Shire).
			TapIDs []string `json:"tap_ids,omitempty"`
			// S27 — crew_ids names the creatures tapped to pay a
			// Vehicle's crew cost (CR 702.122a). Any number of them;
			// what the server checks is the total power.
			CrewIDs []string `json:"crew_ids,omitempty"`
			// #625 — counter_source_ids names the permanent a
			// "remove N counters" cost removes from (omitted for the
			// "from this" form); counter_kind names the kind for "a
			// counter" of any kind.
			// counter_counts is the per-permanent split, added in
			// #789 for a removal spread across several permanents
			// ("from among artifacts you control") or one whose
			// count the activator announces ("Remove X storage
			// counters"). Omitted for a fixed one-permanent cost,
			// which is the shape every earlier client sends.
			// counter_kinds is #943's per-permanent kind, parallel to
			// counter_source_ids, for the one printed cost whose
			// parts may differ in kind (Tekuthal's "three counters
			// from among other artifacts, creatures, and
			// planeswalkers you control"). Omitted whenever the
			// payment is of one kind, which counter_kind still says.
			CounterSourceIDs []string `json:"counter_source_ids,omitempty"`
			CounterCounts    []int    `json:"counter_counts,omitempty"`
			CounterKind      string   `json:"counter_kind,omitempty"`
			CounterKinds     []string `json:"counter_kinds,omitempty"`
			// #660 — discard_ids names the cards paid to a
			// "Discard a creature card" cost on an activated
			// ability (CR 602.2b). Cycling's "Discard this card"
			// needs none: the source IS the payment.
			DiscardIDs []string `json:"discard_ids,omitempty"`
			// #1297 — exile_ids names the cards paid to an "Exile
			// two cards from your graveyard" / "Exile a card from
			// your hand" cost (Grim Lavamancer, Holistic Wisdom).
			// Its own field, as on activate_mana_ability (#1283):
			// an exiled card is not discarded.
			ExileIDs []string `json:"exile_ids,omitempty"`
			// ADR 0109 §7 (#1902) — top_ids names the cards paid to
			// a "Put a card from your hand on top of your library"
			// cost (Penance, Leashling). Its own field: the card is
			// neither discarded nor exiled.
			TopIDs []string `json:"top_ids,omitempty"`
			// #1213 — return_ids names the permanents paid to a
			// "Return a permanent you control to its owner's hand"
			// cost (Quirion Ranger, Master Transmuter, Meloku).
			// Exactly the clause's count, each once, each on the
			// battlefield under the activator's control.
			ReturnIDs []string `json:"return_ids,omitempty"`
			// #1600 — exile_permanent_ids names the permanents paid to
			// an "Exile a creature you control" cost (The Soul Stone's
			// harness, Altar of Bhaal). Its own field rather than
			// exile_ids, which names cards in a hand or a graveyard.
			ExilePermanentIDs []string `json:"exile_permanent_ids,omitempty"`
			// #2598 — reveal_ids names the cards revealed to pay a
			// "Reveal X black cards from your hand" cost (Martyr of
			// Bones). The name a cast's reveal pick rides; the count
			// is the clause's, or the announced x_value for the X form.
			RevealIDs []string `json:"reveal_ids,omitempty"`
			// #1310 — waterbend_ids names the untapped artifacts and
			// creatures tapped to pay part of a "Waterbend {N}" cost
			// (CR 701.67a), each covering {1} of its generic mana.
			// Optional: zero taps pays the whole cost with mana.
			WaterbendIDs []string `json:"waterbend_ids,omitempty"`
			// CR 107.4f / CR 602.2b (#917) — how many of the mana
			// component's Phyrexian symbols are being paid with 2
			// life each instead of mana (Birthing Pod's {1}{G/P}).
			// The SAME field name cast_spell uses, because it is the
			// same announcement one rule number over. Absent (0)
			// pays every symbol with its coloured half, which is
			// what every client that predates it sends.
			PhyrexianLife int  `json:"phyrexian_life,omitempty"`
			Strict        bool `json:"strict,omitempty"`
			AutoTap       bool `json:"auto_tap,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		srcID, err := uuid.Parse(p.SourceCardID)
		if err != nil {
			return fmt.Errorf("activate_ability source_card_id: %w", err)
		}
		// S21 sub-PR 2: when the payload names a catalog ability
		// index, route to the real activation path — cost validated
		// and paid, effect stamped on the stack item. Without an
		// index we keep the S13.1 free-form announce (a labelled
		// item players resolve by hand), which is still how
		// non-catalog cards work.
		if p.AbilityIndex != nil {
			sacIDs := make([]uuid.UUID, 0, len(p.SacrificeIDs))
			for _, raw := range p.SacrificeIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability sacrifice_ids: %w", err)
				}
				sacIDs = append(sacIDs, id)
			}
			tapIDs := make([]uuid.UUID, 0, len(p.TapIDs))
			for _, raw := range p.TapIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability tap_ids: %w", err)
				}
				tapIDs = append(tapIDs, id)
			}
			crewIDs := make([]uuid.UUID, 0, len(p.CrewIDs))
			for _, raw := range p.CrewIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability crew_ids: %w", err)
				}
				crewIDs = append(crewIDs, id)
			}
			counterIDs := make([]uuid.UUID, 0, len(p.CounterSourceIDs))
			for _, raw := range p.CounterSourceIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability counter_source_ids: %w", err)
				}
				counterIDs = append(counterIDs, id)
			}
			discardIDs := make([]uuid.UUID, 0, len(p.DiscardIDs))
			for _, raw := range p.DiscardIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability discard_ids: %w", err)
				}
				discardIDs = append(discardIDs, id)
			}
			exileIDs := make([]uuid.UUID, 0, len(p.ExileIDs))
			for _, raw := range p.ExileIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability exile_ids: %w", err)
				}
				exileIDs = append(exileIDs, id)
			}
			topIDs := make([]uuid.UUID, 0, len(p.TopIDs))
			for _, raw := range p.TopIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability top_ids: %w", err)
				}
				topIDs = append(topIDs, id)
			}
			returnIDs := make([]uuid.UUID, 0, len(p.ReturnIDs))
			for _, raw := range p.ReturnIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability return_ids: %w", err)
				}
				returnIDs = append(returnIDs, id)
			}
			exilePermanentIDs := make([]uuid.UUID, 0, len(p.ExilePermanentIDs))
			for _, raw := range p.ExilePermanentIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability exile_permanent_ids: %w", err)
				}
				exilePermanentIDs = append(exilePermanentIDs, id)
			}
			revealIDs := make([]uuid.UUID, 0, len(p.RevealIDs))
			for _, raw := range p.RevealIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability reveal_ids: %w", err)
				}
				revealIDs = append(revealIDs, id)
			}
			waterbendIDs := make([]uuid.UUID, 0, len(p.WaterbendIDs))
			for _, raw := range p.WaterbendIDs {
				id, err := uuid.Parse(raw)
				if err != nil {
					return fmt.Errorf("activate_ability waterbend_ids: %w", err)
				}
				waterbendIDs = append(waterbendIDs, id)
			}
			refs := make([]game.TargetRef, 0, len(p.Targets))
			for _, t := range p.Targets {
				ref, err := t.toRef()
				if err != nil {
					return fmt.Errorf("activate_ability targets: %w", err)
				}
				refs = append(refs, ref)
			}
			dist, err := parseDistribution(p.Distribution)
			if err != nil {
				return fmt.Errorf("activate_ability: %w", err)
			}
			return g.ActivateCatalogAbility(a.Player, srcID, *p.AbilityIndex, game.ActivateAbilityParams{
				Ref:               p.Ref,
				SacrificeIDs:      sacIDs,
				TapIDs:            tapIDs,
				CrewIDs:           crewIDs,
				CounterSourceIDs:  counterIDs,
				CounterCounts:     p.CounterCounts,
				CounterKind:       p.CounterKind,
				CounterKinds:      p.CounterKinds,
				DiscardIDs:        discardIDs,
				ExileIDs:          exileIDs,
				TopIDs:            topIDs,
				ReturnIDs:         returnIDs,
				ExilePermanentIDs: exilePermanentIDs,
				RevealIDs:         revealIDs,
				WaterbendIDs:      waterbendIDs,
				Targets:           refs,
				// #1563, CR 601.2d via 602.2b: the division, with
				// the targets it divides among.
				Distribution: dist,
				// #764, CR 602.2b: a modal activated ability announces
				// its modes with its targets, in one indivisible step.
				Modes: append([]int(nil), p.Modes...),
				// The same `x_value` the free-form branch below
				// hands to buildAbilityParams, now reaching the real
				// CR 602 path: an ability whose cost carries {X}
				// announces a value here and the engine charges
				// XSlots·X generic for it.
				XValue: p.XValue,
				// CR 107.4f: the life half of the announcement, routed
				// through the same strike-and-pay helper a cast's is.
				PhyrexianLife: p.PhyrexianLife,
				Strict:        p.Strict,
				AutoTap:       p.AutoTap,
			})
		}
		params, err := buildAbilityParams(p.Label, p.Targets, p.Modes, p.XValue, p.Distribution)
		if err != nil {
			return fmt.Errorf("activate_ability: %w", err)
		}
		return g.ActivateAbility(a.Player, srcID, params)

	case TypeActivateLoyalty:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		if err := requirePriorityHolder(g, a.Caller); err != nil {
			return err
		}
		var p struct {
			PlaneswalkerID string `json:"planeswalker_id"`
			Label          string `json:"label,omitempty"`
			Delta          int    `json:"delta"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		pwID, err := uuid.Parse(p.PlaneswalkerID)
		if err != nil {
			return fmt.Errorf("activate_loyalty planeswalker_id: %w", err)
		}
		return g.ActivateLoyalty(a.Player, pwID, p.Label, p.Delta)

	case TypeAddPlayerCounter:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Name  string `json:"name"`
			Delta int    `json:"delta"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		return g.AddPlayerCounter(a.Player, p.Name, p.Delta)

	case TypeDiscardSelection:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		ids := make([]uuid.UUID, 0, len(p.CardIDs))
		for i, raw := range p.CardIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("discard_selection card_ids[%d]: %w", i, err)
			}
			ids = append(ids, id)
		}
		return g.DiscardSelection(a.Player, ids)

	case TypeResolveChoice:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			ChoiceID string   `json:"choice_id"`
			Call     string   `json:"call"`
			CardIDs  []string `json:"card_ids"`
			// Color answers the two colour prompts: a PendingChoiceMana
			// pick (one of "W"/"U"/"B"/"R"/"G"/"C") and a #742
			// PendingChoiceColor ("choose a color", W/U/B/R/G). The
			// two share the field so the client's colour buttons
			// answer both, and are routed by the choice's KIND — the
			// presence of `color` alone no longer says which resolver
			// it belongs to.
			Color string `json:"color"`
			// Order populates an S17 PendingChoiceReplacementOrder
			// pick — the client returns the effect IDs in the
			// chosen order as decimal-string entries. Absent for
			// other kinds. Dispatcher routes by presence. Added in
			// S17 sub-PR 2.
			Order []string `json:"order"`
			// OptionalApply populates an S17 sub-PR 6
			// PendingChoiceOptionalReplacement yes/no pick — the
			// client returns {apply: true|false} for the "may"
			// replacement prompt (today: CR 903.9 commander-zone).
			// Pointer so we can disambiguate "absent" (nil) from
			// "false" (&false) in the routing check.
			OptionalApply *bool `json:"apply"`
			// TapIDs rides a PendingChoicePayUnless "pay" answer whose
			// cost is a waterbend cost (#1311, "Ward—Waterbend {4}"):
			// the untapped artifacts and creatures the chooser taps,
			// each paying {1} of the generic mana. Absent for every
			// other answer, and refused on one that has no waterbend
			// clause. The same wire name a cast's convoke taps use.
			TapIDs []string `json:"tap_ids,omitempty"`
			// PhyrexianLife rides a PendingChoicePayUnless "pay" answer
			// (ADR 0131 §2): how many symbols of the cost the chooser
			// pays 2 life each for instead of the mana — a ward {B}
			// under K'rrik, or a printed {B/P}. The wire name casts and
			// activations use. Refused on a "Don't pay" and on a
			// non-mana payment.
			PhyrexianLife int `json:"phyrexian_life,omitempty"`
			// Assignments populates an S18 sub-PR 3
			// PendingChoiceDamageAssignment pick — each entry is
			// {blocker_id, amount}. The attacker's controller
			// submits the ordered list; optional
			// trample_to_player carries overflow when attacker has
			// trample.
			Assignments     []damageAssignmentParam `json:"assignments"`
			TrampleToPlayer int                     `json:"trample_to_player"`
			// Target populates an S20 sub-PR 2 PendingChoicePickTarget
			// answer: the {kind, id} ref the trigger's controller
			// clicked on the board. Same shape as cast_spell targets.
			Target *castTargetWire `json:"target"`
			// Targets is the multi-slot form (S20 sub-PR 5) for a
			// pick_target prompt whose clause takes several targets
			// ("up to two target creatures"). Ordered as clicked.
			Targets []castTargetWire `json:"targets"`
			// Distribution rides a pick_target answer whose clause
			// divides (#1563, CR 601.2d via CR 603.3d): target id →
			// share. Absent for every other answer, and for a lone
			// target, which takes the whole amount.
			Distribution map[string]int `json:"distribution,omitempty"`
			// Bottom / TopOrder answer a PendingChoiceScry (CR
			// 701.22): the looked-at cards going under the library,
			// and the ones staying on top listed TOP-FIRST. Every
			// looked-at card must appear in exactly one list. A
			// put_in_library (ADR 0088) answers on the same two keys,
			// with Bottom top-first as well.
			Bottom   []string `json:"bottom"`
			TopOrder []string `json:"top_order"`
			// CreatureType answers an S26 PendingChoiceCreatureType
			// ("as this enters, choose a creature type", CR 614.12).
			// A single type name from the CR 205.3m vocabulary; the
			// engine validates and normalises it. Routed by presence,
			// like Color above.
			CreatureType string `json:"creature_type"`
			// CardName answers a PendingChoiceCardName ("as this
			// enters, choose a card name", CR 614.12, #1210). Free
			// text — CR 201.2 lets a player name any card name, so
			// there is no vocabulary to check it against and the
			// engine validates only its shape. Routed by presence,
			// like CreatureType and Color above.
			CardName string `json:"card_name"`
			// Graveyard answers a PendingChoiceSurveil (CR 701.25)
			// alongside TopOrder: the looked-at cards going to the
			// chooser's graveyard. Its PRESENCE is what distinguishes
			// a surveil answer from a scry one, since both carry
			// top_order — so a client must send `graveyard` (even as
			// []) on a surveil and must not send it on a scry.
			// Added in S22.
			Graveyard []string `json:"graveyard"`
			// Iterations answers a PendingChoiceLoopShortcut (#804,
			// CR 732): how many more times the loop's controller
			// wants the repeating ability to resolve before the
			// engine asks again. Zero — the field's own zero value —
			// is "stop here", which is why this branch is routed by
			// the choice's KIND and not by the field's presence.
			Iterations int `json:"iterations"`
			// OptionIndex answers a PendingChoiceOptionPick (#568):
			// which of the prompt's branches the chooser took.
			// Zero — the field's own zero value — is the FIRST
			// option and the commonest answer, which is why this
			// branch is routed by the choice's KIND and not by the
			// field's presence, exactly as Iterations above is.
			OptionIndex int `json:"option_index"`
			// Modes answers a PendingChoiceModePick (#764, CR
			// 603.3c): the chosen OPTION indexes in the order
			// chosen. Routed by the choice's KIND, for the same
			// reason Iterations is — an index list of zeroes
			// ([0, 0, 0] is Mystic Confluence drawing three cards) is
			// an ordinary answer, and so is the empty list on a
			// "choose up to one".
			Modes []int `json:"modes"`
			// Amount answers a PendingChoicePayAmount (ADR 0129 §3):
			// how much energy the chooser pays. Routed by the choice's
			// KIND, for Iterations' reason — paying nothing is zero.
			Amount int `json:"amount"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		choiceID, err := uuid.Parse(p.ChoiceID)
		if err != nil {
			return fmt.Errorf("resolve_choice choice_id: %w", err)
		}
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceCoinCall {
			return g.ResolveCoinCall(choiceID, a.Player, p.Call)
		}
		// #804, CR 732: "resolve it K more times, then stop?" Routed by
		// kind like the coin call above, because its whole payload is
		// an integer whose most meaningful value is zero — there is no
		// presence to route on.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceLoopShortcut {
			return g.ResolveLoopShortcut(choiceID, a.Player, p.Iterations)
		}
		// ADR 0129 §3, CR 118.12: "you may pay any amount of {E}".
		// Routed by kind: the whole payload is an integer whose
		// commonest value may be zero.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoicePayAmount {
			return g.ResolvePayAmount(choiceID, a.Player, p.Amount)
		}
		// #568, CR 608.2: "choose one of the following", addressed to
		// any seat. Routed by kind for the same reason the shortcut
		// above is — the whole payload is an integer whose most
		// meaningful value is zero.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceOptionPick {
			return g.ResolveOptionPick(choiceID, a.Player, p.OptionIndex)
		}
		// ADR 0102, CR 614.12a: "enters under the control of an
		// opponent of your choice". The same {option_index} payload as
		// option_pick, routed by kind for the same reason.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceEntryController {
			return g.ResolveEntryController(choiceID, a.Player, p.OptionIndex)
		}
		// #764, CR 603.3c: the mode of a modal triggered ability,
		// chosen as the ability is put on the stack. Routed by kind
		// for the same reason the two above are.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceModePick {
			return g.ResolveModePick(choiceID, a.Player, p.Modes)
		}
		// ADR 0108 §7, CR 615.7: "divide this shield among the damage".
		// The distribution payload a divided pick_target answer uses,
		// keyed by the prompt's entry IDs, routed by kind.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceDivideShield {
			dist, err := parseDistribution(p.Distribution)
			if err != nil {
				return fmt.Errorf("resolve_choice %w", err)
			}
			return g.ResolveDivideShield(choiceID, a.Player, dist)
		}
		if p.Color != "" {
			// #742: route by kind. A "choose a color" answer sent to
			// ResolveManaChoice would be refused (wrong kind), and a
			// mana pick sent to ResolveColorChoice likewise — so the
			// lookup is what makes the shared field safe.
			if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceColor {
				return g.ResolveColorChoice(choiceID, a.Player, p.Color)
			}
			return g.ResolveManaChoice(choiceID, a.Player, p.Color)
		}
		if p.CreatureType != "" {
			return g.ResolveCreatureTypeChoice(choiceID, a.Player, p.CreatureType)
		}
		// #1210, CR 614.12: "as this enters, choose a card name".
		// Routed by presence like the two above; a non-empty
		// card_name identifies the answer, and ResolveCardNameChoice
		// refuses it against any other kind.
		if p.CardName != "" {
			return g.ResolveCardNameChoice(choiceID, a.Player, p.CardName)
		}
		// The scry family — scry, surveil, "look at the top N and put
		// them back in any order" — routes on the CHOICE'S KIND, not
		// on the payload shape every other branch here keys off.
		//
		// It has to: all three answers carry top_order, and the plain
		// look-at carries nothing else at all, so there is no key
		// whose presence identifies it. Guessing from the payload
		// would send a surveil to ResolveScry and bury cards that
		// should have been binned. One read-lock acquisition is a
		// cheap price for not being able to get that wrong.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && game.IsLookAtTopKind(kind) {
			top, err := parseUUIDs(p.TopOrder, "top_order")
			if err != nil {
				return err
			}
			switch kind {
			case game.PendingChoiceSurveil:
				gy, err := parseUUIDs(p.Graveyard, "graveyard")
				if err != nil {
					return err
				}
				return g.ResolveSurveil(choiceID, a.Player, gy, top)
			case game.PendingChoiceLookAtTop:
				return g.ResolveLookAtTop(choiceID, a.Player, top)
			case game.PendingChoicePutInLibrary:
				// ADR 0088: scry's two keys, each lane top-first;
				// the resolver refuses a lane the prompt's
				// placement does not open.
				bottom, err := parseUUIDs(p.Bottom, "bottom")
				if err != nil {
					return err
				}
				return g.ResolvePutInLibrary(choiceID, a.Player, bottom, top)
			default:
				bottom, err := parseUUIDs(p.Bottom, "bottom")
				if err != nil {
					return err
				}
				return g.ResolveScry(choiceID, a.Player, bottom, top)
			}
		}
		if p.Target != nil {
			ref, err := p.Target.toRef()
			if err != nil {
				return fmt.Errorf("resolve_choice target: %w", err)
			}
			// S27: the legend rule answers with the same {kind, id}
			// ref — the permanent the controller KEEPS — because it is
			// the same question shape (pick one from a server-computed
			// set) and reusing the payload means the client's existing
			// highlight flow answers it with no second picker. It is
			// not targeting; the kind is what keeps them apart.
			if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceLegendRule {
				return g.ResolveLegendRule(choiceID, a.Player, ref.ID)
			}
			// S27: two kinds answer with a single {kind, id} ref.
			// pick_target is a real targeting choice; choose_protector
			// is a battle's controller naming an opponent as it enters
			// (CR 310.9a), which is not targeting at all — it just asks
			// the same question shape, so it reuses the payload and
			// the client's player-highlight flow rather than growing a
			// second one.
			if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceChooseProtector {
				return g.ResolveChooseProtector(choiceID, a.Player, ref.ID)
			}
			// #1196, CR 115.7: the retarget prompt reuses the same
			// {kind, id} payload for the same reason — one ref out of
			// a server-computed set — and is routed apart on the
			// KIND, since what the answer does is rewrite an item on
			// the stack rather than build one.
			if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceRetarget {
				return g.ResolveRetarget(choiceID, a.Player, []game.TargetRef{ref})
			}
			dist, err := parseDistribution(p.Distribution)
			if err != nil {
				return fmt.Errorf("resolve_choice %w", err)
			}
			return g.ResolvePickTargetsDivided(choiceID, a.Player, []game.TargetRef{ref}, dist)
		}
		if p.Targets != nil {
			refs := make([]game.TargetRef, 0, len(p.Targets))
			for _, t := range p.Targets {
				ref, err := t.toRef()
				if err != nil {
					return fmt.Errorf("resolve_choice targets: %w", err)
				}
				refs = append(refs, ref)
			}
			// The client's targeting banner always submits the PLURAL
			// form, so a legend-rule answer arrives here rather than
			// in the singular branch above. Both are routed: the
			// singular one is what gamecli and the tests send.
			if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceLegendRule {
				if len(refs) != 1 {
					return game.ErrInvalidParam
				}
				return g.ResolveLegendRule(choiceID, a.Player, refs[0].ID)
			}
			// S27: the client's targeting banner always submits the
			// PLURAL form, so choose_protector arrives here rather
			// than in the singular branch above. Both are routed:
			// the singular one is what gamecli and the tests send.
			if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceChooseProtector {
				if len(refs) != 1 {
					return game.ErrInvalidParam
				}
				return g.ResolveChooseProtector(choiceID, a.Player, refs[0].ID)
			}
			// #1196: the client's targeting banner always submits the
			// PLURAL form, so a retarget answer normally arrives
			// here. An EMPTY list is the decline ("you may choose new
			// targets" left unchanged, CR 115.7c), which is why this
			// branch is reached at all — `targets: []` is non-nil.
			if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceRetarget {
				return g.ResolveRetarget(choiceID, a.Player, refs)
			}
			dist, err := parseDistribution(p.Distribution)
			if err != nil {
				return fmt.Errorf("resolve_choice %w", err)
			}
			return g.ResolvePickTargetsDivided(choiceID, a.Player, refs, dist)
		}
		if p.OptionalApply != nil {
			// Five yes/no kinds share the {apply: bool} payload
			// shape: PendingChoiceOptionalReplacement (S17),
			// PendingChoiceTriggerPrompt (S19),
			// PendingChoicePayUnless (S19 sub-PR 6),
			// PendingChoiceEntryPayLife (shocklands) and
			// PendingChoiceMayCast (S28 cascade). Disambiguate
			// by looking up the choice's kind on the engine.
			kind, ok := g.PendingChoiceKindFor(choiceID)
			if !ok {
				return game.ErrPendingChoiceNotFound
			}
			switch kind {
			case game.PendingChoiceTriggerPrompt:
				return g.ResolveTriggerPrompt(choiceID, a.Player, *p.OptionalApply)
			case game.PendingChoicePayUnless:
				// S19 sub-PR 6: "unless that player pays {N}" — apply
				// means "I pay".
				// #1311: a waterbend payment names the permanents
				// it taps beside the apply.
				tapIDs := make([]uuid.UUID, 0, len(p.TapIDs))
				for i, raw := range p.TapIDs {
					id, err := uuid.Parse(raw)
					if err != nil {
						return fmt.Errorf("resolve_choice tap_ids[%d]: %w", i, err)
					}
					tapIDs = append(tapIDs, id)
				}
				// ADR 0108 §5: a non-mana payment ("discard a card",
				// "sacrifice two lands") names what it pays with in
				// card_ids beside the apply.
				if len(p.CardIDs) > 0 {
					if len(tapIDs) > 0 {
						return game.ErrInvalidParam
					}
					cardIDs, err := parseUUIDs(p.CardIDs, "card_ids")
					if err != nil {
						return err
					}
					return g.ResolvePayUnlessWithCards(choiceID, a.Player, *p.OptionalApply, cardIDs)
				}
				if p.PhyrexianLife != 0 {
					return g.ResolvePayUnlessWithLife(choiceID, a.Player, *p.OptionalApply, tapIDs, p.PhyrexianLife)
				}
				return g.ResolvePayUnlessWithTaps(choiceID, a.Player, *p.OptionalApply, tapIDs)
			case game.PendingChoiceMayCast:
				// S28 cascade: "you may cast it without paying its
				// mana cost" — apply means "I'll take it", and the
				// engine stamps a free-cast permission on the card.
				return g.ResolveMayCast(choiceID, a.Player, *p.OptionalApply)
			case game.PendingChoiceConfirm:
				// The chained-choice two-way prompt: "do A, or do B."
				// apply takes the accept branch. Both branches are
				// plain continuations supplied by the card, which is
				// what lets either one queue the next link of a
				// chain. See game/chained_choice.go.
				return g.ResolveConfirm(choiceID, a.Player, *p.OptionalApply)
			case game.PendingChoiceEntryPayLife:
				// Shocklands: "as this enters, you may pay 2 life" —
				// apply means "I pay", and paying is what keeps the
				// permanent from entering tapped.
				return g.ResolveEntryPayLife(choiceID, a.Player, *p.OptionalApply)
			case game.PendingChoiceEntryRiot:
				// Riot (CR 702.136a, ADR 0109 §10): apply takes the
				// +1/+1 counter ("you may"), and not applying takes
				// haste ("if you don't").
				return g.ResolveEntryRiot(choiceID, a.Player, *p.OptionalApply)
			case game.PendingChoiceCommanderReturn:
				// CR 903.9a (ADR 0115): apply puts the commander into
				// its owner's command zone.
				return g.ResolveCommanderReturn(choiceID, a.Player, *p.OptionalApply)
			default:
				return g.ResolveOptionalReplacement(choiceID, a.Player, *p.OptionalApply)
			}
		}
		if len(p.Order) > 0 {
			// Two reorder kinds share the {order: []string} payload:
			// S17 replacement_order (effect IDs) and S19 sub-PR 8
			// trigger_order (stack-item UUIDs). Route by kind.
			if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceTriggerOrder {
				ids := make([]uuid.UUID, 0, len(p.Order))
				for i, raw := range p.Order {
					id, err := uuid.Parse(raw)
					if err != nil {
						return fmt.Errorf("resolve_choice order[%d]: %w", i, err)
					}
					ids = append(ids, id)
				}
				return g.ResolveTriggerOrder(choiceID, a.Player, ids)
			}
			order := make([]game.ReplacementEffectID, 0, len(p.Order))
			for i, raw := range p.Order {
				id, err := game.ReplacementEffectIDFromString(raw)
				if err != nil {
					return fmt.Errorf("resolve_choice order[%d]: %w", i, err)
				}
				order = append(order, id)
			}
			return g.ResolveReplacementOrder(choiceID, a.Player, order)
		}
		if len(p.Assignments) > 0 || p.TrampleToPlayer > 0 {
			entries := make([]game.DamageAssignmentEntry, 0, len(p.Assignments))
			for i, raw := range p.Assignments {
				blockerID, err := uuid.Parse(raw.BlockerID)
				if err != nil {
					return fmt.Errorf("resolve_choice assignments[%d].blocker_id: %w", i, err)
				}
				entries = append(entries, game.DamageAssignmentEntry{
					BlockerID: blockerID,
					Amount:    raw.Amount,
				})
			}
			return g.ResolveDamageAssignment(choiceID, a.Player, entries, p.TrampleToPlayer)
		}
		ids := make([]uuid.UUID, 0, len(p.CardIDs))
		for i, raw := range p.CardIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("resolve_choice card_ids[%d]: %w", i, err)
			}
			ids = append(ids, id)
		}
		// Three card-pick kinds share the {card_ids: []string}
		// payload: discard_from_hand (S14), sacrifice_choice ("each
		// player sacrifices a creature") and search_library (S22).
		// Route by kind, as the {apply} and {order} payloads already
		// do, rather than minting another wire field for the same
		// shape.
		//
		// search_library is the one that accepts an EMPTY list: "you
		// may fail to find" (CR 701.23b) arrives as {choice_id} with
		// no card_ids at all, which is why this lookup happens before
		// the count-based guards below rather than after.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok {
			switch kind {
			case game.PendingChoiceSacrifice:
				if len(ids) != 1 {
					return game.ErrInvalidParam
				}
				return g.ResolveSacrificeChoice(choiceID, a.Player, ids[0])
			case game.PendingChoiceSearchLibrary:
				return g.ResolveSearchLibrary(choiceID, a.Player, ids)
			case game.PendingChoiceChooseCards:
				// The chained-choice card-set pick. Like search, its
				// floor can be zero, so an EMPTY list is a real answer
				// rather than a missing field — which is why it is
				// routed here by kind, ahead of the count guards.
				return g.ResolveChooseCards(choiceID, a.Player, ids)
			case game.PendingChoiceUntapChoice:
				// #826, CR 502.3: "choose which of these untap". Same
				// payload and same floor-can-be-zero reason as the
				// line above — a board of nothing but "you may choose
				// not to untap" permanents accepts the empty answer.
				return g.ResolveUntapChoice(choiceID, a.Player, ids)
			case game.PendingChoiceEntryRevealFromHand:
				// #1198, CR 614.1c: "as this land enters, you may
				// reveal an Island or Swamp card from your hand". The
				// third kind on this payload, and the floor is zero
				// by design — an EMPTY list is the decline that makes
				// the land enter tapped, so it is routed here ahead
				// of the count guards like the two above it.
				return g.ResolveEntryRevealFromHand(choiceID, a.Player, ids)
			case game.PendingChoiceEntryDiscardFromHand, game.PendingChoiceEntrySacrifice:
				// ADR 0098: "you may discard a land card instead"
				// (Mox Diamond, floor zero — the empty list is the
				// decline that puts it into its owner's graveyard) and
				// "sacrifice a Forest instead" (floor N). The same
				// payload and resolver as the reveal above.
				return g.ResolveEntryCardChoice(choiceID, a.Player, ids)
			case game.PendingChoiceRevealPick:
				// #1214, CR 608.2 / CR 701.20: "an opponent chooses
				// two of those cards". Same payload and the same
				// floor-can-be-zero reason — a Fact or Fiction pile
				// split is a legal 5-0.
				return g.ResolveRevealPick(choiceID, a.Player, ids)
			case game.PendingChoiceTheirPermanents:
				// #1214: "you choose from among the permanents that
				// player controls" — a pick across the table, routed
				// by kind so it cannot be answered through the
				// chooser's-own-material verb below.
				return g.ResolveTheirPermanents(choiceID, a.Player, ids)
			case game.PendingChoiceOwnPermanents:
				// #1214: "sacrifice any number of lands" — the
				// untargeted self-choice, whose floor really is zero.
				return g.ResolveOwnPermanents(choiceID, a.Player, ids)
			case game.PendingChoiceChooseSource:
				// ADR 0107 §6, CR 609.7a: "a source of your choice" —
				// exactly one, from the prompt's candidates.
				return g.ResolveChooseSource(choiceID, a.Player, ids)
			case game.PendingChoiceProliferate:
				// #2525, CR 701.34a: "choose any number of permanents
				// and/or players with counters". Floor zero, so an
				// EMPTY list is a real answer; seats ride in the same
				// list by player ID.
				return g.ResolveProliferate(choiceID, a.Player, ids)
			case game.PendingChoiceRingBearer:
				// ADR 0114 §4, CR 701.54a: "choose a creature you
				// control" as the Ring tempts you — exactly one, from
				// the prompt's candidates.
				return g.ResolveRingBearer(choiceID, a.Player, ids)
			case game.PendingChoiceCopyTarget:
				// "You may have this enter as a copy of ..." — an
				// EMPTY list is the decline, exactly as it is for
				// search's fail-to-find, because every printed copy
				// effect of this class says "you may" (CR 614.1c).
				if len(ids) == 0 {
					return g.ResolveCopyTarget(choiceID, a.Player, uuid.Nil)
				}
				if len(ids) != 1 {
					return game.ErrInvalidParam
				}
				return g.ResolveCopyTarget(choiceID, a.Player, ids[0])
			}
		}
		return g.ResolvePendingChoice(choiceID, a.Player, ids)

	case TypeActivateManaAbility:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			CardID       string `json:"card_id"`
			AbilityIndex int    `json:"ability_index"`
			// ADR 0093 Decision 5 — the row's stable ref
			// (ManaAbilityView.Ref). Optional, as on activate_ability.
			Ref string `json:"ref,omitempty"`
			// sacrifice_ids names the permanents paid to a
			// sacrifice-another cost (Ashnod's Altar). Same field
			// name and shape as activate_ability's, so the client
			// reuses one picker for both ability kinds.
			SacrificeIDs []string `json:"sacrifice_ids,omitempty"`
			// #758 — the permanents paying a TapOthers component on a
			// mana ability (Springleaf Drum). Same field as the CR 602
			// activation path because it is the same cost component.
			TapIDs []string `json:"tap_ids,omitempty"`
			// #789 — the counter component of a mana ability's cost,
			// with exactly the field names and shapes
			// activate_ability uses. One component, one payload
			// shape, whichever ability kind carries it: Vivid Creek
			// sends nothing at all (the self form with a printed
			// kind and count), Mage-Ring Network sends
			// counter_counts, and #943's counter_kinds.
			CounterSourceIDs []string `json:"counter_source_ids,omitempty"`
			CounterCounts    []int    `json:"counter_counts,omitempty"`
			CounterKind      string   `json:"counter_kind,omitempty"`
			CounterKinds     []string `json:"counter_kinds,omitempty"`
			// #1213 — discard_ids names the cards paid to a discard
			// cost on a MANA ability (Skirge Familiar's "Discard a
			// card: Add {B}"). Same field name and shape as
			// activate_ability's, so the client reuses one picker
			// for both ability kinds.
			DiscardIDs []string `json:"discard_ids,omitempty"`
			// #1283 — exile_ids names the cards paid to an
			// exile-a-card cost on a MANA ability (Cadaverous Bloom's
			// "Exile a card from your hand"). Its own field rather
			// than discard_ids, because it is its own component: an
			// exiled card is not discarded.
			ExileIDs []string `json:"exile_ids,omitempty"`
			// #1600 — exile_permanent_ids names the permanents paid to
			// an "Exile a creature you control" cost on a MANA ability
			// (Food Chain). The same field activate_ability uses.
			ExilePermanentIDs []string `json:"exile_permanent_ids,omitempty"`
			// #1443 — the colour a pipe slot adds, named BEFORE the
			// source is tapped, so no mana_pick is queued. `color` is
			// the one-slot spelling (a painland, Birds of Paradise,
			// Command Tower); `colors` names one per picking slot of
			// a multi-slot output ("{W|U}{W|U}"), in output order.
			// Never both. Each must be in the ability's published
			// `color_options`; absent is the ordinary two-step
			// activation, which the auto-tapper and the bots use.
			Color  string   `json:"color,omitempty"`
			Colors []string `json:"colors,omitempty"`
			// #2215 — auto_tap lets a mana component of the cost
			// (Crystal Quarry's {5}, a Signet's {1}) be paid by tapping
			// the activator's other sources for whatever the pool is
			// missing. See game.ManaAbilityParams.AutoTap.
			AutoTap bool `json:"auto_tap,omitempty"`
			// ADR 0131 §2 — phyrexian_life is how many of the mana
			// component's symbols are paid with 2 life each: a printed
			// {B/P}, or a {B} under K'rrik. The same wire name
			// cast_spell and activate_ability use. See
			// game.ManaAbilityParams.PhyrexianLife.
			PhyrexianLife int `json:"phyrexian_life,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		cardID, err := uuid.Parse(p.CardID)
		if err != nil {
			return fmt.Errorf("activate_mana_ability card_id: %w", err)
		}
		manaColors := p.Colors
		if p.Color != "" {
			if len(p.Colors) > 0 {
				return fmt.Errorf("activate_mana_ability: send color or colors, not both: %w", game.ErrInvalidParam)
			}
			manaColors = []string{p.Color}
		}
		sacIDs := make([]uuid.UUID, 0, len(p.SacrificeIDs))
		for _, raw := range p.SacrificeIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("activate_mana_ability sacrifice_ids: %w", err)
			}
			sacIDs = append(sacIDs, id)
		}
		manaTapIDs := make([]uuid.UUID, 0, len(p.TapIDs))
		for _, raw := range p.TapIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("activate_mana_ability tap_ids: %w", err)
			}
			manaTapIDs = append(manaTapIDs, id)
		}
		manaCounterIDs := make([]uuid.UUID, 0, len(p.CounterSourceIDs))
		for _, raw := range p.CounterSourceIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("activate_mana_ability counter_source_ids: %w", err)
			}
			manaCounterIDs = append(manaCounterIDs, id)
		}
		manaDiscardIDs := make([]uuid.UUID, 0, len(p.DiscardIDs))
		for _, raw := range p.DiscardIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("activate_mana_ability discard_ids: %w", err)
			}
			manaDiscardIDs = append(manaDiscardIDs, id)
		}
		manaExileIDs := make([]uuid.UUID, 0, len(p.ExileIDs))
		for _, raw := range p.ExileIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("activate_mana_ability exile_ids: %w", err)
			}
			manaExileIDs = append(manaExileIDs, id)
		}
		manaExilePermanentIDs := make([]uuid.UUID, 0, len(p.ExilePermanentIDs))
		for _, raw := range p.ExilePermanentIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("activate_mana_ability exile_permanent_ids: %w", err)
			}
			manaExilePermanentIDs = append(manaExilePermanentIDs, id)
		}
		return g.ActivateManaAbility(a.Player, cardID, p.AbilityIndex, game.ManaAbilityParams{
			Ref:               p.Ref,
			SacrificeIDs:      sacIDs,
			TapIDs:            manaTapIDs,
			CounterSourceIDs:  manaCounterIDs,
			CounterCounts:     p.CounterCounts,
			CounterKind:       p.CounterKind,
			CounterKinds:      p.CounterKinds,
			DiscardIDs:        manaDiscardIDs,
			ExileIDs:          manaExileIDs,
			ExilePermanentIDs: manaExilePermanentIDs,
			Colors:            manaColors,
			AutoTap:           p.AutoTap,
			PhyrexianLife:     p.PhyrexianLife,
		})

	case TypeSetMaxHandSize:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			Value int `json:"value"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		return g.SetMaxHandSize(a.Player, p.Value)

	case TypeSpecialAction:
		// CR 116.2: a special action is taken by a player who has
		// priority, and it uses no stack. The priority gate is the
		// same one every other announce runs; there is deliberately
		// no SplitSecondActive check here, because whether split
		// second bars this action is the KIND's question and the
		// engine's timing table answers it (foretell yes, suspend
		// no — CR 702.61b, CR 702.62c).
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		if err := requirePriorityHolder(g, a.Caller); err != nil {
			return err
		}
		var p struct {
			CardID  string `json:"card_id"`
			Kind    string `json:"kind"`
			Strict  bool   `json:"strict,omitempty"`
			AutoTap bool   `json:"auto_tap,omitempty"`
			// #1391: which offer of this kind, by its printed cost,
			// when the card has two (a plot card under Fblthp).
			Cost string `json:"cost,omitempty"`
			// ADR 0103: which door an unlock unlocks — "left" or
			// "right". Required for kind "unlock", refused otherwise.
			Door string `json:"door,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		cardID, err := uuid.Parse(p.CardID)
		if err != nil {
			return fmt.Errorf("special_action card_id: %w", err)
		}
		door := game.ParseDoorName(p.Door)
		if p.Door != "" && door == game.DoorNone {
			return fmt.Errorf("special_action door: %q is not left or right", p.Door)
		}
		return g.PerformSpecialAction(a.Player, cardID, game.SpecialActionKind(p.Kind), game.SpecialActionParams{
			Strict:  p.Strict,
			AutoTap: p.AutoTap,
			Cost:    p.Cost,
			Door:    door,
		})

	case TypeSacrificePermanent:
		var p struct {
			InstanceID string `json:"instance_id"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("sacrifice_permanent instance_id: %w", err)
		}
		if err := requireCardController(g, overrideCaller(a), instanceID); err != nil {
			return err
		}
		return g.SacrificePermanent(a.Player, instanceID)

	case TypeMarkDamage:
		var p struct {
			InstanceID string `json:"instance_id"`
			Delta      int    `json:"delta"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("mark_damage instance_id: %w", err)
		}
		// Caller-gate: positive damage requires controller; negative
		// is allowed without (admin / opponent undoing damage during
		// resolution flow). Use requireCardController which already
		// bypasses for admins.
		if p.Delta > 0 {
			if err := requireCardController(g, overrideCaller(a), instanceID); err != nil {
				return err
			}
		}
		return g.MarkDamage(instanceID, p.Delta)

	case TypeAnnounceTrigger:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			SourceCardID string           `json:"source_card_id"`
			Label        string           `json:"label,omitempty"`
			Targets      []castTargetWire `json:"targets,omitempty"`
			Modes        []int            `json:"modes,omitempty"`
			XValue       int              `json:"x_value,omitempty"`
			Distribution map[string]int   `json:"distribution,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		srcID, err := uuid.Parse(p.SourceCardID)
		if err != nil {
			return fmt.Errorf("announce_trigger source_card_id: %w", err)
		}
		params, err := buildAbilityParams(p.Label, p.Targets, p.Modes, p.XValue, p.Distribution)
		if err != nil {
			return fmt.Errorf("announce_trigger: %w", err)
		}
		return g.AnnounceTrigger(a.Player, srcID, params)

	case TypeSetBattlefieldPosition:
		var p struct {
			InstanceID string  `json:"instance_id"`
			X          float64 `json:"x"`
			Y          float64 `json:"y"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("set_battlefield_position instance_id: %w", err)
		}
		if err := requireCardController(g, overrideCaller(a), instanceID); err != nil {
			return err
		}
		return g.SetBattlefieldPosition(instanceID, p.X, p.Y)
	}

	return fmt.Errorf("%w: %q", ErrUnknownType, a.Type)
}

// zoneRefWire is the wire representation of a game.ZoneRef.
type zoneRefWire struct {
	Kind  string `json:"kind"`
	Owner string `json:"owner,omitempty"`
}

// castTargetWire is the wire representation of a single target slot
// captured at announce time. Mirrors game.TargetRef with stringified
// UUID. Empty / missing ID is allowed for "self" / "none" kinds.
type castTargetWire struct {
	Kind string `json:"kind"`
	ID   string `json:"id,omitempty"`
	// Slot / Mode name the target CLAUSE this pick answers (#764,
	// ADR 0065 §2): the clause index within the clause list, and the
	// index into the announced `modes` whose clause list that is.
	// Both optional and both defaulting to 0, which is what every
	// single-clause non-modal cast has always meant — a client that
	// never sends them keeps working, and the server fills them in
	// by walking the clauses in order.
	Slot int `json:"slot,omitempty"`
	Mode int `json:"mode,omitempty"`
}

// damageAssignmentParam is one {blocker_id, amount} pair from a
// resolve_choice action on a PendingChoiceDamageAssignment entry.
// The attacker's controller submits one entry per blocker; the
// server's ResolveDamageAssignment validates the total and trample's
// lethal-first rule (CR 510.1c, 702.19b). Added in S18 sub-PR 3.
type damageAssignmentParam struct {
	BlockerID string `json:"blocker_id"`
	Amount    int    `json:"amount"`
}

// buildAbilityParams marshals the wire-decoded fields of an
// activate_ability / announce_trigger payload into a
// game.AbilityParams. Returns wrapped target-decode errors so the
// caller's error message can prepend the action name.
func buildAbilityParams(label string, targets []castTargetWire, modes []int, xValue int, distribution map[string]int) (game.AbilityParams, error) {
	out := game.AbilityParams{
		Label:  label,
		Modes:  append([]int(nil), modes...),
		XValue: xValue,
	}
	if len(targets) > 0 {
		out.Targets = make([]game.TargetRef, 0, len(targets))
		for i, t := range targets {
			ref, err := t.toRef()
			if err != nil {
				return game.AbilityParams{}, fmt.Errorf("targets[%d]: %w", i, err)
			}
			out.Targets = append(out.Targets, ref)
		}
	}
	dist, err := parseDistribution(distribution)
	if err != nil {
		return game.AbilityParams{}, err
	}
	out.Distribution = dist
	return out, nil
}

// parseDistribution decodes a wire `distribution` — target id → share,
// the division of a "divided as you choose" clause (#1563, CR 601.2d) —
// into the engine's map. Nil for an absent or empty one.
func parseDistribution(distribution map[string]int) (map[uuid.UUID]int, error) {
	if len(distribution) == 0 {
		return nil, nil
	}
	out := make(map[uuid.UUID]int, len(distribution))
	for k, v := range distribution {
		id, err := uuid.Parse(k)
		if err != nil {
			return nil, fmt.Errorf("distribution key %q: %w", k, err)
		}
		out[id] = v
	}
	return out, nil
}

func (t castTargetWire) toRef() (game.TargetRef, error) {
	ref := game.TargetRef{Kind: game.TargetRefKind(t.Kind), Slot: t.Slot, Mode: t.Mode}
	if ref.Slot < 0 || ref.Mode < 0 {
		return game.TargetRef{}, fmt.Errorf("slot / mode must not be negative")
	}
	switch ref.Kind {
	case game.TargetSelf, game.TargetNone:
		// ID is meaningless / allowed-empty for these kinds. Drop
		// any provided ID silently — the engine ignores it.
		return ref, nil
	case game.TargetPlayer, game.TargetCard:
		if t.ID == "" {
			return game.TargetRef{}, fmt.Errorf("kind %q requires id", t.Kind)
		}
		id, err := uuid.Parse(t.ID)
		if err != nil {
			return game.TargetRef{}, fmt.Errorf("id: %w", err)
		}
		ref.ID = id
		return ref, nil
	default:
		return game.TargetRef{}, fmt.Errorf("unknown target kind %q", t.Kind)
	}
}

// parseUUIDs converts a wire list of IDs, naming the field in any
// error so a malformed entry is traceable to the key it arrived on.
// A nil or empty input yields a nil slice, which the scry-family
// resolvers treat identically.
func parseUUIDs(raw []string, field string) ([]uuid.UUID, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]uuid.UUID, 0, len(raw))
	for i, s := range raw {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("resolve_choice %s[%d]: %w", field, i, err)
		}
		out = append(out, id)
	}
	return out, nil
}

func (z zoneRefWire) toRef() (game.ZoneRef, error) {
	ref := game.ZoneRef{Kind: game.ZoneKind(z.Kind)}
	if z.Owner != "" {
		id, err := uuid.Parse(z.Owner)
		if err != nil {
			return game.ZoneRef{}, fmt.Errorf("owner: %w", err)
		}
		ref.Owner = id
	}
	return ref, nil
}
