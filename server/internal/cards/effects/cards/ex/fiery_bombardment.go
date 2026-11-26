package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiery Bombardment — {1}{R} Enchantment:
//
//	"Chroma — {2}, Sacrifice a creature: This enchantment deals damage to any
//	 target equal to the number of red mana symbols in the sacrificed creature's
//	 mana cost."
//
// The sacrificed creature's cost is read as it last existed (CR 608.2h): its
// copiable values, so a copy counts the copied cost. A token that is no copy
// has no cost and deals none.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85842fa9-e7e8-4380-9034-c93a5497d3d8",
		Name:         "Fiery Bombardment",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Chroma — {2}, Sacrifice a creature: This enchantment deals damage to any target equal to the number of red mana symbols in the sacrificed creature's mana cost.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeACreature()),
			Targets: effects.TargetAny(),
			Effect: effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
				info, ok := ctx.SacrificedPermanent()
				if !ok || info.Copiable == nil {
					return 0
				}
				return effects.S82OldB521Symbols(info.Copiable.ManaCost, ctx.TextColor("R"))
			}),
		}},
	})
}
