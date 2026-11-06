package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Spiral — Sorcery {4}{U} (Bloomburrow, #4467):
//
//	"Gift a tapped Fish
//	 Target player draws three cards. If the gift was promised, tap
//	 target creature an opponent controls and put a stun counter on
//	 it."
//
// The second target exists only when the gift was promised, which is
// gift.go's Instead with the printed player clause and the creature
// clause appended: the engine asks for the creature only after the
// promise. The gift (the opponent's tapped Fish) happens before the
// rest of the spell. The draw goes to the player target; the tap and the
// stun counter go to the creature target, if it is still legal as the
// spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03e356ae-f1e8-4730-8d2b-b9e2f32db915",
		Name:         "Mind Spiral",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Targets:      TargetPlayer("target player"),
		Gift: GiftATappedFish().Instead(Clauses(
			TargetPlayer("target player"),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		)),
		OnResolve: ResolveSteps(
			Instruction(mindSpiralDraw),
			Instruction(mindSpiralStun),
		),
	})
}

func mindSpiralDraw(ctx *Context) error {
	who, ok := ctx.ClauseTarget(0)
	if !ok || who.Kind != game.TargetPlayer {
		return nil
	}
	return DrawCards{Player: who.ID, N: 3}.Apply(ctx)
}

func mindSpiralStun(ctx *Context) error {
	if !ctx.GiftPromised() {
		return nil
	}
	c, ok := ctx.ClauseTarget(1)
	if !ok || c.Kind != game.TargetCard {
		return nil
	}
	if err := (TapTarget{Target: c.ID}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: c.ID, Kind: game.CounterStun, N: 1}.Apply(ctx)
}
