package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// T'Chaka, Venerable King — Legendary Creature — Human Noble Hero {G}{W},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"When T'Chaka enters, mill three cards, then you may put an artifact
//	 or land card from among the milled cards into your hand.
//	 {3}, Exile this card from your graveyard: You become the monarch.
//	 Activate only if you control your commander."
//
// The enter trigger is Scout the City's mill-then-take (MillToZone with a
// tail that offers one matching card from among the milled cards, still in
// the graveyard). The graveyard ability exiles its own card as the cost
// (CR 113.6). "Your commander" is a commander its owner controls
// (b18ControlsYourCommander), read at activation (CR 602.1b) and never
// again at resolution: a commander an opponent has stolen, or one in the
// command zone, does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a9f2d69-328c-46f4-be37-60b85b197b72",
		Name:         "T'Chaka, Venerable King",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("T'Chaka, Venerable King — mill three cards, then you may put an artifact or land card from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(tchakaMilled)}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, Exile this card from your graveyard: You become the monarch. Activate only if you control your commander.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.B18ControlsYourCommander(g, controller)
			},
			Effect: effects.Do(effects.BecomeTheMonarch{}),
		}},
	})
}

var tchakaMilled = effects.CardTail("tchaka/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.Or(effects.Artifact(), effects.Land()),
		"T'Chaka, Venerable King — you may put an artifact or land card from among them into your hand")
})
