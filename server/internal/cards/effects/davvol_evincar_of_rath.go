package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e4aca279-b144-4517-acef-d925f4f552e2",
		Name:         "Davvol, Evincar of Rath",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Davvol, Evincar of Rath — you lose 1 life and add {B}",
				Do(youLoseLife(1), AddMana{Produced: "{B}"})),
		},
	})
}
