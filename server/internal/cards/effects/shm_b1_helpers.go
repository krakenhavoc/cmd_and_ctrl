package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shadowmoor, batch 1 (#5591). Append-only.

// shmB1Cohort is the Cohorts' "This creature gets +1/+1 as long as you
// control another <colour> creature.": a layer 7c pump (CR 613.4c) on the
// source alone, while shmB1WhileYouControlAnother holds. `letter` is the
// colour as printed; `label` is the printed line.
func shmB1Cohort(letter, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: shmB1WhileYouControlAnother(letter),
		Apply:     pumpBy(1, 1),
	}
}

// shmB1WhileYouControlAnother admits the source creature while its
// controller controls a creature OTHER than it that is `letter` as it is
// now. The colour is a word of the source's text (CR 612.2), read
// through its lens as g2OtherCreaturesYouControlOfColor reads it.
func shmB1WhileYouControlAnother(letter string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		if !selfCreature(target, g, source) {
			return false
		}
		color := source.TextColor(letter)
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.InstanceID != source.InstanceID && c.Controller == source.Controller &&
				c.IsCreature() && slices.Contains(c.EffectiveColors(), color) {
				return true
			}
		}
		return false
	}
}
