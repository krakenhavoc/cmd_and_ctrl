package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exhumer Thrull — Creature — Thrull {5}{B}, 3/3 (Guildpact):
//
//	"Haunt (When this creature dies, exile it haunting target creature.)
//	 When this creature enters or the creature it haunts dies, return
//	 target creature card from your graveyard to your hand."
//
// Haunt is CR 702.55 (#6083): one ability with two conditions. "Your
// graveyard" is the card owner's from exile (CR 108.4). A card that left
// the graveyard in response is no longer a legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd2ae326-1a74-4b67-9718-92b6a209303b",
		Name:         "Exhumer Thrull",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Haunt("Exhumer Thrull"),
			effects.Targeting(
				effects.WhenThisEntersOrTheCreatureItHauntsDies("Exhumer Thrull — return target creature card from your graveyard to your hand", effects.ReturnFirstLegalGraveyardTargetToHand),
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
			),
		},
	})
}
