package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Once Upon a Time — Instant {1}{G}:
//
//	"If this spell is the first spell you've cast this game, you may cast it
//	 without paying its mana cost.
//	 Look at the top five cards of your library. You may reveal a creature or
//	 land card from among them and put it into your hand. Put the rest on the
//	 bottom of your library in a random order."
//
// The free cast is an alternative cost offered only while the caster has cast
// no spell this game (Player.FirstSpellCast, stamped as a cast completes, CR
// 601.2i), judged as the spell is cast, as the printed words say. The look is
// Sylvan-style: look at five, take one creature or land, the rest to the
// bottom in random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d00ade82-caab-43c6-9756-6e9e6edfc519",
		Name:         "Once Upon a Time",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		AlternativeCosts: []game.AlternativeCost{{
			Key:   "free",
			Label: "Cast without paying its mana cost (the first spell you've cast this game)",
			Condition: func(g *game.Game, controller uuid.UUID) bool {
				p := g.PlayerByIDForEffect(controller)
				return p != nil && p.FirstSpellCast.ID == uuid.Nil
			},
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    Or(Creature(), Land()),
				Max:      1,
				Optional: true,
				Label:    "Once Upon a Time — reveal a creature or land card and put it into your hand",
				Key:      takeRevealRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
