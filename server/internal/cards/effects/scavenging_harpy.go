package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scavenging Harpy — Creature — Harpy {2}{B}, 2/1 (Theros Beyond Death, #5514):
//
//	"Flying
//	 When this creature enters, exile target card from an opponent's
//	 graveyard."
//
// Flying is the printed keyword; the enters trigger exiles one card from
// a graveyard its controller doesn't own (Carrion Locust's clause,
// Ambush Wolf's exile).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "94ea649c-9d4f-4882-a947-e2df4993b996",
		Name:            "Scavenging Harpy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Scavenging Harpy — exile target card from an opponent's graveyard", b27ExileChosenTarget),
				TargetCardInGraveyard("target card from an opponent's graveyard", Not(YouOwn()))),
		},
	})
}
