package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spot, Felis catus — Legendary Creature — Cat {2}{W}, 1/1 (Star Trek
// Commander, #4974):
//
//	"Whenever Spot or another nontoken legendary permanent you control
//	 enters, create a Food token.
//	 {4}, Sacrifice a Food: Destroy target artifact."
//
// The entry trigger is Fíli the Pathfinder's shape: Spot himself, or
// another permanent that enters under his controller's control and is
// legendary and not a token, so a legendary token copy does not make a
// Food. The activation sacrifices any Food you control as a cost, the
// one Spot just made included, and destroys any artifact (Gyome's cost,
// Abrade's target).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25fe4df1-abf7-4f17-9e0d-57e414697400",
		Name:         "Spot, Felis catus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.TrekB33SpotOrAnotherNontokenLegendaryEntered, "Spot, Felis catus — create a Food token",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, Sacrifice a Food: Destroy target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyEachLegalTarget(item, effects.NewContext(g, item))
			},
		}},
	})
}
