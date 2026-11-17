package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_bro_b17_helpers.go — shared pieces of The Brothers' War batch 17
// (S82): "for each {C}{C} spent to cast it" and entry counters read off
// the mana that paid for the spell.

// pairsSpent is the number of times "{C}{C}" was spent: "for each {U}{U}
// spent to cast it" counts pairs, so four blue mana is two of them (the
// same reading as Deceit's "{U}{U} was spent", per pair rather than once).
func pairsSpent(m game.ManaSpent, color string) int { return m.Count(color) / 2 }

// WhenThisEntersForEachPairSpent is "When this creature enters, for each
// {C}{C} spent to cast it, <effect>": the pairs are counted as the
// trigger is built, off the payment record the permanent carries
// (CR 400.7d), and the effect reads them from item.Params.Amount. A
// permanent that was not cast, or was cast for no mana, makes 0 and the
// effect does nothing. Every card using it declares ReadsManaSpent.
func WhenThisEntersForEachPairSpent(color, label string, effect Effect) game.TriggeredAbility {
	t := WhenThisEnters(label, effect)
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = pairsSpent(source.ManaSpentToCast(), color)
		return item
	}
	return t
}

// SelfEntersWithCountersFromSpentMana is a CR 614.1c self-replacement
// "this creature enters with <counters> on it" whose number is read off
// the mana spent to cast it inside the entry window
// (game.EntrySpentForEffect). An entry that was not a cast spent nothing,
// so count sees the zero ManaSpent. A count of zero or less adds nothing.
func SelfEntersWithCountersFromSpentMana(kind, rowLabel string, count func(game.ManaSpent) int) game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:  rowLabel,
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID &&
				count(g.EntrySpentForEffect(ev)) > 0
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(kind, count(g.EntrySpentForEffect(ev)))
			return nil
		},
	}
}

// drawPerPairSpent is "draw a card" once for each pair counted into the
// item (item.Params.Amount, WhenThisEntersForEachPairSpent). It declares
// the draw's footprint so the row is not an unfootprinted one-sentence
// trigger; the item its Build makes carries the count in Params, which the
// trigger-order check never accepts as plain (builtItemIsPlain), so the
// row keeps its ordering prompt and the declared N of 1 is never trusted.
type drawPerPairSpent struct{}

func (drawPerPairSpent) Apply(ctx *Context) error {
	if ctx.Item == nil {
		return nil
	}
	return DrawCards{N: ctx.Item.Params.Amount}.Apply(ctx)
}

func (drawPerPairSpent) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintDraw, N: 1}, true
}

// counterOnThisIfSpentExceedsPower is Liberator's whole effect: the
// intervening "if" (CR 603.4) rechecked as the item resolves, against the
// mana spent that the trigger carried (item.Params.Amount) and this
// permanent's power now, then a +1/+1 counter on it. Footprint declared as
// the counter on the source; as with drawPerPairSpent the carried amount
// keeps the item out of the order check's plain items.
type counterOnThisIfSpentExceedsPower struct{}

func (counterOnThisIfSpentExceedsPower) Apply(ctx *Context) error {
	if ctx.Item == nil || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) ||
		!liberatorBeaten(ctx.Game, ctx.Item.SourceCardID, ctx.Item.Params.Amount) {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

func (counterOnThisIfSpentExceedsPower) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintCounterOnSource, N: 1, Counter: game.CounterPlusOne}, true
}
