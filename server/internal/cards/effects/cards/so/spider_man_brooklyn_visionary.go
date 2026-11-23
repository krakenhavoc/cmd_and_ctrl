package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spider-Man, Brooklyn Visionary — Legendary Creature — Spider Human Hero
// {4}{G}, 4/3 (Marvel's Spider-Man, #4460):
//
//	"Web-slinging {2}{G} (You may cast this spell for {2}{G} if you also
//	 return a tapped creature you control to its owner's hand.)
//	 When Spider-Man enters, search your library for a basic land card,
//	 put it onto the battlefield tapped, then shuffle."
//
// Web-slinging is Spider-Sense's alternative cost (the returned creature is
// a cost, chosen as the spell is cast). The fetch is Molten Man's basic-land
// search without the Mountain restriction: the land arrives tapped and the
// library is shuffled even when no basic land is found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ca85872-9436-4515-af50-c13e59605627",
		Name:         "Spider-Man, Brooklyn Visionary",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {2}{G}",
			ManaCost:     "{2}{G}",
			ReturnToHand: effects.PermanentYouControl("a tapped creature you control", effects.Creature(), effects.B751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spider-Man, Brooklyn Visionary — search for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Spider-Man, Brooklyn Visionary — a basic land card, onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
