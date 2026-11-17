package effects

// Ice Storm — Sorcery {2}{G} (Limited Edition Alpha, #5642):
//
//	"Destroy target land."
//
// Stone Rain's sentence (stone_rain.go): one target clause and the
// shared "Destroy target [permanent]." body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0b97e33-2d0c-4800-a9f5-9cd9be651ea8",
		Name:         "Ice Storm",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
