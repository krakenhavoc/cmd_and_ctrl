package effects

// Courier Hawk — Creature — Bird {1}{W}, 1/2 (Ravnica: City of Guilds,
// #5600):
//
//	"Flying, vigilance"
//
// Serra Angel's two keywords on a smaller body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7fc904aa-377b-4671-8308-275205f56d81",
		Name:            "Courier Hawk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
