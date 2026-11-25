package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zendikar Incarnate — Creature — Elemental {2}{R}{G}, */4:
//
//	"Zendikar Incarnate's power is equal to the number of lands you control."
//
// A characteristic-defining ability (CR 604.3) in layer 7a that sets the
// power only; the 4 toughness is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecd307e6-9096-401f-bd82-ad190532d389",
		Name:         "Zendikar Incarnate",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB195LandsPower("Zendikar Incarnate's power is equal to the number of lands you control."),
		},
	})
}
