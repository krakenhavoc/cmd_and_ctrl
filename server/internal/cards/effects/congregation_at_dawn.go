package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Congregation at Dawn — Instant {G}{G}{W} (Ravnica: City of Guilds,
// #5600):
//
//	"Search your library for up to three creature cards, reveal them, then
//	 shuffle and put those cards on top in any order."
//
// Worldly Tutor's search with a limit of three: the found cards are
// revealed, the library is shuffled, and they go on top. A search with a
// stated quality may find fewer (CR 701.23b). With two or more found, the
// caster then arranges them (the ordered library placement), and the
// other players, who saw which cards they were, don't learn the order
// (CR 401.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b63dad39-4dd5-4340-953e-b05df1a3c776",
		Name:         "Congregation at Dawn",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsCreature() },
				Dest:      game.ZoneLibrary,
				ToTop:     true,
				Limit:     3,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Congregation at Dawn — up to three creature cards",
				ThenTail:  ctx.ItemTail(congregationAtDawnOrder),
			}.Apply(ctx)
		},
	})
}

// congregationAtDawnOrder is "put those cards on top in any order": the
// found cards are already on top, top-first, and the caster arranges them.
// An on-disk key: never renamed, never reused.
var congregationAtDawnOrder = CardTail("then/congregation-at-dawn-order", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) < 2 {
		return nil
	}
	return PutInLibraryInAnyOrder{
		Cards:     o.IDs,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Label:     "Congregation at Dawn — put those cards on top in any order",
	}.Apply(ctx)
})
