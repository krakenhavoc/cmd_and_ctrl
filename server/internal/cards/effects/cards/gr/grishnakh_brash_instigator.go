package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grishnákh, Brash Instigator — Legendary Creature — Goblin Soldier {2}{R},
// 1/1 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"When Grishnákh enters, amass Orcs 2. When you do, until end of turn,
//	 gain control of target nonlegendary creature an opponent controls with
//	 power less than or equal to the amassed Army's power. Untap that
//	 creature. It gains haste until end of turn."
//
// Foray of Orcs's amass with a reflexive trigger (CR 603.12) carrying the
// amassed Army's power as its amount: the target clause is built from it
// as the trigger goes on the stack, so only a nonlegendary creature an
// opponent controls whose power is at most the Army's can be chosen. The
// steal is Act of Treason's. With no Army to choose or make there is no
// amassed Army and nothing is stolen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f139012-fa71-4f5f-8c57-6847bff95678",
		Name:         "Grishnákh, Brash Instigator",
		Purpose:      game.Purpose{Amass: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Grishnákh, Brash Instigator — amass Orcs 2; when you do, gain control of target nonlegendary creature an opponent controls with power no greater than the Army's until end of turn",
				effects.Do(effects.Amass{Subtype: "Orc", N: 2, ThenTail: effects.GapB126GrishnakhTail})),
		},
	})
}
