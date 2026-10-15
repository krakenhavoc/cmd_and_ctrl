package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hall of the Bandit Lord — Legendary Land (EDHREC rank 2995):
//
//	"Hall of the Bandit Lord enters tapped.
//	 {T}, Pay 3 life: Add {C}. If that mana is spent on a creature
//	 spell, it gains haste."
//
// The haste land. The {C} is unrestricted — it pays for anything — and
// carries a spend rider (#1547) whose filter is "a creature spell".
// When the token pays for one, the rider is stamped on the spell's
// payment record, which CR 400.7d carries onto the permanent as its
// provenance; the layer engine grants haste to every permanent whose
// provenance says so. No end-of-turn clause is printed, so the haste
// lasts for as long as the creature remains that object — a flicker
// makes a new one that never had the mana spent on it.
//
// The enters-tapped clause is a CR 614 self-replacement. The life is a
// COST (validated before the tap), so the auto-tapper never plans the
// Hall — paying life is a decision, the painlands' posture (ADR 0040
// §7).
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:     "32fe7ac4-86f5-44af-9f73-ee8f6a9ce2ba",
		Name:         "Hall of the Bandit Lord",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:        ManaAbilityCost{Tap: true, Life: 3},
			Produced:    "{C}",
			Label:       "Add {C}",
			SpendRiders: []game.ManaSpendRider{SpentCreatureGainsHaste()},
		}},
	})
}
