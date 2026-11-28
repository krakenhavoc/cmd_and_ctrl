package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Derelor — Creature — Thrull {3}{B}, 4/4 (Fallen Empires, S82 OLD b855, #5637):
//
//	"Black spells you cast cost {B} more to cast."
//
// Ruby Leech's static cost increase with a black unit on your own black spells: the {B}
// is added to the total cost (CR 601.2f), and Derelor's own spell is black but is not
// taxed by itself, since it is not on the battlefield as it is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43f2969e-46eb-4fd4-8b2d-ff81aafcc24a",
		Name:         "Derelor",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMoreColored("{B}", "Black spells you cast cost {B} more to cast.", effects.YourSpell(), effects.ColoredSpell("B")),
		},
	})
}
