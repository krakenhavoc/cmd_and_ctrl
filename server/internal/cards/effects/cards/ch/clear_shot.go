package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clear Shot — Instant {2}{G} (Innistrad: Midnight Hunt Commander
// reprint, S82):
//
//	"Target creature you control gets +1/+1 until end of turn. It deals
//	 damage equal to its power to target creature you don't control."
//
// Two target clauses read positionally, Bite Down's shape: slot 0 is the
// pumped creature and the damage SOURCE, slot 1 the recipient. The pump
// resolves first, so "its power" counts the +1 (CR 608.2c), and the damage
// carries the creature's deathtouch and lifelink. If either target is gone
// by resolution the other half still happens (CR 608.2b): a lost recipient
// still gets the pump, and a lost creature deals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b8defd2-8530-47d7-a0c6-27aa5162dff6",
		Name:         "Clear Shot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 1))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 1, Toughness: 1, Label: "Clear Shot — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return effects.OneSidedBiteTimes(ctx, 1)
		},
	})
}
