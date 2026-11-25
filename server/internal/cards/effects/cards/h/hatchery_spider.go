package cardsh

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hatchery Spider — Creature — Spider {5}{G}{G}, 5/7 (Guilds of Ravnica,
// #5525):
//
//	"Reach
//	 Undergrowth — When you cast this spell, reveal the top X cards of your
//	 library, where X is the number of creature cards in your graveyard. You
//	 may put a green permanent card with mana value X or less from among them
//	 onto the battlefield. Put the rest on the bottom of your library in a
//	 random order."
//
// A cast trigger (CR 603.2): it resolves above the spell, so X is counted
// as the cast trigger resolves, and it resolves even if the spell is
// countered. The cards are revealed (public), the choice is optional, and
// the rest go to the bottom in a random order whether or not one was put.
//
// No simplification.
func init() {
	const label = "Hatchery Spider — reveal the top X cards (X is the number of creature cards in your graveyard); you may put a green permanent card with mana value X or less onto the battlefield; the rest on the bottom in a random order"
	effects.Register(effects.Spec{
		OracleID:        "d1950094-2cc9-48ce-ae1a-0ae78ba252c1",
		Name:            "Hatchery Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.B11CreatureCardsInGraveyard(g, ctx.Controller())
				if x <= 0 {
					return nil
				}
				revealed := g.RevealTopOfLibraryForEffect(ctx.Controller(), item.SourceCardID, x,
					"Hatchery Spider — reveal the top X cards")
				return effects.PutFromLibraryOntoBattlefield{
					Player:   ctx.Controller(),
					Cards:    revealed,
					Match:    effects.And(effects.OfColor("G"), effects.ManaValueLE(x)),
					Max:      1,
					Optional: true,
					Label:    "Hatchery Spider — you may put a green permanent card with mana value " + strconv.Itoa(x) + " or less onto the battlefield",
					ThenTail: effects.PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
			}),
		},
	})
}
