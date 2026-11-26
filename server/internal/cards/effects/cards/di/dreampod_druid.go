package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreampod Druid — Creature — Human Druid {1}{G}, 2/2 (Planechase 2012, S82 OLD,
// #5569):
//
//	"At the beginning of each upkeep, if this creature is enchanted, create a
//	 1/1 green Saproling creature token."
//
// Ruthless Winnower's every-upkeep trigger behind an intervening "if" (CR 603.4):
// an Aura is attached to the Druid as each upkeep begins and again as the trigger
// resolves. The Druid's controller gets the Saproling, whoever's upkeep it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db2af2b0-a5aa-4e4a-b5bc-52a5ac3968fc",
		Name:         "Dreampod Druid",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB432InterveningIfEnchanted(
				effects.AtEachUpkeep("Dreampod Druid — create a 1/1 green Saproling creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}))),
		},
	})
}
