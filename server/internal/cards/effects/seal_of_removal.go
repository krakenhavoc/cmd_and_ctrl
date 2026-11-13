package effects

// Seal of Removal — Enchantment {U} (Nemesis, #5618):
//
//	"Sacrifice this enchantment: Return target creature to its owner's
//	 hand."
//
// Aether Spellbomb's bounce with no mana in the cost: SacrificeThis() and
// the shared bounceTheTargetToHand body. The creature goes to its
// OWNER's hand; one that left the battlefield in response is no longer a
// legal target, and the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0801029-bcf7-4bdb-84bf-e88dcaa9dc03",
		Name:         "Seal of Removal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Return target creature to its owner's hand.",
			Cost:    SacrificeThis(),
			Targets: TargetCreature("target creature"),
			Effect:  bounceTheTargetToHand,
		}},
	})
}
