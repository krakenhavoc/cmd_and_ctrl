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
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (MillCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
						return err
					}
					n := g2CardsInYourGraveyard(func(c game.Card) bool { return c.IsCreature() })(item, ctx)
					return Incubate{N: n}.Apply(ctx)
				}),
		},
	})
}
