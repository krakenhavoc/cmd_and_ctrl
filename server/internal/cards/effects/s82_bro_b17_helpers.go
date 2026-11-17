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

// DrawForEach is "draw a card for each <thing>", counted as the item
// resolves: InvestigateForEach's contract for Count, and its footprint with
// the draw step in place of the investigate step (a declared board read,
// then the draw). Bladecoil Serpent's count is the pairs its Build carried
// on the item.
type DrawForEach struct {
	// Name says what it counts, for a reader of the card file.
	Name  string
	Count func(ctx *Context) int
}

func (s DrawForEach) Apply(ctx *Context) error {
	if s.Count == nil {
		return nil
	}
	return DrawCards{N: s.Count(ctx)}.Apply(ctx)
}

func (s DrawForEach) footprints() ([]game.FootprintStep, bool) {
	if s.Count == nil {
		return nil, false
	}
	step, _ := DrawCards{N: 1}.footprint()
	return []game.FootprintStep{condStep(game.FootprintCondBoard, 0), step}, true
}

// itemPairsSpent is the pair count WhenThisEntersForEachPairSpent put on
// the item.
func itemPairsSpent(ctx *Context) int {
	if ctx.Item == nil {
		return 0
	}
	return ctx.Item.Params.Amount
}
