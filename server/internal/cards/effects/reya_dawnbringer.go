package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reya Dawnbringer — Legendary Creature — Angel {6}{W}{W}{W}, 4/6:
//
//	"Flying
//	 At the beginning of your upkeep, you may return target creature
//	 card from your graveyard to the battlefield."
//
// The target is chosen as the ability goes on the stack; the "may" is
// answered as it resolves (CR 603.5).
func init() {
	Register(Spec{
		OracleID:        "08d1a5d1-141d-48e3-ac22-c2529191cb13",
		Name:            "Reya Dawnbringer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				AtYourUpkeep("Reya Dawnbringer — return target creature card from your graveyard to the battlefield",
					returnFirstLegalGraveyardTargetToBattlefield),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
				"Reya Dawnbringer — return the creature card to the battlefield?"),
		},
	})
}
