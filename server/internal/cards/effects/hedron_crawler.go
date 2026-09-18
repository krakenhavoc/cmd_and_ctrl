package effects

// Hedron Crawler — 0/1 Artifact Creature — Construct for {2} (EDHREC
// rank 3966):
//
//	"{T}: Add {C}. ({C} represents colorless mana.)"
//
// A two-mana body that taps for one colourless. It is in the batch
// because the Eldrazi decks that want it want a CREATURE that makes
// {C} rather than a rock — it can be tutored with a creature tutor,
// sacrificed to Ashnod's Altar, and pumped by a lord — and because it
// is the cleanest possible check that a colourless-only mana ability
// on a creature is offered from the battlefield at all.
//
// Summoning sickness applies: the Crawler's ability has {T} in its
// cost, so it makes no mana the turn it lands (CR 302.6). That is the
// engine's rule, not something this file declares.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b9a4e06-b21d-4cbe-906f-9dbd08dbe5d3",
		Name:         "Hedron Crawler",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
