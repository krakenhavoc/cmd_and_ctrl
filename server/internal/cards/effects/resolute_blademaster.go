package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Resolute Blademaster — Creature — Human Soldier Ally {3}{R}{W}, 2/2
// (Battle for Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control
//	 enters, creatures you control gain double strike until end of
//	 turn."
//
// Rally is one trigger per Ally entering under your control, this
// creature included (bfzB1Rally, Appa, the Vigilant's condition). The
// creatures it affects are fixed as it resolves (CR 611.2c), so a
// creature that enters later in the turn gets nothing from this
// instance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fec7b50a-7ed3-461d-9e91-e5c6ca9f774f",
		Name:         "Resolute Blademaster",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{bfzB1RallyTeamGains("Resolute Blademaster", "double strike")},
	})
}
