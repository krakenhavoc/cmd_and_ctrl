package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loyal Fire Sage — Creature — Human Cleric Ally {2}{R}, 3/3 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 {5}: Create a 1/1 white Ally creature token."
//
// Firebending is the set's keyword trigger (firebending), its mana marked
// to last until end of combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c052420-2d8b-4bbf-8197-b8a36e7208c9",
		Name:         "Loyal Fire Sage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Loyal Fire Sage — firebending 1", effects.Do(effects.NewFirebending(1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}: Create a 1/1 white Ally creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{5}"),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Ally"), N: 1}),
		}},
	})
}
