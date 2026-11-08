package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demonic Ruckus — Enchantment — Aura {1}{R} (Outlaws of Thunder Junction,
// #4468):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has menace and trample.
//	 When this Aura is put into a graveyard from the battlefield, draw a
//	 card.
//	 Plot {R}"
//
// Reach for the Sky's shape: PumpAttached and GrantToAttached for the
// static, and a battlefield-to-graveyard leave trigger (cardDied reads
// the EventLTB's destination, so a bounce or an exile draws nothing).
// The Aura going to the graveyard as a state-based action when its
// creature leaves (CR 704.5m) draws the card too, as printed. Plot is the
// ordinary special action (CR 702.170); the later free cast still
// targets a creature as it is cast (CR 303.4a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "a861ee41-a2c6-4530-9974-82f55ce712b1",
		Name:           "Demonic Ruckus",
		Completeness:   CompletenessFull,
		Targets:        EnchantCreature(),
		Purpose:        game.Purpose{Aura: game.AuraBenefit},
		SpecialActions: []game.SpecialAction{Plot("{R}")},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("menace", "trample"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			}, "Demonic Ruckus — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
