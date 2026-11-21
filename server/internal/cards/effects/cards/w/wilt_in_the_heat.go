package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wilt in the Heat — Instant {2}{R}{W}:
//
//	"This spell costs {2} less to cast if one or more cards left your
//	 graveyard this turn.
//	 Wilt in the Heat deals 5 damage to target creature. If that creature
//	 would die this turn, exile it instead."
//
// The reduction is a self cost modifier read at CR 601.2f off this
// turn's events (g2CardLeftYourGraveyardThisTurn): any card you own
// moving out of your graveyard — exiled, returned, reanimated, cast. The
// replacement is the spell's (ADR 0108 §1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bed0b60-5944-44bb-9ddd-c82f323e6d20",
		Name:         "Wilt in the Heat",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if one or more cards left your graveyard this turn.",
				effects.G2CardLeftYourGraveyardThisTurn()),
		},
		OnResolve: effects.DamageFirstTargetExileIfItDies(effects.FixedAmount(5)),
	})
}
