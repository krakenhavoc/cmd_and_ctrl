package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapB103EnchantmentEnteredUnderYourControl is "whenever an enchantment you
// control enters" (constellation without "this creature or another"): the
// card is read as it entered, through the layers.
func GapB103EnchantmentEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := EnteredUnderYourControl(ev, source, g, false)
	return ok && c.IsEnchantment()
}

// gapB103NoncreatureEnchantmentValue is "the total mana value of
// noncreature enchantments you control" (Nyxborn Behemoth), counted as the
// spell is cast (CR 601.2f). A card whose cost has {X} counts X as 0 on the
// battlefield (CR 202.3e).
func GapB103NoncreatureEnchantmentValue(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	total := 0
	for _, c := range q.Game.BattlefieldCardsForEffect() {
		if c.Controller != q.Controller || !c.IsEnchantment() || c.IsCreature() {
			continue
		}
		if mv, ok := c.ParsedManaValue(); ok {
			total += mv
		}
	}
	return total
}

// gapB103CreatureYouControlDealtCombatDamageToAPlayer is "whenever a creature
// you control deals combat damage to a player", the dealer included.
func GapB103CreatureYouControlDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return CombatDamageToPlayerBy(ev, source.Controller, g)
}
