package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Go Nuts! — {G} Sorcery:
//
//	"Teamwork 3 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 3 or more.)
//	 Choose one. If this spell was cast using teamwork, choose both
//	 instead.
//	 • Put a +1/+1 counter on target creature.
//	 • Target creature you control fights target creature an opponent
//	   controls."
//
// #1703: Teamwork(3) with InsteadIf(2, TeamworkUsed). Printed order
// matters here (CR 608.2c): with both bullets the counter lands before
// the fight, so a creature that gets the counter and then fights
// fights with the extra power. The counter placement can pause on a
// CR 616 ordering prompt (Doubling Season beside Hardened Scales), so
// the fight is the placement's continuation rather than the next line
// (#1282). The fight is b10Fight. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "0efe1357-bfbb-42c0-9cc9-6a919d686c66",
		Name:          "Go Nuts!",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Teamwork(3)},
		Modes: effects.ChooseOne(
			effects.Mode("Put a +1/+1 counter on target creature.",
				effects.TargetCreature("target creature")),
			effects.Mode("Target creature you control fights target creature an opponent controls.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				)),
		).InsteadIf(2, effects.TeamworkUsed),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			counterOcc, fightOcc := -1, -1
			for occ, m := range ctx.Modes() {
				switch m {
				case 0:
					counterOcc = occ
				case 1:
					fightOcc = occ
				}
			}
			if counterOcc >= 0 {
				if t, ok := effects.ModeTarget(ctx, counterOcc); ok {
					// The fight waits for the counter (#4253 slice 7): the mode's
					// occurrence rides the tail as its N.
					return ctx.Game.AddCounterByKeyedForEffect(uuid.Nil, t.ID, game.CounterPlusOne, 1,
						effects.WithN(ctx.ItemTail(goNutsFight), fightOcc))
				}
			}
			return goNutsFightClause(ctx, fightOcc)
		},
	})
}

// goNutsFightClause is "target creature you control fights target creature an
// opponent controls", if that mode was chosen (occ >= 0).
func goNutsFightClause(c *effects.Context, fightOcc int) error {
	if fightOcc < 0 {
		return nil
	}
	mine, ok := effects.ModeClauseTarget(c, fightOcc, 0)
	if !ok {
		return nil
	}
	theirs, ok := effects.ModeClauseTarget(c, fightOcc, 1)
	if !ok {
		return nil
	}
	return effects.B10Fight(c, mine.ID, theirs.ID)
}

// goNutsFight is the fight behind the counter: the tail's N is the fight
// mode's occurrence, -1 for none. An on-disk key: never renamed, never reused.
var goNutsFight = effects.CardTail("go-nuts/fight", func(ctx *effects.Context, o game.TailOutcome) error {
	return goNutsFightClause(ctx, o.Tail.N)
})
