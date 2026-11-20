package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Joraga Bard — Creature — Elf Rogue Bard Ally {3}{G}, 1/4 (Zendikar, #5583):
//
//	"Whenever this creature or another Ally you control enters, you may
//	 have Ally creatures you control gain vigilance until end of turn."
//
// The Zendikar Ally trigger (zen_b2_helpers.go) with Indelicate
// Healer's team grant: the Ally creatures its controller controls as
// the ability resolves gain the keyword until end of turn (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7eef84a-0b56-41d5-a298-dc3cf199e125",
		Name:         "Joraga Bard",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{zenB2AllyGrant("Joraga Bard", "vigilance")},
	})
}
