package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jackal, Genius Geneticist — Legendary Creature — Human Scientist Villain
// {G}{U}, 1/1 (Marvel's Spider-Man, #4460):
//
//	"Trample
//	 Whenever you cast a creature spell with mana value equal to Jackal's
//	 power, copy that spell, except the copy isn't legendary. Then put a
//	 +1/+1 counter on Jackal. (The copy becomes a token.)"
//
// The condition is read as the spell is cast (it is not an intervening "if"):
// the spell's mana value on the stack, with an announced X counted
// (CR 202.3e), against Jackal's power through the layers as he is then. The
// copy is Double Major's (CopySpell with the legendary supertype removed; it
// becomes a token as it resolves, CR 707.10); it names the spell and does
// not target it, so a spell countered in response is still copied from
// last-known information (CR 608.2h). The counter is a separate sentence
// ("Then"), put on Jackal if he is still on the battlefield, whether or not
// a copy was made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5998f07e-01f9-4d02-94dd-867c7166463f",
		Name:            "Jackal, Genius Geneticist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, jackalCastCreatureSpellWithManaValueEqualToPower,
				"Jackal, Genius Geneticist — copy that spell, except the copy isn't legendary; then put a +1/+1 counter on Jackal",
				Do(Instruction(jackalCopyThatSpell), CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// jackalCastCreatureSpellWithManaValueEqualToPower is the trigger condition.
func jackalCastCreatureSpellWithManaValueEqualToPower(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !spell.IsCreature() {
		return false
	}
	mv, ok := g.ManaValueForEffect(spell)
	return ok && mv == source.CurrentPower()
}

// jackalCopyThatSpell copies the cast spell, the copy not legendary.
func jackalCopyThatSpell(ctx *Context) error {
	return CopySpell{
		StackID:       ctx.Trigger().Event.CardID,
		Controller:    ctx.Controller(),
		FromLastKnown: true,
		Except: func(v *game.PrintedValues) {
			v.RemoveSupertype("Legendary")
		},
	}.Apply(ctx)
}
