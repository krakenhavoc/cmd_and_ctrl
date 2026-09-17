package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreamdew Entrancer —
//
// "Reach When this creature enters, tap up to one target creature and put
// three stun counters on it. If you control that creature, draw two cards."
func init() {
	Register(Spec{
		OracleID:        "b6b2d63f-5b8c-47ee-8712-58596e0e9e94",
		Name:            "Dreamdew Entrancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Dreamdew Entrancer — tap and stun target creature", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				id := ts[0].ID
				if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: id, Kind: game.CounterStun, N: 3}).Apply(ctx); err != nil {
					return err
				}
				c, ok := g.LookupCardForEffect(id)
				if ok && c.Controller == item.Controller {
					return (DrawCards{Player: item.Controller, N: 2}).Apply(ctx)
				}
				return nil
			}),
			TargetCreature("up to one target creature").WithCount(0, 1),
		)},
	})
}
