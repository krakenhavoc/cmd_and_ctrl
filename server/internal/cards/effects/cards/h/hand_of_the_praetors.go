package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hand of the Praetors — Creature — Phyrexian Zombie {3}{B}, 3/2 (Scars of
// Mirrodin, #5577):
//
//	"Infect
//	 Other creatures you control with infect get +1/+1.
//	 Whenever you cast a creature spell with infect, target player gets a
//	 poison counter."
//
// Infect is the engine's damage tail (CR 702.90). The anthem is a layer 7c
// static over the other creatures its controller controls that have infect
// after layer 6, so a creature that gains infect later gets the bonus too.
// The cast trigger reads the spell's keywords as it is cast and gives the
// target player a poison counter as Ichor Rats does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f595c7f6-783f-4825-9c1a-e7f8edd380fd",
		Name:            "Hand of the Praetors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Static: []game.StaticAbility{{
			Label:    "Other creatures you control with infect get +1/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.InstanceID != source.InstanceID && game.HasKeyword(target, "infect")
			},
			Apply: effects.PumpBy(1, 1),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.And(effects.Creature(), effects.HasKeyword("infect")),
					"Hand of the Praetors — target player gets a poison counter",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetPlayer {
								continue
							}
							if err := g.AddPlayerCounterByForEffect(item.Controller, t.ID, game.CounterPoison, 1); err != nil {
								return err
							}
						}
						return nil
					}),
				effects.TargetPlayer("target player")),
		},
	})
}
