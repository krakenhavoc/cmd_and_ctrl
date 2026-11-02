package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merciless Enforcers — Creature — Human Mercenary Villain {1}{B}, 2/1
// (Marvel's Spider-Man, #4460):
//
//	"Lifelink
//	 {3}{B}: This creature deals 1 damage to each opponent."
//
// Lifelink is a printed keyword. The activated ability is the Thermo-Alchemist
// pinger on a mana cost: the creature is the damage source (one event per
// opponent), so lifelink gains its controller the life and a Fog-class
// shield stops it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b32a7b97-0883-4c9f-a23d-ad269efd8cc3",
		Name:            "Merciless Enforcers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}: This creature deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			},
		}},
	})
}
