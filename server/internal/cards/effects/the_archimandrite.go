package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Archimandrite — Legendary Creature — Human Advisor {2}{U}{R}{W}, 0/5
// (The Brothers' War Commander):
//
//	"At the beginning of your upkeep, you gain X life, where X is the
//	 number of cards in your hand minus 4.
//	 Whenever you gain life, each Advisor, Artificer, and Monk you control
//	 gains vigilance and gets +X/+0 until end of turn, where X is the
//	 amount of life you gained.
//	 Tap three untapped Advisors, Artificers, and/or Monks you control:
//	 Draw a card."
//
// The upkeep X is Iron Maiden's hand size minus 4 (cardsInHand), read as the
// trigger resolves; zero or less gains nothing (CR 107.1b). The second
// trigger reads the amount of the life-change event; the creatures it affects
// are those you control as it resolves (CR 611.2c). The cost is Skystrike
// Officer's tap-three over three creature types and does not target.
//
// No simplification.
func init() {
	const label = "three untapped Advisors, Artificers, and/or Monks you control"
	isKin := Or(OfCreatureType("Advisor"), OfCreatureType("Artificer"), OfCreatureType("Monk"))
	Register(Spec{
		OracleID:     "e5e8368d-43af-4f83-8565-25d03da64360",
		Name:         "The Archimandrite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("The Archimandrite — you gain X life, where X is the number of cards in your hand minus 4",
				func(g *game.Game, item *game.StackItem) error {
					x := cardsInHand(g, item.Controller) - 4
					if x <= 0 {
						return nil
					}
					return GainLife{Player: item.Controller, Amount: x}.Apply(NewContext(g, item))
				}),
			On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && b10YouGainedLife(ev, source)
			},
				"The Archimandrite — each Advisor, Artificer, and Monk you control gains vigilance and gets +X/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					ctx := NewContext(g, item)
					match := And(Creature(), YouControl(), isKin)
					if err := (BoostUntilEOT{Match: match, Power: item.Trigger.Event.Amount, Label: "The Archimandrite — +X/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: match, Keywords: []string{"vigilance"}, Label: "The Archimandrite — vigilance"}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Tap three untapped Advisors, Artificers, and/or Monks you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: TargetPermanent(label, Creature(), isKin),
				Label:  label,
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
