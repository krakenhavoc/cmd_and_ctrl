package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "53fb8118-f332-407b-add8-1b329e420226",
		Name:         "Mwonvuli Acid-Moss",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Lands: 1},
		Targets:      TargetPermanent("target land", Land()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     IsLandWithSubtype("Forest"),
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Mwonvuli Acid-Moss — a Forest card, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
