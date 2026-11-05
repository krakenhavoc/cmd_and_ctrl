package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// All Good Things... — Enchantment — Saga {3}{G}{U} (Star Trek, #4973):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III — Draw two cards, then exile a card from your hand.
//	 IV — If all cards exiled with this Saga share a card type, you gain
//	 10 life."
//
// "I, II, III —" is three chapter abilities with one text (CR 714.2c).
// Each draws two, then asks for one card of the hand to exile, linked to
// the Saga object (CR 607.2a). Chapter IV reads the cards in exile that
// this Saga exiled and gains 10 life when one card type is common to all
// of them (trekB23AllGoodThingsFinale); a card that has left exile is a
// new object and no longer one of them (CR 400.7).
//
// No simplification.
func init() {
	const name = "All Good Things..."
	Register(Spec{
		OracleID:     "61dc9770-1e29-49fa-a31b-b63b5f0496fb",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 6},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "draw two cards, then exile a card from your hand"), trekB23DrawTwoThenExileFromHand),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "draw two cards, then exile a card from your hand"), trekB23DrawTwoThenExileFromHand),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "draw two cards, then exile a card from your hand"), trekB23DrawTwoThenExileFromHand),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "if all cards exiled with this Saga share a card type, you gain 10 life"), trekB23AllGoodThingsFinale),
		},
	})
}
