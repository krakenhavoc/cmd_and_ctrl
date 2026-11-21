package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gaddock Teeg — Legendary Creature — Kithkin Advisor {G}{W}, 2/2 (Lorwyn,
// #5593):
//
//	"Noncreature spells with mana value 4 or greater can't be cast.
//	 Noncreature spells with {X} in their mana costs can't be cast."
//
// Aether Storm's symmetrical "can't be cast" twice, binding every player,
// Teeg's controller included (CR 101.2). The first line reads the mana
// value of the spell being cast; the second reads its printed mana cost, so
// a noncreature spell with {X} is refused whatever X would be, while a
// creature spell is never refused by either line. Abilities are not spells,
// so neither line stops one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98dd6d43-cc56-4656-9e0c-240d0679b404",
		Name:         "Gaddock Teeg",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.PlayersCantCast("Noncreature spells with mana value 4 or greater can't be cast.",
				effects.And(effects.Not(effects.Creature()), effects.ManaValueGE(4))),
			effects.PlayersCantCast("Noncreature spells with {X} in their mana costs can't be cast.",
				effects.And(effects.Not(effects.Creature()), lrwB1XInManaCost)),
		},
	})
}

// lrwB1XInManaCost passes for a card whose mana cost has an {X} in it.
func lrwB1XInManaCost(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	cost, err := game.ParseCost(c.ManaCost)
	return err == nil && cost.XSlots > 0
}
