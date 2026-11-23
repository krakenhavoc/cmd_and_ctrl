package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain America's Motorcycle — Artifact — Vehicle {1}{R}, 3/1 (Marvel
// Super Heroes Commander, #4455):
//
//	"Flash
//	 When this Vehicle enters, target creature or Vehicle gets +2/+0
//	 until end of turn.
//	 Crew 1"
//
// The target is any creature or Vehicle (Hellish Sideswipe's predicate),
// itself included: the trigger can pump the Motorcycle, which is a
// Vehicle even while it is not a creature.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19e60344-57a8-41e0-8eca-87dbedf5a81a",
		Name:            "Captain America's Motorcycle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Captain America's Motorcycle — target creature or Vehicle gets +2/+0",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Captain America's Motorcycle — +2/+0"}.Apply(ctx)
						}
					}
					return nil
				}), effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle")))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Captain America's Motorcycle"),
		}},
	})
}
