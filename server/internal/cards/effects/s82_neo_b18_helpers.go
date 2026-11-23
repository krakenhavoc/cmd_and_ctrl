package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b18_helpers.go — pieces shared by the Kamigawa: Neon Dynasty
// transforming Sagas of S82 batch 18 (#5493).

// s82NeoB18ArtifactsAndEnchantments is "the number of artifacts and/or
// enchantments `player` controls": a permanent that is both counts once.
func S82NeoB18ArtifactsAndEnchantments(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && (c.IsArtifact() || c.IsEnchantment()) {
			n++
		}
	}
	return n
}

// s82NeoB18SelfPT is a layer 7c "this creature gets +n/+n" static whose n
// is read off the board each time the layers run.
func S82NeoB18SelfPT(label string, n func(g *game.Game, source *game.Card) int) game.StaticAbility {
	return game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			k := n(g, source)
			c.Power += k
			c.Toughness += k
		},
	}
}
