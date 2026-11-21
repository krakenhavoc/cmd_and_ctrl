package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flying Octobot — Artifact Creature — Robot Villain {1}{U}, 1/1 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 Whenever another Villain you control enters, put a +1/+1 counter on this
//	 creature. This ability triggers only once each turn."
//
// Marwyn's "another <type> you control enters" (any permanent with the type,
// a token included) gated by the per-turn trigger record Ant-Man, Colony
// Commander reads: once this ability has been put on the stack in a turn, no
// later Villain triggers it, even if the first counter never lands. "Villain"
// is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	const label = "Flying Octobot — put a +1/+1 counter on it"
	effects.Register(effects.Spec{
		OracleID:        "5ef8925c-5f43-49ca-83f0-e2ca8e655600",
		Name:            "Flying Octobot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Villain")) &&
					!effects.B11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
