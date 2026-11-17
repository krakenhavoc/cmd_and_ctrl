package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Begin the Invasion — Sorcery {X}{W}{U}{B}{R}{G} (Commander: March of the
// Machine, S82 batch 23):
//
//	"Search your library for up to X battle cards with different names, put
//	 them onto the battlefield, then shuffle."
//
// A library search with the different-names set rule, so the prompt refuses
// two battles of one name (CR 201.2) and a library short of battles finds
// fewer, none included (CR 701.23b). The cards enter as one simultaneous
// batch; each battle then runs its own entry (CR 310.4, CR 310.9a): the
// printed defense counters and its controller's choice of an opponent to
// protect it, one prompt per battle, answered in turn. A Siege's own "When
// this Siege enters" ability triggers for each. X is the value announced
// for the spell (CR 107.3m). The library is shuffled even when X is 0 or
// nothing is found (CR 701.24d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4feca1da-dcc8-440c-8933-bab91782b015",
		Name:         "Begin the Invasion",
		XMatters:     true,
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
			}
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsBattle() },
				Dest:      game.ZoneBattlefield,
				Limit:     x,
				Shuffle:   true,
				Rule:      searchRuleDifferentNames,
				Reason:    "Begin the Invasion — up to X battle cards with different names",
			}.Apply(ctx)
		},
	})
}
