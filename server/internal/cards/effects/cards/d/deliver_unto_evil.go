package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deliver Unto Evil — Sorcery {2}{B} (War of the Spark, #5520):
//
//	"Choose up to four target cards in your graveyard. If you control a Bolas
//	 planeswalker, return those cards to your hand. Otherwise, an opponent
//	 chooses two of them. Leave the chosen cards in your graveyard and put
//	 the rest into your hand.
//	 Exile Deliver Unto Evil."
//
// Up to four targets in your graveyard. With a Bolas planeswalker (read as the
// spell resolves) every still-legal target returns. Without one you name an
// opponent (Extrapolate the Impossible's choice, skipped with one opponent
// left), who picks two of the cards to leave, or all of them when fewer than
// two remain; the rest go to your hand. The spell exiles itself before the
// choice, so it cannot be one of the cards (it is on the stack, not in the
// graveyard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e73f938-0aa4-4639-a97a-f7e68f645623",
		Name:         "Deliver Unto Evil",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("up to four target cards in your graveyard", effects.YouOwn()).WithCount(0, 4),
		OnResolve:    effects.S82OldB170DeliverUntoEvil,
	})
}
