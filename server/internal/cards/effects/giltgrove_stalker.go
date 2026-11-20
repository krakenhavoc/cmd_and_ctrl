package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giltgrove Stalker — Creature — Merfolk Warrior {1}{G}, 2/1 (Rivals of
// Ixalan, #5531):
//
//	"This creature can't be blocked by creatures with power 2 or less."
//
// Enraged Ceratok's block rule: a blocker's power is read as blockers are
// declared (CR 509.1b), so a pumped creature may block it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2be3969b-a427-47be-a49c-47b98e3d6b5f",
		Name:         "Giltgrove Stalker",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
