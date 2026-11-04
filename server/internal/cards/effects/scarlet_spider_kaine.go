package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scarlet Spider, Kaine — Legendary Creature — Spider Human Hero {B}{R},
// 2/1 (Marvel's Spider-Man, #4460):
//
//	"Menace
//	 When Scarlet Spider enters, you may discard a card. If you do, put a
//	 +1/+1 counter on him.
//	 Mayhem {B/R} (You may cast this card from your graveyard for {B/R}
//	 if you discarded it this turn. Timing rules still apply.)"
//
// The discard is the controller's choice of one card or none; "if you do"
// is the discard-then continuation, which runs once the card has really
// left their hand (CR 701.9a). The counter goes on only if Kaine is still
// on the battlefield. Mayhem (CR 702.187, #4628).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4ec8a5ea-b750-4711-8843-2bc98f4ce060",
		Name:             "Scarlet Spider, Kaine",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{B/R}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Scarlet Spider — you may discard a card. If you do, put a +1/+1 counter on him",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						UpTo:     true,
						Question: "Scarlet Spider — discard a card to put a +1/+1 counter on him?",
					}, scarletSpiderCounterIfDiscarded, nil)
				}),
		},
	})
}

// scarletSpiderCounterIfDiscarded is "If you do, put a +1/+1 counter on
// him": a registered keyed continuation (#4264), so the discard prompt is
// a restore point. Nothing happens if Kaine has left the battlefield.
var scarletSpiderCounterIfDiscarded = game.RegisterDiscardThen("choose-cards/discard-then-scarlet-spider-counter",
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 {
			return nil
		}
		return nothingIfGone(g.AddCounterByForEffect(r.Player, r.Source, game.CounterPlusOne, 1))
	})
