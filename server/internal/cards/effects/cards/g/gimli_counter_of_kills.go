package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gimli, Counter of Kills — Legendary Creature — Dwarf Warrior {3}{R}, 4/3
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Trample
//	 Whenever a creature an opponent controls dies, Gimli deals 1 damage to
//	 that creature's controller."
//
// Trample is a printed keyword. The trigger is Haunted Library's "a
// creature an opponent controls dies" watch; the damage goes to the player
// who controlled the creature as it died (its last-known controller,
// CR 603.10a), and Gimli is the source, which is last-known if he has left
// by the time it resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "96ebcd8b-ae3c-47b0-b895-bfc9848713e5",
		Name:            "Gimli, Counter of Kills",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B18OpponentsCreatureDied(ev, source, g)
			}, "Gimli, Counter of Kills — 1 damage to that creature's controller",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ev := item.Trigger.Event
					dead, ok := effects.DiedCreature(ev, g)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: effects.LeftUnderControlOf(ev, dead), Amount: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
