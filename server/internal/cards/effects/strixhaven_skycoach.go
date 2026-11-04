package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strixhaven Skycoach — Artifact — Vehicle {3}, 3/2 (Secrets of Strixhaven,
// #4456):
//
//	"Flying
//	 When this Vehicle enters, you may search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle.
//	 Crew 2 (Tap any number of creatures you control with total power 2 or
//	 more: This Vehicle becomes an artifact creature until end of turn.)"
//
// The search is Yavimaya Granger's "you may" trigger with Horizon Seeker's
// reveal-to-hand search; the crew ability is Aethersphere Harvester's.
//
// No simplification.
func init() {
	const label = "Strixhaven Skycoach — search for a basic land card"
	Register(Spec{
		OracleID:        "e286a191-39d4-4073-b844-11d2aba911b4",
		Name:            "Strixhaven Skycoach",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Strixhaven Skycoach — a basic land card",
				}.Apply(NewContext(g, item))
			}), label+"?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Strixhaven Skycoach"),
		}},
	})
}
