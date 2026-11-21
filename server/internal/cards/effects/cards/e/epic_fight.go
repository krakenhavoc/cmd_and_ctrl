package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Epic Fight — {2}{G} Sorcery:
//
//	"Choose one or both —
//	 • Double target creature's power and toughness until end of turn.
//	 • Target creature you control fights target creature an opponent
//	   controls."
//
// The bullets resolve in printed order (CR 608.2c), so a creature doubled by
// the first bullet fights with the doubled power when both are chosen. Each
// bullet reads its own targets and skips them if they are gone (CR 608.2b).
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4bdfc580-de51-445b-b170-8c6832e44704",
		Name:         "Epic Fight",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Double target creature's power and toughness until end of turn.",
				effects.TargetCreature("target creature"), epicFightDouble),
			effects.ModeDoing("Target creature you control fights target creature an opponent controls.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				), epicFightFight),
		),
	})
}

func epicFightDouble(_ *game.StackItem, ctx *effects.Context, occurrence int) error {
	target, ok := effects.ModeTarget(ctx, occurrence)
	if !ok || target.Kind != game.TargetCard {
		return nil
	}
	return effects.DoublePowerAndToughnessOfEachUntilEOT(ctx, []uuid.UUID{target.ID}, "Epic Fight — double power and toughness")
}

func epicFightFight(_ *game.StackItem, ctx *effects.Context, occurrence int) error {
	mine, ok := effects.ModeClauseTarget(ctx, occurrence, 0)
	if !ok {
		return nil
	}
	theirs, ok := effects.ModeClauseTarget(ctx, occurrence, 1)
	if !ok {
		return nil
	}
	return effects.B10Fight(ctx, mine.ID, theirs.ID)
}
