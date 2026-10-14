package actions

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// controlled_player.go — acting as a player you control (CR 723, ADR
// 0164 §2).
//
// The engine keeps every prompt's real chooser and every action's real
// player, so a controller's action is re-addressed here, once, before
// any per-seat gate reads it:
//
//   - a controlled player makes no decisions (CR 723.5), except to
//     concede (CR 723.6);
//   - an action from the controller that names the controlled player,
//     or answers a prompt the controlled player owes, runs as that
//     player: Caller and Player both become the controlled seat, and
//     Actor keeps the controller;
//   - the controller can't concede for them (CR 723.6);
//   - while Word of Command's forced play is under way, the controller
//     may do only what the card allows the player: make the play (a
//     cast_spell naming the prompt, checked in the cast branch),
//     activate their lands' mana abilities (the engine refuses any
//     other source), and answer the prompts that raises.

// routeControlledPlayer applies the rules above. Admin and spectator
// connections (Caller uuid.Nil) are untouched.
func routeControlledPlayer(g *game.Game, a Action) (Action, error) {
	a.Actor = a.Caller
	if a.Caller == uuid.Nil {
		return a, nil
	}
	if g.ControllerOfPlayer(a.Caller) != a.Caller {
		if a.Type == TypeConcede {
			return a, nil
		}
		return a, game.ErrPlayerControlled
	}
	target := a.Player
	if a.Type == TypeResolveChoice {
		if chooser, ok := choiceChooserOf(g, a.Params); ok {
			target = chooser
		}
	}
	if target == uuid.Nil || target == a.Caller || g.ControllerOfPlayer(target) != a.Caller {
		return a, nil
	}
	if a.Type == TypeConcede {
		return a, ErrPlayerCallerMismatch
	}
	if g.LandManaOnly(target) {
		switch a.Type {
		case TypeCastSpell, TypeActivateManaAbility, TypeResolveChoice:
		default:
			return a, game.ErrPlayerControlled
		}
	}
	a.Player = target
	a.Caller = target
	return a, nil
}

// choiceChooserOf reads a resolve_choice's choice_id and returns that
// prompt's chooser.
func choiceChooserOf(g *game.Game, raw json.RawMessage) (uuid.UUID, bool) {
	var p struct {
		ChoiceID string `json:"choice_id"`
	}
	if len(raw) == 0 || json.Unmarshal(raw, &p) != nil {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(p.ChoiceID)
	if err != nil {
		return uuid.Nil, false
	}
	return g.ChoiceChooser(id)
}
