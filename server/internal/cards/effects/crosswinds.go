package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crosswinds — Enchantment {1}{G} (Urza's Saga):
//
//	"Creatures with flying get -2/-0."
//
// Every player's fliers, the reverse of Serra Aviary. Layer 7c, after
// layer 6 has settled which creatures have flying, so a creature granted
// flying is shrunk too and one that lost it is not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68d7fafd-15f4-46fe-8233-09c3c873d1a1",
		Name:         "Crosswinds",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(crosswindsCreaturesWithFlying, -2, 0),
		},
	})
}

// crosswindsCreaturesWithFlying is "creatures with flying", every
// player's, read off what layer 6 left on the creature.
func crosswindsCreaturesWithFlying(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsCreature() && game.HasKeyword(target, "flying")
}
