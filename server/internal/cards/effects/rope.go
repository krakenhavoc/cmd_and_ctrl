package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rope — Artifact — Clue Equipment {G} (Murders at Karlov Manor, #4469):
//
//	"Equipped creature gets +1/+2, has reach, and can't be blocked by more
//	 than one creature.
//	 {2}, Sacrifice this Equipment: Draw a card.
//	 Equip {3}"
//
// Lead Pipe's shape with Vorrac Battlehorns' block limit: the equipped
// creature is the scope of MaxBlockers, so menace on it makes it
// unblockable in effect (the two rules both hold). The draw is printed on
// the card, and sacrificing it is sacrificing a Clue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3f09898-6c77-4d68-8d10-8440c75f884c",
		Name:         "Rope",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
			GrantToAttached("reach"),
		},
		BlockRules: []game.BlockRule{MaxBlockers(OnAttached(), 1)},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, Sacrifice this Equipment: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
				Effect:  Do(DrawCards{N: 1}),
			},
			EquipAbility("{3}"),
		},
	})
}
