package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zagras, Thief of Heartbeats — Legendary Creature — Vampire Rogue
// {4}{B}{R}, 4/4 (Zendikar Rising, S82 OLD #5508):
//
//	"This spell costs {1} less to cast for each creature in your party.
//	 Flying, deathtouch, haste
//	 Other creatures you control have deathtouch.
//	 Whenever a creature you control deals combat damage to a planeswalker,
//	 destroy that planeswalker."
//
// The party is counted as the spell is cast (CR 700.8). The destroy trigger
// names the planeswalker that was dealt the damage; it is destroyed only if
// it is still that object (CR 400.7).
//
// No simplification.
func init() {
	deathtouch := effects.KeywordGrant(effects.OtherCreaturesYouControl, "deathtouch")
	deathtouch.Label = "Other creatures you control have deathtouch."
	effects.Register(effects.Spec{
		OracleID:          "a2375a23-a1f0-4514-87a0-e699f91acd7d",
		Name:              "Zagras, Thief of Heartbeats",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"flying", "deathtouch", "haste"},
		SelfCostModifiers: []game.CostModifier{effects.S82OldB091PartyCostReduction()},
		Static:            []game.StaticAbility{deathtouch},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.S82OldB091CreatureYouControlDealtCombatDamageToAPlaneswalker,
				"Zagras, Thief of Heartbeats — destroy that planeswalker",
				effects.S82OldB091DestroyTheDamagedPermanent),
		},
	})
}
