package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "0efe1357-bfbb-42c0-9cc9-6a919d686c66",
		Name:          "Go Nuts!",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(3)},
		Modes: ChooseOne(
			Mode("Put a +1/+1 counter on target creature.",
				TargetCreature("target creature")),
			Mode("Target creature you control fights target creature an opponent controls.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature an opponent controls", OpponentControls()),
				)),
		).InsteadIf(2, TeamworkUsed),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			counterOcc, fightOcc := -1, -1
			for occ, m := range ctx.Modes() {
				switch m {
				case 0:
					counterOcc = occ
				case 1:
					fightOcc = occ
				}
			}
			fight := func(g *game.Game) error {
				if fightOcc < 0 {
					return nil
				}
				c := NewContext(g, item)
				mine, ok := ModeClauseTarget(c, fightOcc, 0)
				if !ok {
					return nil
				}
				theirs, ok := ModeClauseTarget(c, fightOcc, 1)
				if !ok {
					return nil
				}
				return b10Fight(c, mine.ID, theirs.ID)
			}
			if counterOcc >= 0 {
				if t, ok := ModeTarget(ctx, counterOcc); ok {
					return ctx.Game.AddCounterThenForEffect(t.ID, game.CounterPlusOne, 1,
						func(g *game.Game, _ int) error { return fight(g) })
				}
			}
			return fight(ctx.Game)
		},
	})
}
