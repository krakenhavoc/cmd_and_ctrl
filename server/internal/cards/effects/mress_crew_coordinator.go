package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// M'Ress, Crew Coordinator — Legendary Creature — Caitian Officer {4}{W},
// 4/4 (Star Trek Commander, #4974):
//
//	"Landfall — Whenever a land you control enters, create two 1/1 red
//	 Officer creature tokens."
//
// The shared Landfall trigger with the printed Officer row, once for each
// land that enters under M'Ress's controller's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a33f6ec8-5992-43a1-a16c-d8e83c4017fa",
		Name:         "M'Ress, Crew Coordinator",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			Landfall("M'Ress, Crew Coordinator — create two 1/1 red Officer creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 red Officer"), N: 2})),
		},
	})
}
