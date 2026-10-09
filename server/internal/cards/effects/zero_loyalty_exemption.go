package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// zero_loyalty_exemption.go — the card-side sentence over
// game/zero_loyalty_exemption.go (CR 704.5i, #2797).

// PlaneswalkersSurviveZeroLoyalty is "Planeswalkers you control aren't
// put into their owners' graveyards for having 0 loyalty." (Sanctum
// Lurker). It covers the source's controller's planeswalkers only, as
// printed, so a planeswalker an opponent controls still dies to
// CR 704.5i.
func PlaneswalkersSurviveZeroLoyalty() []game.ZeroLoyaltyExemption {
	return []game.ZeroLoyaltyExemption{{
		Label: "Planeswalkers you control aren't put into their owners' graveyards for having 0 loyalty.",
		Whose: game.ZeroLoyaltyExemptYours,
	}}
}
