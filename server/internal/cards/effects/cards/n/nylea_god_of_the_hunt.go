package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nylea, God of the Hunt — Legendary Enchantment Creature — God {3}{G}, 6/6
// (Secret Lair Drop):
//
//	"Indestructible
//	 As long as your devotion to green is less than five, Nylea isn't a
//	 creature.
//	 Other creatures you control have trample.
//	 {3}{G}: Target creature gets +2/+2 until end of turn."
//
// The God clause is the layer-4 devotion static every God shares; the trample
// is a layer-6 grant to the other creatures its controller has; the pump
// targets any creature (CR 115.1) and is fixed on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "43674fdf-496f-42dc-8a0a-5c42a7c7d910",
		Name:            "Nylea, God of the Hunt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{
			effects.GodUnlessDevotion("G"),
			effects.KeywordGrant(effects.OtherCreaturesYouControl, "trample"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: Target creature gets +2/+2 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
			Cost:    effects.ManaCost("{3}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Nylea, God of the Hunt — +2/+2"}.Apply(ctx)
			},
		}},
	})
}
