package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wreck Remover — Artifact Creature — Construct {4}, 3/4 (Aetherdrift,
// #4464):
//
//	"Whenever this creature enters or attacks, exile up to one target card
//	 from a graveyard. You gain 1 life.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// One trigger with two conditions (WhenThisEntersOrAttacks), Graveyard
// Trespasser's "up to one target card from a graveyard" clause. With no
// target chosen it still resolves and the life is gained; with a chosen
// target that left its graveyard in response it does nothing
// (CR 608.2b). Cycling is the hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89c89499-23a8-412d-9406-1a7ffc88c64d",
		Name:         "Wreck Remover",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersOrAttacks("Wreck Remover — exile up to one target card from a graveyard, gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						if err := ExileTargetCards(g, item); err != nil {
							return err
						}
						return GainLife{Amount: 1}.Apply(NewContext(g, item))
					}),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
