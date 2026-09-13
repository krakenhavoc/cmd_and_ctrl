package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Despark — Instant {W}{B}:
//
//	"Exile target permanent with mana value 4 or greater."
//
// Two mana to exile anything expensive, which in Commander is most of
// what matters — and exile rather than destroy means indestructible
// bombs are fair game. The mana-value floor is the cost: it cannot
// answer a cheap threat or a land (mana value 0).
//
// ManaValueGE reads the printed cost, so {X} counts as 0 on the
// battlefield (CR 202.3b) — a creature cast for X=8 is NOT a legal
// Despark target unless its printed cost is otherwise 4+, exactly as
// in paper.
func init() {
	Register(Spec{
		OracleID:     "bd16434d-55ea-4c5a-a9ef-752971a4af16",
		Name:         "Despark",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent with mana value 4 or greater", ManaValueGE(4)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
