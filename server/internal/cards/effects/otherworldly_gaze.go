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
// Flashback {1}{U} (S29). The card predates the graveyard cast path
// and used to declare "no flashback" in this comment. It kept that
// gap after #409/#411 shipped the path, because nothing re-read it:
// the coverage probe reads Caveats, and this card had none. It now
// declares the same pair Faithless Looting does — the graveyard as a
// cast source and the flashback offer bound to it — so the second
// Surveil 3 costs {1}{U} and the card is exiled when it leaves the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "be668c2d-71ea-4346-8980-1fbf5e4cbed3",
		Name:             "Otherworldly Gaze",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
