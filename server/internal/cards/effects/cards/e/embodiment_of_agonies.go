package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embodiment of Agonies — Creature — Demon {1}{B}{B}, 0/0 (Core Set 2020,
// #5518):
//
//	"Flying, deathtouch
//	 This creature enters with a +1/+1 counter on it for each different mana
//	 cost among nonland cards in your graveyard. (For example, {2}{B} and
//	 {1}{B}{B} are different mana costs.)"
//
// A self-replacement on entry (CR 614.1c) whose count is read as the creature
// enters, from the graveyard of the player who cast it; the counters are on it
// before any state check sees the printed 0/0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d6686a0-1648-42f6-a606-16d2a84138a7",
		Name:            "Embodiment of Agonies",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return effects.OldB156DifferentManaCostsInGraveyard(g, src.Controller)
			}, "Embodiment of Agonies enters with a +1/+1 counter on it for each different mana cost among nonland cards in your graveyard."),
		},
	})
}
