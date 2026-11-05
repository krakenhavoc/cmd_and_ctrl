package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "25fe4df1-abf7-4f17-9e0d-57e414697400",
		Name:         "Spot, Felis catus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, trekB33SpotOrAnotherNontokenLegendaryEntered, "Spot, Felis catus — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, Sacrifice a Food: Destroy target artifact.",
			Cost:    Plus(ManaCost("{4}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return destroyEachLegalTarget(item, NewContext(g, item))
			},
		}},
	})
}
