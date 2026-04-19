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
	TypeSetPoison: {},
	TypeSetEnergy: {},
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

	case TypeMoveCard:
		var p struct {
			Src        zoneRefWire `json:"src"`
			Dst        zoneRefWire `json:"dst"`
			InstanceID string      `json:"instance_id"`
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
		return g.MoveCardByID(srcRef, dstRef, instanceID)

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
