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
// One declared simplification, weaker than printed: with strict mana
// off the pool is never spent (ADR 0068 §3), so the spell can still be
// countered.
func init() {
	Register(Spec{
		OracleID:     "36937483-30cb-449a-8028-75017a124922",
		Name:         "Boseiju, Who Shelters All",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so an instant or sorcery cast with Boseiju's mana can still be countered."},
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
