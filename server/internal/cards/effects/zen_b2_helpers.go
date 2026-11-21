package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// zen_b2_helpers.go — shared bodies for Zendikar batch 2 (#5583): the
// Ally cycle's "Whenever this creature or another Ally you control
// enters, you may …" trigger.
//
// The trigger is Harvestrite Host's shape with an Ally in place of a
// Rabbit: the card's own entry, or another permanent with the subtype
// entering under its controller's control (a token, a changeling), read
// post-layer. "Ally" is a word of the card's text (CR 612.2), so a
// text-changing effect reaches it. Every Zendikar Ally asks its "you
// may" as the ability resolves (CR 603.5).

// zenB2AllyEntered is "this creature or another Ally you control
// enters".
func zenB2AllyEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID == source.InstanceID ||
		B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
}

// zenB2AllyTrigger is the optional Ally trigger with its body.
func ZenB2AllyTrigger(label, question string, effect Effect) game.TriggeredAbility {
	return Optional(On(game.EventETB, zenB2AllyEntered, label, effect), question)
}

// zenB2AllyCounter is "…, you may put a +1/+1 counter on this
// creature." The counter goes on the creature only while it is still
// the same object on the battlefield (CR 400.7).
func ZenB2AllyCounter(name string) game.TriggeredAbility {
	return ZenB2AllyTrigger(name+" — you may put a +1/+1 counter on it",
		"Put a +1/+1 counter on "+name+"?",
		Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))
}

// zenB2AllyGrant is "…, you may have Ally creatures you control gain
// <keyword> until end of turn." The set is locked as the ability
// resolves (CR 611.2c): an Ally that enters afterwards does not gain it.
func ZenB2AllyGrant(name, keyword string) game.TriggeredAbility {
	return ZenB2AllyTrigger(name+" — Ally creatures you control gain "+keyword+" until end of turn",
		"Have Ally creatures you control gain "+keyword+" until end of turn?",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			return GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl(), Subtype(ctx.TextSubtype("Ally"))),
				Keywords: []string{keyword},
				Label:    name + " — " + keyword + " until end of turn",
			}.Apply(ctx)
		})
}

// zenB2AlliesYouControl counts the Allies (any permanent with the
// subtype, not only creatures) the ability's controller controls as it
// resolves.
func ZenB2AlliesYouControl(ctx *Context) int {
	ally := ctx.TextSubtype("Ally")
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.HasSubtype(ally) {
			n++
		}
	}
	return n
}
