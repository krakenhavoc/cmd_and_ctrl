package roadmap

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// declaresRevealCards is the "Reveal cards from your hand" cost probe
// (#2598): an activated ability whose cost has a reveal component
// (game.AbilityCost.RevealCards, built by effects.RevealX / RevealN).
func declaresRevealCards(s effects.Spec) bool {
	for _, a := range s.Activated {
		if a.Cost.RevealCards != nil {
			return true
		}
	}
	return false
}
