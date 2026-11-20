package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kor Bladewhirl — Creature — Kor Soldier Ally {1}{W}, 2/2 (Battle for
// Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control
//	 enters, creatures you control gain first strike until end of
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
		OracleID:     "7024d9db-df2e-4eba-9fa0-76ebe3df05a7",
		Name:         "Kor Bladewhirl",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{bfzB1RallyTeamGains("Kor Bladewhirl", "first strike")},
	})
}
