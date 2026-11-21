package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b191_helpers.go — shared pieces of the S82 GAP batch 191 cards (#5948).

// S82GapB191OpponentCreatureDiesTrigger is "Whenever a creature an opponent
// controls dies, that player loses N life" (Assault Intercessor). The player
// is the creature's controller as it last existed (CR 603.10a).
func S82GapB191OpponentCreatureDiesTrigger(label string, n int) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventLTB},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			dead, ok := DiedCreature(ev, g)
			return ok && LeftUnderControlOf(ev, dead) != source.Controller
		},
		Key: label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			dead, ok := DiedCreature(ev, g)
			if !ok {
				return nil
			}
			item := game.NewTriggeredItem(source, label)
			item.Params.Player = LeftUnderControlOf(ev, dead)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Params.Player, -n)
		},
	}
}

// S82GapB191DivideSourceCastX is "distribute X <counters> among …" on an enters
// trigger, where X is the X announced for the spell that became the source
// (Broodlord, CR 107.3m): Card.CastX, read as the trigger is put on the stack.
var S82GapB191DivideSourceCastX = game.RegisterDivideAmount("s82-gap-b191/source-cast-x",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		if c, ok := g.LookupCardForEffect(a.Source); ok {
			return game.DivideSpec{Total: c.CastX()}
		}
		return game.DivideSpec{}
	})
