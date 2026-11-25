package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darigaaz Reincarnated — Legendary Creature — Dragon {4}{B}{R}{G}, 7/7
// (Dominaria):
//
//	"Flying, trample, haste
//	 If Darigaaz would die, instead exile it with three egg counters on it.
//	 At the beginning of your upkeep, if this card is exiled with an egg counter
//	 on it, remove an egg counter from it. Then if this card has no egg counters
//	 on it, return it to the battlefield."
//
// The first is a replacement of Darigaaz's own death (CR 614.1a): it is exiled
// as it would have gone to the graveyard, landing with three egg counters
// (ReplacementEvent.ExileCounterCount), and nothing is put into a graveyard, so
// no "dies" trigger fires. The last functions from exile (CR 113.6): an
// intervening "if" (CR 603.4) that it is there with an egg counter, checked
// as your upkeep begins and again as the trigger resolves, removes one counter
// and, when none is left, returns the card to the battlefield under its owner's
// control. Three upkeeps later it comes back, as a new object.
//
// No simplification.
func init() {
	upkeep := effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Actor == source.Owner && source.Counters[effects.OldB209EggCounter] > 0
	}, "Darigaaz Reincarnated — remove an egg counter; if it has none, return it to the battlefield", effects.OldB209DarigaazUpkeep)
	upkeep.Zones = []game.ZoneKind{game.ZoneExile}
	effects.Register(effects.Spec{
		OracleID:        "c2f42c67-ff88-4248-bfa5-79a18c6473d5",
		Name:            "Darigaaz Reincarnated",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "haste"},
		Replacements:    []game.ReplacementEffect{effects.OldB209DarigaazDies()},
		Triggered:       []game.TriggeredAbility{upkeep},
	})
}
