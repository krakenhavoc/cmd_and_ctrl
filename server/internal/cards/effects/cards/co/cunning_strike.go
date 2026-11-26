package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cunning Strike — {3}{U}{R} Instant (Fate Reforged):
//
//	"Cunning Strike deals 2 damage to target creature and 2 damage to target
//	 player or planeswalker.
//	 Draw a card."
//
// Two target clauses, each re-checked as the spell resolves (CR 608.2b): one
// that has become illegal takes no damage and the other still does. The card
// is drawn whatever happens to the targets, unless both are illegal and the
// spell does not resolve (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a01491b7-89c8-4d6e-a516-431440ff3e88",
		Name:         "Cunning Strike",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPlayerOrPlaneswalker(),
		),
		Purpose: game.Purpose{Draws: 1, Targets: game.ForTargets(effects.DamageToTarget(0, 2), effects.DamageToTarget(1, 2))},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.S82OldB338DamageEachClauseTarget(ctx, 2, 2); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
