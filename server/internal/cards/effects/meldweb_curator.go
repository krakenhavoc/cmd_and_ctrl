package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meldweb Curator — Creature — Phyrexian Wizard {3}{U}, 3/4:
//
//	"When this creature enters, put up to one target instant or sorcery
//	 card from your graveyard on top of your library."
//
// Mystic Sanctuary's put-on-top, with an "up to one" target that is not
// optional: choosing no target does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26d6a089-b867-4b89-be8e-2ad11bff0d64",
		Name:         "Meldweb Curator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Meldweb Curator — put up to one target instant or sorcery card from your graveyard on top of your library", PutChosenTargetOnTopOfLibrary),
				TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard", Or(Instant(), Sorcery()), YouOwn()).WithCount(0, 1)),
		},
	})
}
