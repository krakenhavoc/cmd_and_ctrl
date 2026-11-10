package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b11_helpers.go — helpers for the wave-2 Commander batch 11
// (Faeries, #5368). Append-only; every name starts cmdr2B11.

// cmdr2B11FlashByTappingKey is the optional cost's identity on the card.
const cmdr2B11FlashByTappingKey = "flash-by-tapping"

// cmdr2B11FlashByTapping is "You may cast this spell as though it had
// flash by tapping <n> untapped creatures you control with <keyword> in
// addition to paying its other costs" (Tegwyll's Scouring). CR 601.3c: an
// effect that allows the flash only if an additional cost is paid lets
// the caster begin the cast at instant speed; announcing this optional
// cost opens the window for that cast alone, and the taps ride
// cast_spell's teamwork_ids like conspire's. Tapping is not the {T}
// symbol, so a creature that arrived this turn may pay (CR 302.6).
func cmdr2B11FlashByTapping(n int, label string, preds ...CardPredicate) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:      true,
		Key:           cmdr2B11FlashByTappingKey,
		TapOthers:     tapOthersPrice(n, label, preds...),
		Label:         "Tap " + label + " to cast it as though it had flash",
		AsThoughFlash: true,
	}
}
