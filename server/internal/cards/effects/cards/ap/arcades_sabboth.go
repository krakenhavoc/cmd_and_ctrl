package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcades Sabboth — Legendary Creature — Elder Dragon {2}{G}{G}{W}{W}{U}{U},
// 7/7 (Dominaria Remastered, #5948):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice Arcades Sabboth unless you
//	 pay {G}{W}{U}.
//	 Each untapped creature you control gets +0/+2 as long as it's not
//	 attacking.
//	 {W}: Arcades Sabboth gets +0/+1 until end of turn."
//
// The upkeep tax is the shared UpkeepSacrificeUnlessYouPay. The static is
// a layer 7c modification of every untapped, non-attacking creature its
// controller controls, Arcades Sabboth included (it is "each untapped
// creature you control"); it is read live, so a creature that taps or is
// declared as an attacker loses the bonus at once. A vigilant attacker is
// untapped but attacking, and gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ead94c5-8447-4f20-87ec-efdccde689fc",
		Name:            "Arcades Sabboth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.UpkeepSacrificeUnlessYouPay("Arcades Sabboth", "{G}{W}{U}"),
		},
		Static: []game.StaticAbility{{
			Label:                    "Each untapped creature you control gets +0/+2 as long as it's not attacking.",
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7C_Modify,
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					!target.Tapped && target.AttackingTarget == uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Toughness += 2
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    effects.ManaCost("{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Toughness: 1, Label: "Arcades Sabboth — +0/+1"}.Apply(ctx)
			},
		}},
	})
}
