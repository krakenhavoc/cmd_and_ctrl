package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imskir Iron-Eater — Legendary Creature — Demon {6}{B}{R}, 5/5 (Modern
// Horizons 3):
//
//	"Affinity for artifacts
//	 When Imskir enters, you draw X cards and you lose X life, where X is half
//	 the number of artifacts you control, rounded down.
//	 {3}{R}, Sacrifice an artifact: Imskir deals damage equal to the sacrificed
//	 artifact's mana value to any target."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2494101a-2c6b-45d8-95fe-e62bf9a25130",
		Name:         "Imskir Iron-Eater",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Imskir Iron-Eater — you draw X cards and you lose X life, where X is half the number of artifacts you control, rounded down",
				func(g *game.Game, item *game.StackItem) error {
					x := len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Artifact())) / 2
					if x <= 0 {
						return nil
					}
					if err := (effects.DrawCards{Player: item.Controller, N: x}).Apply(effects.NewContext(g, item)); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -x)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}, Sacrifice an artifact: Imskir deals damage equal to the sacrificed artifact's mana value to any target.",
			Cost:    effects.Plus(effects.ManaCost("{3}{R}"), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsComputedDamageToEachLegalTarget(effects.SacrificedManaValue),
		}},
	})
}
