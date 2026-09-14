package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mass Hysteria — Enchantment {R} (EDHREC rank 3083):
//
//	"All creatures have haste."
//
// Concordant Crossroads without the world supertype. A layer 6
// keyword grant over every creature on the battlefield — every
// player's, as printed — so the engine's summoning-sickness checks
// (attacking, {T} abilities, mana abilities) read the granted haste
// exactly as they read a printed one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4500131b-7417-4f30-a1b0-97d51b2e6458",
		Name:         "Mass Hysteria",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b29AllCreatures, "haste"),
		},
	})
}
