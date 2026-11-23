package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82Rpr29VampireDealtCombatDamage is "a Vampire you control deals combat
// damage to a player", the creature type read post-layer so a changeling
// counts.
func s82Rpr29VampireDealtCombatDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return B24CombatDamageToPlayerByYourCreatureOfSubtypes(ev, source, g, source.TextSubtype("Vampire"))
}

// s82Rpr29VampireCounterTrigger is "Whenever a Vampire you control deals
// combat damage to a player, put a +1/+1 counter on it" (Rakish Heir,
// Stensia Masquerade).
func S82Rpr29VampireCounterTrigger(name string) game.TriggeredAbility {
	return On(game.EventDealDamage, s82Rpr29VampireDealtCombatDamage,
		name+" — put a +1/+1 counter on that Vampire", S82Mom09CounterOnThatCreature)
}
