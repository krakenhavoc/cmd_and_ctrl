package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d4409b15-b9f8-4bdb-9857-5cd70a0d3193",
		Name:         "Nantuko Calmer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice this creature: Destroy target enchantment.",
			Cost:    Plus(ManaCost("{G}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect:  destroyFirstLegalTarget,
		}},
		Static: []game.StaticAbility{fraThresholdSelfPT(1, 1)},
	})
}
