package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savage Ventmaw — Creature — Dragon {4}{R}{G}, 4/4:
//
//	"Flying
//	 Whenever this creature attacks, add {R}{R}{R}{G}{G}{G}. Until end
//	 of turn, you don't lose this mana as steps and phases end."
//
// The six mana carry the keep mark (KeepManaUntilEndOfTurn, #2166): they
// survive the combat steps and are lost when the cleanup step begins.
func init() {
	Register(Spec{
		OracleID:        "f743d85f-754e-4e70-9907-af54acb9e8d4",
		Name:            "Savage Ventmaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Savage Ventmaw — add {R}{R}{R}{G}{G}{G}", Do(AddMana{
				Produced: "{R}{R}{R}{G}{G}{G}",
				Riders:   []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
			})),
		},
	})
}
