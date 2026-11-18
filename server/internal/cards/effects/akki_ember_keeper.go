package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8df42f4a-022c-4fcb-a953-fb165fdbfc13",
		Name:         "Akki Ember-Keeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == source.Controller &&
					msc17WasModified(g, ev.CardID)
			}, "Akki Ember-Keeper — create a 1/1 colorless Spirit creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: 1})),
				game.Purpose{Tokens: 1, CreatureTokens: 1, DeathPayoff: true}),
		},
	})
}
