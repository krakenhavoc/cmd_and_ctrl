package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnashing of Teeth — Sorcery {1}{B}{B}:
//
//	"Choose one —
//	 • Target creature gets -5/-5 until end of turn. If that creature
//	   would die this turn, exile it instead.
//	 • Creatures target player controls get -1/-1 until end of turn."
//
// The first bullet marks its target as the spell resolves (ADR 0108
// §1), before the state-based check that kills it, so a creature the
// -5/-5 kills is exiled. The second bullet's set is the targeted
// player's creatures as it resolves (CR 611.2c), Shields of Velis Vel's
// shape.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "188a3f61-3f7c-450d-8ee7-e12ad13b86c6",
		Name:         "Gnashing of Teeth",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -5/-5 until end of turn. If that creature would die this turn, exile it instead.",
				effects.TargetCreature("target creature"), gnashingShrinkAndExileIfItDies), effects.ForTargets(effects.HarmUntilEOT(0, -5, -5))),
			effects.ModeWithPurpose(effects.ModeDoing("Creatures target player controls get -1/-1 until end of turn.",
				effects.TargetPlayer("target player"), gnashingShrinkPlayersCreatures), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}),
		),
	})
}

func gnashingShrinkAndExileIfItDies(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: t.ID, Power: -5, Toughness: -5, Label: "Gnashing of Teeth — -5/-5"}).Apply(ctx); err != nil {
		return err
	}
	return effects.ExileIfItWouldDieThisTurn{Target: t.ID}.Apply(ctx)
}

func gnashingShrinkPlayersCreatures(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.BoostUntilEOT{
		Match: effects.And(effects.Creature(), effects.ControlledByPlayer(t.ID)), Power: -1, Toughness: -1,
		Label: "Gnashing of Teeth — -1/-1",
	}.Apply(ctx)
}
