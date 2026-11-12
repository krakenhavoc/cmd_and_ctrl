package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza Assembles the Titans — Enchantment — Saga {3}{W}{W} (#5665):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I — Scry 4, then you may reveal the top card of your library. If a
//	     planeswalker card is revealed this way, put it into your hand.
//	 II — You may put a planeswalker card with mana value 6 or less from
//	      your hand onto the battlefield.
//	 III — You may activate the loyalty abilities of planeswalkers you
//	       control twice this turn rather than only once."
//
// Read ahead is the engine's (CR 702.155, game/read_ahead.go).
//
// Chapter I's reveal comes after "then", so it is the scry's tail: the
// player is asked once the cards are back in the order they chose. The
// "you may" is the reveal, not the move: a revealed planeswalker card
// goes to the hand, and any other card stays on top, revealed. An empty
// library asks nothing.
//
// Chapter II is put, not cast, and the choice is made as it resolves.
// The mana value is the card's in hand, where {X} is 0 (CR 202.3e).
//
// Chapter III is Game.AllowLoyaltyTwiceThisTurnForEffect: Oath of
// Teferi's static as an effect that lasts the turn. It is the player's,
// so a planeswalker that arrives later this turn is covered; a loyalty
// ability activated before the chapter resolved counts as the first of
// the two; and the timing of each activation is still a loyalty
// ability's (CR 606.3).
//
// No simplification.
func init() {
	const name = "Urza Assembles the Titans"
	Register(Spec{
		OracleID:        "0d406496-e1c3-4e1c-976c-618d5ad67d0b",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "scry 4, then you may reveal the top card of your library"),
				urzaAssemblesScry),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "you may put a planeswalker card with mana value 6 or less from your hand onto the battlefield"),
				Do(PutFromHandOntoBattlefield{
					Match:    And(Planeswalker(), ManaValueLE(6)),
					Optional: true,
					Label:    name + " — you may put a planeswalker card with mana value 6 or less from your hand onto the battlefield",
				})),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "you may activate the loyalty abilities of planeswalkers you control twice this turn"),
				urzaAssemblesLoyaltyTwice),
		},
	})
}

// urzaAssemblesScry is chapter I: the scry, with the reveal as its tail.
func urzaAssemblesScry(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return Scry{N: 4, ThenTail: ctx.Tail(urzaAssemblesMayReveal)}.Apply(ctx)
}

// urzaAssemblesMayReveal is "then you may reveal the top card of your
// library", asked once the scry is done. Keyed, so the scry prompt is a
// restore point.
var urzaAssemblesMayReveal = CardTail("then/urza-assembles-the-titans-may-reveal", func(ctx *Context, _ game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
		return nil
	}
	return MayChoice{
		Question: "Urza Assembles the Titans — reveal the top card of your library?",
		Then:     urzaAssemblesReveal,
	}.Apply(ctx)
})

// urzaAssemblesReveal is the "yes": reveal the top card, and a
// planeswalker card revealed this way goes to the hand.
var urzaAssemblesReveal = MayYes("confirm/urza-assembles-the-titans-reveal", func(ctx *Context, r game.ConfirmAnswer) error {
	ids := ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1,
		"Urza Assembles the Titans — revealed from the top of the library")
	if len(ids) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(ids[0])
	if !ok || !isPermanentCard(c) || !c.IsPlaneswalker() {
		return nil
	}
	return ctx.Game.BounceToHandForEffect(ids[0])
})

// urzaAssemblesLoyaltyTwice is chapter III.
func urzaAssemblesLoyaltyTwice(g *game.Game, item *game.StackItem) error {
	g.AllowLoyaltyTwiceThisTurnForEffect(item.Controller)
	return nil
}
