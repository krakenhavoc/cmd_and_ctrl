package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beastie Beatdown — Sorcery {R}{G} (Duskmourn: House of Horror, #4466):
//
//	"Choose target creature you control and target creature an opponent
//	 controls.
//	 Delirium — If there are four or more card types among cards in your
//	 graveyard, put two +1/+1 counters on the creature you control.
//	 The creature you control deals damage equal to its power to the
//	 creature an opponent controls."
//
// Bite Down's two clauses (slot 0 yours, slot 1 theirs; both are required,
// so the spell cannot be cast without both). Delirium is an ability word
// (CR 207.2c) and is read as the spell resolves, so the spell itself, still
// on the stack, is not among the graveyard's types. The counters go on the
// creature you control while it is still a legal target (CR 608.2b), even
// when the opponent's creature has gone; the damage is then read from its
// power with those counters, and is dealt only if both targets are still
// legal. The counters may stop for a replacement-order prompt, so the
// damage waits for it (CR 608.2c, #4155).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3da5fbe2-e47b-4a1d-9e73-393f37314fa3",
		Name:         "Beastie Beatdown",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // a one-sided fight at a creature
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				mine, ok := ctx.ClauseTarget(0)
				if !ok || mine.Kind != game.TargetCard || effects.B16CardTypesInGraveyard(ctx.Game, ctx.Controller()) < 4 {
					return nil
				}
				return effects.AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OneSidedBite(nil, ctx)
			}),
		),
	})
}
