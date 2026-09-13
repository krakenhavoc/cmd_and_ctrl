package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Otherworldly Gaze — Instant {U}:
//
//	"Surveil 3."
//	"Flashback {1}{U}"
//
// One mana, three cards deep, and everything you do not want goes
// into the graveyard rather than under the library — which is the
// point of the card in a deck that wants a full graveyard.
//
// # Declared sandbox simplification: NO FLASHBACK
//
// The flashback half is not implemented. Casting a card from a
// graveyard for an alternative cost, then exiling it instead of
// letting it return there, is an alternative CAST PATH — a different
// zone to cast from plus a replacement on where it goes afterwards.
// Spec.AlternativeCosts covers alternative costs paid from the hand
// (overload, evoke, cleave); it has no "cast this from your
// graveyard" leg, and inventing one for a single card would be the
// wrong place to design it.
//
// The consequence is honest and small: this is a one-shot Surveil 3
// rather than a two-for-one. Casting it from the graveyard by hand is
// still possible in the sandbox — the card just does not offer the
// button, and the engine will not exile it afterwards.
func init() {
	Register(Spec{
		OracleID: "be668c2d-71ea-4346-8980-1fbf5e4cbed3",
		Name:     "Otherworldly Gaze",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
