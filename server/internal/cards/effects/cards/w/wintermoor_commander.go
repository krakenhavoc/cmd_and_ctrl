package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wintermoor Commander — Creature — Human Knight {W}{B}, 2/*:
//
//	"Deathtouch
//	 Wintermoor Commander's toughness is equal to the number of Knights you
//	 control.
//	 Whenever this creature attacks, another target Knight you control gains
//	 indestructible until end of turn."
//
// The toughness is a Layer 7a CDA (Daxos' shape) counting the Knights its
// controller has through the effective view, itself included (it is a
// Knight), so it is at least 1 while it is on the battlefield. The attack
// trigger is a targeted keyword grant; with no other Knight there is no
// legal target and the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b1e4cf31-9efe-4aca-a3ba-cc2681109a2f",
		Name:            "Wintermoor Commander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{{
			Label:     "Toughness is equal to the number of Knights you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Toughness = effects.CountPermanentsYouControl(g, source, func(o *game.Card) bool {
					return o.IsCreature() && o.HasSubtypeWord("Knight")
				})
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Wintermoor Commander — another target Knight you control gains indestructible until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"},
									Label: "Wintermoor Commander — another target Knight you control gains indestructible"}.Apply(ctx)
							}
						}
						return nil
					}),
				effects.Another(effects.TargetCreature("another target Knight you control", effects.Subtype("Knight"), effects.YouControl())),
			),
		},
	})
}
