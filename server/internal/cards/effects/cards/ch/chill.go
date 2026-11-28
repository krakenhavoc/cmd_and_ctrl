package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chill — Enchantment {1}{U} (Tempest, S82 OLD batch 765, #5628):
//
//	"Red spells cost {2} more to cast."
//
// A cost increase over every red spell, whoever casts it (CR 601.2f). The
// colour is a word of this card's text (CR 612.2) and is read off the spell
// through its layered colours, so a multicolour spell that is red is taxed
// once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70a93d0f-9a02-4e60-a2b8-827fd2410604",
		Name:         "Chill",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(2, "Red spells cost {2} more to cast.", effects.ColoredSpell("R")),
		},
	})
}
