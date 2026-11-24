package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ManaValueNoGreaterThanTarget is "<permanent> with mana value equal to or
// less than <that target>'s": the clause's candidates must have a mana value no
// greater than the earlier target in `slot` (Puca's Mischief's "target
// nonland permanent an opponent controls with equal or lesser mana value").
// Judged as the ability is put on the stack (CR 601.2c) and again as it
// resolves (CR 608.2b); the anchor is read for the mana value it has, wherever
// it is.
func ManaValueNoGreaterThanTarget(spec *game.TargetSpec, slot int) *game.TargetSpec {
	spec.Relation = &game.TargetRelation{Slot: slot, Kind: game.TargetManaValueNoGreater}
	return spec
}

// AddTheEnchantmentTypeOnce is "<permanent> is an enchantment in addition to
// its other types" (Enchanted Evening's layer 4 body): idempotent, so a
// permanent that already is an enchantment is unchanged.
func AddTheEnchantmentTypeOnce(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
	for _, t := range c.Types {
		if t == "Enchantment" {
			return
		}
	}
	c.Types = append(c.Types, "Enchantment")
}
