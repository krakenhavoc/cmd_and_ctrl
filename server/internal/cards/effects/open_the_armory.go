package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Open the Armory — Sorcery {1}{W} (EDHREC rank 640):
//
//	"Search your library for an Aura or Equipment card, reveal it,
//	 put it into your hand, then shuffle."
//
// The Voltron deck's tutor. A library search to hand, Demonic Tutor's
// shape with a predicate: the Aura and Equipment SUBTYPES, read off
// the card's type line (a library card is off the battlefield, so
// HasSubtype falls through to the printed line, which is the right
// read for a card in a hidden zone). Equipment and Auras themselves
// still wait on #280; the tutor for them does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ecf7771-8061-4710-ad2e-80b092ae0b4b",
		Name:         "Open the Armory",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.HasSubtype("Aura") || c.HasSubtype("Equipment") },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Open the Armory — an Aura or Equipment card",
			}.Apply(ctx)
		},
	})
}
