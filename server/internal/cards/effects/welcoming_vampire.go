package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Welcoming Vampire — Creature — Vampire {2}{W}, 2/3 (EDHREC top 1,000,
// #3654):
//
//	"Flying
//	 Whenever one or more other creatures you control with power 2 or
//	 less enter, draw a card. This ability triggers only once each turn."
//
// The gate is three conditions: another creature of yours entered, its
// power is 2 or less as it entered, and this ability has not already
// been put on the stack this turn (b11TriggeredThisTurn, the Exemplar of
// Light memory). The once-a-turn rule already collapses a batch to its
// first creature, so no separate "one or more" batching is needed. A
// trigger that was countered still counts as having triggered.
//
// No simplification.
func init() {
	const label = "Welcoming Vampire — draw a card"
	Register(Spec{
		OracleID:        "605c1ee0-5e8a-4e0a-a99b-42a38873f822",
		Name:            "Welcoming Vampire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(
				AnotherCreatureEnteredUnderYourControl,
				func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Effective().Power <= 2
				},
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return source != nil && !b11TriggeredThisTurn(g, source.InstanceID, label)
				},
			), label, Do(DrawCards{N: 1})),
		},
	})
}
