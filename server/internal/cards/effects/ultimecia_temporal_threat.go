package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultimecia, Temporal Threat — Legendary Creature — Human Warlock
// {4}{U}{U}, 4/4 (Final Fantasy, #4462):
//
//	"When Ultimecia enters, tap all creatures your opponents control.
//	 Whenever a creature you control deals combat damage to a player,
//	 draw a card."
//
// The enters trigger is Summon: Alexander's third chapter
// (tapAllCreaturesYourOpponentsControl), read as it resolves. The draw is
// Ohran Frostfang's: one trigger per creature you control that deals
// combat damage to a player, Ultimecia included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a8e7e35-e9ea-4808-8aa7-7580a23b9d9d",
		Name:         "Ultimecia, Temporal Threat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ultimecia, Temporal Threat — tap all creatures your opponents control",
				tapAllCreaturesYourOpponentsControl),
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Ultimecia, Temporal Threat — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
