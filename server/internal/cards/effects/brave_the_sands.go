package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brave the Sands — Enchantment {1}{W} (EDHREC rank 1805):
//
//	"Creatures you control have vigilance.
//	 Each creature you control can block an additional creature each
//	 combat."
//
// The two-mana team vigilance. The grant is a Layer 6 static over
// "creatures you control" (b16GrantKeywords), and vigilance is
// engine-honoured: an attacker with it stays untapped.
//
// The second line is CanBlockAdditional over the same set (#1706): each
// of your creatures may block two attackers, and one that does divides
// its combat damage between them (CR 510.1d). Two Brave the Sands add
// up — each creature blocks three.
func init() {
	Register(Spec{
		OracleID:     "4e89bd75-f59d-4f08-be51-5660fbbba3c2",
		Name:         "Brave the Sands",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "vigilance"),
			CanBlockAdditional(b16CreaturesYouControl, 1),
		},
	})
}
