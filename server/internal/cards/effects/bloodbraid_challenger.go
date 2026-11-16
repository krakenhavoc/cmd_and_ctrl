package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodbraid Challenger — Creature — Elf Berserker {3}{R}{G}, 4/3 (Modern
// Horizons 3 Commander, #5468):
//
//	"Cascade
//	 Haste
//	 Escape—{3}{R}{G}, Exile three other cards from your graveyard."
//
// Bloodbraid Elf's cascade and haste, with escape. Cascade is a cast
// trigger, so it cascades however the Challenger is cast, from the
// graveyard by its escape cost included. Escape is the shared
// alternative cost: the graveyard opens as a zone to cast it from, and
// the three exiled cards are a cost, never the Challenger itself. It
// escapes with no counters and no other change.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "6e111622-c7cc-49cf-83e8-3555e06518c1",
		Name:             "Bloodbraid Challenger",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"haste"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{3}{R}{G}", 3)},
		Triggered:        []game.TriggeredAbility{Cascade()},
	})
}
