package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pitfall Trap — Instant — Trap {2}{W} (Zendikar, S82 OLD b495):
//
//	"If exactly one creature is attacking, you may pay {W} rather than pay this
//	 spell's mana cost.
//	 Destroy target attacking creature without flying."
//
// Arrow Volley Trap's alternative cost, read over the attacking creatures of the
// table as the spell is cast; the target must be attacking and not have flying
// (through the layered view) as it is chosen and again as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26e75f0b-3a02-4989-8429-b6276e98ac33",
		Name:         "Pitfall Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature without flying", effects.IsAttackingCreature(), effects.WithoutKeyword("flying")),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If exactly one creature is attacking, pay {W}",
			ManaCost:  "{W}",
			Condition: effects.Old495ExactlyOneCreatureIsAttacking,
		}},
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
