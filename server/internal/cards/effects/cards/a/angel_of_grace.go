package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angel of Grace — Creature — Angel {3}{W}{W}, 5/4 (Ravnica Allegiance):
//
//	"Flash
//	 Flying
//	 When this creature enters, until end of turn, damage that would reduce your life
//	 total to less than 1 reduces it to 1 instead.
//	 {4}{W}{W}, Exile this card from your graveyard: Your life total becomes 10."
//
// The enters trigger grants Angel's Grace's floor until end of turn: the damage stays
// whole, only the life it costs is capped (CR 120.3a). The graveyard ability exiles
// the card as a cost and sets the life total to 10 as it resolves (CR 119.5), as a
// life gain or loss of the difference.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcfcac6b-6d26-4ffe-ae9e-b0ee7fa49787",
		Name:            "Angel of Grace",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Angel of Grace — until end of turn, damage that would reduce your life total to less than 1 reduces it to 1 instead",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					g.GrantLifeFloorForEffect(ctx.Controller(), "Angel of Grace: damage can't reduce your life total below 1", ctx.Source(), g.UntilEndOfTurnDuration())
					return nil
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}{W}, Exile this card from your graveyard: Your life total becomes 10.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}{W}{W}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.B31LifeBecomes(ctx, ctx.Controller(), 10)
			},
		}},
	})
}
