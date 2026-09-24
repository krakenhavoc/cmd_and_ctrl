package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beastbond Outcaster — {2}{G} Creature — Human Druid 3/3:
//
//	"When this creature enters, if you control a creature with power 4
//	 or greater, draw a card.
//	 Plot {1}{G} (You may pay {1}{G} and exile this card from your
//	 hand. Cast it as a sorcery on a later turn without paying its mana
//	 cost. Plot only as a sorcery.)"
//
// A plot proof card (#1342). The enters trigger is an intervening-if
// (CR 603.4), checked when the creature enters AND again on
// resolution, the read Colossal Majesty makes: current power, so a
// counter or an anthem counts. The Outcaster itself is a 3/3 and does
// not satisfy its own condition unless something pumps it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "19a32e29-45ec-433e-9cb4-b2c32cac8f80",
		Name:           "Beastbond Outcaster",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && youControlPowerFourOrGreater(g, source.Controller)
			}, "Beastbond Outcaster — draw a card", drawIfYouControlPowerFourOrGreater),
		},
	})
}
