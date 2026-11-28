package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b775_helpers.go — shared piece of S82 OLD batch 775 (#5629).

// OldB775ReturnTargetCreatureCardToBattlefield is "return target creature card
// from your graveyard to the battlefield" as an ability's whole body (Strands of
// Night): the first target still legal as it resolves comes back under the
// controller's control (CR 608.2b).
func OldB775ReturnTargetCreatureCardToBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
		}
	}
	return nil
}
