package effects

// Rag Dealer — Creature — Human Rogue {B}, 1/1:
//
//	"{2}{B}, {T}: Exile up to three target cards from a single
//	 graveyard."
//
// Carrion Beetles' ability on a Human Rogue (#1807, ADR 0106 §5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "493acba2-92b0-43a4-9f4b-d49a91ca77f5",
		Name:         "Rag Dealer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			ExileFromASingleGraveyardAbility("{2}{B}, {T}", Plus(ManaCost("{2}{B}"), TapCost()), 3),
		},
	})
}
