package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rubblehulk — Creature — Elemental {4}{R}{G}, */*:
//
//	"Rubblehulk's power and toughness are each equal to the number of lands
//	 you control.
//	 Bloodrush — {1}{R}{G}, Discard this card: Target attacking creature gets
//	 +X/+X until end of turn, where X is the number of lands you control."
//
// The first line is Molimo's lands-controlled characteristic-defining
// ability. Bloodrush is an activated ability that works from the hand with
// the discard as part of its cost (GtcBloodrush's shape); X is counted for
// the ability's controller as it resolves (CR 608.2h) and a target that
// stopped attacking is illegal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "063960e0-3f9c-4697-ba5e-81c56ac1b5d2",
		Name:         "Rubblehulk",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.LandsYouControlPT("Rubblehulk's power and toughness are each equal to the number of lands you control."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Bloodrush — {1}{R}{G}, Discard this card: Target attacking creature gets +X/+X until end of turn, where X is the number of lands you control.",
			Cost:    game.AbilityCost{Mana: "{1}{R}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				legal := effects.LegalTargetCards(item, g)
				if len(legal) == 0 {
					return nil
				}
				x := effects.B10LandsControlled(g, item.Controller)
				return effects.BoostUntilEOT{Target: legal[0], Power: x, Toughness: x, Label: "Rubblehulk — bloodrush"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
