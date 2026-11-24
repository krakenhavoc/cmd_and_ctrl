package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urban Daggertooth — Creature — Dinosaur {2}{G}{G}, 4/3 (Modern Horizons 2):
//
//	"Vigilance
//	 Enrage — Whenever this creature is dealt damage, proliferate."
//
// One trigger per batch of damage dealt at once (CR 120.2), as every enrage
// card; proliferate asks which permanents and players to add counters to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cc545c68-829f-4c57-94d2-cd1edf2667a7",
		Name:            "Urban Daggertooth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Urban Daggertooth — Enrage: proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
