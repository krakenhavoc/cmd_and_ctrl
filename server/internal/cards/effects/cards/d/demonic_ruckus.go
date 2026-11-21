package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Ruckus — Enchantment — Aura {1}{R} (Outlaws of Thunder Junction):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has menace and trample.
//	 When this Aura is put into a graveyard from the battlefield, draw a
//	 card.
//	 Plot {R}"
//
// An Aura's static pump and keyword grant (Alien Symbiosis' shape), the
// Rancor dies-to-graveyard trigger with a draw in place of the return, and
// plot, the ordinary special action (CR 702.170). A plotted Aura is cast
// for free on a later turn and chooses what it enchants as it is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "a861ee41-a2c6-4530-9974-82f55ce712b1",
		Name:           "Demonic Ruckus",
		Completeness:   effects.CompletenessFull,
		Targets:        effects.EnchantCreature(),
		Purpose:        game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		SpecialActions: []game.SpecialAction{effects.Plot("{R}")},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("menace", "trample"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source)
			},
			Key:    "Demonic Ruckus — draw a card",
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
