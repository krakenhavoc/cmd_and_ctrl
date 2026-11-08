package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Virulent Silencer — Artifact Creature — Robot Assassin {3}, 2/3 (Edge of
// Eternities, #4461):
//
//	"Whenever a nontoken artifact creature you control deals combat damage
//	 to a player, that player gets two poison counters."
//
// Sharding Sphinx's watch (an artifact creature its controller controls
// dealt combat damage to a player, read post-layer off the damage source)
// narrowed to nontoken creatures. One trigger per creature that connects
// (CR 603.2c), the Silencer itself included, and two in a double striker's
// two damage steps. "That player" is the damaged player, read off the
// triggering event, and the poison is placed by the trigger's controller,
// as Glowing One's rad counters are. Ten poison counters lose the game as
// a state-based action (CR 704.5c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d70a2f4-92cb-40c3-9e5a-f6c57d4a64c2",
		Name:         "Virulent Silencer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, virulentSilencerConnects,
				"Virulent Silencer — that player gets two poison counters", virulentSilencerPoison),
		},
	})
}

// virulentSilencerConnects is "a nontoken artifact creature you control
// deals combat damage to a player".
func virulentSilencerConnects(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b36ArtifactCreatureYouControlDealtCombatDamageToPlayer(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && !c.IsToken()
}

// virulentSilencerPoison gives the damaged player two poison counters.
func virulentSilencerPoison(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player := item.Trigger.Event.Target
	if g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return g.AddPlayerCounterByForEffect(item.Controller, player, game.CounterPoison, 2)
}
