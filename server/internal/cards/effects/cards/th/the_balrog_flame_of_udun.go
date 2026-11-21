package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Balrog, Flame of Udûn — Legendary Creature — Avatar Demon {3}{B}{R},
// 7/7 (The Lord of the Rings: Tales of Middle-earth):
//
//	"Trample
//	 When a legendary creature an opponent controls dies, put The Balrog on
//	 the bottom of its owner's library."
//
// A dies trigger over any creature (diedCreature), narrowed to a legendary
// one that left the battlefield under an opponent's control (CR 603.10a's
// last-known information, leftUnderControlOf). The ability does nothing if
// the Balrog is no longer on the battlefield as it resolves (CR 608.2b,
// 400.7); otherwise it is put on the bottom of its owner's library, through
// the tuck route so a commander is offered the command zone (CR 903.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de12ca82-4824-4975-9a31-309a6d9101a7",
		Name:            "The Balrog, Flame of Udûn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				if !ok || !dead.IsLegendary() {
					return false
				}
				who := effects.LeftUnderControlOf(ev, dead)
				return who != source.Controller
			}, "The Balrog, Flame of Udûn — put The Balrog on the bottom of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.PutIntoLibrary{Card: item.SourceCardID, ToBottom: true}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
