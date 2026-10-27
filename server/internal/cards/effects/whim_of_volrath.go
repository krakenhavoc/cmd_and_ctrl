package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whim of Volrath — Instant {U}:
//
//	"Buyback {2} (You may pay an additional {2} as you cast this spell.
//	 If you do, put this card into your hand as it resolves.)
//	 Change the text of target permanent by replacing all instances of
//	 one color word with another or one basic land type with another
//	 until end of turn. (For example, you may change "nonred creature"
//	 to "nongreen creature" or "plainswalk" to "swampwalk.")"
//
// Mind Bend for the turn, with Capsize's mana buyback: the engine
// returns the spell to its owner's hand as it resolves when buyback was
// paid. CR 612.2, text_words.go, ADR 0160's 2026-10-15 amendment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "51a20c9f-68d2-471b-8622-67d481b53b92",
		Name:          "Whim of Volrath",
		Completeness:  CompletenessFull,
		Targets:       TargetPermanent("target permanent"),
		OptionalCosts: []game.AdditionalCost{Buyback("{2}")},
		OnResolve:     changeTextOfTarget(textEitherWords, true, nothingAfterTheChange),
	})
}
