package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Electro's Bolt — Sorcery {2}{R}:
//
//	"Electro's Bolt deals 4 damage to target creature.
//	 Mayhem {1}{R} (You may cast this card from your graveyard for
//	 {1}{R} if you discarded it this turn. Timing rules still apply.)"
//
// Mayhem (CR 702.187, #4628): the graveyard opening and the cheaper
// alternative cost. Timing rules still apply, so it is a sorcery from
// the graveyard too. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e9222f89-e495-4502-bb0c-f5e802686b01",
		Name:             "Electro's Bolt",
		Purpose:          effects.ForTargets(effects.DamageToTarget(0, 4)),
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		CastableZonesIf:  []game.ConditionalCastZone{effects.MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{effects.Mayhem("{1}{R}")},
		OnResolve:        effects.DamageToFirstTargetStep(4),
	})
}
