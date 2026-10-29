package effects

// Agent of Atlas — Creature — Human Spy Hero {1}{W}, 2/2 (Marvel Super
// Heroes, #4455):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Prowess is a keyword the engine derives its trigger from
// (game/prowess.go); the card only declares it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cd403866-bef1-413a-85e2-c86e295b3589",
		Name:            "Agent of Atlas",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
	})
}
