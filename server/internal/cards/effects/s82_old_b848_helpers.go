package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b848_helpers.go — shared pieces of Ice Age batch 848 (#5636, S82).

// OldB848CreaturesCantAttackUnlessLandSacrificed is "<Color> creatures can't
// attack unless their controller sacrifices a land of their choice for each
// <color> creature they control that's attacking" (Flooded Woodlands: green,
// Reclamation: black): the #6190 attack cost, one land per taxed attacker, paid
// by that creature's controller as attackers are declared (CR 508.1h). The color
// is read off the attacker through the layers at declaration.
func OldB848CreaturesCantAttackUnlessLandSacrificed(colorLetter string) game.AttackTax {
	return CreaturesCantAttackUnlessControllerSacrifices(
		func(q game.AttackTaxQuery) bool { return q.Attacker.HasColor(colorLetter) }, 1, "a land", Land())
}

// OldB848UpkeepPayOrSacrifice is "At the beginning of your upkeep, sacrifice
// this enchantment unless you pay <cost>" (Glaciers, Drought): Aura Flux's
// pay-or-sacrifice, asked of the enchantment's controller.
func OldB848UpkeepPayOrSacrifice(name, cost string) game.TriggeredAbility {
	return AtYourUpkeep(name+" — sacrifice this enchantment unless you pay "+cost, func(g *game.Game, item *game.StackItem) error {
		return UpkeepPayUnless{
			Chooser:  item.Controller,
			Cost:     cost,
			Question: name + " — pay " + cost + " or sacrifice this enchantment?",
			Else:     UpkeepSacrificeSource,
		}.Apply(NewContext(g, item))
	})
}
