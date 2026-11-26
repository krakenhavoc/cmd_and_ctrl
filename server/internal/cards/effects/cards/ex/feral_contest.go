package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Feral Contest — Sorcery {3}{G} (Worldwake, S82 OLD b487):
//
//	"Put a +1/+1 counter on target creature you control. Another target creature
//	 blocks it this turn if able."
//
// Two target clauses, the second a different creature from the first (Monstrous
// Step's shape). Each instruction skips a target that is illegal as the spell
// resolves (CR 608.2b); the requirement names the first creature as the object it
// is now (Hunt Down's record), and asks nothing of a blocker that can't legally
// block it (CR 509.1c). With the first creature gone there is no "it", so nothing
// is required of the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efbb5b87-d175-477e-a490-85dd98b62de6",
		Name:         "Feral Contest",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			blocker, ok := ctx.ClauseTarget(1)
			if !ok || blocker.Kind != game.TargetCard {
				return nil
			}
			ref, ok := ctx.Game.PermanentRefForEffect(mine.ID)
			if !ok {
				return nil
			}
			return effects.BlocksAttackerUntilEOT{Blocker: blocker.ID, Attacker: ref, Label: "Feral Contest — blocks the creature that got the counter if able"}.Apply(ctx)
		},
	})
}
