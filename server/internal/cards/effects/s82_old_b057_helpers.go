package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b057_helpers.go — shared pieces of S82 OLD batch 057 (#5504).

// CopyTargetedSpellWithNewTargets is the body of "copy target instant or
// sorcery spell. You may choose new targets for the copy" (Dualcaster Mage,
// Naru Meha, Master Wizard): the first target is the spell, copied under
// the ability's controller with the CR 707.10c prompt for new targets. A
// spell that left the stack in response is nothing to copy (CR 608.2b).
func CopyTargetedSpellWithNewTargets(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return CopySpell{
		StackID:          item.Targets[0].ID,
		Controller:       item.Controller,
		ChooseNewTargets: true,
	}.Apply(NewContext(g, item))
}
