package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c19f4a43-85bb-4070-a237-171597c94ec3",
		Name:         "Undercity Upheaval",
		Completeness: CompletenessFull,
		Targets: TargetCreature("any number of target creatures you control", YouControl()).WithCount(0, 0).
			Dividing(DivideBy(matDivideCreatureCardsInYourGraveyard)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := PutDividedCounters(ctx, game.CounterPlusOne); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"vigilance"},
				Label:    "Undercity Upheaval — vigilance",
			}.Apply(ctx)
		},
	})
}
