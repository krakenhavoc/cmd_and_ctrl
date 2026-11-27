package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Academy Rector — Creature — Human Cleric {3}{W}, 1/2 (Urza's Destiny, S82
// #5621):
//
//	"When this creature dies, you may exile it. If you do, search your
//	 library for an enchantment card, put that card onto the battlefield, then
//	 shuffle."
//
// Arena Rector's shape: the trigger's own yes/no, then the Rector is exiled from
// the graveyard as it resolves. A Rector that has left the graveyard in response
// cannot be exiled, so "if you do" fails and there is no search. The search is
// for any enchantment card, put onto the battlefield; the searcher may find
// nothing, and the library is shuffled either way.
//
// No simplification.
func init() {
	t := effects.WhenThisDies("Academy Rector — exile it, then search your library for an enchantment card and put it onto the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			z := g.FindCardZoneForEffect(item.SourceCardID)
			if z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			return effects.ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(academyRectorExiled)}.Apply(ctx)
		})
	t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Academy Rector — exile it to search your library for an enchantment card?"}
	effects.Register(effects.Spec{
		OracleID:     "e3c85068-b4b6-40b9-a16c-5c3b2d059ec4",
		Name:         "Academy Rector",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered:    []game.TriggeredAbility{t},
	})
}

// academyRectorExiled is the clause after the exile: only when the Rector was
// exiled is the library searched. An on-disk key: never renamed, never
// reused.
var academyRectorExiled = effects.CardTail("s82-old-b705/academy-rector-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return effects.SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return c.IsEnchantment() },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Academy Rector — an enchantment card",
	}.Apply(ctx)
})
