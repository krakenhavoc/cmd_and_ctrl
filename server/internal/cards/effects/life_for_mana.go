package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// life_for_mana.go — the card-side sentence over game/life_for_mana.go
// (ADR 0131, #2531): a PERMANENT's "for each {B} in a cost, you may pay
// 2 life rather than pay that mana", declared on Spec.LifeForMana.
//
// The engine owns the rest: it reads the grant at every mana payment the
// permanent's controller makes, marks each {B} (and the {B} half of a
// hybrid symbol) as payable with life like a printed {B/P}, and leaves
// generic mana, {C} and the price shown alone. The player claims it at
// announcement through `phyrexian_life`; auto-tap never pays it.

// YouMayPayLifeForMana is "For each {<color>} in a cost, you may pay 2
// life rather than pay that mana." — every mana payment the
// permanent's controller makes (K'rrik, Son of Yawgmoth, with "B").
func YouMayPayLifeForMana(color string) []game.LifeForManaStatic {
	return []game.LifeForManaStatic{{
		Label: "For each {" + color + "} in a cost, you may pay 2 life rather than pay that mana.",
		Color: color,
	}}
}
