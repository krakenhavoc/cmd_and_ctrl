package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yavimaya Granger — Creature — Elf, {2}{G}, 2/2:
//
//	"Echo {2}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, you may search your library for a basic land card, put that card onto the battlefield tapped, then shuffle."
//
// The search is Solemn Simulacrum's: a basic land card onto the battlefield
// tapped, then a shuffle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fac9ff8d-8342-4099-b224-d48a044bcc9a",
		Name:         "Yavimaya Granger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Yavimaya Granger", "{2}{G}"),
			Optional(WhenThisEnters("Yavimaya Granger — search for a basic land", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     b30IsBasicLandCard,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Yavimaya Granger — a basic land",
				}.Apply(NewContext(g, item))
			}), "Yavimaya Granger — search for a basic land?"),
		},
	})
}
