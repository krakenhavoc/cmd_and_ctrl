package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raph & Mikey, Troublemakers — Legendary Creature — Mutant Ninja Turtle
// {5}{R/G}{R/G}, 7/7 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Trample, haste
//	 Whenever Raph & Mikey attack, reveal cards from the top of your
//	 library until you reveal a creature card. Put that card onto the
//	 battlefield tapped and attacking and the rest on the bottom of your
//	 library in a random order."
//
// The reveal is the reveal-until family's (revealUntil, The Regalia's
// sentence); the entry is the shared tapped-and-attacking put
// (putAttackingYourChoice) from the library: its controller chooses
// which player, planeswalker or battle it attacks (CR 508.4), with no
// question when there is one choice. It was never declared as an
// attacker, so its own "whenever this attacks" does not trigger. The
// rest go to the bottom first, so the put that may wait on that
// question finds the creature card still in the library. A library
// with no creature card reveals itself whole and every card goes to the
// bottom in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "82563a5f-cae4-4ece-876c-0df245d2dc58",
		Name:            "Raph & Mikey, Troublemakers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Raph & Mikey, Troublemakers — reveal until a creature card and put it onto the battlefield tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() }, "Raph & Mikey, Troublemakers")
					rest := slices.DeleteFunc(slices.Clone(run), func(id uuid.UUID) bool { return id == hit })
					if err := g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, rest); err != nil {
						return err
					}
					if hit == uuid.Nil {
						return nil
					}
					return putAttackingYourChoice(ctx, hit, game.ZoneLibrary, "Raph & Mikey, Troublemakers")
				}),
		},
	})
}
