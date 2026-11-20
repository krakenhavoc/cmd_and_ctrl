package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steel Wrecking Ball — Artifact {5} (Marvel's Spider-Man, #4460):
//
//	"When this artifact enters, it deals 5 damage to target creature.
//	 {1}{R}, Discard this card: Destroy target artifact."
//
// The enters trigger shares Magmatic Galleon's body (sourceDealsFiveToTheChosenTarget), aimed at any creature. The
// discard ability works only from the hand (CR 113.6), the way cycling
// does and as Proft's does: the discard is part of the cost, so the card is
// already in the graveyard when the ability resolves, and the destruction
// takes the target chosen as it was activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a07e22c-71cd-4723-b07e-0ff02009a878",
		Name:         "Steel Wrecking Ball",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Steel Wrecking Ball — 5 damage to target creature",
				effects.SourceDealsFiveToTheChosenTarget),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 5))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Discard this card: Destroy target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
