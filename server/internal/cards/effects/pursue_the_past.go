package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pursue the Past — Sorcery {R}{W} (Secrets of Strixhaven, #4456):
//
//	"You gain 2 life. You may discard a card. If you do, draw two cards.
//	 Flashback {2}{R}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The life comes first, then the discard is one instruction with a
// conditional second half (b39MayDiscardThenDraw, drawTwoIfDiscarded): the
// card is pitched before the replacements are seen, and discarding nothing
// draws nothing. Flashback is the shared constructor (CR 702.34a); the cast
// from the graveyard exiles the spell as it leaves the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4384996f-aeaa-41ac-aa0e-12d72c887731",
		Name:             "Pursue the Past",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GainLife{Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return b39MayDiscardThenDraw(1, false,
				"Pursue the Past — discard a card to draw two cards?",
				drawTwoIfDiscarded)(ctx)
		},
	})
}
