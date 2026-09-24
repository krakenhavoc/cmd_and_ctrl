package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noble Hierarch — Creature — Human Druid {G}, 0/1:
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)
//	 {T}: Add {G}, {W}, or {U}."
//
// Ignoble Hierarch's Bant twin. Exalted rides the shared Exalted()
// constructor (exalted.go); the mana ability is the same plain
// three-colour pipe shape, no commander-identity narrowing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98aa9424-5912-4bd6-9300-b3972a31d8af",
		Name:         "Noble Hierarch",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Exalted()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|W|U}",
			Label:    "Add {G}, {W}, or {U}",
		}},
	})
}
