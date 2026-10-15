package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	enters := WhenThisEnters("Angel of the Ruins — exile up to two target artifacts and/or enchantments",
		exileEveryLegalCardTarget)
	enters.Targets = TargetPermanent("up to two target artifacts and/or enchantments",
		Or(Artifact(), Enchantment())).WithCount(0, 2)
	Register(Spec{
		OracleID:        "14878a5f-7983-4133-a1a7-9d539531e092",
		Name:            "Angel of the Ruins",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
		Activated: []ActivatedAbility{
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
		},
	})
}
