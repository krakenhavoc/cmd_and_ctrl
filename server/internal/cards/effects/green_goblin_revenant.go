package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Green Goblin, Revenant — Legendary Creature — Goblin Human Villain
// {3}{B}{R}, 3/3:
//
//	"Flying, deathtouch
//	 Whenever Green Goblin attacks, discard a card. Then draw a card for
//	 each card you've discarded this turn."
//
// The discard is required (the controller picks the card), and "then"
// is the discard-then continuation, which runs even when the hand was
// empty: the draw is not conditional on this discard. It reads the
// turn's discard tally after the discard, so the card just discarded
// counts with every card discarded earlier this turn (Change of
// Fortune's count, CR 701.9a). No simplification.
func init() {
	Register(Spec{
		OracleID:        "67175889-02f9-4ad8-a9ab-11db44328b67",
		Name:            "Green Goblin, Revenant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Green Goblin, Revenant — discard a card, then draw a card for each card you've discarded this turn",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Green Goblin, Revenant — discard a card",
					}, greenGoblinRevenantDraw, nil)
				}),
		},
	})
}

// greenGoblinRevenantDraw is "Then draw a card for each card you've
// discarded this turn", keyed so a table waiting on the discard is a
// restore point. An on-disk key: never renamed, never reused.
var greenGoblinRevenantDraw = game.RegisterDiscardThen("choose-cards/green-goblin-revenant-draw", func(g *game.Game, r game.DiscardsDone) error {
	n := g.TurnTallyFor(r.Player).CardsDiscarded
	if n == 0 {
		return nil
	}
	return g.DrawNForEffect(r.Player, n)
})
