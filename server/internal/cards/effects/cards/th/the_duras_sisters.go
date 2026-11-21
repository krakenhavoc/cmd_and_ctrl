package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Duras Sisters — Legendary Creature — Klingon Warrior {B/R}{B/R},
// 3/2 (Star Trek, #4973):
//
//	"When The Duras Sisters die, create a 2/1 red Klingon Warrior
//	 creature token with haste."
//
// A dies trigger (CR 700.4): it fires only when the Sisters go from the
// battlefield to a graveyard, so an exile or a bounce makes no token. The
// token is the set's printed Klingon Warrior (Cha'DIch Investigator's
// row).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03973bad-1b3c-4488-a23e-232b192c95cc",
		Name:         "The Duras Sisters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("The Duras Sisters — create a 2/1 red Klingon Warrior with haste",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 red Klingon Warrior with haste"), N: 1})),
		},
	})
}
