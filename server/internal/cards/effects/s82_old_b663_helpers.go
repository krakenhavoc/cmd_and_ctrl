package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b663_helpers.go — shared piece of Invasion batch 663 (#5616).

// OldB663SacrificeLand is the whole Spec of one of Invasion's five tapped
// lands that tap for one color and can be cashed in for another pair:
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {T}, Sacrifice this land: Add {W}{B}."
//
// TappedSacrificeLand's shape (a tapped entry and two mana abilities that
// resolve without the stack, CR 605.3b; the land is sacrificed as part of the
// second one's cost) with the cash-in producing two other colors. `tap` is the
// letter of the color it taps for and `cashIn` the two letters it makes.
func OldB663SacrificeLand(oracleID, name, tap, cashIn string) Spec {
	one := "{" + tap + "}"
	two := "{" + cashIn[:1] + "}{" + cashIn[1:] + "}"
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: one,
				Label:    "Add " + one,
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: two,
				Label:    "{T}, Sacrifice this land: Add " + two,
			},
		},
	}
}
