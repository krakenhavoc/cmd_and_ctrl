package effects

// Sinkhole — Sorcery {B}{B} (Limited Edition Alpha, #5642):
//
//	"Destroy target land."
//
// Stone Rain's sentence (stone_rain.go): one target clause and the
// shared "Destroy target [permanent]." body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a46ad2a-35b1-4dd5-b7c3-fec36b7c67ab",
		Name:         "Sinkhole",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
