package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ray Fillet, Man Ray — {3}{U} Legendary Creature — Fish Mutant 3/3
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 When Ray Fillet enters, create a Mutagen token. (It's an artifact with
//	 "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")
//	 {2}, Remove a +1/+1 counter from a creature you control: Draw a card."
//
// The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676). The draw's cost is Heart of Kiran's "remove a counter from a
// permanent you control" (RemoveCountersFrom): the creature is chosen as
// the cost is paid, not targeted, and the ability has no tap, so Ray
// Fillet may activate it the turn it arrives and as often as it can pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "698098b8-b8e4-4e88-b13a-6b039579d192",
		Name:            "Ray Fillet, Man Ray",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ray Fillet, Man Ray — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Remove a +1/+1 counter from a creature you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{2}"), RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", Creature())),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
