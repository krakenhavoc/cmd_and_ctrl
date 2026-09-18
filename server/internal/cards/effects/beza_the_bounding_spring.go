package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beza, the Bounding Spring — Legendary Creature — Elemental Elk
// {2}{W}{W}, 4/5 (EDHREC rank 4110):
//
//	"When Beza enters, create a Treasure token if an opponent
//	 controls more lands than you. You gain 4 life if an opponent has
//	 more life than you. Create two 1/1 blue Fish creature tokens if
//	 an opponent controls more creatures than you. Draw a card if an
//	 opponent has more cards in hand than you."
//
// The catch-up card. Four mana for a 4/5 body is already playable;
// the four riders mean that in a four-player pod, where somebody is
// almost always ahead of you on each axis, Beza usually lands with
// two or three of them live. In a duel where you are winning she is
// a vanilla 4/5, which is the printed design and the reason she is
// a Commander card rather than a Standard one.
//
// FOUR INDEPENDENT "IF AN OPPONENT" CHECKS, EACH AGAINST ANY ONE
// OPPONENT. This is the half a reader gets wrong: it is not "the
// opponent who is ahead" and it is not "all opponents". Each clause
// asks whether SOME single opponent beats you on that one axis, and
// they can be four different players. Hence four separate reads
// rather than one comparison.
//
// STRICTLY MORE, NEVER TIED. "More lands than you" is a strict
// inequality, so an opponent with exactly as many lands gives
// nothing. That is the difference between a Beza that triggers off
// the table's average and one that triggers off a real deficit.
//
// PRINTED ORDER, EVALUATED AS EACH CLAUSE IS CARRIED OUT (CR 608.2).
// The order is observable: the Fish clause counts creatures BEFORE
// the two Fish exist, so Beza never fails her own creature check by
// having already made the tokens; and the draw clause counts your
// hand before the draw. The Treasure and the life gain cannot affect
// any later count.
//
// An eliminated seat is not an opponent (CR 800.4a) and is skipped by
// every one of the four.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "020de6d7-f5a2-4036-ad25-451e5977b4d4",
		Name:         "Beza, the Bounding Spring",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Beza, the Bounding Spring — the four catch-up clauses", b39BezaEntry),
		},
	})
}

// b39BezaEntry is Beza's entry body: the four printed clauses, each
// with its own independent "an opponent beats me" test, carried out
// in printed order.
func b39BezaEntry(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := item.Controller

	if b39AnOpponentControlsMoreLands(g, you) {
		if err := (CreateToken{Controller: you, Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if b39AnOpponentHasMoreLife(g, you) {
		if err := (GainLife{Player: you, Amount: 4}).Apply(ctx); err != nil {
			return err
		}
	}
	if b39AnOpponentControlsMoreCreatures(g, you) {
		if err := (CreateToken{Controller: you, Template: TokenCard("1/1 blue Fish"), N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	if b39AnOpponentHasMoreCardsInHand(g, you) {
		return DrawCards{Player: you, N: 1}.Apply(ctx)
	}
	return nil
}
