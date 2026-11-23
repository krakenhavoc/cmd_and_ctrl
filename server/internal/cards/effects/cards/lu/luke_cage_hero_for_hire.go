package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luke Cage, Hero for Hire — Legendary Creature — Human Mercenary Hero
// {2}{R}{R}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 At the beginning of combat on your turn, create a Treasure token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d776a25-c341-4058-9b5f-84e70edb25a4",
		Name:            "Luke Cage, Hero for Hire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Luke Cage — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
