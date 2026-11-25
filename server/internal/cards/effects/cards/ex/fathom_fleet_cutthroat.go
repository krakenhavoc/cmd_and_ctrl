package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fathom Fleet Cutthroat — Creature — Human Pirate {3}{B}, 3/3 (Core Set 2020):
//
//	"When this creature enters, destroy target creature an opponent controls
//	 that was dealt damage this turn."
//
// Elektra's targeted destroy on enter with Case of the Gorgon's Kiss's
// damage test: the creature must have been dealt damage this turn, judged as
// the target is chosen and again as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e90e3f58-1197-41c1-81c5-008e946eb539",
		Name:         "Fathom Fleet Cutthroat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Fathom Fleet Cutthroat — destroy target creature an opponent controls that was dealt damage this turn", effects.DestroyChosenPermanent),
				effects.TargetCreature("target creature an opponent controls that was dealt damage this turn", effects.OpponentControls(), effects.DealtDamageThisTurn())),
		},
	})
}
