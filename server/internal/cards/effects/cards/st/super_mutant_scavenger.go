package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Super Mutant Scavenger — Creature — Mutant Warrior {4}{G}, 5/5
// (Fallout, #5469):
//
//	"Trample
//	 When this creature enters or dies, return up to one target Aura or
//	 Equipment card from your graveyard to your hand."
//
// Hunting Moa's one ability with two events (WhenThisEntersOrDies) over
// Forum Filibuster's "up to one target Aura or Equipment card from your
// graveyard" clause, resolved by the shared
// returnFirstLegalGraveyardTargetToHand. The card is returned only if it is still in the
// graveyard as the trigger resolves (CR 608.2b); choosing no target is
// legal and does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2972b9c8-d9b3-4bb3-b9c5-9aa70fb1ac71",
		Name:            "Super Mutant Scavenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEntersOrDies("Super Mutant Scavenger — return up to one Aura or Equipment card from your graveyard to your hand",
				effects.ReturnFirstLegalGraveyardTargetToHand),
				effects.TargetCardInGraveyard("up to one target Aura or Equipment card from your graveyard",
					effects.YouOwn(), effects.Or(effects.HasSubtype("Aura"), effects.HasSubtype("Equipment"))).WithCount(0, 1)),
		},
	})
}
