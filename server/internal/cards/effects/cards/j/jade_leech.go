package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jade Leech — Creature — Leech {2}{G}{G}, 5/5 (Invasion, S82 older sets batch 664, #5616):
//
//	"Green spells you cast cost {G} more to cast."
//
// Geist-Fueled Scarecrow's tax on its own controller's spells, for green
// spells, with {G} (a colored symbol, CostIncrease with a Unit) in place of
// generic mana. The Leech taxes itself too, while it is a spell it is not on the
// battlefield, so only the spells cast after it are charged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "495e08a8-e654-4a0b-a25f-1e344ae2a0fa",
		Name:         "Jade Leech",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.OldB664CostsMoreMana("{G}", "Green spells you cast cost {G} more to cast.", effects.YourSpell(), effects.ColoredSpell("G")),
		},
	})
}
