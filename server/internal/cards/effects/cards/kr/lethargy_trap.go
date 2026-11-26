package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lethargy Trap — Instant — Trap {3}{U} (Zendikar, S82 OLD b496):
//
//	"If three or more creatures are attacking, you may pay {U} rather than pay this
//	 spell's mana cost.
//	 Attacking creatures get -3/-0 until end of turn."
//
// Arrow Volley Trap's alternative cost (four there, three here) and Morale's
// "attacking creatures" boost, fixed at resolution (CR 611.2c): a creature that
// starts attacking later is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7356319d-1485-426b-80a7-9700d6684ac9",
		Name:         "Lethargy Trap",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If three or more creatures are attacking, pay {U}",
			ManaCost:  "{U}",
			Condition: func(g *game.Game, _ uuid.UUID) bool { return len(effects.AttackingCreatures(g)) >= 3 },
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: -3, Label: "Lethargy Trap — attacking creatures get -3/-0"}.Apply(ctx)
		},
	})
}
