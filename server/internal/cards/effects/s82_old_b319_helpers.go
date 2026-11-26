package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB319TapTheTargetCreature is the body of "tap target creature" for an
// ability whose target clause is set on the trigger (Grasp of the
// Hieromancer, Web-Shooters): the first legal target is tapped, and a
// target that is gone or illegal as it resolves does nothing (CR 608.2b).
func OldB319TapTheTargetCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return TapTarget{Target: target}.Apply(ctx)
}
