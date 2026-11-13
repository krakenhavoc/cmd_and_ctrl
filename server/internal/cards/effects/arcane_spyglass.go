package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcane Spyglass — Artifact {4} (Darksteel, #5605):
//
//	"{2}, {T}, Sacrifice a land: Draw a card and put a charge counter on
//	 this artifact.
//	 Remove three charge counters from this artifact: Draw a card."
//
// The land is any land its controller controls, sacrificed as a cost
// (Zuran Orb's clause). The charge counter goes on the Spyglass only if it
// is still on the battlefield as the ability resolves; the card is drawn
// either way. The second ability removes the counters as its cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2fbdc80-ad5d-4fe0-a860-3e4f452e6302",
		Name:         "Arcane Spyglass",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}, Sacrifice a land: Draw a card and put a charge counter on this artifact.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    Plus(ManaCost("{2}"), TapCost(), b08SacrificeALand()),
				Effect:  Do(DrawCards{N: 1}, CounterOnThis{Kind: game.CounterCharge, N: 1}),
			},
			{
				Label:   "Remove three charge counters from this artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    RemoveCountersFromThis(game.CounterCharge, 3),
				Effect:  Do(DrawCards{N: 1}),
			},
		},
	})
}
