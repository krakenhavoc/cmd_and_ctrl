package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forerunner of the Coalition — Creature — Human Pirate {2}{B}, 2/2:
//
//	"When this creature enters, you may search your library for a Pirate card,
//	 reveal it, then shuffle and put that card on top.
//	 Whenever another Pirate you control enters, each opponent loses 1 life."
//
// The first ability is the Harbinger cycle's. The second triggers for each
// other Pirate entering under its controller's control; each opponent still in
// the game loses 1 life (a life change, not damage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6458cf6-eb1f-448b-8205-1b5d887ff2bf",
		Name:         "Forerunner of the Coalition",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LrwB1Harbinger("Forerunner of the Coalition", "a Pirate card", effects.LrwB1CardOfType("Pirate")),
			effects.On(game.EventETB, effects.S82OldB216TribeEnters("Pirate", true),
				"Forerunner of the Coalition — each opponent loses 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
