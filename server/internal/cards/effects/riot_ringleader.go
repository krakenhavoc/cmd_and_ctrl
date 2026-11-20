package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riot Ringleader — {2}{R} Creature — Human Warrior 2/2 (Avacyn Restored, #5570):
//
//	"Whenever this creature attacks, Human creatures you control get
//	 +1/+0 until end of turn."
//
// Wandering Musicians' attack trigger, narrowed to Humans. The set
// is read as the ability resolves (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "420d09b5-aced-4f71-914c-90d19f7476dc",
		Name:         "Riot Ringleader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Riot Ringleader — Human creatures you control get +1/+0 until end of turn",
				Do(BoostUntilEOT{Match: And(Creature(), YouControl(), HasSubtype("Human")), Power: 1, Label: "Riot Ringleader — +1/+0"})),
		},
	})
}
