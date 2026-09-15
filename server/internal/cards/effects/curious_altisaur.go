package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curious Altisaur — Creature — Dinosaur {3}{G}, 2/5 (EDHREC rank
// 2937):
//
//	"Reach, vigilance
//	 Whenever a Dinosaur you control deals combat damage to a player,
//	 draw a card."
//
// The Dinosaur deck's card draw. Reach and vigilance ride
// PrintedKeywords; the draw is one trigger per Dinosaur that
// connects (b24CombatDamageToPlayerByYourCreatureOfSubtypes — the
// Altisaur included, effective subtypes so a changeling counts), no
// "one or more" batching because the printed card has none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "80686908-abb5-4728-a5f6-71baca27f467",
		Name:            "Curious Altisaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24CombatDamageToPlayerByYourCreatureOfSubtypes(ev, source, g, "Dinosaur")
			}, "Curious Altisaur — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
