package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mwonvuli Acid-Moss — Sorcery {2}{G}{G} (Time Spiral, #5596):
//
//	"Destroy target land. Search your library for a Forest card, put that
//	 card onto the battlefield tapped, then shuffle."
//
// Stone Rain, then Binding the Old Gods' Forest search: any land card
// with the Forest type, onto the battlefield tapped. A target that became
// illegal counters the whole spell, search included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53fb8118-f332-407b-add8-1b329e420226",
		Name:         "Mwonvuli Acid-Moss",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Lands: 1},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.SearchLibrary{
				Player:        item.Controller,
				Predicate:     effects.IsLandWithSubtype("Forest"),
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Mwonvuli Acid-Moss — a Forest card, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
