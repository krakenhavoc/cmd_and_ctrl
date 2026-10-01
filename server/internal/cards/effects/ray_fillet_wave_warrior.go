package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ray Fillet, Wave Warrior — Legendary Creature — Fish Mutant {2}{U},
// 0/2:
//
//	"Flying
//	 Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Whenever a creature you control with a counter on it deals combat
//	 damage to a player, draw a card."
//
// Toski's trigger with a condition on the dealer: one trigger per
// creature that connects, read as the damage is dealt. "A counter" is
// any kind (CR 122.1), so a creature carrying only a shield or a stun
// counter counts too. Ray Fillet's own evolve counters make it one of
// those creatures. Evolve is the engine's keyword trigger
// (game/evolve.go, #1805).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "206913aa-4d73-4465-8311-8054a22c0743",
		Name:            "Ray Fillet, Wave Warrior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordEvolve},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !combatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && b41HasACounter()(g, source.Controller, dealer)
			}, "Ray Fillet, Wave Warrior — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
