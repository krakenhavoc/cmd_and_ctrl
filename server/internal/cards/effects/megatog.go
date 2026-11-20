package effects

// Megatog — Creature — Atog {4}{R}{R}, 3/4 (Mirrodin, #5606):
//
//	"Sacrifice an artifact: This creature gets +3/+3 and gains trample
//	 until end of turn."
//
// Atog's sacrifice cost (any artifact you control, the Megatog itself
// if it is one), with the boost and the trample grant both landing on
// the source as the ability resolves (mrdB5SacrificeAnArtifactFor).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bde427fa-8a00-4eff-9371-81324bc75364",
		Name:         "Megatog",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{mrdB5SacrificeAnArtifactFor("Megatog", 3, 3, "trample")},
	})
}
