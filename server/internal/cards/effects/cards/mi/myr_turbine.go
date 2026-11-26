package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myr Turbine — Artifact {5} (Mirrodin Besieged, S82 OLD b466, #5576):
//
//	"{T}: Create a 1/1 colorless Myr artifact creature token.
//	 {T}, Tap five untapped Myr you control: Search your library for a Myr
//	 creature card, put it onto the battlefield, then shuffle."
//
// Two abilities. The first is the {T} symbol and a token. The second adds a
// cost of five other untapped Myr the controller names as it is announced
// (not the {T} symbol, CR 302.6: Myr that arrived this turn may pay); the
// Turbine is not a Myr, so it can't pay with itself. The search names a
// quality, so the controller may find nothing (CR 701.23b), and the library
// is shuffled either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9043dcd-234f-4144-80db-1cb903c009ea",
		Name:         "Myr Turbine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Create a 1/1 colorless Myr artifact creature token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 colorless Myr artifact"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{T}, Tap five untapped Myr you control: Search your library for a Myr creature card, put it onto the battlefield, then shuffle.",
				Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
				Cost: effects.Plus(effects.TapCost(), game.AbilityCost{TapOthers: &game.TapOthersCost{
					Count:         5,
					Filter:        effects.TargetPermanent("five untapped Myr you control", effects.OfCreatureType("Myr")),
					ExcludeSource: true,
					Label:         "five untapped Myr you control",
				}}),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					myr := ctx.TextSubtype("Myr")
					return effects.SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(myr) },
						Dest:      game.ZoneBattlefield,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Myr Turbine — a Myr creature card",
					}.Apply(ctx)
				},
			},
		},
	})
}
