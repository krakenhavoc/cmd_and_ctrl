package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Buccaneer's Bravado — Instant {1}{R}:
//
//	"Choose one —
//	 • Target creature gets +1/+1 and gains first strike until end of turn.
//	 • Target Pirate gets +1/+1 and gains double strike until end of turn."
//
// Each mode has its own target clause, chosen as the spell is cast (CR 601.2b);
// the pump and the keyword are one until-end-of-turn effect, and a target that
// has left or stopped being a legal one is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e85421f3-37f6-49ee-9fc4-4424fd4f5236",
		Name:         "Buccaneer's Bravado",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +1/+1 and gains first strike until end of turn.",
				effects.TargetCreature("target creature"), buccaneersBravadoPump("first strike")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "first strike"))),
			effects.ModeWithPurpose(effects.ModeDoing("Target Pirate gets +1/+1 and gains double strike until end of turn.",
				effects.TargetCreature("target Pirate", effects.Subtype("Pirate")), buccaneersBravadoPump("double strike")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "double strike"))),
		),
	})
}

func buccaneersBravadoPump(keyword string) func(*game.StackItem, *effects.Context, int) error {
	return effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
		if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Buccaneer's Bravado — +1/+1"}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{keyword}, Label: "Buccaneer's Bravado — " + keyword}.Apply(ctx)
	})
}
