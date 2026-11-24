package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kamahl, Heart of Krosa — Legendary Creature — Human Druid {6}{G}{G}, 5/5
// (Commander Legends, #5507):
//
//	"At the beginning of combat on your turn, creatures you control get +3/+3
//	 and gain trample until end of turn.
//	 {1}{G}: Until end of turn, target land you control becomes a 1/1
//	 Elemental creature with vigilance, indestructible, and haste. It's still a
//	 land.
//	 Partner"
//
// The pump and the trample are one until-end-of-turn effect over the
// creatures you control as the trigger resolves, the set locked then (CR
// 611.2c). The animation is Obuun's land animation (ScopedEffectFor): the
// land gains the Creature type and the Elemental subtype, its base power and
// toughness are set to 1/1 (layer 7b) and it gains the three keywords until
// end of turn, still a land; it does nothing if the land has left. Partner is
// a deck-construction rule and needs nothing on the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "343fd080-c2c2-4dcd-84dd-c065addbaebe",
		Name:         "Kamahl, Heart of Krosa",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Kamahl, Heart of Krosa — creatures you control get +3/+3 and gain trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UntilEndOfTurn(effects.NewContext(g, item), uuid.Nil, effects.And(effects.Creature(), effects.YouControl()),
						"Kamahl, Heart of Krosa — +3/+3 and trample", game.ModifyPTMod(3, 3), game.AddKeywordsMod("trample"))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Until end of turn, target land you control becomes a 1/1 Elemental creature with vigilance, indestructible, and haste. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0))},
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				land := effects.FirstLegalCardTarget(ctx)
				if land == uuid.Nil {
					return nil
				}
				return effects.ScopedEffectFor{
					Target: land,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
						game.SetBasePowerMod(1),
						game.SetBaseToughnessMod(1),
						game.AddKeywordsMod("vigilance", "indestructible", "haste"),
					},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Kamahl, Heart of Krosa — a 1/1 Elemental creature with vigilance, indestructible and haste until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
