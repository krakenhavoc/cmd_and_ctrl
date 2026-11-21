package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Davvol, Evincar of Rath — Legendary Creature — Phyrexian Noble {2}{B}{G},
// 2/2 (Mystery Booster Commander, #5948):
//
//	"Whenever another creature you control enters, you lose 1 life and add
//	 {B}."
//
// Alliance-shaped (AnotherCreatureEnteredUnderYourControl). The mana is
// produced as the trigger resolves, so it empties with the pool at the end
// of the step (CR 106.4), and the life is lost first, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4aca279-b144-4517-acef-d925f4f552e2",
		Name:         "Davvol, Evincar of Rath",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEnteredUnderYourControl,
				"Davvol, Evincar of Rath — you lose 1 life and add {B}",
				effects.Do(effects.YouLoseLife(1), effects.AddMana{Produced: "{B}"})),
		},
	})
}
