package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "c4d1a78f-0cca-4042-862c-8f322ee25d50",
		Name:            "Jukai Naturalist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Enchantment spells you cast cost {1} less to cast.", YourSpell(), socB13EnchantmentSpell()),
		},
	})
}
