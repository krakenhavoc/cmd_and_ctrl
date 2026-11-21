package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Konstrari Charm — Instant {R}{G}:
//
//	"Choose one —
//	 • Konstrari Charm deals 6 damage to target creature with flying.
//	 • Put two +1/+1 counters on target creature. It gains trample
//	   until end of turn.
//	 • Add {C}{C}{C}."
//
// Mode one's "with flying" is a target restriction read through
// HasKeyword, so a granted flying counts. Mode three is a spell
// adding mana (it uses the stack and can be countered), not a mana
// ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40961b23-b153-434a-aeb9-b9c2ff613dcb",
		Name:         "Konstrari Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Konstrari Charm deals 6 damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				effects.SocB20DamageToTheModeTarget(6)),
			// The trample waits for a prompt the counters ask (a CR 616 order; CR 608.2c, #4155).
			effects.ModeSteps("Put two +1/+1 counters on target creature. It gains trample until end of turn.",
				effects.TargetCreature("target creature"),
				effects.Instruction(effects.PutCountersOnTheModesTargetStep(game.CounterPlusOne, 2)),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"trample"},
						Label:    "Konstrari Charm — trample",
					}.Apply(ctx)
				})),
			effects.ModeDoing("Add {C}{C}{C}.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.AddMana{Produced: "{C}{C}{C}"}.Apply(ctx)
				}),
		),
	})
}
