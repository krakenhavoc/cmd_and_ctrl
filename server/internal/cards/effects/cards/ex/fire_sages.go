package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Sages — Creature — Human Cleric {1}{R}, 2/2 (Avatar: The Last
// Airbender, #4459):
//
//	"Firebending 1
//	 {1}{R}{R}: Put a +1/+1 counter on this creature."
//
// Firebending's {R} lasts until end of combat (KeepManaUntilEndOfCombat).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08d1940a-ddc6-4b6d-b6c9-9d2aa470b906",
		Name:         "Fire Sages",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Fire Sages — firebending 1", effects.Do(effects.NewFirebending(1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}{R}"),
			Effect:  effects.PlusOneCountersOnThis(1),
		}},
	})
}
