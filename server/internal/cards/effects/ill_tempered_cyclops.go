package effects

// Ill-Tempered Cyclops — Creature — Cyclops {3}{R}, 3/3 (Theros, #5561):
//
//	"Trample
//	 {5}{R}: Monstrosity 3. (If this creature isn't monstrous, put three +1/+1 counters on it and it becomes monstrous.)"
//
// The shared monstrosity row (CR 701.37).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "07f62413-92cc-4f44-a827-a7c3a7318602",
		Name:            "Ill-Tempered Cyclops",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{5}{R}"), 3)},
	})
}
