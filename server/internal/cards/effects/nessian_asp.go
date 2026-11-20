package effects

// Nessian Asp — Creature — Snake {4}{G}, 4/5 (Theros, #5561):
//
//	"Reach
//	 {6}{G}: Monstrosity 4. (If this creature isn't monstrous, put four +1/+1 counters on it and it becomes monstrous.)"
//
// The shared monstrosity row (CR 701.37).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8b46e077-f667-432f-ab5b-8256a7378311",
		Name:            "Nessian Asp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []ActivatedAbility{Monstrosity(ManaCost("{6}{G}"), 4)},
	})
}
