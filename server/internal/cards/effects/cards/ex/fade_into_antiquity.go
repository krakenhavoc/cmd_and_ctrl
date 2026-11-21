package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fade into Antiquity — Sorcery {2}{G} (Edge of Eternities Commander reprint,
// S82 reprint pass):
//
//	"Exile target artifact or enchantment."
//
// A target that is no longer legal as the spell resolves leaves it with
// nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.ExileTargetArtifactOrEnchantmentSpell("3bd3156f-dbc7-48f7-9e50-1f2b8b227f97", "Fade into Antiquity"))
}
