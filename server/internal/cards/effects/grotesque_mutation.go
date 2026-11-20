package effects

// Grotesque Mutation — Instant {1}{B} (Shadows over Innistrad, #5543):
//
//	"Target creature gets +3/+1 and gains lifelink until end of turn.
//	 (Damage dealt by the creature also causes its controller to gain
//	 that much life.)"
//
// Interjection's pump-and-keyword body (socB27TargetGetsAndGains) at +3/+1
// and lifelink.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "058a8475-649f-433a-b258-42bc03bf40ba",
		Name:         "Grotesque Mutation",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 1, "lifelink")),
		OnResolve:    socB27TargetGetsAndGains(3, 1, "lifelink", "Grotesque Mutation"),
	})
}
