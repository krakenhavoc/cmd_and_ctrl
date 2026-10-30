package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molly Hayes, Runaway — {2}{R} Legendary Creature — Mutant Hero 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Power-up — {5}{R}: Put two +1/+1 counters on Molly Hayes. Exile the top
//	 card of your library. Until the end of your next turn, you may play that
//	 card. (Activate each power-up ability only once. Reduce the cost by her
//	 mana cost if she entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The impulse exile is the shared
// ExileTopNUntilYourNextTurn (Crimson Operative).
//
// No simplification.
func init() {
	ab := PowerUp("{5}{R}: Put two +1/+1 counters on Molly Hayes. Exile the top card of your library. Until the end of your next turn, you may play that card.",
		ManaCost("{5}{R}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(2)(g, item); err != nil {
				return err
			}
			return ExileTopNUntilYourNextTurn(NewContext(g, item), 1)
		})
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:     "e3f0a2a6-7a42-46bc-9736-d0f6a0ce0b37",
		Name:         "Molly Hayes, Runaway",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{ab},
	})
}
