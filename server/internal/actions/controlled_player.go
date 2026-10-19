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
//   - what the controller may send as the controlled player depends on
//     the stage of the control (controlledStageVerbs): Word of Command's
//     forced play allows the play (a cast_spell naming the prompt,
//     checked in the cast branch), their lands' mana abilities (the
//     engine refuses any other source) and the prompts that raises;
//     while the forced spell resolves, and while a search is under way,
//     only the prompts; while the controller plays a whole turn
//     (Emrakul, ADR 0167 §4) every rules verb; and while they play one
//     combat phase (Secret of Bloodbending, #3999) every rules verb of
//     that phase;
//   - no stage allows a sandbox verb as the controlled player, and a
//     controller who plays a whole turn or a combat phase has the
//     sandbox closed as themselves too (CR 723.5b: those are not choices the rules or an
//     object call for).

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
	TypeOutsidePlay:         {},
	TypeDeclareAttacker:     {},
	TypeDeclareAttackers:    {},
	TypeResolveChoice:       {},
	TypeDiscardSelection:    {},
}

// combatRulesVerbs are the verbs a combat phase's controller may send as
// the player whose combat phase they control (#3999, Secret of
// Bloodbending): turnRulesVerbs less the two that reach past the phase.
// end_turn would walk the controlled player's turn on to its end, after
// the control is over, and the cleanup discard (CR 514.1) is never in a
// combat phase.
var combatRulesVerbs = func() map[Type]struct{} {
	out := make(map[Type]struct{}, len(turnRulesVerbs))
	for t := range turnRulesVerbs {
		if t == TypeEndTurn || t == TypeDiscardSelection {
			continue
		}
		out[t] = struct{}{}
	}
	return out
}()

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

// controlledStageVerbs is what a controller may send as the player they
// control, per stage of the control (CR 723.5b: no choice the rules or an
// object don't call for; ADR 0166 amendment of 2026-10-19, #4025). A
// stage with no row allows nothing, so a stage added later is closed
// until it is given one.
//
//   - ControlPlaying (Word of Command, the forced play): the play itself,
//     the controlled player's lands' mana abilities, and the prompts that
//     raises.
//   - ControlSpellResolving (Word of Command, the forced spell resolving):
//     only the prompts the resolution raises. Nobody holds priority during
//     a resolution, so no rules verb is a choice either.
//   - ControlSearching (Opposition Agent): only the search prompt.
//   - ControlTurn (Emrakul): every rules verb (turnRulesVerbs).
//   - ControlCombat (Secret of Bloodbending, #3999): the rules verbs of
//     a combat phase (combatRulesVerbs).
//   - ControlDeclaring (Master Warcraft, #4154): the attack or block
//     declaration being chosen for the player (declaringVerbs).
//
// No stage allows a sandboxVerbs entry.
var controlledStageVerbs = map[game.ControlStage]map[Type]struct{}{
	game.ControlPlaying: {
		TypeCastSpell:           {},
		TypeActivateManaAbility: {},
		TypeResolveChoice:       {},
	},
	game.ControlSpellResolving: {TypeResolveChoice: {}},
	game.ControlSearching:      {TypeResolveChoice: {}},
	game.ControlTurn:           turnRulesVerbs,
	game.ControlCombat:         combatRulesVerbs,
	game.ControlDeclaring:      declaringVerbs,
}

// declaringVerbs are what the declarer of a "you choose which creatures
// attack / block" effect may send as the player whose declaration they
// are making (#4154, Master Warcraft; ADR 0167's second amendment of
// 2026-10-19): the declaration verbs, and the pass that finishes an
// attack declaration (the active player holds priority while it is
// made, and the engine spends that pass on finishing it). Nothing
// else: the player is not controlled, only their declaration is
// chosen for them.
var declaringVerbs = map[Type]struct{}{
	TypeDeclareAttacker:  {},
	TypeDeclareAttackers: {},
	TypePassPriority:     {},
	TypeDeclareBlocker:   {},
	TypeDeclareBlockers:  {},
	TypeFinishBlocks:     {},
}

// controlledVerbAllowed reports whether a controller may send a verb of
// type t as a player whose control is in stage.
func controlledVerbAllowed(stage game.ControlStage, t Type) bool {
	_, ok := controlledStageVerbs[stage][t]
	return ok
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
	// ADR 0167 §4: the sandbox is closed to a turn's (or a combat
	// phase's, #3999) controller, as themselves too, for as long as
	// they play it.
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
	// ADR 0166 amendment (2026-10-19, #4025): what the controller may
	// send as the controlled player depends on the stage of the control,
	// and a stage with no row, or no live record, allows nothing.
	if !controlledVerbAllowed(g.ControlStageOfPlayer(target), a.Type) {
		return a, game.ErrPlayerControlled
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
