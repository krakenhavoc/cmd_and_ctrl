package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreamtide Whale — Creature — Whale {2}{U}, 7/5 (Modern Horizons 3):
//
//	"Vanishing 2 (This creature enters with two time counters on it. At
//	 the beginning of your upkeep, remove a time counter from it. When the
//	 last is removed, sacrifice it.)
//	 Whenever a player casts their second spell each turn, proliferate.
//	 (Choose any number of permanents and/or players, then give each
//	 another counter of each kind already there.)"
//
// Vanishing is Deep Forest Hermit's three parts (cmdr2B17Vanishing*). The
// trigger is Ledger Shredder's "any player's second spell each turn"
// (CastTallyFor(actor).Total == 2) and resolves the shared Proliferate for
// the Whale's controller, whoever cast the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee8c0b4c-03c5-4da5-b4f1-028c845384af",
		Name:         "Dreamtide Whale",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(2)},
		Triggered: append(effects.Cmdr2B17VanishingTriggers(),
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && g.CastTallyFor(ev.Actor).Total == 2
			}, "Dreamtide Whale — proliferate", effects.Do(effects.Proliferate{}))),
	})
}
