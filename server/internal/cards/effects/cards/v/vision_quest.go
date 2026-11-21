package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vision Quest — {X}{U}{R} Sorcery (Marvel Super Heroes, #4455):
//
//	"Search your library and/or graveyard for an artifact creature card
//	 with mana value X or less and put it onto the battlefield with X
//	 additional +1/+1 counters on it. If X is 4 or greater, it gains haste
//	 until end of turn. If you search your library this way, shuffle."
//
// Finale of Devastation's X search over the library and/or graveyard
// (SearchLibrary.AlsoGraveyard: the zones are asked first and the shuffle
// follows only a searched library, #4516/#4519), with the found creature
// entering with X counters (SearchLibrary.EntersWithCounters, so they are
// on it before it enters). X rides the continuation's tail; the haste is
// granted to the creature that came in, and only when X is 4 or more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "240410a0-3288-4336-b1ad-34936e64ccf2",
		Name:         "Vision Quest",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			tail := ctx.ItemTail(visionQuestHaste)
			tail.N = x
			search := effects.SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					return c.IsCreature() && c.IsArtifact() && c.ManaValue() <= x
				},
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				AlsoGraveyard: true,
				Reason:        "Vision Quest — an artifact creature card with mana value X or less",
				ThenTail:      tail,
			}
			if x > 0 {
				search.EntersWithCounters = map[string]int{game.CounterPlusOne: x}
			}
			return search.Apply(ctx)
		},
	})
}

// visionQuestHaste is "if X is 4 or greater, it gains haste until end of
// turn" (X is Tail.N), for the creature the search put onto the battlefield.
var visionQuestHaste = effects.CardTail("vision-quest/haste", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N < 4 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (effects.GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    "Vision Quest — haste",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
