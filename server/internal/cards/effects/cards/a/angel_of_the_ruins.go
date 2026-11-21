package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel of the Ruins — Artifact Creature — Angel {5}{W}{W}, 5/7:
//
//	"Flying
//	 When this creature enters, exile up to two target artifacts and/or
//	 enchantments.
//	 Plainscycling {2}"
//
// The entry trigger takes up to two targets (CR 601.2c / 115.4) and
// exiles each one still legal as it resolves (CR 608.2b). Plainscycling
// is the shared Typecycling constructor with "a Plains card" (any land
// with the Plains type, not only a basic one).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Angel of the Ruins — exile up to two target artifacts and/or enchantments",
		effects.ExileEveryLegalCardTarget)
	enters.Targets = effects.TargetPermanent("up to two target artifacts and/or enchantments",
		effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 2)
	effects.Register(effects.Spec{
		OracleID:        "14878a5f-7983-4133-a1a7-9d539531e092",
		Name:            "Angel of the Ruins",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
	})
}
