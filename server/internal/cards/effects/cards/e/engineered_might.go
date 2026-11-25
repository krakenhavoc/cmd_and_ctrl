package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Engineered Might — Sorcery {3}{G}{W} (Kaladesh, #5539):
//
//	"Choose one —
//	 • Target creature gets +5/+5 and gains trample until end of turn.
//	 • Creatures you control get +2/+2 and gain vigilance until end of turn."
//
// Modal. The first bullet is Moment of Valor's targeted mode (a target gone
// is skipped, CR 608.2b) with a +5/+5 pump and a trample grant; the second
// is Myrkul's Invoker's team pump with vigilance: the creatures you control
// as it resolves are the ones affected (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38462d26-f252-4055-9f1a-59bc3ab2c75e",
		Name:         "Engineered Might",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +5/+5 and gains trample until end of turn.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(engineeredMightTarget)),
				effects.ForTargets(effects.BenefitUntilEOT(0, 5, 5, "trample"))),
			effects.ModeDoing("Creatures you control get +2/+2 and gain vigilance until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					mine := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: mine, Power: 2, Toughness: 2, Label: "Engineered Might — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"vigilance"}, Label: "Engineered Might — vigilance"}.Apply(ctx)
				}),
		),
	})
}

// engineeredMightTarget is the first bullet on its still-legal target.
func engineeredMightTarget(ctx *effects.Context, id uuid.UUID) error {
	if err := (effects.BoostUntilEOT{Target: id, Power: 5, Toughness: 5, Label: "Engineered Might — +5/+5"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Engineered Might — trample"}.Apply(ctx)
}
