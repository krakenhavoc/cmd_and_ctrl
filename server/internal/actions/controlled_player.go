package actions

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// controlled_player.go — acting as a player you control (CR 723, ADR
// 0166 §2, ADR 0167 §4).
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
//     other source), and answer the prompts that raises;
//   - while the controller plays a whole turn (Emrakul, ADR 0167 §4),
//     they may send every rules verb as the controlled player, and no
//     sandbox verb at all, as either player (CR 723.5b: those are not
//     choices the rules or an object call for).

// turnRulesVerbs are the verbs a turn's controller may send as the
// player whose turn they control (ADR 0167 §4): the action types
// internal/legal enumerates, plus the dock's turn walk (end_turn) and the
// cleanup discard (CR 514.1). Blocks are not here: the active player
// never declares them, and the defenders are not controlled.
var turnRulesVerbs = map[Type]struct{}{
	TypePassPriority:        {},
	TypeEndTurn:             {},
	TypeCastSpell:           {},
	TypeActivateAbility:     {},
	TypeActivateLoyalty:     {},
	TypeActivateManaAbility: {},
	TypeSpecialAction:       {},
	TypeDeclareAttacker:     {},
	TypeDeclareAttackers:    {},
	TypeResolveChoice:       {},
	TypeDiscardSelection:    {},
}

// sandboxVerbs are the manual table verbs (ADR 0167 §4) closed to both
// players of a controlled turn for as long as it is played: a sandbox
// edit is not a choice the rules or an object call for (CR 723.5b), and
// it would let the controller move the controlled player's cards by
// hand. Every other verb (settings, the table die, concede, the
// controller's own rules verbs and blocks) is unaffected. An admin
// connection (Caller uuid.Nil) keeps them all, for repairs.
var sandboxVerbs = map[Type]struct{}{
	TypeDrawCard:               {},
	TypePlayCard:               {},
	TypeMoveCard:               {},
	TypeTap:                    {},
	TypeUntap:                  {},
	TypeUntapAll:               {},
	TypePassTurn:               {},
	TypeShuffleLibrary:         {},
	TypeChangeLife:             {},
	TypeAddCounter:             {},
	TypeAddPlayerCounter:       {},
	TypeSetCommanderDamage:     {},
	TypeSetBattlefieldPosition: {},
	TypeSetMonarch:             {},
	TypeSetInitiative:          {},
	TypeSetGoaded:              {},
	TypeSetPoison:              {},
	TypeSetEnergy:              {},
	TypeSetPromise:             {},
	TypeSetMaxHandSize:         {},
	TypeClearCombat:            {},
	TypeRemoveFromCombat:       {},
	TypeAdvanceStep:            {},
	TypeRevealCard:             {},
	TypeCounterSpell:           {},
	TypeCounterAbility:         {},
	TypeAnnounceTrigger:        {},
	TypeMarkDamage:             {},
	TypeSacrificePermanent:     {},
	TypeStartVote:              {},
	TypeCastVote:               {},
	TypeEndVote:                {},
}

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
	// ADR 0167 §4: the sandbox is closed to a turn's controller, as
	// themselves too, for as long as they play it.
	if _, sandbox := sandboxVerbs[a.Type]; sandbox && g.ControlsATurn(a.Caller) {
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
	if g.ControlUntilOfPlayer(target) == game.ControlUntilTurn {
		if _, ok := turnRulesVerbs[a.Type]; !ok {
			return a, game.ErrPlayerControlled
		}
	} else if g.LandManaOnly(target) {
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
