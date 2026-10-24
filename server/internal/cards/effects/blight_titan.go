package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blight Titan — Creature — Phyrexian Giant {4}{B}{B}, 6/6 (March of the
// Machine):
//
//	"Deathtouch
//	 Whenever this creature enters or attacks, mill two cards, then
//	 incubate X, where X is the number of creature cards in your
//	 graveyard."
//
// One ability with two conditions. The count is taken after the mill,
// as the instruction says "then", and includes the creature cards that
// were already there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f2c75577-c9be-48f3-8398-3020c79bd34e",
		Name:            "Blight Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Blight Titan — mill two cards, then incubate X, where X is the number of creature cards in your graveyard",
				// The incubate waits for a prompt the mill asks (a replacement offer; CR 608.2c, #4155)
				// and counts the graveyard when it runs.
				Do(
					Instruction(func(ctx *Context) error {
						return MillCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						n := g2CardsInYourGraveyard(func(c game.Card) bool { return c.IsCreature() })(ctx.Item, ctx)
						return Incubate{N: n}.Apply(ctx)
					}),
				)),
		},
	})
}
