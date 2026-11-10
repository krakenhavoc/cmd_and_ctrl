package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdrB25DestroyChosenTarget is "Destroy target <permanent>" for a spell
// with one target clause: the first target, if it is still a card as the
// spell resolves (CR 608.2b; the single-target verb honours indestructible).
func cmdrB25DestroyChosenTarget(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
}
