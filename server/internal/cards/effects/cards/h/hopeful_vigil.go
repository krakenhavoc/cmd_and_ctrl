package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hopeful Vigil — Enchantment {1}{W} (Wilds of Eldraine, #4471):
//
//	"When this enchantment enters, create a 2/2 white Knight creature
//	 token with vigilance.
//	 When this enchantment is put into a graveyard from the battlefield,
//	 scry 2.
//	 {2}{W}: Sacrifice this enchantment."
//
// The Knight is History of Benalia's token. The graveyard trigger is
// Fallen Ideal's: any trip from the battlefield to a graveyard, by
// this enchantment's own ability or anything else (destroyed, or
// sacrificed as a bargain), and not a trip to exile or a hand.
//
// The activated ability sacrifices on RESOLUTION, not as a cost, so it
// can be answered; it sacrifices only the object that activated it,
// still under its controller's control (sacrificeDelayedTriggerSource,
// CR 400.7, CR 701.21a). The scry then comes from the graveyard
// trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33fff722-5cc5-4d0e-b788-166b3edb1223",
		Name:         "Hopeful Vigil",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hopeful Vigil — create a 2/2 white Knight creature token with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1})),
			effects.On(game.EventLTB, effects.ThisDied, "Hopeful Vigil — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Sacrifice this enchantment.",
			Cost:    effects.ManaCost("{2}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.SacrificeDelayedTriggerSource,
		}},
	})
}
