package effects

// Knight of Dusk's Shadow — Creature — Human Knight {1}{B}, 2/2:
//
//	"Menace
//	 Your opponents can't gain life.
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// Menace rides PrintedKeywords; "Your opponents can't gain life" is
// ADR 0107 §5's battlefield static (CR 119.7, #1880); the pump is the
// shared until-end-of-turn boost on the Knight itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "815ae30f-f455-4795-96eb-4bcc41d415e8",
		Name:            "Knight of Dusk's Shadow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		CantGainLife:    OpponentsCantGainLife(),
		Activated: []ActivatedAbility{{
			Label:  "{1}{B}: This creature gets +1/+1 until end of turn.",
			Cost:   ManaCost("{1}{B}"),
			Effect: thisGetsUntilEndOfTurn(1, 1, "Knight of Dusk's Shadow — +1/+1 until end of turn"),
		}},
	})
}
