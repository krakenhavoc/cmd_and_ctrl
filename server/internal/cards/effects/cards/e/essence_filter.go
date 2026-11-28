package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Filter — Sorcery {1}{G}{G} (Ice Age, S82 OLD, #5636):
//
//	"Destroy all enchantments or all nonwhite enchantments."
//
// The caster chooses one of the two (as a choose-one spell, CR 700.2) as it is cast: every enchantment,
// or every enchantment that is not white.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9f63388-d1ac-4895-84c5-e5b1db187463",
		Name:         "Essence Filter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy all enchantments.", nil, effects.S82OldB842DestroyAllEnchantments),
			effects.ModeDoing("Destroy all nonwhite enchantments.", nil, effects.S82OldB842DestroyAllNonwhiteEnchantments),
		),
	})
}
