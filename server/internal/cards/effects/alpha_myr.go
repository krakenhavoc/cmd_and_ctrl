package effects

// Alpha Myr — Artifact Creature — Myr {2}, 2/1 (Mirrodin, #5606):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see bear_cub.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4caa8e68-e599-416f-b839-3d009569ff29",
		Name:         "Alpha Myr",
		Completeness: CompletenessFull,
	})
}
