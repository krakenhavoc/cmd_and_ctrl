package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Parting of the Ways — Enchantment — Saga {4}{R}{R} (Doctor Who):
//
//	"I — Exile the top five cards of your library. For each nonland card
//	 exiled this way, put a number of time counters on that card equal to
//	 its mana value. If it doesn't have suspend, it gains suspend.
//	 II — Time travel, then time travel.
//	 III — For each opponent, destroy up to one target artifact that
//	 player controls."
//
// I is Kang Prime's exile with time counters and a granted suspend
// (game.GainSuspendForEffect, CR 702.62a), with the amount the card's own
// mana value. II is two time travels in a row (CR 701.56a). III is
// Dismantling Wave's per-opponent "up to one" clause list.
//
// No simplification.
func init() {
	const name = "The Parting of the Ways"
	destroy := ChapterTrigger(3, SagaChapterLabel(name, 3, "for each opponent, destroy up to one target artifact that player controls"),
		func(g *game.Game, item *game.StackItem) error {
			return destroyEachLegalTarget(item, NewContext(g, item))
		})
	destroy.Targets = perOpponentClauses(game.MaxPlayers-1, "artifact", Artifact())
	Register(Spec{
		OracleID:     "58d5b285-eb0c-4478-b137-26a897075f70",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "exile the top five cards of your library; time counters equal to mana value, and suspend"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{Player: ctx.Controller(), N: 5, To: game.ZoneExile, ThenTail: ctx.Tail(partingOfTheWaysCounted)}.Apply(ctx)
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "time travel, then time travel"), Do(TimeTravel{}, TimeTravel{})),
			destroy,
		},
	})
}

// partingOfTheWaysCounted is chapter I's rest: each nonland card that
// reached exile gets time counters equal to its mana value, and suspend if
// it has none. An on-disk key.
var partingOfTheWaysCounted = CardTail("the-parting-of-the-ways/time-counters", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		if err := ctx.Game.GainSuspendForEffect(ctx.Controller(), id, c.ManaValue()); err != nil {
			return err
		}
	}
	return nil
})
