package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ojutai Exemplars — Creature — Human Monk {2}{W}{W}, 4/4 (Foundations
// Jumpstart):
//
//	"Whenever you cast a noncreature spell, choose one —
//	 • Tap target creature.
//	 • This creature gains first strike and lifelink until end of turn.
//	 • Exile this creature, then return it to the battlefield tapped
//	   under its owner's control."
//
// A modal cast trigger (Gandalf the Grey's shape): the bullet is chosen
// as the trigger goes on the stack, and only the tap bullet targets. The
// third bullet is a flicker of the Exemplars themselves: it comes back
// as a new object, tapped, under its owner's control.
//
// No simplification.
func init() {
	t := WheneverYouCast(Noncreature(), "Ojutai Exemplars — choose one",
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Tap target creature.", TargetCreature("target creature"), TapTheModesTarget),
		ModeDoing("This creature gains first strike and lifelink until end of turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike", "lifelink"}, Label: "Ojutai Exemplars — first strike and lifelink"}.Apply(ctx)
			}),
		ModeDoing("Exile this creature, then return it to the battlefield tapped under its owner's control.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				c, ok := ctx.Game.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				return Flicker{Target: item.SourceCardID, Controller: c.Owner, Tapped: true}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "4405db13-7b36-4dc2-9a5d-1d31d568abf5",
		Name:         "Ojutai Exemplars",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
