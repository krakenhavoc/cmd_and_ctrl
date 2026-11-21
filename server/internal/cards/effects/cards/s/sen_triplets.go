package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sen Triplets — Legendary Artifact Creature — Human Wizard {2}{W}{U}{B},
// 3/3 (#5922, #5908):
//
//	"At the beginning of your upkeep, choose target opponent. This turn,
//	 that player can't cast spells or activate abilities and plays with
//	 their hand revealed. You may play lands and cast spells from that
//	 player's hand this turn."
//
// The opponent is chosen as the trigger goes on the stack. As it
// resolves, four grants that last until end of turn (CR 611.2a):
//
//   - a cast ban on that player (game.GrantCastBanForEffect), so they
//     cast nothing, from any zone;
//   - an activation ban on them (game.GrantCantActivateAbilitiesForEffect),
//     every ability, mana abilities included: the clause names no
//     exception;
//   - their hand is revealed to every player
//     (game.GrantHandRevealedForEffect), including cards that arrive in
//     it later this turn;
//   - a standing permission over that hand for you (a stored ScopeStanding
//     hand permission naming the opponent, game.CastPermission.ForeignHandRule).
//     You play its lands with your own land drop, in your main phase with
//     the stack empty (CR 305.1, CR 305.2), and cast its spells at their
//     own timing, paying their costs. You control those spells (CR
//     601.2a); each card is still its owner's and goes to its owner's
//     graveyard.
//
// A target that has become illegal (it left the game) makes the trigger
// do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0c5864f-83c3-4d99-8f97-0ab92608fff3",
		Name:         "Sen Triplets",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourUpkeep(senTripletsLabel, senTripletsUpkeep),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

const senTripletsLabel = "Sen Triplets — choose target opponent; this turn they can't cast spells or activate abilities, their hand is revealed, and you may play lands and cast spells from it"

func senTripletsUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	opp := effects.FirstLegalTargetPlayer(ctx)
	if opp == uuid.Nil {
		return nil
	}
	source := item.SourceCardID
	const name = "Sen Triplets"
	eot := g.UntilEndOfTurnDuration()
	g.GrantCastBanForEffect(opp, game.CastBanRule{Kind: game.CastBanOutright},
		"Sen Triplets — can't cast spells this turn", source, eot)
	g.GrantCantActivateAbilitiesForEffect(opp, "Sen Triplets — can't activate abilities this turn", source, eot)
	g.GrantHandRevealedForEffect(opp, name, source, eot)
	g.GrantCastPermissionForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneHand,
		ZoneOwner:  opp,
		Scope:      game.ScopeStanding,
		Duration:   eot,
		Source:     source,
		SourceName: name,
		Label:      "Sen Triplets — you may play lands and cast spells from that player's hand this turn",
	})
	return nil
}
