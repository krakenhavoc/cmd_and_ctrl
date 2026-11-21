package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky Cycle — {4}{R} Artifact — Vehicle 4/3 (Marvel Super Heroes Commander,
// #4455):
//
//	"Flying
//	 When this Vehicle enters, it deals X damage to up to one target
//	 creature, where X is twice the number of Vehicles you control.
//	 Crew 2"
//
// An up-to-one targeted enter trigger. X is counted as the trigger resolves
// (Sky Cycle itself is a Vehicle you control, so it is always at least 2),
// and the damage comes from the Vehicle, so a Sky Cycle that has left by
// then still deals it from its last-known information (CR 608.2h). A target
// that became illegal is not damaged (CR 608.2b). Crew is Dependable
// Quinjet's CrewCost / CrewEffect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "599402d9-2ea6-4709-b010-b35636be8c8c",
		Name:            "Sky Cycle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Sky Cycle — it deals X damage to up to one target creature, X being twice the number of Vehicles you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						vehicle := ctx.TextSubtype("Vehicle")
						x := 2 * effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(vehicle) })
						return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: x}.Apply(ctx)
					}),
				effects.UpToOneTargetCreature("up to one target creature")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Sky Cycle"),
		}},
	})
}
