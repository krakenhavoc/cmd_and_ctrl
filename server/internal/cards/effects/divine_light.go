package effects

// Divine Light — Sorcery {W}:
//
//	"Prevent all damage that would be dealt this turn to creatures you
//	 control."
//
// #2045: creatures you control, and not you. "You control" is read as
// the damage would be dealt (CR 611.2c): a creature you gain control of
// later this turn is protected, and one you lose control of is not.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "924bf3bb-9f06-4a2c-a17f-3fe11551bcf2",
		Name:         "Divine Light",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldCreaturesYouControl}),
	})
}
