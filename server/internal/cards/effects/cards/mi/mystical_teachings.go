package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystical Teachings — Instant {3}{U} (Foundations, #4465):
//
//	"Search your library for an instant card or a card with flash,
//	 reveal it, put it into your hand, then shuffle.
//	 Flashback {5}{B} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Waterlogged Teachings' tutor (b06TutorToHand with the same "an instant
// card or a card with flash" predicate) and the shared Flashback
// alternative cost, which casts it from the graveyard and exiles it as it
// leaves the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "81d72511-88b4-45ec-a5a9-e722caf55577",
		Name:             "Mystical Teachings",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{5}{B}")},
		OnResolve: effects.B06TutorToHand("Mystical Teachings — an instant card or a card with flash, revealed, to your hand",
			effects.WaterloggedTeachingsFinds),
	})
}
