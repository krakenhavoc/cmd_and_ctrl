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
	"log/slog"

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
	TypeDeclareBlocker         Type = "declare_blocker"
	TypeClearCombat            Type = "clear_combat"
	TypeAdvanceStep            Type = "advance_step"
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
		// S17 diagnostic: log the raw params bytes on every cast_spell
		// so we can see exactly what the wire delivered. Helps catch
		// client bugs where targets are stripped mid-flight.
		slog.Warn("cast_spell raw params", "bytes", string(a.Params))
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
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		instanceID, err := uuid.Parse(p.InstanceID)
		if err != nil {
			return fmt.Errorf("cast_spell instance_id: %w", err)
		}
		params := game.CastSpellParams{
			FromZone:     p.FromZone,
			Modes:        append([]int(nil), p.Modes...),
			XValue:       p.XValue,
			HoldPriority: p.HoldPriority,
			SplitSecond:  p.SplitSecond,
			Strict:       p.Strict,
			ForceCast:    p.ForceCast,
			AutoTap:      p.AutoTap,
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
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		srcID, err := uuid.Parse(p.SourceCardID)
		if err != nil {
			return fmt.Errorf("activate_ability source_card_id: %w", err)
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
			CardIDs  []string `json:"card_ids"`
			// Color populates a PendingChoiceMana pick — one of
			// "W"/"U"/"B"/"R"/"G"/"C". Absent for discard_from_hand
			// picks. The dispatcher routes by the presence of this
			// field rather than round-tripping the PendingChoice to
			// check its Kind (saves a lock acquisition).
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
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		choiceID, err := uuid.Parse(p.ChoiceID)
		if err != nil {
			return fmt.Errorf("resolve_choice choice_id: %w", err)
		}
		if p.Color != "" {
			return g.ResolveManaChoice(choiceID, a.Player, p.Color)
		}
		if p.OptionalApply != nil {
			return g.ResolveOptionalReplacement(choiceID, a.Player, *p.OptionalApply)
		}
		if len(p.Order) > 0 {
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
		ids := make([]uuid.UUID, 0, len(p.CardIDs))
		for i, raw := range p.CardIDs {
			id, err := uuid.Parse(raw)
			if err != nil {
				return fmt.Errorf("resolve_choice card_ids[%d]: %w", i, err)
			}
			ids = append(ids, id)
		}
		return g.ResolvePendingChoice(choiceID, a.Player, ids)

	case TypeActivateManaAbility:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		var p struct {
			CardID       string `json:"card_id"`
			AbilityIndex int    `json:"ability_index"`
		}
		if err := unmarshalParams(a.Params, a.Type, &p); err != nil {
			return err
		}
		cardID, err := uuid.Parse(p.CardID)
		if err != nil {
			return fmt.Errorf("activate_mana_ability card_id: %w", err)
		}
		return g.ActivateManaAbility(a.Player, cardID, p.AbilityIndex)

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
	ref := game.TargetRef{Kind: game.TargetRefKind(t.Kind)}
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
