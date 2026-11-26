package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warden of the Eye — Creature — Djinn Wizard {2}{U}{R}{W}, 3/3 (Khans of
// Tarkir):
//
//	"When this creature enters, return target noncreature, nonland card from
//	 your graveyard to your hand."
//
// Regrowth's return over a card the trigger targets in your own graveyard as it
// goes on the stack (CR 603.3d) and re-checks as it resolves (CR 608.2b): any
// card that is neither a creature nor a land, so an instant, a sorcery or a
// noncreature permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5342ce88-4afc-431c-814a-054e5fca4b9e",
		Name:         "Warden of the Eye",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Warden of the Eye — return target noncreature, nonland card from your graveyard to your hand",
				effects.ReturnFirstLegalGraveyardTargetToHand),
				effects.TargetCardInGraveyard("target noncreature, nonland card in your graveyard", effects.YouOwn(), effects.Not(effects.Creature()), effects.Not(effects.Land()))),
		},
	})
}
