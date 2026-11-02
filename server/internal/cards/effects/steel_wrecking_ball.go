package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steel Wrecking Ball — Artifact {5} (Marvel's Spider-Man, #4460):
//
//	"When this artifact enters, it deals 5 damage to target creature.
//	 {1}{R}, Discard this card: Destroy target artifact."
//
// The enters trigger is Magmatic Galleon's, aimed at any creature. The
// discard ability works only from the hand (CR 113.6), the way cycling
// does and as Proft's does: the discard is part of the cost, so the card is
// already in the graveyard when the ability resolves, and the destruction
// takes the target chosen as it was activated.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a07e22c-71cd-4723-b07e-0ff02009a878",
		Name:         "Steel Wrecking Ball",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Steel Wrecking Ball — 5 damage to target creature",
				steelWreckingBallDamage),
				TargetCreature("target creature")), ForTargets(DamageToTarget(0, 5))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Discard this card: Destroy target artifact.",
			Cost:    Plus(ManaCost("{1}{R}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  destroyFirstLegalCardTarget,
		}},
	})
}

// steelWreckingBallDamage is "it deals 5 damage to target creature": the
// first target still legal at resolution (CR 608.2b).
func steelWreckingBallDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: 5}.Apply(ctx)
}
