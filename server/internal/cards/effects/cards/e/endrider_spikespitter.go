package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Endrider Spikespitter — Creature — Human Mercenary {3}{R}, 3/4:
//
//	"Reach
//	 Start your engines!
//	 Max speed — At the beginning of your upkeep, exile the top card of
//	 your library. You may play that card this turn."
//
// ADR 0138 (#2711). The body is Outpost Siege's Khans impulse exile: the
// PLAY grant, so an exiled land can be played.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "817f3b8b-4399-4c28-9453-7f57df4a7360",
		Name:            "Endrider Spikespitter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.MaxSpeedTrigger(effects.AtYourUpkeep("Endrider Spikespitter — exile the top card of your library; you may play it this turn",
				effects.OutpostSiegeImpulse)),
		},
	})
}
