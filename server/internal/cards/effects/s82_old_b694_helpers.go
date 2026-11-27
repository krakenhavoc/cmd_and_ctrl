package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b694_helpers.go — shapes of the Mercadian Masques life cards
// (#5619, S82).

// AnyPlayerCastsASpellOfTheChosenColor is "Whenever a player casts a spell
// of the chosen color": any caster, the spell's colours read as the cast
// happens, the colour the one stored on the source (Jeweled Torque).
func AnyPlayerCastsASpellOfTheChosenColor(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || source.ChosenColor == "" {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.HasColor(source.ChosenColor)
}

// oldB694GainTwoLife is the paid branch of "you may pay {2}. If you do,
// you gain 2 life". An on-disk key: never renamed, never reused.
var oldB694GainTwoLife = PayUnlessBranches("pay-unless/old-b694-gain-2-life",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
	}, nil)

// MayPayTwoToGainTwoLife is the effect of Jeweled Torque's trigger: its
// controller may pay {2}; if they do, they gain 2 life.
func MayPayTwoToGainTwoLife(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayPay{
			Chooser:  item.Controller,
			Cost:     "{2}",
			Question: label + " — pay {2} to gain 2 life?",
			Then:     oldB694GainTwoLife,
		}.Apply(NewContext(g, item))
	}
}
