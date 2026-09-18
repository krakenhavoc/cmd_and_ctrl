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
	TypeClearCombat      Type = "clear_combat"
	TypeAdvanceStep      Type = "advance_step"
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
	TypeSetUndoLimit Type = "set_undo_limit"
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
	// S21 sub-PR 1 — sacrifice a permanent you control (CR 701.21).
	// Params carry `{instance_id}`. Distinct from a manual
	// move_card to the graveyard: it emits EventSacrifice, so
	// "whenever you sacrifice" payoffs fire. Only the permanent's
	// controller may sacrifice it.
	TypeSacrificePermanent Type = "sacrifice_permanent"
)

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

// ErrEmptyAttackerSet is returned when declare_attackers arrives with
// an empty `attackers` list. "Attack with nobody" is the default
// state of the step, not an action — a player who wants it passes
// priority instead.
var ErrEmptyAttackerSet = errors.New("actions: declare_attackers needs at least one attacker")

// ErrTooManyAttackers is returned when a declare_attackers batch
// exceeds MaxBulkAttackers entries.
var ErrTooManyAttackers = errors.New("actions: declare_attackers batch is too large")

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
type Action struct {
	Type   Type
	Player uuid.UUID
	Caller uuid.UUID
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
		return nil
	}
	holder := snap.Seats[snap.Turn.PriorityHolder]
	if holder == nil || holder.ID != caller {
		return ErrNotPriorityHolder
	}
	return nil
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
	// Cast vote on behalf of self only — the seated voter is the
	// authoritative caller. start_vote is also self-driven (the
	// initiator is `Player`) but the "any caller may start a vote"
	// posture matches set_monarch / set_initiative — not scoped.
	TypeCastVote: {},
}

// Dispatch applies an action to a game. Returns nil on success, an
// action-specific error on failure. Dispatch itself is stateless; all
// state lives on the game.
func Dispatch(g *game.Game, a Action) error {
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
			// ADR 0034 — which printed face of a multi-face card is
			// being cast or played. Absent (0) is the front face,
			// which is the right answer for every single-faced card
			// and for any client that predates the face picker.
			Face int `json:"face,omitempty"`
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
		if len(p.Distribution) > 0 {
			params.Distribution = make(map[uuid.UUID]int, len(p.Distribution))
			for k, v := range p.Distribution {
				id, err := uuid.Parse(k)
				if err != nil {
					return fmt.Errorf("cast_spell distribution key %q: %w", k, err)
				}
				params.Distribution[id] = v
			}
		}
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
		if err := requireCardController(g, a.Caller, instanceID); err != nil {
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
		if err := requireCardController(g, a.Caller, instanceID); err != nil {
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
		if err := requireCardController(g, a.Caller, instanceID); err != nil {
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

	case TypeDeclareAttacker:
		var p struct {
			Attacker string `json:"attacker"`
			Target   string `json:"target"`
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
		return g.DeclareAttacker(attackerID, targetID)

	case TypeDeclareAttackers:
		var p struct {
			Attackers []struct {
				Attacker string `json:"attacker"`
				Target   string `json:"target"`
			} `json:"attackers"`
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
			decls = append(decls, game.AttackDeclaration{Attacker: attackerID, Target: targetID})
		}
		_, err := g.DeclareAttackers(decls)
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
		// Sandbox — any seated player or admin may raise/lower the
		// limit. The table self-polices abuse.
		return g.SetUndoLimit(p.Limit)

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
			AbilityIndex *int     `json:"ability_index,omitempty"`
			SacrificeIDs []string `json:"sacrifice_ids,omitempty"`
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
			CounterSourceIDs []string `json:"counter_source_ids,omitempty"`
			CounterCounts    []int    `json:"counter_counts,omitempty"`
			CounterKind      string   `json:"counter_kind,omitempty"`
			Strict           bool     `json:"strict,omitempty"`
			AutoTap          bool     `json:"auto_tap,omitempty"`
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
			refs := make([]game.TargetRef, 0, len(p.Targets))
			for _, t := range p.Targets {
				ref, err := t.toRef()
				if err != nil {
					return fmt.Errorf("activate_ability targets: %w", err)
				}
				refs = append(refs, ref)
			}
			return g.ActivateCatalogAbility(a.Player, srcID, *p.AbilityIndex, game.ActivateAbilityParams{
				SacrificeIDs:     sacIDs,
				CrewIDs:          crewIDs,
				CounterSourceIDs: counterIDs,
				CounterCounts:    p.CounterCounts,
				CounterKind:      p.CounterKind,
				Targets:          refs,
				// #764, CR 602.2b: a modal activated ability announces
				// its modes with its targets, in one indivisible step.
				Modes: append([]int(nil), p.Modes...),
				// The same `x_value` the free-form branch below
				// hands to buildAbilityParams, now reaching the real
				// CR 602 path: an ability whose cost carries {X}
				// announces a value here and the engine charges
				// XSlots·X generic for it.
				XValue:  p.XValue,
				Strict:  p.Strict,
				AutoTap: p.AutoTap,
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
			// client returns {apply: true|false} for the CR 614.10
			// "may" prompt (today: CR 903.9 commander-zone).
			// Pointer so we can disambiguate "absent" (nil) from
			// "false" (&false) in the routing check.
			OptionalApply *bool `json:"apply"`
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
			// Bottom / TopOrder answer a PendingChoiceScry (CR
			// 701.22): the looked-at cards going under the library,
			// and the ones staying on top listed TOP-FIRST. Every
			// looked-at card must appear in exactly one list.
			Bottom   []string `json:"bottom"`
			TopOrder []string `json:"top_order"`
			// CreatureType answers an S26 PendingChoiceCreatureType
			// ("as this enters, choose a creature type", CR 614.12).
			// A single type name from the CR 205.3m vocabulary; the
			// engine validates and normalises it. Routed by presence,
			// like Color above.
			CreatureType string `json:"creature_type"`
			// Graveyard answers a PendingChoiceSurveil (CR 701.25)
			// alongside TopOrder: the looked-at cards going to the
			// chooser's graveyard. Its PRESENCE is what distinguishes
			// a surveil answer from a scry one, since both carry
			// top_order — so a client must send `graveyard` (even as
			// []) on a surveil and must not send it on a scry.
			// Added in S22.
			Graveyard []string `json:"graveyard"`
			// Iterations answers a PendingChoiceLoopShortcut (#804,
			// CR 726): how many more times the loop's controller
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
		// #804, CR 726: "resolve it K more times, then stop?" Routed by
		// kind like the coin call above, because its whole payload is
		// an integer whose most meaningful value is zero — there is no
		// presence to route on.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceLoopShortcut {
			return g.ResolveLoopShortcut(choiceID, a.Player, p.Iterations)
		}
		// #568, CR 608.2: "choose one of the following", addressed to
		// any seat. Routed by kind for the same reason the shortcut
		// above is — the whole payload is an integer whose most
		// meaningful value is zero.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceOptionPick {
			return g.ResolveOptionPick(choiceID, a.Player, p.OptionIndex)
		}
		// #764, CR 603.3c: the mode of a modal triggered ability,
		// chosen as the ability is put on the stack. Routed by kind
		// for the same reason the two above are.
		if kind, ok := g.PendingChoiceKindFor(choiceID); ok && kind == game.PendingChoiceModePick {
			return g.ResolveModePick(choiceID, a.Player, p.Modes)
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
			return g.ResolvePickTarget(choiceID, a.Player, ref)
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
			return g.ResolvePickTargets(choiceID, a.Player, refs)
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
				return g.ResolvePayUnless(choiceID, a.Player, *p.OptionalApply)
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
			// sacrifice_ids names the permanents paid to a
			// sacrifice-another cost (Ashnod's Altar). Same field
			// name and shape as activate_ability's, so the client
			// reuses one picker for both ability kinds.
			SacrificeIDs []string `json:"sacrifice_ids,omitempty"`
			// #789 — the counter component of a mana ability's cost,
			// with exactly the field names and shapes
			// activate_ability uses. One component, one payload
			// shape, whichever ability kind carries it: Vivid Creek
			// sends nothing at all (the self form with a printed
			// kind and count), Mage-Ring Network sends
			// counter_counts.
			CounterSourceIDs []string `json:"counter_source_ids,omitempty"`
			CounterCounts    []int    `json:"counter_counts,omitempty"`
			CounterKind      string   `json:"counter_kind,omitempty"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		cardID, err := uuid.Parse(p.CardID)
		if err != nil {
			return fmt.Errorf("activate_mana_ability card_id: %w", err)
		}
		sacIDs := make([]uuid.UUID, 0, len(p.SacrificeIDs))
		for _, raw := range p.SacrificeIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("activate_mana_ability sacrifice_ids: %w", err)
			}
			sacIDs = append(sacIDs, id)
		}
		manaCounterIDs := make([]uuid.UUID, 0, len(p.CounterSourceIDs))
		for _, raw := range p.CounterSourceIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("activate_mana_ability counter_source_ids: %w", err)
			}
			manaCounterIDs = append(manaCounterIDs, id)
		}
		return g.ActivateManaAbility(a.Player, cardID, p.AbilityIndex, game.ManaAbilityParams{
			SacrificeIDs:     sacIDs,
			CounterSourceIDs: manaCounterIDs,
			CounterCounts:    p.CounterCounts,
			CounterKind:      p.CounterKind,
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
		if err := requireCardController(g, a.Caller, instanceID); err != nil {
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
			if err := requireCardController(g, a.Caller, instanceID); err != nil {
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
		if err := requireCardController(g, a.Caller, instanceID); err != nil {
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
// The attacker's controller submits an ordered list; the server's
// ResolveDamageAssignment validates prefix-lethal and total. Added
// in S18 sub-PR 3.
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
	if len(distribution) > 0 {
		out.Distribution = make(map[uuid.UUID]int, len(distribution))
		for k, v := range distribution {
			id, err := uuid.Parse(k)
			if err != nil {
				return game.AbilityParams{}, fmt.Errorf("distribution key %q: %w", k, err)
			}
			out.Distribution[id] = v
		}
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
