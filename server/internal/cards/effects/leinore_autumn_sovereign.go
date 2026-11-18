package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leinore, Autumn Sovereign — Legendary Creature — Human Noble {2}{G}{W},
// 0/4 (Midnight Hunt Commander, S82 MID batch 13):
//
//	"Coven — At the beginning of combat on your turn, put a +1/+1 counter on
//	 up to one target creature you control. Then if you control three or more
//	 creatures with different powers, draw a card."
//
// The target is optional ("up to one"): with none chosen the trigger still
// resolves and the draw is still checked. Coven is read after the counter is
// placed (the "Then" sequence, CR 608.2c), through the layers
// (trekB48Coven), so the counter can be what makes a third power. It is not
// an intervening "if": the trigger always goes on the stack. If the one
// target is gone by resolution the ability does not resolve (CR 608.2b), so
// nothing is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd47eedd-57cb-460d-b339-4785f63a5475",
		Name:         "Leinore, Autumn Sovereign",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(
				"Leinore, Autumn Sovereign — put a +1/+1 counter on up to one target creature you control; then if you control three or more creatures with different powers, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					if !trekB48Coven(g, item.Controller, item.SourceCardID) {
						return nil
					}
					return g.DrawNForEffect(item.Controller, 1)
				}),
				TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1)),
		},
	})
}
