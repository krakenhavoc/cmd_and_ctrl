package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indoraptor, the Perfect Hybrid — Legendary Creature — Dinosaur Mutant
// {1}{B/G}{R}, 3/1 (Jurassic World Collection, #5948):
//
//	"Bloodthirst X (This creature enters with X +1/+1 counters on it, where
//	 X is the damage dealt to your opponents this turn.)
//	 Menace
//	 Enrage — Whenever Indoraptor is dealt damage, choose an opponent at
//	 random. Indoraptor deals damage equal to its power to that player
//	 unless they sacrifice a nontoken creature of their choice."
//
// Bloodthirst X (CR 702.54b) is an entry replacement counting the total
// damage dealt to opponents this turn. Enrage fires once per batch of damage;
// the bodies are in s82_gap_b82_helpers.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0dd4bc7b-2ddf-4ca4-912c-59685c41c3b5",
		Name:            "Indoraptor, the Perfect Hybrid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return effects.GapB82DamageDealtToOpponents(g, src.Controller)
			}, "Indoraptor enters with X +1/+1 counters, X the damage dealt to your opponents this turn"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Indoraptor — Enrage: choose an opponent at random; damage equal to its power unless they sacrifice a nontoken creature",
				effects.GapB82IndoraptorEnrage),
		},
	})
}
