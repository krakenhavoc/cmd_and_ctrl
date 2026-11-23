package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ondu Champion — Creature — Minotaur Warrior Ally {2}{R}{R}, 4/3
// (Battle for Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control
//	 enters, creatures you control gain trample until end of turn."
//
// Rally is one trigger per Ally entering under your control, this
// creature included (bfzB1Rally, Appa, the Vigilant's condition). The
// creatures it affects are fixed as it resolves (CR 611.2c), so a
// creature that enters later in the turn gets nothing from this
// instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b638407b-4749-41fa-8054-f9db57b849ea",
		Name:         "Ondu Champion",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BfzB1RallyTeamGains("Ondu Champion", "trample")},
	})
}
