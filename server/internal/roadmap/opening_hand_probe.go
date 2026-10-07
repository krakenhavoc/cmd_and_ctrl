package roadmap

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// declaresOpeningHand is the "opening-hand actions" probe (#2190, ADR
// 0133): a card that offers its owner the CR 103.6a action of beginning
// the game with it on the battlefield (Spec.OpeningHand). Kept out of
// registry.go so the seam's entry there stays one hunk.
func declaresOpeningHand(s effects.Spec) bool { return s.OpeningHand != nil }
