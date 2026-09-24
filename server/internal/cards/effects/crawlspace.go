package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crawlspace — Artifact {3} (EDHREC rank 1403):
//
//	"No more than two creatures can attack you each combat."
//
// The per-defender half of #1507's attack limit. "You" is the
// Crawlspace's controller — the PLAYER: an attack on a planeswalker
// they control does not count, the reading ADR 0080 gives the same
// word on Propaganda. In a four-player game the other opponents may
// still be attacked by any number of creatures in the same combat,
// which is the whole point of the card at a Commander table.
func init() {
	Register(Spec{
		OracleID:     "2296370c-fe34-4df6-92a5-260f1634bede",
		Name:         "Crawlspace",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{NoMoreThanNCanAttackYouEachCombat(2)},
	})
}
