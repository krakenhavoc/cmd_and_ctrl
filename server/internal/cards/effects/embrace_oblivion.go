package effects

// Embrace Oblivion — Sorcery {B} (Edge of Eternities, #4461):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Destroy target creature or Spacecraft."
//
// Bone Splinters with Deadly Dispute's "an artifact or creature" cost
// and Gravkill's target clause: "Spacecraft" is the artifact subtype, so
// a Spacecraft below its station threshold can be destroyed too. The
// sacrifice is paid as the spell is cast (CR 601.2h), so it is gone even
// if the target becomes illegal and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "0fad5713-c09c-43d3-9cc0-10c65aead488",
		Name:           "Embrace Oblivion",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		Targets:        TargetPermanent("target creature or Spacecraft", Or(Creature(), HasSubtype("Spacecraft"))),
		OnResolve:      destroyTheTargetPermanent,
	})
}
