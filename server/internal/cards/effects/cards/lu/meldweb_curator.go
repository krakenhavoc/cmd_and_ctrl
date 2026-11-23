package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "26d6a089-b867-4b89-be8e-2ad11bff0d64",
		Name:         "Meldweb Curator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Meldweb Curator — put up to one target instant or sorcery card from your graveyard on top of your library", effects.PutChosenTargetOnTopOfLibrary),
				effects.TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()).WithCount(0, 1)),
		},
	})
}
