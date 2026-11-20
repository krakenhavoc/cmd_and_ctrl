package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arena Rector — Creature — Human Cleric {3}{W}, 1/2 (Battlebond, #5529):
//
//	"When this creature dies, you may exile it. If you do, search your
//	 library for a planeswalker card, put it onto the battlefield, then
//	 shuffle."
//
// Cavalier of Thorns' "you may exile it. If you do": the trigger's own
// yes/no, then the Rector is exiled from the graveyard as it resolves. A
// Rector that has left the graveyard in response (reanimated, or exiled by
// something else) cannot be exiled, so "if you do" fails and there is no
// search. The search is for any planeswalker card, put onto the
// battlefield; the searcher may find nothing, and the library is shuffled
// either way.
//
// No simplification.
func init() {
	t := WhenThisDies("Arena Rector — exile it, then search your library for a planeswalker card and put it onto the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			z := g.FindCardZoneForEffect(item.SourceCardID)
			if z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			return ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(arenaRectorExiled)}.Apply(ctx)
		})
	t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Arena Rector — exile it to search your library for a planeswalker card?"}
	Register(Spec{
		OracleID:     "6a23963c-21dc-4d21-8d28-5abad7f859a1",
		Name:         "Arena Rector",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// arenaRectorExiled is the clause after the exile: only when the Rector was
// exiled is the library searched. An on-disk key: never renamed, never
// reused.
var arenaRectorExiled = CardTail("bbd/arena-rector-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return c.IsPlaneswalker() },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Arena Rector — a planeswalker card",
	}.Apply(ctx)
})
