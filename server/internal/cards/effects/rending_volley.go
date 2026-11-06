package effects

// Rending Volley — Instant {R} (Avatar: The Last Airbender Eternal, #5075):
//
//	"This spell can't be countered.
//	 Rending Volley deals 4 damage to target white or blue creature."
//
// Combust's shape without its "can't be prevented": the spell's own
// CantBeCountered flag and 4 damage to a creature that is white or blue
// as it is targeted and again as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0324941f-484c-4656-8d1c-66e919c2bc7c",
		Name:            "Rending Volley",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetCreature("target white or blue creature", Or(OfColor("W"), OfColor("U"))),
		Purpose:         ForTargets(DamageToTarget(0, 4)),
		OnResolve:       damageToFirstTarget(4),
	})
}
