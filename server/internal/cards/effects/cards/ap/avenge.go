package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avenge — Sorcery {4}{W}{W}:
//
//	"This spell costs {2} less to cast if a player attacked you during
//	 their last turn.
//	 Destroy all creatures. You gain 1 life for each creature destroyed
//	 this way."
//
// Fumigate's resolution under a self cost reduction (CR 601.2f, 118.7a —
// the reduction touches only the generic part, {4}{W}{W} becoming
// {2}{W}{W}) that reads ADR 0108 §6's record: some other player's
// declared attack on YOU during THEIR last turn. An attack on one of your
// planeswalkers or battles is not an attack on you. In a four-player game
// "their last turn" is each player's own, so an opponent who attacked
// someone else during theirs does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65a64759-3a93-4542-84fc-3af5ccb741f4",
		Name:         "Avenge",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if a player attacked you during their last turn.", effects.APlayerAttackedYouDuringTheirLastTurn()),
		},
		OnResolve: effects.DestroyAllCreaturesGainLifeForEach,
	})
}
