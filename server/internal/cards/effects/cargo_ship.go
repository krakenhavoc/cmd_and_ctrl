package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cargo Ship — Artifact — Vehicle {1}{U}, 2/3 (Final Fantasy, #4462):
//
//	"Flying, vigilance
//	 {T}: Add {C}. Spend this mana only to cast an artifact spell or
//	 activate an ability of an artifact source.
//	 Crew 1 (Tap any number of creatures you control with total power 1
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// The mana ability is Isolinear Processor's: {C} restricted to casting
// an artifact spell or activating an ability whose source is an artifact
// (CR 106.6), and any other use is refused. Crew 1 is Adventurer's
// Airship's ordinary crew ability (CR 702.122a). While it is crewed the
// Ship is a creature, so its {T} ability then needs it to have been
// under its controller's control since the turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "93e336e3-2be7-4581-8ab3-9bea422d81c3",
		Name:            "Cargo Ship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Cargo Ship"),
		}},
	})
}
