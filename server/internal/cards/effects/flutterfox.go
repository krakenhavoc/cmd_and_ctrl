package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flutterfox — Creature — Fox {1}{W}, 2/2:
//
//	"As long as you control an artifact or enchantment, this creature has
//	 flying."
//
// A layer 6 keyword grant on the Fox alone, gated by a check read after the
// layers (so a permanent made an artifact or enchantment by another effect
// counts) that ends the moment the last one leaves.
//
// No simplification.
func init() {
	grant := KeywordGrant(whileYouHave(s82B30ControlsArtifactOrEnchantment, selfCreature), "flying")
	grant.Label = "As long as you control an artifact or enchantment, this creature has flying."
	Register(Spec{
		OracleID:     "58805c22-29cc-4ab2-a95c-b1798eb7a90d",
		Name:         "Flutterfox",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
