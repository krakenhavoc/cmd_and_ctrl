package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Borg Queen, Perfection Manifest — Legendary Artifact Creature — Borg
// Noble {4}{B}{B}, 1/4 (Star Trek Commander, #4974):
//
//	"Artifact creatures you control get +2/+0.
//	 When Borg Queen enters, assimilate target creature card from an
//	 opponent's graveyard. (Put it onto the battlefield under your
//	 control with a +1/+1 counter. It's a Borg artifact creature and
//	 loses all other creature types.)"
//
// A layer 7c anthem over the artifact creatures her controller controls,
// read on post-layer types, so the Queen pumps herself and anything she
// assimilates. The enters trigger targets a creature card in a graveyard
// an opponent owns and assimilates it (trekB11Assimilate): it enters
// under her controller with a +1/+1 counter as a Borg artifact creature
// that keeps its other card types (CR 205.1b). A target that has left
// its graveyard by resolution makes the trigger do nothing (CR 608.2b).
//
// No simplification.
func init() {
	anthem := b16Anthem(trekB47ArtifactCreaturesYouControl, 2, 0)
	anthem.Label = "Artifact creatures you control get +2/+0."
	Register(Spec{
		OracleID:     "f9b46a1a-474f-4fac-8d71-131c1720e4c0",
		Name:         "Borg Queen, Perfection Manifest",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{anthem},
		Triggered: []game.TriggeredAbility{TriggerWithPurpose(Targeting(
			WhenThisEnters("Borg Queen, Perfection Manifest — assimilate target creature card from an opponent's graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					card := firstLegalCardTarget(ctx)
					if card == uuid.Nil {
						return nil
					}
					return trekB11Assimilate(ctx, card, "Borg Queen, Perfection Manifest — assimilated: a Borg artifact creature")
				}),
			TargetCardInGraveyard("target creature card from an opponent's graveyard", Creature(), Not(YouOwn())),
		), ForTargets(Reanimates(0)))},
	})
}
