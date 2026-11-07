package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hopeless Nightmare — Enchantment {B} (Wilds of Eldraine, #4471):
//
//	"When this enchantment enters, each opponent discards a card and
//	 loses 2 life.
//	 When this enchantment is put into a graveyard from the battlefield,
//	 scry 2.
//	 {2}{B}: Sacrifice this enchantment."
//
// The discard is Burglar Rat's: each opponent chooses a card from their
// own hand (CR 701.9a), and one with an empty hand discards nothing but
// still loses the life. The graveyard trigger and the sacrifice ability
// are Hopeful Vigil's: any trip from the battlefield to a graveyard
// scries 2, and the ability sacrifices on resolution, not as a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e63b69c-d901-4a01-8dfa-88e044856239",
		Name:         "Hopeless Nightmare",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hopeless Nightmare — each opponent discards a card and loses 2 life",
				Do(
					Instruction(func(ctx *Context) error { return eachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
					EachOpponentLosesLife{N: 2},
				)),
			On(game.EventLTB, ThisDied, "Hopeless Nightmare — scry 2", Do(Scry{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Sacrifice this enchantment.",
			Cost:    ManaCost("{2}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  sacrificeDelayedTriggerSource,
		}},
	})
}
