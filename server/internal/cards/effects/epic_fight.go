package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "4bdfc580-de51-445b-b170-8c6832e44704",
		Name:         "Epic Fight",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Double target creature's power and toughness until end of turn.",
				TargetCreature("target creature"), epicFightDouble),
			ModeDoing("Target creature you control fights target creature an opponent controls.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature an opponent controls", OpponentControls()),
				), epicFightFight),
		),
	})
}

func epicFightDouble(_ *game.StackItem, ctx *Context, occurrence int) error {
	target, ok := ModeTarget(ctx, occurrence)
	if !ok || target.Kind != game.TargetCard {
		return nil
	}
	return DoublePowerAndToughnessOfEachUntilEOT(ctx, []uuid.UUID{target.ID}, "Epic Fight — double power and toughness")
}

func epicFightFight(_ *game.StackItem, ctx *Context, occurrence int) error {
	mine, ok := ModeClauseTarget(ctx, occurrence, 0)
	if !ok {
		return nil
	}
	theirs, ok := ModeClauseTarget(ctx, occurrence, 1)
	if !ok {
		return nil
	}
	return b10Fight(ctx, mine.ID, theirs.ID)
}
