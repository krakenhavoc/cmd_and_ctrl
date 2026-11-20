package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nezumi Bladeblesser — Creature — Rat Samurai {2}{B}, 3/2 (Neon Dynasty):
//
//	"This creature has deathtouch as long as you control an artifact.
//	 This creature has menace as long as you control an enchantment."
//
// Two independent layer 6 keyword grants (CR 613.1f) on the Bladeblesser
// alone, each gated by its own condition read after the layers (Gravblade
// Heavy's shape), so each ends the moment the last such permanent leaves.
//
// No simplification.
func init() {
	deathtouch := effects.KeywordGrant(effects.WhileYouHave(effects.CloudsculptControlsAnArtifact, effects.SelfCreature), "deathtouch")
	deathtouch.Label = "Has deathtouch as long as you control an artifact."
	menace := effects.KeywordGrant(effects.WhileYouHave(effects.NeoB01ControlsAnEnchantment, effects.SelfCreature), "menace")
	menace.Label = "Has menace as long as you control an enchantment."
	effects.Register(effects.Spec{
		OracleID:     "408512c7-5016-4df4-b708-b7467bcdf0ae",
		Name:         "Nezumi Bladeblesser",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{deathtouch, menace},
	})
}
