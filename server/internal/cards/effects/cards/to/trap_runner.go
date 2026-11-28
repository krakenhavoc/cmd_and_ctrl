package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trap Runner — Creature — Human Soldier {2}{W}{W}, 2/3 (Mercadian Masques,
// S82 OLD b681, #5619):
//
//	"{T}: Target unblocked attacking creature becomes blocked. Activate only
//	 during combat after blockers are declared. (This ability works on
//	 creatures that can't be blocked.)"
//
// Curtain of Light's effect (Game.BecomeBlockedForEffect, CR 509.1h) as an
// activated ability: the target must be an attacking creature no creature is
// blocking, checked at activation and again on resolution (CR 608.2b), and it
// becomes blocked whatever its evasion. The window is the declare blockers
// step and the combat steps after it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "730fbd6a-096c-4d25-9372-c05f90088f83",
		Name:         "Trap Runner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Target unblocked attacking creature becomes blocked. Activate only during combat after blockers are declared.",
			Purpose:   game.Purpose{Answers: game.AnswerRestrict},
			Cost:      effects.TapCost(),
			Targets:   effects.TargetCreature("target unblocked attacking creature", effects.OldB585UnblockedAttacker()),
			Condition: effects.OldB681AfterBlockersAreDeclared,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					g.BecomeBlockedForEffect([]uuid.UUID{id})
				}
				return nil
			},
		}},
	})
}
