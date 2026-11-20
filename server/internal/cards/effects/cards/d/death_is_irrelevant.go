package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death Is Irrelevant — Enchantment {6}{B}{B}:
//
//	"Artifact creatures you control have indestructible.
//	 At the beginning of your end step, assimilate target creature card
//	 from a graveyard. (Put it onto the battlefield under your control
//	 with a +1/+1 counter. It's a Borg artifact creature and loses all
//	 other creature types.)"
//
// A layer 6 grant over the artifact creatures its controller controls,
// and an end-step trigger that targets a creature card in any graveyard
// and assimilates it (trekB11Assimilate): it enters under this
// enchantment's controller with a +1/+1 counter as a Borg artifact
// creature, keeping its other card types (CR 205.1b), so it is an
// artifact creature the first ability makes indestructible. The type
// change lasts as long as that object stays on the battlefield. A target
// that has left its graveyard by resolution makes the trigger do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aeec05f1-43ea-4012-82d8-cded5494535d",
		Name:         "Death Is Irrelevant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
		}, "indestructible")},
		Triggered: []game.TriggeredAbility{effects.TriggerWithPurpose(effects.Targeting(
			effects.AtYourEndStep("Death Is Irrelevant — assimilate target creature card from a graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					card := effects.FirstLegalCardTarget(ctx)
					if card == uuid.Nil {
						return nil
					}
					return effects.TrekB11Assimilate(ctx, card, "Death Is Irrelevant — assimilated: a Borg artifact creature")
				}),
			effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
		), effects.ForTargets(effects.Reanimates(0)))},
	})
}
