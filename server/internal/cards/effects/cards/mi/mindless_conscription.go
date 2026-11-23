package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindless Conscription — Enchantment {2}{B} (Modern Horizons 3):
//
//	"When this enchantment enters and whenever you draw your third card
//	 each turn, amass Zombies 3. (Put three +1/+1 counters on an Army you
//	 control. It's also a Zombie. If you don't control an Army, create a
//	 0/0 black Zombie Army creature token first.)"
//
// One printed ability with two conditions is one registered trigger over
// two events (Astrologian's Planisphere's shape): the enchantment's own
// entry, and the third draw of the turn (the tally is bumped before the
// harvester runs, so the third draw reports exactly three). Amass is the
// shared keyword action (CR 701.47a): the Army token is made only if you
// control none, and the Army becomes a Zombie in addition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf65e547-9d81-4f4b-8adb-46aee9088906",
		Name:         "Mindless Conscription",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 3},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventDrawCard},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB {
						return ev.CardID == source.InstanceID
					}
					return ev.Actor == source.Controller && g.TurnTallyFor(source.Controller).CardsDrawn == 3
				},
				"Mindless Conscription — amass Zombies 3",
				effects.Do(effects.Amass{Subtype: "Zombie", N: 3})),
		},
	})
}
