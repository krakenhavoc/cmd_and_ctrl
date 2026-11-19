package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imskir Iron-Eater — Legendary Creature — Demon {6}{B}{R}, 5/5 (Modern
// Horizons 3):
//
//	"Affinity for artifacts
//	 When Imskir enters, you draw X cards and you lose X life, where X is half
//	 the number of artifacts you control, rounded down.
//	 {3}{R}, Sacrifice an artifact: Imskir deals damage equal to the sacrificed
//	 artifact's mana value to any target."
func init() {
	Register(Spec{
		OracleID:     "2494101a-2c6b-45d8-95fe-e62bf9a25130",
		Name:         "Imskir Iron-Eater",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Imskir Iron-Eater — you draw X cards and you lose X life, where X is half the number of artifacts you control, rounded down",
				func(g *game.Game, item *game.StackItem) error {
					x := len(permanentsControlledByMatching(g, item.Controller, Artifact())) / 2
					if x <= 0 {
						return nil
					}
					if err := (DrawCards{Player: item.Controller, N: x}).Apply(NewContext(g, item)); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -x)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, Sacrifice an artifact: Imskir deals damage equal to the sacrificed artifact's mana value to any target.",
			Cost:    Plus(ManaCost("{3}{R}"), SacrificeN(1, "an artifact", Artifact())),
			Targets: TargetAny(),
			Effect:  sourceDealsComputedDamageToEachLegalTarget(sacrificedManaValue),
		}},
	})
}
