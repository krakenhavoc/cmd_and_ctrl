package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunspear Shikari — Creature — Cat Soldier {1}{W}, 2/2 (Foundations Jumpstart):
//
//	"As long as this creature is equipped, it has first strike and lifelink."
//
// Armed Assailant's "as long as this creature is equipped" gate
// (isEquippedCreature, read on every layer recompute) on two layer 6
// keyword grants to the creature alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c49db060-2d4b-4dc8-8e54-7606e079c9fe",
		Name:         "Sunspear Shikari",
		Completeness: CompletenessFull,
		Static:       sunspearShikariStatics(),
	})
}

func sunspearShikariStatics() []game.StaticAbility {
	whileEquipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && isEquippedCreature(g, *target)
	}
	fs := KeywordGrant(whileEquipped, "first strike")
	fs.Label = "As long as this creature is equipped, it has first strike."
	ll := KeywordGrant(whileEquipped, "lifelink")
	ll.Label = "As long as this creature is equipped, it has lifelink."
	return []game.StaticAbility{fs, ll}
}
