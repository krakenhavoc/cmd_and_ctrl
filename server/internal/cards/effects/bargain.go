package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bargain.go — CR 702.166, bargain: "You may sacrifice an artifact,
// enchantment, or token as you cast this spell." Append-only.
//
// An optional cost (ADR 0073) with the ordinary sacrifice component, so
// it is validated and paid like Gatekeeper of Malakir's sacrifice
// kicker, with the spell already on the stack. The engine reads nothing
// of it; "if this spell was bargained" is
// ctx.OptionalCostTimes(BargainKey) > 0.
//
//	OptionalCosts: []game.AdditionalCost{Bargain()},   // Beseech the Mirror

// BargainKey is the optional-cost key bargain is announced under.
const BargainKey = "bargain"

// Bargain is the keyword's optional cost.
func Bargain() game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       BargainKey,
		Sacrifice: sacrificeSpec("an artifact, enchantment, or token", Or(Artifact(), Enchantment(), IsTokenPredicate())),
		Label:     "Bargain",
	}
}
