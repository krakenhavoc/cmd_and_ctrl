package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blighted Woodland — Land (EDHREC rank 915):
//
//	"{T}: Add {C}.
//	 {3}{G}, {T}, Sacrifice this land: Search your library for up to
//	 two basic land cards, put them onto the battlefield tapped, then
//	 shuffle."
//
// Myriad Landscape without the enters-tapped and the shared-type
// clause, for a mana more. The same activated shape: mana, tap and
// sacrifice paid at announce, so the land is gone before the ability
// resolves and a bounce in response gets nothing back; the search is
// the S22 chooser with a limit of two and no set constraint.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "02679a2e-303d-412f-87d8-0a37a8ca259c",
		Name:          "Blighted Woodland",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, {T}, Sacrifice this land: Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         2,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Blighted Woodland — up to two basic lands, onto the battlefield tapped",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
