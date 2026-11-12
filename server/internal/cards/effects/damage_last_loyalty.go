package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// damage_last_loyalty.go — the card-side sentences over
// game/damage_last_loyalty.go (CR 120.3c, 120.4c, #5646), and the
// "choose a planeswalker type" as-enters choice they read.

// ChoosePlaneswalkerTypeAsEnters builds the `Spec.AsEnters` for "As this
// permanent enters, choose a planeswalker type." (CR 614.12). The options
// are CR 205.3j's planeswalker types, those of the planeswalkers the
// chooser controls first, through the ordinary option pick; the answer
// lands on the permanent's Card.ChosenOption, and
// PlaneswalkersYouControlOfTheChosenType reads it. `name` is the card's
// name, the prompt's header.
func ChoosePlaneswalkerTypeAsEnters(name string) func(*game.Card, *Context) error {
	question := name + " — choose a planeswalker type"
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueChooseOptionAsEntersForEffect(card.Controller, card.InstanceID, question,
			leoriPlaneswalkerTypeOptions(ctx.Game, card.Controller))
		return nil
	}
}

// PlaneswalkersYouControlOfTheChosenType is "planeswalkers you control of
// the chosen type": a planeswalker the source's controller controls that
// has the planeswalker type chosen for the source as it entered
// (ChoosePlaneswalkerTypeAsEnters). Before the choice is answered it
// matches nothing.
func PlaneswalkersYouControlOfTheChosenType(target *game.Card, g *game.Game, source *game.Card) bool {
	return source.ChosenOption != "" && PlaneswalkersYouControl(target, g, source) &&
		target.HasSubtype(source.ChosenOption)
}

// DamageLeavesChosenTypeAPlaneswalkersLastLoyaltyWhileYouControlACreature
// is "As long as you control a creature, if damage dealt to a
// planeswalker you control of the chosen type would result in all loyalty
// counters on it being removed, instead all but one of those counters are
// removed." (Deification). The damage is still dealt in full; only the
// loyalty it removes is capped. Pair it with
// ChoosePlaneswalkerTypeAsEnters.
func DamageLeavesChosenTypeAPlaneswalkersLastLoyaltyWhileYouControlACreature() []game.DamageLastLoyaltyStatic {
	return []game.DamageLastLoyaltyStatic{{
		Label:                    "As long as you control a creature, if damage dealt to a planeswalker you control of the chosen type would result in all loyalty counters on it being removed, instead all but one of those counters are removed.",
		OfChosenType:             true,
		WhileYouControlACreature: true,
	}}
}
