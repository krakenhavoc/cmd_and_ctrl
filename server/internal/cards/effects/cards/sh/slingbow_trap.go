package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slingbow Trap — Instant — Trap {3}{G} (Worldwake, S82 OLD b488):
//
//	"If a black creature with flying is attacking, you may pay {G} rather than
//	 pay this spell's mana cost.
//	 Destroy target attacking creature with flying."
//
// The trap clause is an alternative cost (CR 118.9) offered while any attacking creature,
// whoever controls it, is black and has flying as it stands. The target is an attacking
// creature with flying, checked as it is announced and again as the spell resolves
// (CR 608.2b), so one that has left combat or lost flying in response is not destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66ce0ae5-7000-4e4a-891f-f571f0e58cd7",
		Name:         "Slingbow Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature with flying", effects.IsAttackingCreature(), effects.HasKeyword("flying")),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If a black creature with flying is attacking, pay {G}",
			ManaCost:  "{G}",
			Condition: effects.S82OldB488ABlackCreatureWithFlyingIsAttacking,
		}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
