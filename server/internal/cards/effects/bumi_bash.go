package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bumi Bash — Sorcery {3}{R}:
//
//	"Choose one —
//	 • Bumi Bash deals damage equal to the number of lands you control
//	   to target creature.
//	 • Destroy target land creature or nonbasic land."
//
// Abrade's two-mode shape. The first mode counts the lands its
// controller controls as it resolves (an earthbent land creature is
// still a land and counts). The second mode's target is a permanent
// that is both a land and a creature, or any land without the basic
// supertype; a basic land that is not a creature is not a legal target.
// An illegal target makes the spell do nothing (CR 608.2b).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "629f247c-1e64-4c32-8892-0716ed469d1c",
		Name:         "Bumi Bash",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			Mode("Bumi Bash deals damage equal to the number of lands you control to target creature.",
				TargetCreature("target creature")),
			Mode("Destroy target land creature or nonbasic land.",
				TargetPermanent("target land creature or nonbasic land",
					And(Land(), Or(Creature(), notABasicLandCard())))),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ModeTarget(ctx, 0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			target := t.ID
			switch {
			case ctx.HasMode(0):
				lands := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
				return DealDamage{Source: ctx.Source(), Target: target, Amount: lands}.Apply(ctx)
			case ctx.HasMode(1):
				return DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
