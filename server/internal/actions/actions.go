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

// Action is the decoded form of a wire ActionPayload, with the player
// ID already parsed as a uuid.UUID and Params still raw JSON for the
// individual handler to decode.
type Action struct {
	Type   Type
	Player uuid.UUID
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

// Dispatch applies an action to a game. Returns nil on success, an
// action-specific error on failure. Dispatch itself is stateless; all
// state lives on the game.
func Dispatch(g *game.Game, a Action) error {
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
		return g.TapCard(instanceID, a.Type == TypeTap)

	case TypeUntapAll:
		if a.Player == uuid.Nil {
			return ErrInvalidPlayer
		}
		return g.UntapAll(a.Player)

	case TypePassPriority:
		return g.PassPriority()

	case TypePassTurn:
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
