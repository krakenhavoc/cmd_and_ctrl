package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syphon Essence — Instant {2}{U} (Innistrad: Crimson Vow):
//
//	"Counter target creature or planeswalker spell. Create a Blood token."
//
// If the spell is gone on resolution the target is illegal and this does
// nothing, Blood token included (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "e6e7b5ba-ea10-4f53-8afc-045aa1906c6b",
		Name:         "Syphon Essence",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature or planeswalker spell", Or(Creature(), Planeswalker())),
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return s82Vow03CreateBlood(1).Apply(ctx)
		},
	})
}
