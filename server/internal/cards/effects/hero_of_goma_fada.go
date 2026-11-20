package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hero of Goma Fada — Creature — Human Knight Ally {4}{W}, 4/3 (Battle
// for Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control
//	 enters, creatures you control gain indestructible until end of
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
		OracleID:     "f15908ea-2be0-4035-933d-4412fe032164",
		Name:         "Hero of Goma Fada",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{bfzB1RallyTeamGains("Hero of Goma Fada", "indestructible")},
	})
}
