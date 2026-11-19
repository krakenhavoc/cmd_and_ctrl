package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Traverse the Ulvenwald — Sorcery {G} (Innistrad Remastered):
//
//	"Search your library for a basic land card, reveal it, put it into your
//	 hand, then shuffle.
//	 Delirium — If there are four or more card types among cards in your
//	 graveyard, instead search your library for a creature or land card,
//	 reveal it, put it into your hand, then shuffle."
//
// Delirium is an ability word (CR 207.2c), read as the spell resolves, with no
// rules meaning of its own: Demonic Counsel's way, the types come off the
// printed type lines of the cards in your graveyard. The spell itself is on
// the stack, not in the graveyard, so it is not counted.
func init() {
	Register(Spec{
		OracleID:     "c9dd2bf5-32e0-47f7-9a49-501c262d745b",
		Name:         "Traverse the Ulvenwald",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			search := SearchLibrary{
				Player:    item.Controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Traverse the Ulvenwald — a basic land card",
			}
			if b16CardTypesInGraveyard(ctx.Game, item.Controller) >= 4 {
				search.Predicate = func(c game.Card) bool { return c.IsCreature() || c.IsLand() }
				search.Reason = "Traverse the Ulvenwald (delirium) — a creature or land card"
			}
			return search.Apply(ctx)
		},
	})
}
