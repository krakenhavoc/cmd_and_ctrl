package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// CopyTheFirstLegalTargetSpell is "Copy target spell. You may choose new
// targets for the copy." for the first target still legal as the ability
// resolves (CR 707.10c): Izzet Guildmage's and League Guildmage's body.
func CopyTheFirstLegalTargetSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return CopySpell{StackID: t.ID, Controller: item.Controller, ChooseNewTargets: true}.Apply(ctx)
	}
	return nil
}
