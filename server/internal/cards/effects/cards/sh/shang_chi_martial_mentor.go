package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shang-Chi, Martial Mentor — {3}{G}{G} Legendary Creature — Human Warrior
// Hero, 4/4 (Marvel Super Heroes, #4455):
//
//	"If one or more +1/+1 counters would be put on a creature you control,
//	 twice that many +1/+1 counters are put on that creature instead.
//	 Power-up — {5}{G}{G}: Put three +1/+1 counters on Shang-Chi."
//
// The first line is a Corpsejack Menace replacement (CR 614.1) on
// EventCounterPlaced; Shang-Chi's own power-up goes through it, so it puts six.
// Power-up is the engine's (CR 702.193).
//
// No simplification.
func init() {
	powerUp := effects.PowerUp("{5}{G}{G}: Put three +1/+1 counters on Shang-Chi.", effects.ManaCost("{5}{G}{G}"), effects.PlusOneCountersOnThis(3))
	powerUp.Purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:     "14d3e014-9c4f-4864-8f1e-a45ac27e4dbf",
		Name:         "Shang-Chi, Martial Mentor",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventCounterPlaced},
			AppliesTo: effects.PlusOneCountersOnACreatureYouControl,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Shang-Chi, Martial Mentor: twice that many +1/+1 counters",
		}},
		Activated: []effects.ActivatedAbility{powerUp},
	})
}
