package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Biblioplex Assistant — Artifact Creature — Gargoyle {4}, 2/1
// (Strixhaven: School of Mages, #5502):
//
//	"Flying
//	 When this creature enters, put up to one target instant or sorcery
//	 card from your graveyard on top of your library."
//
// Flying is the printed keyword; the enters trigger is Meldweb Curator's
// put-on-top with an "up to one" target: choosing no target does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3dadb654-6191-4ad3-8315-46f5613fbc01",
		Name:            "Biblioplex Assistant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Biblioplex Assistant — put up to one target instant or sorcery card from your graveyard on top of your library", PutChosenTargetOnTopOfLibrary),
				TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard", Or(Instant(), Sorcery()), YouOwn()).WithCount(0, 1)),
		},
	})
}
