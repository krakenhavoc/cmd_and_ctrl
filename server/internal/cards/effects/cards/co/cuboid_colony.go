package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cuboid Colony — Creature — Insect {G}{U}, 1/1 (Secrets of Strixhaven,
// #4456):
//
//	"Flash
//	 Flying, trample
//	 Increment (Whenever you cast a spell, if the amount of mana you spent
//	 is greater than this creature's power or toughness, put a +1/+1
//	 counter on this creature.)"
//
// The three evergreen keywords are printed keywords; Increment is the
// keyword helper (CR 702.191, Increment), which reads the mana spent as the
// spell is cast and rechecks it as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20f3388f-0931-4a09-bcda-96deb01e1787",
		Name:            "Cuboid Colony",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying", "trample", effects.KeywordIncrement},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{effects.Increment()},
	})
}
