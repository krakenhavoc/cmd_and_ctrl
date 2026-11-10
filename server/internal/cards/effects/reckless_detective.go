package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Detective — Creature — Devil Detective {1}{R}, 0/3 (Murders at
// Karlov Manor, #4469):
//
//	"Whenever this creature attacks, you may sacrifice an artifact or
//	 discard a card. If you do, draw a card and this creature gets +2/+0
//	 until end of turn."
//
// Contract Hero's attack trigger (#4566's either/or may-pay) with a draw
// beside the pump. The creature rides the prompt as an object with its
// epoch, so one that left and came back meanwhile is a new object and gets
// no pump (CR 400.7); the draw still happens, since the cost was paid.
//
// No simplification.
func init() {
	const label = "Reckless Detective — you may sacrifice an artifact or discard a card; if you do, draw a card and +2/+0 until end of turn"
	Register(Spec{
		OracleID:     "4debd0e4-3bb6-4dd1-addc-08e4bda69d11",
		Name:         "Reckless Detective",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				ref, ok := g.PermanentRefForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				return mayPayArtifactOrCard(NewContext(g, item),
					"Reckless Detective — sacrifice an artifact or discard a card to draw a card and get +2/+0?", recklessDetectivePaid, ref)
			}),
		},
	})
}

// recklessDetectivePaid is "If you do, draw a card and this creature gets
// +2/+0 until end of turn", the pump only while it is still the creature
// that attacked.
var recklessDetectivePaid = PayUnlessBranches("pay-unless/reckless-detective-draw-pump",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
			return err
		}
		if len(r.Objects) != 1 {
			return nil
		}
		ref := r.Objects[0]
		c, ok := ctx.Game.LookupCardForEffect(ref.ID)
		if !ok || !onBattlefield(ctx.Game, ref.ID) || c.ObjectEpoch != ref.Epoch {
			return nil
		}
		return BoostUntilEOT{Target: ref.ID, Power: 2, Label: "Reckless Detective — +2/+0 until end of turn"}.Apply(ctx)
	}, nil)
