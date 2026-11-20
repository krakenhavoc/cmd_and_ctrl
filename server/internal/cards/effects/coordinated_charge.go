package effects

// Coordinated Charge — Instant {4}{W} (Ikoria: Lair of Behemoths, #5512):
//
//	"Creatures you control get +2/+1 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Heroic Charge's team pump over the creatures you control as it
// resolves, and cycling from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "688a0b7d-bf00-45ab-ab3f-9d19f52f019a",
		Name:         "Coordinated Charge",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(BoostUntilEOT{
			Match: And(Creature(), YouControl()), Power: 2, Toughness: 1,
			Label: "Coordinated Charge — creatures you control get +2/+1",
		}),
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
