package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Green Goblin, Back for More — Legendary Creature — Goblin Human
// Villain {3}{B}{B}, 4/4:
//
//	"Flying, menace
//	 At the beginning of combat on your turn, you may discard a card.
//	 If you do, each opponent discards a card.
//	 Mayhem {3}{B}{B} (You may cast this card from your graveyard for
//	 {3}{B}{B} if you discarded it this turn. Timing rules still apply.)"
//
// The discard is the controller's choice of one card or none; "if you
// do" is the discard-then continuation, which runs once the card has
// really left their hand, and then every opponent chooses and discards
// one (CR 701.9a). Mayhem (CR 702.187, #4628).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "622fbc01-9099-4b6c-b2ce-e41247fc7355",
		Name:             "Green Goblin, Back for More",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying", "menace"},
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{3}{B}{B}")},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Green Goblin — you may discard a card. If you do, each opponent discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						UpTo:     true,
						Question: "Green Goblin — discard a card to make each opponent discard a card?",
					}, eachOpponentDiscardsIfYouDid, nil)
				}),
		},
	})
}

// eachOpponentDiscardsIfYouDid is "If you do, each opponent discards a
// card" after a "you may discard a card": a registered keyed
// continuation (#4264), so the discard prompt is a restore point. The
// opponents are the discarding player's, and each chooses their own card.
var eachOpponentDiscardsIfYouDid = game.RegisterDiscardThen("choose-cards/discard-then-each-opponent-discards",
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 {
			return nil
		}
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == r.Player {
				continue
			}
			g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: p.ID,
				Source: r.Source,
				N:      1,
			})
		}
		return nil
	})
