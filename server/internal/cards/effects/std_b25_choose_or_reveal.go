package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b25_choose_or_reveal.go — the seam behind Monstrous Emergence
// (Duskmourn, #4452): "As an additional cost to cast this spell, choose a
// creature you control or reveal a creature card from your hand", read back
// by the spell's text.
//
// The cost is the engine's behold component (game.RevealCost) over a card
// type instead of a creature type, with Chosen set so the payment records
// its pick (game.PaidCost.Chosen, game/chosen_cost.go). The caster names the
// one object on cast_spell as reveal_ids, exactly as for a Champion's
// behold, so the client's existing behold picker, the view's reveal_options
// and the bot enumerator all serve it with no protocol change.

// ChooseOrRevealCost is "As an additional cost to cast this spell, choose
// a <card type> you control or reveal a <card type> card from your hand":
// the mandatory slot's one component. `cardType` is the lower-case card
// type ("creature"); `label` is the printed clause without the verb, shown
// as the picker's heading.
func ChooseOrRevealCost(cardType, label string) *game.AdditionalCost {
	return &game.AdditionalCost{
		Reveal: &game.RevealCost{CardType: cardType, Behold: true, Chosen: true},
		Label:  label,
	}
}

// b25ChosenPower is the power of the object the spell's Chosen cost named,
// as the spell resolves (game.ChosenPowerForEffect): the creature on the
// battlefield as it is now, or as it last existed there once it has left
// (CR 608.2h), or the revealed card's printed power. Clamped at zero, since
// a creature with negative power deals no damage (CR 120.3).
func b25ChosenPower(ctx *Context) (int, bool) {
	n, ok := ctx.Game.ChosenPowerForEffect(ctx.Item)
	return max(n, 0), ok
}
