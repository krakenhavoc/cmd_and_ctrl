package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jukai Naturalist — Enchantment Creature — Human Monk {G}{W}, 2/2 (Wilds of
// Eldraine Commander, #5368):
//
//	"Lifelink
//	 Enchantment spells you cast cost {1} less to cast."
//
// Lifelink is the printed keyword. The discount is Starfield Mystic's
// (CostsLess over YourSpell and an enchantment spell, CR 601.2f): it reduces
// generic mana only and never touches an opponent's spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c4d1a78f-0cca-4042-862c-8f322ee25d50",
		Name:            "Jukai Naturalist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Enchantment spells you cast cost {1} less to cast.", effects.YourSpell(), effects.SocB13EnchantmentSpell()),
		},
	})
}
