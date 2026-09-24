package effects

// Grim Lavamancer — Creature — Human Wizard {R}, 1/1 (EDHREC rank 15361):
//
//	"{R}, {T}, Exile two cards from your graveyard: This creature deals
//	 2 damage to any target."
//
// The card #1297 was filed under. "Exile two cards from your graveyard"
// is AbilityCost.ExileCards read against the graveyard (game.ExileCost,
// From: ZoneGraveyard): the activator names the two cards at announce
// (CR 602.2b), they leave for exile before the ability is on the stack,
// and nothing about the move is a discard or a sacrifice. Any two cards
// — the clause has no predicate — and never the Lavamancer itself,
// which is on the battlefield anyway.
//
// The damage comes from the creature (CR 609.7a), by last-known
// information if it has left in response, which is what
// b33DamageChosenTargetFromSource reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37445e06-88a1-4e2e-a432-383736c9b977",
		Name:         "Grim Lavamancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}, Exile two cards from your graveyard: This creature deals 2 damage to any target.",
			Cost:    Plus(ManaCost("{R}"), TapCost(), ExileFromGraveyard(2, "two cards", nil)),
			Targets: TargetAny(),
			Effect:  b33DamageChosenTargetFromSource(2),
		}},
	})
}
