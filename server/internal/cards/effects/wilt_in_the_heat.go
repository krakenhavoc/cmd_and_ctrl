package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3bed0b60-5944-44bb-9ddd-c82f323e6d20",
		Name:         "Wilt in the Heat",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if one or more cards left your graveyard this turn.",
				g2CardLeftYourGraveyardThisTurn()),
		},
		OnResolve: damageFirstTargetExileIfItDies(fixedAmount(5)),
	})
}
