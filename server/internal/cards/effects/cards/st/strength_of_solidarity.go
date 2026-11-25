package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strength of Solidarity — Sorcery {G} (Zendikar Rising, S82 OLD #5508):
//
//	"Choose target creature you control. Put a +1/+1 counter on it for each
//	 creature in your party. (Your party consists of up to one each of
//	 Cleric, Rogue, Warrior, and Wizard.)"
//
// The party is counted as the spell resolves (CR 700.8a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f903b76f-d405-427a-b115-5066eb311e10",
		Name:         "Strength of Solidarity",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := effects.S82OldB092PartySize(ctx)
			if n <= 0 {
				return nil
			}
			return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
		},
	})
}
