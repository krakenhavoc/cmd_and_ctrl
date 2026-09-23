package effects

// Deadbridge Goliath — Creature — Insect {2}{G}{G}, 5/5:
//
//	"Scavenge {4}{G}{G} ({4}{G}{G}, Exile this card from your
//	 graveyard: Put a number of +1/+1 counters equal to this card's
//	 power on target creature. Scavenge only as a sorcery.)"
//
// #1221's scavenge proof, and the first card in the catalog to pay
// `AbilityCost.ExileSelf`. A 5/5 with no other text, so the five
// counters a test counts are the keyword's own arithmetic: "this
// card's power" is read at resolution off the card the COST has
// already moved to exile, where its printed power is still 5 because
// nothing outside the battlefield has layers.
//
// See scavenge.go and exile_cost.go.
func init() {
	Register(Spec{
		OracleID:     "1498f5a1-6df7-4f80-9470-c93528b64a9c",
		Name:         "Deadbridge Goliath",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Scavenge("{4}{G}{G}")},
	})
}
