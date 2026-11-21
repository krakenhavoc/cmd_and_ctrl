package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nantuko Calmer — Creature — Insect Druid {2}{G}{G}, 2/3 (Torment,
// #5612):
//
//	"{G}, {T}, Sacrifice this creature: Destroy target enchantment.
//	 Threshold — This creature gets +1/+1 as long as there are seven or
//	 more cards in your graveyard."
//
// Elf Replica's ability with a tap added to the cost: the tap and the
// sacrifice are paid as it is activated (CR 602.2b, 601.2h), so it needs
// the Calmer untapped and free of summoning sickness, and the Calmer is
// gone before the destruction resolves. The target is checked again as
// it resolves (CR 608.2b), and an indestructible enchantment survives.
// Threshold is an ability word (CR 207.2c): Seton's Scout's layer 7c
// self-modifier (CR 613.4c) at +1/+1, read live on every recompute.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4409b15-b9f8-4bdb-9857-5cd70a0d3193",
		Name:         "Nantuko Calmer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice this creature: Destroy target enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
		Static: []game.StaticAbility{effects.FraThresholdSelfPT(1, 1)},
	})
}
