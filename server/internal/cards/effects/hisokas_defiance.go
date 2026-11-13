package effects

// Hisoka's Defiance — Instant {1}{U} (Champions of Kamigawa, #5603):
//
//	"Counter target Spirit or Arcane spell."
//
// Essence Scatter's shape: a spell is legal when its type line carries
// the Spirit creature type or the Arcane spell type, checked again as
// this resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ac6f1c7-f4c5-4d45-9644-da49d8fe4758",
		Name:         "Hisoka's Defiance",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target Spirit or Arcane spell", chkB1SpiritOrArcane()),
		OnResolve:    counterTheTargetSpell,
	})
}
