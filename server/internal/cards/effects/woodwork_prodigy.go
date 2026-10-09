package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodwork Prodigy // Soul Tether — Creature — Cat Druid {2}{R/G}, 3/3 //
// Sorcery {2}{R/G} (preparation card, CR 722):
//
//	"At the beginning of your upkeep, if this creature isn't prepared,
//	 it becomes prepared. (While it's prepared, you may cast a copy of
//	 its spell. Doing so unprepares it.)"
//
//	Soul Tether — "Create a Heartwood token. (It's a red and green
//	 artifact with '{T}: Add {R} or {G}.')"
//
// It does not enter prepared; the upkeep trigger is its only source of
// the designation.
//
// No simplification.
func init() {
	const id = "f3ed157f-a09b-405f-aec4-0e7a81d80fc8"
	Register(Spec{
		OracleID:     id,
		Name:         "Woodwork Prodigy",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{fraBecomesPreparedAtUpkeep("Woodwork Prodigy")},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Soul Tether",
		Completeness: CompletenessFull,
		OnResolve:    soulTetherResolve,
	})
}
