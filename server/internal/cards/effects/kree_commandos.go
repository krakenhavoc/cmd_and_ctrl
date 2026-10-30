package effects

// Kree Commandos — Creature — Kree Soldier Villain {2}{W}, 2/1 (Marvel
// Super Heroes, #4455):
//
//	"Flying, vigilance
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// All three are keywords the engine derives its behaviour from
// (game/prowess.go for the trigger); the card only declares them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "09eb988a-e2eb-42e1-a786-01be63d2263f",
		Name:            "Kree Commandos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "prowess"},
	})
}
