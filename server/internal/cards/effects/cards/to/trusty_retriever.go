package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trusty Retriever — Creature — Dog {3}{W}, 2/3:
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • Return target artifact or enchantment card from your graveyard
//	   to your hand."
//
// A modal enters trigger (CR 700.2): the second bullet carries its own
// target clause, so it can be chosen only with such a card in your
// graveyard, and a card that left the graveyard in response is no
// longer legal (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Trusty Retriever — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on this creature.", nil, effects.PutAPlusOneCounterOnThisCreatureMode),
		effects.ModeDoing("Return target artifact or enchantment card from your graveyard to your hand.",
			effects.TargetCardInGraveyard("target artifact or enchantment card in your graveyard", effects.YouOwn(), effects.Or(effects.Artifact(), effects.Enchantment())),
			effects.ReturnTheModesGraveyardTargetToHand),
	)
	effects.Register(effects.Spec{
		OracleID:     "06a2a50d-f54f-4eb7-9cf8-ce37d802794a",
		Name:         "Trusty Retriever",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
