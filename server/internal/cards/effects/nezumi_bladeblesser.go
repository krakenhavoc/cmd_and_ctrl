package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	deathtouch := KeywordGrant(whileYouHave(cloudsculptControlsAnArtifact, selfCreature), "deathtouch")
	deathtouch.Label = "Has deathtouch as long as you control an artifact."
	menace := KeywordGrant(whileYouHave(neoB01ControlsAnEnchantment, selfCreature), "menace")
	menace.Label = "Has menace as long as you control an enchantment."
	Register(Spec{
		OracleID:     "408512c7-5016-4df4-b708-b7467bcdf0ae",
		Name:         "Nezumi Bladeblesser",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{deathtouch, menace},
	})
}
