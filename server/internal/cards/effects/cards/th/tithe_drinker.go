package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tithe Drinker — Creature — Vampire {W}{B}, 2/1:
//
//	"Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)
//	 Extort (Whenever you cast a spell, you may pay {W/B}. If you do,
//	 each opponent loses 1 life and you gain that much life.)"
//
// Lifelink is a printed keyword (CR 702.15); extort is the shared Extort
// trigger (CR 702.101), Crypt Ghast's: each spell its controller casts
// offers a {W/B} payment that drains each opponent for 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "06b1fd96-a3ad-4ebd-9ea4-b3cfb4c9c43d",
		Name:            "Tithe Drinker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered:       []game.TriggeredAbility{effects.Extort("Tithe Drinker")},
	})
}
