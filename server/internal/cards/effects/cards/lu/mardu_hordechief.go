package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Hordechief — Creature — Human Warrior {2}{W}, 2/3 (Khans of Tarkir):
//
//	"Raid — When this creature enters, if you attacked this turn, create a
//	 1/1 white Warrior creature token."
//
// Shipwreck Looter's intervening-"if" raid trigger (CR 603.4) with Blaring
// Recruiter's Warrior token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9ea5b9e-a23d-48cf-8601-56e79ca844f9",
		Name:         "Mardu Hordechief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB223RaidWhenThisEnters("Mardu Hordechief — raid: create a 1/1 white Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Warrior"), N: 1})),
		},
	})
}
