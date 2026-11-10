package effects

// Twincast — Instant {U}{U}:
//
//	"Copy target instant or sorcery spell. You may choose new
//	 targets for the copy."
//
// Word-for-word Reverberate in blue. Both are registered because
// the catalog keys on oracle ID and they are two different cards —
// a Commander deck runs whichever colour it is in, and several run
// both.
func init() {
	Register(Spec{
		OracleID:     "8f878efc-850f-43d2-a6fe-5ea8d1dd5afb",
		Name:         "Twincast",
		Completeness: CompletenessFull,
		Targets:      instantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve:    cmdrB21CopyTheTargetSpell,
	})
}
