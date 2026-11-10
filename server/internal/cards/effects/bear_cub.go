package effects

// Bear Cub — Creature — Bear {1}{G}, 2/2 (Foundations, #4465):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see ordinary_bear.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed206651-3d9b-4546-8d45-0682817192fd",
		Name:         "Bear Cub",
		Completeness: CompletenessFull,
	})
}
