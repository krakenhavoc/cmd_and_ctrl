package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stronghold Gambit — Sorcery {1}{R} (Nemesis, S82 OLD batch 679, #5618):
//
//	"Each player chooses a card in their hand. Then each player reveals their
//	 chosen card. The owner of each creature card revealed this way with the
//	 lowest mana value puts it onto the battlefield."
//
// Every player with a card in hand chooses one (a player with an empty hand
// chooses nothing), in turn order from the active player. No choice is shown
// until the last is made, and then every chosen card is revealed together
// (effects.OldB679StrongholdGambit). Among the creature cards revealed, those
// of the lowest mana value (a tie puts them all) go onto the battlefield, one
// simultaneous entry, under their owners' control (CR 110.2). A noncreature
// card revealed is not put anywhere, and does not count toward "lowest".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "797bb53d-b111-41fe-95d7-64cb0f833b22",
		Name:         "Stronghold Gambit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OldB679StrongholdGambit(ctx)
		},
	})
}
