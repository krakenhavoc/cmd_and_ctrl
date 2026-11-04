package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cost of Brilliance — Sorcery {2}{B} (Secrets of Strixhaven, #4456):
//
//	"Target player draws two cards and loses 2 life. Put a +1/+1 counter
//	 on up to one target creature."
//
// Two target clauses: the player (slot 0) and up to one creature (slot 1,
// any creature, not only yours). The draw and the life loss are one clause
// of the effect; the counter is a second step, so a draw or a life loss
// that waits on a question still gets it afterwards (CR 608.2c). Each part
// reads its own slot as it resolves (CR 608.2b): a player who is no longer
// a legal target neither draws nor loses life, and a creature that is gone
// takes no counter, without stopping the other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b45afde2-eaaf-4174-8082-d344efa54c44",
		Name:         "Cost of Brilliance",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target player"),
			TargetCreature("up to one target creature").WithCount(0, 1),
		),
		Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2}),
		OnResolve: ResolveSteps(
			Instruction(costOfBrillianceDrawAndLose),
			Instruction(costOfBrillianceCounter),
		),
	})
}

// costOfBrillianceDrawAndLose is "target player draws two cards and loses
// 2 life".
func costOfBrillianceDrawAndLose(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	if err := (DrawCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -2)
}

// costOfBrillianceCounter is "put a +1/+1 counter on up to one target
// creature".
func costOfBrillianceCounter(ctx *Context) error {
	t, ok := ctx.ClauseTarget(1)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
