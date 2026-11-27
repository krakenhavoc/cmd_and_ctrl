package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sustenance — Enchantment {1}{G} (Mercadian Masques, S82 OLD, #5619):
//
//	"{1}, Sacrifice a land: Target creature gets +1/+1 until end of turn."
//
// The land is a cost, any land its controller controls, tapped or not, paid on
// activation, so the bonus is given even if the target is gone only when the
// target is still legal as it resolves (CR 608.2b). Activated as often as the
// controller has mana and lands to give up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d32a32d2-203d-4be1-8a33-e037747053c7",
		Name:         "Sustenance",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a land: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a land", effects.Land())}),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Sustenance — +1/+1 until end of turn"}.Apply(ctx)
			},
		}},
	})
}
