package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "e9222f89-e495-4502-bb0c-f5e802686b01",
		Name:             "Electro's Bolt",
		Purpose:          ForTargets(DamageToTarget(0, 4)),
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{1}{R}")},
		OnResolve:        damageToFirstTarget(4),
	})
}
