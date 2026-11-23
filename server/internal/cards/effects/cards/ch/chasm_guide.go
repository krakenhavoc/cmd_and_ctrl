package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chasm Guide — Creature — Goblin Scout Ally {3}{R}, 3/2 (Battle for
// Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control
//	 enters, creatures you control gain haste until end of turn."
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
		OracleID:     "446d3207-bff8-4edf-9b20-62abcc56427e",
		Name:         "Chasm Guide",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.BfzB1RallyTeamGains("Chasm Guide", "haste")},
	})
}
