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
// One declared simplification, weaker than printed: with strict mana
// off the pool is never spent (ADR 0068 §3), so no token pays and the
// creature never gains haste.
func init() {
	Register(Spec{
		OracleID:     "32fe7ac4-86f5-44af-9f73-ee8f6a9ce2ba",
		Name:         "Hall of the Bandit Lord",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so the creature never gains haste."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:        ManaAbilityCost{Tap: true, Life: 3},
			Produced:    "{C}",
			Label:       "Add {C}",
			SpendRiders: []game.ManaSpendRider{SpentCreatureGainsHaste()},
		}},
	})
}
