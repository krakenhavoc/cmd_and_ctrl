package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unscythe, Killer of Kings — Legendary Artifact — Equipment {U}{B}{B}{R}
// (Alara Reborn, #5586):
//
//	"Equipped creature gets +3/+3 and has first strike.
//	 Whenever a creature dealt damage by equipped creature this turn dies, you
//	 may exile that card. If you do, create a 2/2 black Zombie creature token.
//	 Equip {2}"
//
// The trigger is read as the creature dies (CR 603.10a): the equipped creature
// is the one the Equipment is attached to then, and the damage is the turn's
// record of damage that permanent dealt to the dying one (Zurgo Helmsmasher's
// reader, objects as they were, CR 400.7). The exile is optional and takes the
// card out of the graveyard; the Zombie follows only once it has reached exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d16976b8-4eda-4415-bc16-c20d958b6f59",
		Name:         "Unscythe, Killer of Kings",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("first strike"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventLTB, effects.S82OldB504UnscytheDamagedCreatureDied,
					"Unscythe, Killer of Kings — you may exile that card; if you do, create a 2/2 black Zombie creature token",
					effects.S82OldB504UnscytheExile),
				"Unscythe, Killer of Kings — exile that card?"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
