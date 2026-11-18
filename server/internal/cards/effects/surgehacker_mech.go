package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bf9421d2-54e2-44d8-8b39-2e41192a6517",
		Name:            "Surgehacker Mech",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Surgehacker Mech — damage equal to twice the number of Vehicles you control to target creature or planeswalker an opponent controls",
				sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
					return 2 * countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype("Vehicle") })
				})),
				TargetPermanent("target creature or planeswalker an opponent controls", Or(Creature(), Planeswalker()), OpponentControls())),
				game.Purpose{Answers: game.AnswerRemove}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 4",
			Cost:    CrewCost(4),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Surgehacker Mech"),
		}},
	})
}
