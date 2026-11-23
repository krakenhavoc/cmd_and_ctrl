package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	grant := effects.KeywordGrant(effects.WhileYouHave(effects.S82B30ControlsArtifactOrEnchantment, effects.SelfCreature), "flying")
	grant.Label = "As long as you control an artifact or enchantment, this creature has flying."
	effects.Register(effects.Spec{
		OracleID:     "58805c22-29cc-4ab2-a95c-b1798eb7a90d",
		Name:         "Flutterfox",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
