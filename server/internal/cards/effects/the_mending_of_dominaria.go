package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Mending of Dominaria — Enchantment — Saga {3}{G}{G} (Outlaws of Thunder
// Junction Commander, #5368):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Mill two cards, then you may return a creature card from your
//	 graveyard to your hand.
//	 III — Return all land cards from your graveyard to the battlefield, then
//	 shuffle your graveyard into your library."
//
// I and II mill, then ask for a creature card from the whole graveyard (a
// creature just milled is a legal pick; the pick is optional). III returns
// every land card together (a land that can't enter stays where it is) and,
// once they have entered, shuffles what is left of the graveyard into the
// library: the shuffle is the return's keyed tail, so a replacement prompt on
// an entering land is a restore point and the shuffle waits for it.
//
// No simplification.
func init() {
	chapter12 := func(n int, label string) game.TriggeredAbility {
		return ChapterTrigger(n, label, Do(MillCards{N: 2}, Instruction(func(ctx *Context) error {
			return ReturnChosenFromGraveyard{
				Player:   ctx.Controller(),
				Question: "The Mending of Dominaria — you may return a creature card from your graveyard to your hand",
				Match:    game.Card.IsCreature,
				Max:      1,
				ToHand:   true,
			}.Apply(ctx)
		})))
	}
	Register(Spec{
		OracleID:     "9b0b8f00-6367-4dc9-b33e-680470b3a36c",
		Name:         "The Mending of Dominaria",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			chapter12(1, "The Mending of Dominaria — I: mill two cards, then you may return a creature card from your graveyard to your hand"),
			chapter12(2, "The Mending of Dominaria — II: mill two cards, then you may return a creature card from your graveyard to your hand"),
			ChapterTrigger(3, "The Mending of Dominaria — III: return all land cards from your graveyard to the battlefield, then shuffle your graveyard into your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ReturnFromGraveyardTogether{
						Targets:  graveyardCardIDs(ctx, ctx.Controller(), game.Card.IsLand),
						ThenTail: mendingOfDominariaShuffle,
					}.Apply(ctx)
				}),
		},
	})
}

// mendingOfDominariaShuffle is chapter III's "then shuffle your graveyard into
// your library", run after the lands have entered. An on-disk key.
var mendingOfDominariaShuffle = TogetherTail("cmdr2-b15/the-mending-of-dominaria-shuffle",
	func(ctx *Context, _ []uuid.UUID, _ int) error {
		return finaleShuffleGraveyardIntoLibrary(ctx, ctx.Controller())
	})
