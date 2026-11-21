package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akki Ember-Keeper — Enchantment Creature — Goblin Warrior {1}{R}, 2/1
// (Kamigawa: Neon Dynasty):
//
//	"Whenever a nontoken modified creature you control dies, create a 1/1
//	 colorless Spirit creature token. (Equipment, Auras you control, and
//	 counters are modifications.)"
//
// Modified is read as the creature last existed on the battlefield
// (CR 603.10a, msc17WasModified); Akki itself is one of the creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8df42f4a-022c-4fcb-a953-fb165fdbfc13",
		Name:         "Akki Ember-Keeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead) && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					effects.Msc17WasModified(g, ev.CardID)
			}, "Akki Ember-Keeper — create a 1/1 colorless Spirit creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Spirit"), N: 1})),
				game.Purpose{Tokens: 1, CreatureTokens: 1, DeathPayoff: true}),
		},
	})
}
