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
// Sandbox simplification, declared: the second ability is not
// modelled. A blocker holds exactly one BlockingTarget, so no static
// can let it block a second attacker. (The combat-wide COUNT limits
// that were once skipped beside it — Silent Arbiter, Crawlspace —
// shipped with #1507; this is the other side, and still open.)
// Every creature still blocks one attacker. Weaker than printed,
// never stronger, and the vigilance half is the half the card is
// played for.
func init() {
	Register(Spec{
		OracleID:     "4e89bd75-f59d-4f08-be51-5660fbbba3c2",
		Name:         "Brave the Sands",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Your creatures can't block an additional creature — each still blocks only one."},
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "vigilance"),
		},
	})
}
