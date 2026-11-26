package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Realm Razer — Creature — Beast {3}{R}{G}{W}, 4/2 (Shards of Alara, S82 OLD
// batch 518, #5588):
//
//	"When this creature enters, exile all lands.
//	 When this creature leaves the battlefield, return the exiled cards to the
//	 battlefield tapped under their owners' control."
//
// Worldgorger Dragon's two linked triggers (CR 607.2a) over every land, each
// player's and the Razer controller's own. The enters trigger exiles them in
// one move; the leaves trigger brings back what that ability exiled and is
// still in exile, tapped and as one entry, under their owners' control.
// Tokens cease to exist in exile and do not return. If the Razer leaves
// before its enters trigger resolves, the leaves trigger finds nothing and the
// lands are exiled for good, as the rulings have it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90646d77-9b35-4f87-94ee-255c9f150456",
		Name:         "Realm Razer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepExile, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(realmRazerExileLabel, effects.Do(effects.ExileAllMatching{Match: effects.Land()})),
			effects.On(game.EventLTB, effects.Self, "Realm Razer — return the exiled cards to the battlefield tapped under their owners' control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnFromExileTogether{
						Targets: effects.B27ExiledWith(g, item.SourceCardID, realmRazerExileLabel),
						Tapped:  true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

const realmRazerExileLabel = "Realm Razer — exile all lands"
