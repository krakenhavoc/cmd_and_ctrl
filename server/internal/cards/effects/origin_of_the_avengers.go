package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Origin of the Avengers — Enchantment — Saga {1}{W} (Marvel Super Heroes,
// #4455):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Scry 2.
//	 II — You may put a Hero creature card with mana value 3 or less from
//	 your hand onto the battlefield. If you don't, draw a card.
//	 III — Put a +1/+1 counter on each creature you control."
//
// I is the Scry primitive. II is PutFromHandOntoBattlefield, optional and
// limited to Hero creature cards with mana value 3 or less; the "if you
// don't" is its registered tail, which draws when nothing arrived: a
// declined offer, a hand with no such card, and an entry a replacement
// cancelled all count as not putting a card onto the battlefield. III
// snapshots the creatures first, so a counter payoff cannot hand a counter
// to a token made mid-loop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7f3178d-7b7c-4346-af50-a7b0206bee22",
		Name:         "Origin of the Avengers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("Origin of the Avengers", 1, "scry 2"), Do(Scry{N: 2})),
			ChapterTrigger(2, SagaChapterLabel("Origin of the Avengers", 2, "put a Hero from your hand onto the battlefield or draw a card"), mshA6OriginChapterTwo),
			ChapterTrigger(3, SagaChapterLabel("Origin of the Avengers", 3, "a +1/+1 counter on each creature you control"), b11PutCounterOnEachCreatureYouControl),
		},
	})
}

// mshA6OriginChapterTwo is chapter II: the optional put, then the draw if
// no card arrived.
func mshA6OriginChapterTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return PutFromHandOntoBattlefield{
		Match:    And(Creature(), Subtype("Hero"), ManaValueLE(3)),
		Optional: true,
		Label:    "Origin of the Avengers — you may put a Hero creature card with mana value 3 or less from your hand onto the battlefield",
		ThenTail: mshA6OriginDrawIfNone,
	}.Apply(ctx)
}

// mshA6OriginDrawIfNone is "If you don't, draw a card": the rider on the
// put. An on-disk key: never renamed, never reused.
var mshA6OriginDrawIfNone = PutFromHandTail("msh-a6/origin-of-the-avengers-draw", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered != uuid.Nil {
		return nil
	}
	return DrawCards{Player: res.Player, N: 1}.Apply(ctx)
})
