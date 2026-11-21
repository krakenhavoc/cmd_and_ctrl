package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undercity Upheaval — Sorcery {1}{G}{G} (March of the Machine: The
// Aftermath, #5476):
//
//	"Undergrowth — Distribute X +1/+1 counters among any number of target
//	 creatures you control, where X is the number of creature cards in your
//	 graveyard as you cast this spell. Creatures you control gain vigilance
//	 until end of turn."
//
// Undergrowth is an ability word with no rules meaning. X is the divided
// clause's amount, read once as the targets are announced (CR 601.2d,
// matDivideCreatureCardsInYourGraveyard); the caster names the targets and
// the split with the cast, each target at least one, and a departed
// target's share is lost (PutDividedCounters). With no creature card in
// the graveyard X is 0, no target can be chosen and only the vigilance
// happens. Vigilance goes to the creatures you control as the spell
// resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c19f4a43-85bb-4070-a237-171597c94ec3",
		Name:         "Undercity Upheaval",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("any number of target creatures you control", effects.YouControl()).WithCount(0, 0).
			Dividing(effects.DivideBy(effects.MatDivideCreatureCardsInYourGraveyard)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.PutDividedCounters(ctx, game.CounterPlusOne); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"vigilance"},
				Label:    "Undercity Upheaval — vigilance",
			}.Apply(ctx)
		},
	})
}
