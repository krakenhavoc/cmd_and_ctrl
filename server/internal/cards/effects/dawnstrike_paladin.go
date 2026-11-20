package effects

// Dawnstrike Paladin — Creature — Human Knight {3}{W}{W}, 2/4 (Magic 2014, #5563):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	 Lifelink (Damage dealt by this creature also causes you to gain that much life.)"
//
// Two printed keywords the engine reads: attacking does not tap it, and its
// damage gains its controller that much life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c2bfb230-6949-4d5e-8cac-9d7dafdc5e9a",
		Name:            "Dawnstrike Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
	})
}
