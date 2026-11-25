package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quicksmith Rebel — Creature — Human Artificer {3}{R}, 3/2 (Aether Revolt, S82
// older sets, #5536):
//
//	"When this creature enters, target artifact you control gains "{T}: This
//	 artifact deals 2 damage to any target" for as long as you control this
//	 creature."
//
// Quicksmith Spy's grant with a ping: the artifact, not the Rebel, is the
// source of the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d69594fc-9250-41e7-a170-d4705615312c",
		Name:         "Quicksmith Rebel",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  quicksmithRebelPingGrant,
			Text: "{T}: This artifact deals 2 damage to any target.",
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This artifact deals 2 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
			}},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Quicksmith Rebel — target artifact you control gains \"{T}: This artifact deals 2 damage to any target\" for as long as you control this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					d, ok := effects.DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
					if !ok {
						return nil
					}
					return effects.GrantAbilitiesFor{
						Target:   id,
						Keys:     []string{quicksmithRebelPingGrant},
						Duration: d,
						Label:    "Quicksmith Rebel — {T}: 2 damage to any target",
					}.Apply(ctx)
				}),
				effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl())),
		},
	})
}

// quicksmithRebelPingGrant is the granted bundle's key, an on-disk identity:
// never renamed.
const quicksmithRebelPingGrant = "quicksmith-rebel/ping"
