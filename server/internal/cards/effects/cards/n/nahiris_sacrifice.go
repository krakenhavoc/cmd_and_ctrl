package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nahiri's Sacrifice — Sorcery {1}{R} (Phyrexia: All Will Be One):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature with mana value X.
//	 Nahiri's Sacrifice deals X damage divided as you choose among any number
//	 of target creatures."
//
// X is announced with the cast (CR 107.3a, 601.2b, as for Sidisi, Regent of
// the Mire's sacrifice cost) and the permanent sacrificed must have exactly
// that mana value (CR 202.3: an {X} in a permanent's cost counts as 0), which
// the additional cost's validator checks against the announced X. The same X
// is the amount divided as the spell is cast (CR 601.2d, Rolling Thunder's
// DivideX), among any number of target creatures, each getting at least 1; at
// X = 0 there is nothing to divide and the spell targets nothing. The damage
// is dealt as it resolves to each target still legal (CR 608.2b).
//
// No simplification.
func init() {
	cost := effects.SacrificeCost("an artifact or creature with mana value X", effects.Or(effects.Artifact(), effects.Creature()))
	cost.Sacrifice.ManaValueEqualsX = true
	effects.Register(effects.Spec{
		OracleID:       "a84cc828-8d1b-4e61-a13f-60124e2ee35f",
		Name:           "Nahiri's Sacrifice",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: cost,
		Targets:        effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.DivideX()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
