package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathbellow War Cry — Sorcery {5}{R}{R}{R} (Theros Beyond Death, S82 OLD,
// #5514):
//
//	"Search your library for up to four Minotaur creature cards with different
//	 names, put them onto the battlefield, then shuffle."
//
// Shared Summons' search with the different-names rule
// (SearchRuleDifferentNames): up to four creature cards with the subtype
// Minotaur, no two sharing a name, found as the spell resolves (CR 701.23b: a
// library with fewer finds fewer), put onto the battlefield together; the
// library is shuffled afterwards. The cards are not revealed and not
// tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "947835a2-a72f-44e2-a183-99dcb019d6ff",
		Name:         "Deathbellow War Cry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 4},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			// "Minotaur" is a word of the spell's text (CR 612.2).
			minotaur := ctx.TextSubtype("Minotaur")
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord(minotaur) },
				Dest:      game.ZoneBattlefield,
				Limit:     4,
				UpTo:      true,
				Shuffle:   true,
				Rule:      effects.SearchRuleDifferentNames,
				Reason:    "Deathbellow War Cry — up to four Minotaur creature cards with different names, onto the battlefield",
			}.Apply(ctx)
		},
	})
}
