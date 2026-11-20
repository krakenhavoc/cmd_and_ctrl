package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tobias, Doomed Conqueror — Legendary Creature — Human Soldier {2}{W}{U},
// 3/2 (Dominaria United Commander, S82):
//
//	"Flash
//	 When Tobias dies, for each nontoken creature you controlled that died
//	 this turn, create a 2/2 black Zombie creature token."
//
// Flash is a printed keyword. The count is taken as the trigger resolves
// (CR 608.2h) from the turn's battlefield-to-graveyard moves, each read as
// the creature last existed, under whose control it left, and not a token;
// Tobias is one of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4544369-910a-41b4-b835-eee341bb48c6",
		Name:            "Tobias, Doomed Conqueror",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Tobias, Doomed Conqueror — for each nontoken creature you controlled that died this turn, create a 2/2 Zombie",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("2/2 black Zombie"),
					Name:     "the nontoken creatures you controlled that died this turn",
					Count: func(ctx *effects.Context) int {
						return effects.NontokenCreaturesYouControlledThatDiedThisTurn(ctx.Game, ctx.Controller())
					},
				})),
		},
	})
}
