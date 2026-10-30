package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrificeATreasureToPlayTheTop is "Sacrifice a Treasure: Exile the top
// card of your library. You may play that card this turn." — Professional
// Face-Breaker's and Heroes for Hire's activated ability, which differ only
// in the label they print.
func sacrificeATreasureToPlayTheTop(label string) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerValue, Impulse: 1},
		Cost:    game.AbilityCost{SacrificeOther: sacrificeSpec("a Treasure", isTreasure)},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return ExileTopWithPermission{
				From:    item.Controller,
				GrantTo: item.Controller,
				N:       1,
			}.Apply(NewContext(g, item))
		},
	}
}

// msc12FlyingTheModesTarget is "target creature gains flying until end of
// turn" as a modal bullet's body (Graviton's first bullet), on this
// occurrence's own target while it is still legal (CR 608.2b).
func msc12FlyingTheModesTarget(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Graviton, Fundamental Force — flying until end of turn"}.Apply(ctx)
}

// msc12CreaturesWithPowerAtLeast counts the creatures `controller` controls
// whose power is at least n (Giant-Man's mana).
func msc12CreaturesWithPowerAtLeast(g *game.Game, controller uuid.UUID, n int) int {
	count := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.CurrentPower() >= n {
			count++
		}
	}
	return count
}
