package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murasa Sproutling — Creature — Plant Elemental {2}{G}, 3/3 (Zendikar
// Rising, S82 OLD #5508):
//
//	"Kicker {1}{G} (You may pay an additional {1}{G} as you cast this
//	 spell.)
//	 When this creature enters, if it was kicked, return target card with a
//	 kicker ability from your graveyard to your hand."
//
// A card has a kicker ability if its catalog row declares a kicker cost.
// The target is chosen as the trigger goes on the stack; a card that left
// the graveyard by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "dfa8c557-f2f5-4763-85c7-e2ec14d9e6b6",
		Name:          "Murasa Sproutling",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Murasa Sproutling — kicked: return target card with a kicker ability from your graveyard to your hand",
					effects.S82Mom16ReturnTargetToHand),
				effects.TargetCardInGraveyard("target card with a kicker ability from your graveyard", effects.YouOwn(),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return effects.S82OldB090CardHasKicker(c) })),
		},
	})
}
