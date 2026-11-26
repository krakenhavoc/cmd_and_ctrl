package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// From the Ashes — Sorcery {3}{R} (Commander 2013, S82 older sets):
//
//	"Destroy all nonbasic lands. For each land destroyed this way, its
//	 controller may search their library for a basic land card and put it
//	 onto the battlefield. Then each player who searched their library this
//	 way shuffles."
//
// Ruination's sweep with Blood Money's "for each … destroyed this way" tail:
// the tail is handed the pre-move copies of the lands that were actually
// destroyed (an indestructible land is not among them), counted per
// controller as they last controlled them. Each such player, in turn order,
// searches once for up to that many basic land cards, which stands for one
// "may search" per land destroyed, and shuffles if they searched. The lands
// come onto the battlefield untapped.
//
// No simplification.
var fromTheAshesSearch = effects.MassTail("mass/from-the-ashes-search", func(ctx *effects.Context, swept []game.Card, _ int) error {
	per := map[string]int{}
	for _, c := range swept {
		per[c.Controller.String()]++
	}
	for _, seat := range effects.ApnapPlayers(ctx.Game) {
		n := per[seat.String()]
		if n <= 0 {
			continue
		}
		if err := (effects.SearchLibrary{
			Player:    seat,
			Predicate: effects.IsBasicLand,
			Dest:      game.ZoneBattlefield,
			Limit:     n,
			Reveal:    true,
			Optional:  true,
			UpTo:      true,
			Shuffle:   true,
			Reason:    "From the Ashes — up to one basic land card for each land destroyed, onto the battlefield",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e229329-65e4-4240-959a-b97b26908c0e",
		Name:         "From the Ashes",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Lands:   1,
			Sweep:   game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.NonbasicLand(), ThenTail: ctx.ItemTail(fromTheAshesSearch)}.Apply(ctx)
		},
	})
}
