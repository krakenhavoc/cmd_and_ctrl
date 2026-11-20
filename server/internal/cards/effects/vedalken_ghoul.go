package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vedalken Ghoul — Creature — Vedalken Zombie {U}{B}, 1/1 (Alara Reborn,
// #5586):
//
//	"Whenever this creature becomes blocked, defending player loses 4
//	 life."
//
// The trigger afflict prints (CR 702.130a) under its own words: the
// shared Afflict row, which fires once however many creatures block it
// (CR 509.3c) and resolves in the declare blockers step. The Ghoul has no
// afflict keyword, so only the row's label is its own.
//
// No simplification.
func init() {
	ghoul := Afflict(4, "Vedalken Ghoul")
	ghoul.Key = "Vedalken Ghoul — defending player loses 4 life"
	Register(Spec{
		OracleID:     "708b13ef-f297-4176-9d09-e633c4d600da",
		Name:         "Vedalken Ghoul",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{ghoul},
	})
}
