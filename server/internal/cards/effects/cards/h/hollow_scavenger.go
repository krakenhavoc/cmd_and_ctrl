package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hollow Scavenger // Bakery Raid — Creature — Wolf {2}{G}, 3/2 //
// Sorcery — Adventure {G} (Wilds of Eldraine, #4471):
//
//	Hollow Scavenger — "{1}, Sacrifice a Food: This creature gets +2/+2
//	 until end of turn. Activate only once each turn."
//	Bakery Raid — "Create a Food token."
//
// The pump is Bosco's "Sacrifice a Food" cost (any Food artifact you
// control) with the once-each-turn limit counted per object (CR 602.1b).
// The Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const hollowScavengerOracleID = "4fad3816-41af-4fa4-85b4-ba37614ec848"

func init() {
	label := "{1}, Sacrifice a Food: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     hollowScavengerOracleID,
		Name:         "Hollow Scavenger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.Plus(effects.ManaCost("{1}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.ThisGetsUntilEndOfTurn(2, 2, "Hollow Scavenger — +2/+2 until end of turn")(g, item)
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     hollowScavengerOracleID + "#1",
		Name:         "Bakery Raid",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.FoodToken(), N: 1}),
	})
}
