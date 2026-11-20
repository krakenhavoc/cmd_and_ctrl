package effects

// Walk the Plank — Sorcery {B}{B} (Ixalan, #5532):
//
//	"Destroy target non-Merfolk creature."
//
// Rend Flesh's body with Merfolk for Spirit: the subtype is read off the
// creature as it is now (CR 205.3), at announce and at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9093e6f1-12bc-4dfb-bc78-6ead0752ad61",
		Name:         "Walk the Plank",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target non-Merfolk creature", Creature(), Not(HasSubtype("Merfolk"))),
		OnResolve:    destroyTheTargetPermanent,
	})
}
