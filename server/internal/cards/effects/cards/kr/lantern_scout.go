package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lantern Scout — Creature — Human Scout Ally {2}{W}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control
//	 enters, creatures you control gain lifelink until end of turn."
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
		OracleID:     "803f2f95-533e-4b5a-8ede-d00fad779430",
		Name:         "Lantern Scout",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BfzB1RallyTeamGains("Lantern Scout", "lifelink")},
	})
}
