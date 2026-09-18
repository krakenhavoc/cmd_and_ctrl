package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Confluence — Instant {3}{U}{U} (EDHREC rank 1462):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Counter target spell unless its controller pays {3}.
//	 • Return target creature to its owner's hand.
//	 • Draw a card."
//
// CR 700.2d, the rule the engine could not express before #764:
// validateModes rejected a repeated index outright, so "draw three
// cards" — the mode this card is most often cast for — was
// unannounceable. ChooseNRepeating sets ModeSpec.Repeatable, the
// announced list becomes a MULTISET in the order chosen ([2, 2, 2]),
// and each occurrence gets its own target group, so "bounce two
// creatures and draw a card" is [1, 1, 2] with two separate bounce
// targets rather than one target the second bounce has to share.
//
// Each bullet's body runs once per occurrence in announce order
// (CR 700.2c), which is what makes three draws three draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cd11c27b-9368-4622-bc5a-2e0b993cc42b",
		Name:         "Mystic Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			ModeDoing("Counter target spell unless its controller pays {3}.",
				TargetSpell("target spell"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					stackID := t.ID
					// Read the victim's controller BEFORE anything
					// touches the stack — the Daze posture.
					target := ctx.Game.StackItemForEffect(stackID)
					if target == nil {
						return nil
					}
					return PayUnless{
						Chooser:  target.Controller,
						Cost:     "{3}",
						Question: "Mystic Confluence — pay {3} or your spell is countered",
						OnDecline: func(ctx *Context) error {
							return CounterTarget{StackID: stackID}.Apply(ctx)
						},
					}.Apply(ctx)
				}),
			ModeDoing("Return target creature to its owner's hand.",
				TargetCreature("target creature"),
				BounceTheModesTarget),
			ModeDoing("Draw a card.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				}),
		),
	})
}
