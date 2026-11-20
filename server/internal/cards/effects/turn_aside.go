package effects

// Turn Aside — Instant {U} (Scars of Mirrodin, #5577):
//
//	"Counter target spell that targets a permanent you control."
//
// Rebuff the Wicked's text and target clause: a spell at least one of
// whose targets is a permanent the caster controls on the battlefield,
// judged at announce and again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4325ae73-af43-4a7d-a552-9909f0dc77db",
		Name:         "Turn Aside",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell that targets a permanent you control", b14SpellTargetsAPermanentYouControl()),
		OnResolve:    somB1CounterTheTargetSpells,
	})
}
