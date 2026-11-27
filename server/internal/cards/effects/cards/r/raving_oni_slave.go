package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raving Oni-Slave — Creature — Ogre Warrior {1}{B}, 3/3 (Saviors of
// Kamigawa, #5601):
//
//	"When this creature enters or leaves the battlefield, you lose 3 life if
//	 you don't control a Demon."
//
// One triggered ability for two events (Subterranean Shambler's shape). The
// "if" is the clause's own, so it is an intervening if (CR 603.4): checked as
// the ability triggers and again as it resolves. On leaving, the Demon is
// looked for as the table stands once the Oni-Slave is gone, and the
// subtype is read through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "024ea14f-48b7-485d-8426-317d59f8262d",
		Name:         "Raving Oni-Slave",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you don't control a Demon", effects.OldB586YouDontControlA("Demon"),
				effects.WhenThisEntersOrLeaves("Raving Oni-Slave — you lose 3 life", effects.Do(effects.YouLoseLife(3)))),
		},
	})
}
