package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grand Ball Guest — Creature — Human Peasant {1}{R}, 2/2 (Wilds of
// Eldraine, #4471):
//
//	"Celebration — This creature gets +1/+1 and has trample as long as two
//	 or more nonland permanents entered the battlefield under your control
//	 this turn."
//
// One printed ability, two layers: the +1/+1 is Armory Mice's layer 7c
// pump (SelfPumpWhileCelebration) and the trample a layer 6 grant
// (SelfKeywordWhileCelebration), both on while its controller has
// celebration (CR 207.2c, celebration.go) and off at the next turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b0fdb28-f7b1-4c30-9f2c-14a9074da22d",
		Name:         "Grand Ball Guest",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SelfPumpWhileCelebration(1, 1),
			effects.SelfKeywordWhileCelebration("trample"),
		},
	})
}
