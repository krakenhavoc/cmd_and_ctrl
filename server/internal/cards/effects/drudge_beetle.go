package effects

// Drudge Beetle — Creature — Insect {1}{G}, 2/2 (Return to Ravnica, #5566):
//
//	"Scavenge {5}{G} ({5}{G}, Exile this card from your graveyard: Put
//	 a number of +1/+1 counters equal to this card's power on target
//	 creature. Scavenge only as a sorcery.)"
//
// Scavenge is the shared graveyard ability (Deadbridge Goliath,
// scavenge.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5126c01-f765-44f4-9ced-0fc157f8b897",
		Name:         "Drudge Beetle",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Scavenge("{5}{G}")},
	})
}
