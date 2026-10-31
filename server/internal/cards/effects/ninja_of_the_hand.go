package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ninja of the Hand — Creature — Human Ninja Villain {2}{B}, 2/2 (Marvel
// Super Heroes, #4455):
//
//	"Deathtouch
//	 Power-up — {4}{B}: Each opponent discards a card. Put a +1/+1 counter
//	 on this creature. (Activate each power-up ability only once. Reduce
//	 the cost by its mana cost if it entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The effect is Tinybones's "each
// opponent discards a card" (every opponent chooses their own discard at
// once) followed by the counter, which does not wait on the discards.
//
// No simplification.
func init() {
	ab := PowerUp("{4}{B}: Each opponent discards a card. Put a +1/+1 counter on this creature.",
		ManaCost("{4}{B}"), func(g *game.Game, item *game.StackItem) error {
			if err := eachOpponentDiscardsOne(g, item); err != nil {
				return err
			}
			return plusOneCountersOnThis(1)(g, item)
		})
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "c38f7447-e646-438d-81bd-fe2a71c8a79c",
		Name:            "Ninja of the Hand",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []ActivatedAbility{ab},
	})
}
