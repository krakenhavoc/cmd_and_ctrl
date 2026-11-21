package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Szarekh, the Silent King — Legendary Artifact Creature — Necron {1}{B}{B}{B},
// 3/4 (Warhammer 40,000 Commander, #5948):
//
//	"Flying
//	 My Will Be Done — Whenever Szarekh attacks, mill three cards. You may put
//	 an artifact creature card or Vehicle card from among the cards milled this
//	 way into your hand."
//
// Bristlebud Farmer's mill-then-take: only the cards this mill put into the
// graveyard are offered, filtered to artifact creature cards and Vehicle cards,
// and the pick is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a63f950-2e8a-4617-896c-1d0af538dbc7",
		Name:            "Szarekh, the Silent King",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Szarekh — mill three cards, then you may put an artifact creature card or Vehicle card from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(szarekhMilled)}.Apply(ctx)
				}),
		},
	})
}

var szarekhMilled = effects.CardTail("szarekh-the-silent-king/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.MayTakeOneFromAmongThem(ctx, o.IDs, effects.Or(effects.And(effects.Artifact(), effects.Creature()), effects.HasSubtype("Vehicle")),
		"Szarekh — you may put an artifact creature card or Vehicle card from among them into your hand")
})
