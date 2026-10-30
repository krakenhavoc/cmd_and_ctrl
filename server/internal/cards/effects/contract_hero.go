package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contract Hero — Creature — Human Mercenary Hero {2}{R}, 2/3:
//
//	"When this creature enters, create a Treasure token.
//	 Whenever this creature attacks, you may sacrifice an artifact or
//	 discard a card. If you do, this creature gets +2/+0 until end of
//	 turn."
//
// The attack trigger is #4566's either/or may-pay. The creature it pumps
// rides the prompt as an object with its epoch, so a Contract Hero that
// left and came back while the question was open is a new object and
// gets nothing (CR 400.7).
//
// No simplification.
func init() {
	const label = "Contract Hero — you may sacrifice an artifact or discard a card; if you do, +2/+0 until end of turn"
	Register(Spec{
		OracleID:     "a5c9b167-5824-4fc8-ba7a-ffa64363c029",
		Name:         "Contract Hero",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Contract Hero — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
			WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				ref, ok := g.PermanentRefForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				return mayPayArtifactOrCard(NewContext(g, item),
					"Contract Hero — sacrifice an artifact or discard a card for +2/+0 until end of turn?", contractHeroPaid, ref)
			}),
		},
	})
}

// contractHeroPaid is "If you do, this creature gets +2/+0 until end of
// turn", while it is still the creature that attacked.
var contractHeroPaid = PayUnlessBranches("pay-unless/contract-hero-pump",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Objects) != 1 {
			return nil
		}
		ref := r.Objects[0]
		c, ok := ctx.Game.LookupCardForEffect(ref.ID)
		if !ok || !onBattlefield(ctx.Game, ref.ID) || c.ObjectEpoch != ref.Epoch {
			return nil
		}
		return BoostUntilEOT{Target: ref.ID, Power: 2, Label: "Contract Hero — +2/+0 until end of turn"}.Apply(ctx)
	}, nil)
