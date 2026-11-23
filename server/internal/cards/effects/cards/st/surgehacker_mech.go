package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surgehacker Mech — Artifact — Vehicle {4}, 5/5 (Kamigawa: Neon Dynasty,
// S82):
//
//	"Menace
//	 When this Vehicle enters, it deals damage equal to twice the number of
//	 Vehicles you control to target creature or planeswalker an opponent
//	 controls.
//	 Crew 4"
//
// Annie Joins Up's targeted enters trigger with the amount read once at
// resolution (CR 608.2h: the Mech's own last-known information when it has
// gone; it counts itself only while it is still a Vehicle you control). Crew 4
// is the ordinary CR 702.122 crew ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf9421d2-54e2-44d8-8b39-2e41192a6517",
		Name:            "Surgehacker Mech",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Surgehacker Mech — damage equal to twice the number of Vehicles you control to target creature or planeswalker an opponent controls",
				effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
					return 2 * effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype("Vehicle") })
				})),
				effects.TargetPermanent("target creature or planeswalker an opponent controls", effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())),
				game.Purpose{Answers: game.AnswerRemove}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 4",
			Cost:    effects.CrewCost(4),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Surgehacker Mech"),
		}},
	})
}
