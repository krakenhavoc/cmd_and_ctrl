package effects

// Pay No Heed — Instant {W}:
//
//	"Prevent all damage a source of your choice would deal this turn."
//
// ADR 0108 §7 (#1904): a shield against a source chosen as it resolves
// (CR 609.7a) that is not spent by the first instance of damage: every
// instance from that source this turn, to anything, is prevented
// (CR 615.1a). A chosen permanent spell covers the permanent it becomes.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "71da5acb-78bc-453a-ae28-d22faa5a4e43",
		Name:         "Pay No Heed",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromChosenSource(ShieldAnything)),
	})
}
