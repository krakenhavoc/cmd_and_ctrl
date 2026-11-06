package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Qo'noS Quartermaster — Creature — Klingon Warrior {1}{R}, 2/2 (Star
// Trek Commander, #4974):
//
//	"Menace
//	 Whenever this creature attacks, add one mana of any color. Until end
//	 of turn, you don't lose this mana as steps and phases end."
//
// Kessig Naturalist's attack mana with a five-colour pick: the trigger
// uses the stack (it is not a mana ability, CR 605.1b), and the mana
// carries the keep-until-end-of-turn mark, so it survives the end of
// combat to the second main phase and empties at the end of the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b7115ed6-526c-491d-828f-f60b18437361",
		Name:            "Qo'noS Quartermaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Qo'noS Quartermaster — add one mana of any color; you don't lose it until end of turn", Do(AddMana{
				Produced: "{W|U|B|R|G}",
				Riders:   []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
			})),
		},
	})
}
