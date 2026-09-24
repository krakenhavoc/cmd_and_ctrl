package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ignoble Hierarch — Creature — Goblin Shaman {G}, 0/1:
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)
//	 {T}: Add {B}, {R}, or {G}."
//
// Exalted rides the shared Exalted() constructor (exalted.go). The
// mana ability is a plain three-colour pipe — no "any color", no
// commander-identity narrowing, exactly the three colours printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8de43a3-ebd3-4000-b343-a6ffed11d34d",
		Name:         "Ignoble Hierarch",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Exalted()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|R|G}",
			Label:    "Add {B}, {R}, or {G}",
		}},
	})
}
