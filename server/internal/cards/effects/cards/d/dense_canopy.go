package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dense Canopy — Enchantment {1}{G} (Saviors of Kamigawa, #5601):
//
//	"Creatures with flying can block only creatures with flying."
//
// Chaosphere's block rule on its own: every creature with flying, whoever
// controls it, may be declared as a blocker only of an attacker with flying
// (CR 509.1b). Judged on the blocker's and attacker's current keywords as
// blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d182ab7d-77bd-49d1-9001-9b4016350181",
		Name:         "Dense Canopy",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnMatching(effects.HasKeyword("flying")), effects.OnMatching(effects.WithoutKeyword("flying")),
				"creatures with flying can block only creatures with flying"),
		},
	})
}
