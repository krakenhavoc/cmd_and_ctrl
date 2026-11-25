package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tentative Connection — Sorcery {3}{R} (Ikoria: Lair of Behemoths, #5512):
//
//	"This spell costs {3} less to cast if you control a creature with
//	 menace.
//	 Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn."
//
// The discount reads the caster's battlefield as the cost is worked out
// (Winged Words' shape); the rest is the shared threaten line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d74de74f-be0f-4809-8b7a-1bd58ee85551",
		Name:         "Tentative Connection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(3, "This spell costs {3} less to cast if you control a creature with menace.",
				effects.YouControlA(effects.And(effects.Creature(), effects.HasKeyword("menace")))),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.ThreatenAndGrant(ctx, id, "Tentative Connection", "haste")
		},
	})
}
