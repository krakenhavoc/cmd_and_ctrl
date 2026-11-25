package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos Roustabout — Creature — Ogre Warrior {1}{B}{R}, 3/2 (Ravnica Allegiance):
//
//	"Whenever this creature becomes blocked, it deals 1 damage to the player or
//	 planeswalker it's attacking."
//
// One trigger per becoming blocked (CR 509.3c), however many creatures block. The
// damage goes to what the Roustabout is attacking as the trigger resolves; a
// battle takes none (it is neither a player nor a planeswalker).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c696d82e-6dee-4766-bc08-33af00f61717",
		Name:         "Rakdos Roustabout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Rakdos Roustabout — 1 damage to the player or planeswalker it's attacking",
				effects.OldB175HitWhatItAttacks(1)),
		},
	})
}
