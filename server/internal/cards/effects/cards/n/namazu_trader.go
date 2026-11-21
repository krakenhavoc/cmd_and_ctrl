package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Namazu Trader — Creature — Fish Citizen {3}{B}, 3/4 (Final Fantasy,
// #4462):
//
//	"When this creature enters, you lose 1 life and create a Treasure
//	 token.
//	 Whenever this creature attacks, you may sacrifice another creature
//	 or artifact. If you do, surveil 2."
//
// The attack trigger is Automated Warfare System's (maySacrificeOne,
// asked at resolution over what you control then, the source excluded)
// with a surveil 2 in place of the draw. Declining, or having nothing
// else to sacrifice, does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d01c351-6dbd-4d63-9843-3aa1a3c039a6",
		Name:         "Namazu Trader",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Namazu Trader — you lose 1 life and create a Treasure token", effects.Do(
				effects.GainLife{Amount: -1},
				effects.CreateToken{Template: effects.TreasureToken(), N: 1},
			)),
			effects.WheneverThisAttacks("Namazu Trader — you may sacrifice another creature or artifact to surveil 2",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Namazu Trader — sacrifice another creature or artifact to surveil 2?",
						effects.Or(effects.Creature(), effects.Artifact()), true, namazuTraderSurveil)
				}),
		},
	})
}

var namazuTraderSurveil = effects.RegisterSacrificedOne("prompt-run/namazu-trader-surveil", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.Surveil{Player: ctx.Controller(), N: 2}.Apply(ctx)
})
