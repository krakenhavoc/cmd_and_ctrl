package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silversmote Ghoul — Creature — Zombie Vampire {2}{B}, 3/1 (Jumpstart):
//
//	"At the beginning of your end step, if you gained 3 or more life
//	 this turn, return this card from your graveyard to the battlefield
//	 tapped.
//	 {1}{B}, Sacrifice this creature: Draw a card."
//
// The end-step trigger watches from the GRAVEYARD only (Squee's
// InGraveyard shape), and only at its owner's end step. The "if" is an
// intervening-if (CR 603.4): it is checked when the end step begins and
// again on resolution, off the per-turn tally of life GAINED, so losing
// the life back does not turn it off. The return is Reassembling
// Skeleton's body: a card that has left the graveyard by resolution
// stays where it is. The sacrifice is an ordinary sac-outlet draw, and
// the two halves loop: sacrifice it, gain 3, and it comes back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67d55b4f-944e-4376-93ed-1a240bca1eae",
		Name:         "Silversmote Ghoul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InGraveyard(InterveningIf(CondLifeGained{AtLeast: 3},
				On(game.EventBeginEndStep, AllOf(ByYou, jmpB3YouGainedThreeOrMoreLife),
					"Silversmote Ghoul — return it from your graveyard to the battlefield tapped",
					returnThisFromGraveyardTapped))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeThis()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
