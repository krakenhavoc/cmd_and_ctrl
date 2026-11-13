package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonstorm (the card) — Sorcery {8}{R} (Scourge):
//
//	"Search your library for a Dragon permanent card, put it onto the
//	 battlefield, then shuffle.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn.)"
//
// The file is named for the spell because dragonstorm.go already holds the
// Tarkir: Dragonstorm cycle's shared trigger.
//
// The shared library search straight to the battlefield, over permanent
// cards with the Dragon subtype (a Dragon creature, or a kindred Dragon
// artifact or enchantment), plus the storm keyword (CR 702.40a). Each copy
// is its own search and puts its own Dragon onto the battlefield. The
// searcher need not find a card even when one is there (CR 701.23b), and
// the library is shuffled either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d74c3871-018a-4450-a613-f67703bb0d72",
		Name:         "Dragonstorm",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: scgB1PermanentCardOfSubtype(ctx.TextSubtype("Dragon")),
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Dragonstorm — a Dragon permanent card, onto the battlefield",
			}.Apply(ctx)
		},
	})
}
