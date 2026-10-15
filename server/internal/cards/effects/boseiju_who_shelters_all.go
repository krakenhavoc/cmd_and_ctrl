package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boseiju, Who Shelters All — Legendary Land (EDHREC rank 2743):
//
//	"Boseiju enters tapped.
//	 {T}, Pay 2 life: Add {C}. If that mana is spent on an instant or
//	 sorcery spell, that spell can't be countered."
//
// Cavern of Souls for spells. The {C} is unrestricted and carries a
// can't-be-countered spend rider (#1547) whose filter is the printed
// "an instant or sorcery spell"; paid into anything else, the rider
// does nothing and the mana is ordinary {C}.
//
// The enters-tapped clause is a CR 614 self-replacement; the life is a
// cost, so the auto-tapper never plans it (ADR 0040 §7).
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:     "36937483-30cb-449a-8028-75017a124922",
		Name:         "Boseiju, Who Shelters All",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Life: 2},
			Produced: "{C}",
			Label:    "Add {C}",
			SpendRiders: []game.ManaSpendRider{
				SpentSpellCantBeCountered(ManaRestrictCast, ManaRestrictAnyType("Instant", "Sorcery")),
			},
		}},
	})
}
