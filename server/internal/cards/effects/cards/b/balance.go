package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balance — Sorcery {1}{W} (Secret Lair Drop):
//
//	"Each player chooses a number of lands they control equal to the number of
//	 lands controlled by the player who controls the fewest, then sacrifices the
//	 rest. Players discard cards and sacrifice creatures the same way."
//
// Magus of the Balance's ability as a spell: three steps in printed order (lands,
// cards in hand, creatures), each reading every player's count as it begins. Each
// player over the fewest sacrifices or discards the difference, of their own
// choice (choosing what to keep is choosing what to lose), the players answering
// in APNAP order (CR 101.4). The spell is on the stack, not the battlefield, so
// it is no creature or land; the hand is read after the lands, so a Balance cast
// from a hand of one card counts the other cards (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17fa98cd-ed8f-483f-9525-7e989a82ebb2",
		Name:         "Balance",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true},
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Do(effects.OldB016BalanceSteps()...)(ctx.Game, item)
		},
	})
}
