package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inquisitive Glimmer — Enchantment Creature — Fox Glimmer {W}{U}, 2/3
// (Duskmourn: House of Horror, #4466):
//
//	"Enchantment spells you cast cost {1} less to cast.
//	 Unlock costs you pay cost {1} less."
//
// The first is Starfield Mystic's discount (CostsLess, CR 601.2f). The
// second is the same reduction on a CR 116.2m special action: the unlock
// action is priced through the special-action pass (CostModifier.
// SpecialActions, Ranar the Ever-Watchful's partition), kind unlock, paid
// by the Glimmer's controller. Each reduces generic mana only, so the two
// can never combine into more than {1} off anything (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4abb227a-534e-4300-9009-065a10f10e8a",
		Name:         "Inquisitive Glimmer",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Enchantment spells you cast cost {1} less to cast.", YourSpell(), socB13EnchantmentSpell()),
			SpecialActionCostsLess(1, "Unlock costs you pay cost {1} less.",
				YourSpell(), ASpecialActionOfKind(game.SpecialActionUnlock)),
		},
	})
}
