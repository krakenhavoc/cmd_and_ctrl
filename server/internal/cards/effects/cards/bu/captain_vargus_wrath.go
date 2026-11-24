package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captain Vargus Wrath — Legendary Creature — Orc Pirate {U}{R}, 1/1
// (Commander Legends):
//
//	"Whenever Captain Vargus Wrath attacks, Pirates you control get +1/+1 until
//	 end of turn for each time you've cast a commander from the command zone
//	 this game."
//
// Echo Storm's count (the sum of Player.CommanderCasts, which only a cast from
// the command zone bumps), read as the trigger resolves. The Pirates are the
// creatures with the Pirate type you control as it resolves, Vargus Wrath
// included, locked then (CR 611.2c); none cast means nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e73d72f6-5572-47e0-90c9-dc1c146029d7",
		Name:         "Captain Vargus Wrath",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Captain Vargus Wrath — Pirates you control get +1/+1 until end of turn for each time you've cast a commander from the command zone",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.SocB3CommanderCastsFromCommandZone(g, item.Controller)
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.Subtype("Pirate"), effects.YouControl()),
						Power:     n,
						Toughness: n,
						Label:     "Captain Vargus Wrath — Pirates get +1/+1 for each commander cast from the command zone",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
