package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grendel, Spawn of Knull — Legendary Creature — Symbiote Dragon Villain
// {3}{B}, 3/3 (Marvel's Spider-Man, #4460):
//
//	"Flying
//	 Deathtouch"
//
// Two printed keywords, both read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a93f0735-9854-487d-b646-d55e2bee00c0",
		Name:            "Grendel, Spawn of Knull",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
}

var _ = game.Card{}
