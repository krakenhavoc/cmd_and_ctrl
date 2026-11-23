package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Despark — Instant {W}{B}:
//
//	"Exile target permanent with mana value 4 or greater."
//
// Two mana to exile anything expensive, which in Commander is most of
// what matters — and exile rather than destroy means indestructible
// bombs are fair game. The mana-value floor is the cost: it cannot
// answer a cheap threat or a land (mana value 0).
//
// ManaValueGE counts {X} as 0 on the battlefield (CR 202.3e) — a
// creature cast for X=8 is NOT a legal Despark target unless its
// printed cost is otherwise 4+, exactly as in paper.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd16434d-55ea-4c5a-a9ef-752971a4af16",
		Name:         "Despark",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: exiles a permanent
		Targets:      effects.TargetPermanent("target permanent with mana value 4 or greater", effects.ManaValueGE(4)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
