package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rangers' Aetherhive — Artifact — Vehicle {1}{G}{U}, 3/5 (Aetherdrift,
// #4464):
//
//	"Vigilance
//	 Whenever you activate an exhaust ability, create a 1/1 colorless
//	 Thopter artifact creature token with flying.
//	 Crew 1 (Tap any number of creatures you control with total power 1
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// Rangers' Refueler's trigger (WheneverYouActivateAnExhaustAbility) with
// a Thopter for the draw: it watches its controller's activations of any
// exhaust ability (CR 702.177a), mana abilities included, and goes on
// the stack above the ability that made it (CR 603.3b). The Aetherhive
// prints no exhaust ability of its own. Vigilance is printed; crew is the
// engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "49d149d6-de94-48cf-974b-6bb4a37ce77c",
		Name:            "Rangers' Aetherhive",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverYouActivateAnExhaustAbility(
				"Rangers' Aetherhive — create a 1/1 Thopter with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Rangers' Aetherhive"),
		}},
	})
}
