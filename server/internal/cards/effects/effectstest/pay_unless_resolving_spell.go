package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pay_unless_resolving_spell_test.go — #5793: a PayUnless asked by a
// RESOLVING SPELL holds the table while "that player" decides.
//
// The spell below is Rhystic Deluge's and Withdraw's clause on a sorcery:
// "Return target creature to its owner's hand unless its controller pays
// {1}. Draw a card." The draw is a later instruction (CR 608.2c), so it
// shows whether the rest of the resolution waits for the answer.
//
// The report was reproduced on f2353f118: through the plain
// effects.PayUnless the spell finished before the payer answered (the draw
// ran, the sorcery went to the graveyard, any seat could pass and the step
// could end), and the bounce landed whenever the payer answered. The owner
// decided (2026-11-19, option 1) that the engine derives the halt: a
// pay-unless a resolution asks is owed in the step it is asked in (CR
// 118.12: the payment is a cost paid as the spell resolves; CR 608.2g;
// CR 117.3b). These tests are the same scenarios with the assertions
// turned round.

const PayUnlessSpellOracle = "test-5793-pay-unless-spell"

func init() {
	effects.Register(effects.Spec{
		OracleID: PayUnlessSpellOracle,
		Name:     "Fixture Rhystic Return",
		Targets:  effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(askTheTargetsControllerForOne),
			effects.DrawCards{N: 1},
		),
	})
}

// askTheTargetsControllerForOne asks the target creature's controller for
// {1} through the plain PayUnless; the decline is Crystal Shard's
// registered bounce, so no test-only key is added to the ledger. Nothing
// here says "hold the table": the engine works that out.
func askTheTargetsControllerForOne(ctx *effects.Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		question := "pay {1}, or " + c.Name + " returns to its owner's hand?"
		return effects.PayUnless{Chooser: c.Controller, Cost: "{1}", Question: question, Else: effects.CrystalShardBounce}.Apply(ctx)
	}
	return nil
}
