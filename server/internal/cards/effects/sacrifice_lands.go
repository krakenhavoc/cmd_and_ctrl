package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sacrifice_lands.go — Fallen Empires' five tapped lands that can be
// cashed in for two mana of their colour (#5637): Dwarven Ruins, Ebon
// Stronghold, Havenwood Battleground, Ruins of Trokair and Svyelunite
// Temple.
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {T}, Sacrifice this land: Add {C}{C}."
//
// Crystal Vein's two mana abilities in one colour, behind the tapped
// entry. Both resolve without the stack (CR 605.3b); the land is
// sacrificed as part of the second one's cost.
//
// Append-only: add a builder, never change what one means.

// tappedSacrificeLand is the whole Spec of one of those lands. `color`
// is the mana symbol's letter ("W", "U", …).
func tappedSacrificeLand(oracleID, name, color string) Spec {
	one := "{" + color + "}"
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
				Produced: one + one,
				Label:    "{T}, Sacrifice this land: Add " + one + one,
			},
		},
	}
}
