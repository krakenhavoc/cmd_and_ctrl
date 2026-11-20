package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm the Festival — Sorcery {3}{G}{G}{G} (Innistrad: Midnight Hunt,
// #5497):
//
//	"Look at the top five cards of your library. You may put up to two
//	 permanent cards with mana value 5 or less from among them onto the
//	 battlefield. Put the rest on the bottom of your library in a random
//	 order.
//	 Flashback {7}{G}{G}{G}"
//
// LookAtTopThenMayPutOntoBattlefield with a maximum of two: the look is
// private to the caster, the put is optional ("up to two", so none is
// allowed), and everything not put goes to the bottom in a random order.
// "Permanent card" is any card that is a permanent card in a library
// (an instant or sorcery is not), and the mana value is the card's own
// (an X in a cost is 0 there, CR 202.3e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "5d03188b-fbc4-4e16-9fe6-db98dbdb7f68",
		Name:             "Storm the Festival",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{7}{G}{G}{G}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.LookAtTopThenMayPutOntoBattlefield(5, effects.And(effects.Permanent(), effects.ManaValueLE(5)), 2,
				"Storm the Festival — put up to two permanent cards with mana value 5 or less onto the battlefield")(ctx.Game, item)
		},
	})
}
