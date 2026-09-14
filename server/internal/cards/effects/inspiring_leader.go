package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiring Leader — Legendary Enchantment — Background {2}{W}
// (EDHREC rank 2838):
//
//	"Commander creatures you own have "Creature tokens you control
//	 get +2/+2.""
//
// The go-wide Background. The printed card GRANTS a static to the
// commander; the layer engine can grant keywords but not a whole
// anthem with its own "you", so the anthem lives on the Background
// itself, gated the way the grant implies: a creature token gets
// +2/+2 when ITS controller controls a creature that is a commander
// the Background's controller OWNS (b26ControlsCommanderCreatureOwnedBy).
// That reads the printed card exactly, stolen commander included: a
// commander an opponent has taken carries the granted ability for
// the thief, so the thief's tokens grow and the owner's do not. With
// no such commander on the battlefield the ability does not exist,
// as printed. "Choose a Background" is a deck-construction rule (CR
// 702.124), the deck importer's business.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d46dcc75-55ff-4226-a392-a49755d269d2",
		Name:         "Inspiring Leader",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && IsToken(*target) &&
				b26ControlsCommanderCreatureOwnedBy(g, target.Controller, source.Controller)
		}, 2, 2)},
	})
}
