package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a3f09898-6c77-4d68-8d10-8440c75f884c",
		Name:         "Rope",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
			effects.GrantToAttached("reach"),
		},
		BlockRules: []game.BlockRule{effects.MaxBlockers(effects.OnAttached(), 1)},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, Sacrifice this Equipment: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			effects.EquipAbility("{3}"),
		},
	})
}
