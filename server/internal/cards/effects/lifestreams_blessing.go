package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lifestream's Blessing — Instant {4}{G}{G} (#3656):
//
//	"Draw X cards, where X is the greatest power among creatures you
//	 controlled as you cast this spell. If this spell was cast from
//	 exile, you gain twice X life.
//	 Foretell {4}{G}"
//
// Foretell is one special action. The life half reads the spell's cast
// zone, so any cast from exile gains the life, as printed.
//
// Simplification, declared: X is the greatest power among your
// creatures as the spell resolves, not as you cast it. The engine does
// not record the board at cast time, so a creature that died or grew in
// response changes the answer.
func init() {
	Register(Spec{
		OracleID:     "7cbdba72-ed33-40c6-a9c0-4d65c280a4d0",
		Name:         "Lifestream's Blessing",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The cards drawn count the greatest power among your creatures when the spell resolves, not when you cast it.",
		},
		SpecialActions: []game.SpecialAction{
			Foretell("{4}{G}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := b42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
			if x < 0 {
				x = 0
			}
			if err := (DrawCards{Player: ctx.Controller(), N: x}).Apply(ctx); err != nil {
				return err
			}
			if item.CastFromZone != game.ZoneExile {
				return nil
			}
			return GainLife{Player: ctx.Controller(), Amount: 2 * x}.Apply(ctx)
		},
	})
}
