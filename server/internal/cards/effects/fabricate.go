package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fabricate — Sorcery {2}{U}:
//
//	"Search your library for an artifact card, reveal it, put it into
//	 your hand, then shuffle."
//
// The blue artifact tutor, and the counterpart to the one-mana cycle
// next door: three mana buys the card in HAND rather than on top, so
// there is no draw step to wait through. Straight SearchLibrary with
// an artifact predicate.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "422e1869-134f-463d-9fa1-86b66a998b3e",
		Name:     "Fabricate",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsArtifact() },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Fabricate — an artifact card",
			}.Apply(ctx)
		},
	})
}
