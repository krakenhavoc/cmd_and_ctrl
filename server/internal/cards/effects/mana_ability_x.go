package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_ability_x.go — #6125: X in a mana ability's cost. Wizard's
// Rockets' "{X}, {T}, Sacrifice this artifact: Add X mana in any
// combination of colors." and Springjack Pasture's "{T}, Sacrifice X
// Goats: Add X mana of any one color. You gain X life." The activator
// announces X with the activation (`x_value`, CR 107.3a, 602.2b via
// 605.3a), and the row reads it back as PaidCost.XValue in
// ProducedForPaid (and in its Rider, through ManaAbility.RiderForPaid).

// XManaAnyCombination is "Add X mana in any combination of colors" (#6125,
// Wizard's Rockets): X mana, each of a colour the activator picks.
func XManaAnyCombination(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	return game.ManaOutput{Amount: game.ManaAmount{Options: append([]string(nil), game.AllColors...), N: paid.XValue}}
}

// checkManaAbilityX is the boot-time rule for X in a mana ability's cost
// (#6125): one X claimant (an {X} in the mana, or a sacrifice-X clause,
// never both), and a ProducedForPaid to read it, since a fixed output
// would ignore the X the player paid.
func checkManaAbilityX(name string, i int, ma ManaAbility) {
	manaX := false
	if ma.Cost.Mana != "" {
		if cost, err := game.ParseCost(ma.Cost.Mana); err == nil && cost.XSlots > 0 {
			manaX = true
		}
	}
	sacX := game.SacrificeCountFromX(ma.Cost.SacrificeOther)
	switch {
	case ma.GainLifeX && !manaX && !sacX:
		panic(fmt.Sprintf("effects.Register: %q mana ability %d gains X life but its cost has no X", name, i))
	case manaX && sacX:
		panic(fmt.Sprintf("effects.Register: %q mana ability %d has {X} in its mana AND sacrifices X permanents — one announced X cannot pay both", name, i))
	case (manaX || sacX) && ma.ProducedForPaid == nil:
		panic(fmt.Sprintf("effects.Register: %q mana ability %d announces X but nothing reads it — declare ProducedForPaid over PaidCost.XValue", name, i))
	}
}

// XManaOfAnyOneColor is "Add X mana of any one color" (#6125, Springjack
// Pasture): X mana, all of the one colour the activator picks.
func XManaOfAnyOneColor(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	return game.ManaOutput{Amount: game.ManaAmount{Options: append([]string(nil), game.AllColors...), N: paid.XValue, OneColor: true}}
}
