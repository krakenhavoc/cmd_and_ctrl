package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reya Dawnbringer — Legendary Creature — Angel {6}{W}{W}{W}, 4/6:
//
//	"Flying
//	 At the beginning of your upkeep, you may return target creature
//	 card from your graveyard to the battlefield."
//
// The target is chosen as the ability goes on the stack; the "may" is
// answered as it resolves (CR 603.5).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "08d1a5d1-141d-48e3-ac22-c2529191cb13",
		Name:            "Reya Dawnbringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.AtYourUpkeep("Reya Dawnbringer — return target creature card from your graveyard to the battlefield",
					effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())),
				"Reya Dawnbringer — return the creature card to the battlefield?"),
		},
	})
}
