package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honeymoon Hearse — Artifact — Vehicle {2}{R}, 5/5 (Innistrad: Crimson
// Vow, #5495):
//
//	"Trample
//	 Tap two untapped creatures you control: This Vehicle becomes an
//	 artifact creature until end of turn."
//
// Not crew: the cost is tapping two untapped creatures you control (the
// validator keeps them untapped and yours, and a creature with summoning
// sickness can pay, since it is not the {T} symbol, CR 302.6), so Crew's
// power sum, Crew-only effects and "crewed it" memory do not apply. The
// effect is Bespoke Battlewagon's energy animation: the Vehicle's printed
// 5/5 and an artifact creature type until end of turn (BecomeCreatureUntilEOT).
// The Hearse itself may be one of the two creatures once it is animated.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "87925405-072b-49d6-97c0-fc3a13c10d07",
		Name:            "Honeymoon Hearse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "Tap two untapped creatures you control: This Vehicle becomes an artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent("two untapped creatures you control", Creature()),
				Label:  "two untapped creatures you control",
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BecomeCreatureUntilEOT{Label: "Honeymoon Hearse — becomes an artifact creature"}.Apply(NewContext(g, item))
			},
		}},
	})
}
