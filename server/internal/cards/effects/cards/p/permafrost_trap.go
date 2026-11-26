package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Permafrost Trap — Instant — Trap {2}{U}{U} (Worldwake, S82 OLD b488):
//
//	"If an opponent had a green creature enter the battlefield under their
//	 control this turn, you may pay {U} rather than pay this spell's mana cost.
//	 Tap up to two target creatures. Those creatures don't untap during their
//	 controller's next untap step."
//
// The trap clause is an alternative cost (CR 118.9) offered while one opponent's entry tally
// for the turn holds a green creature (the per-colour creature tally, recorded as each
// entered: a token counts, one that has since left or changed colour still counts). The
// body is Decision Paralysis': tap each target still legal as it resolves, and keep it
// tapped through its controller's next untap step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f037b319-e559-42ad-bcfe-d43d07ce05e7",
		Name:         "Permafrost Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent had a green creature enter the battlefield under their control this turn, pay {U}",
			ManaCost:  "{U}",
			Condition: effects.S82OldB488AnOpponentHadAGreenCreatureEnter,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B751TapFreezeTargets(ctx, uuid.Nil, "Permafrost Trap")
		},
	})
}
