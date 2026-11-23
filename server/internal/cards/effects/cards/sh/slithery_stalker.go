package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slithery Stalker — Creature — Nightmare Horror {1}{B}{B}, 1/1
// (Torment, #5612):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)
//	 When this creature enters, exile target green or white creature an
//	 opponent controls.
//	 When this creature leaves the battlefield, return the exiled card to
//	 the battlefield under its owner's control."
//
// Swampwalk is an engine keyword carried as printed data (CR 702.14c).
// The rest is Faceless Butcher's linked pair (CR 607.2a) with a narrower
// target: green or white, read as the trigger is put on the stack and
// again as it resolves (CR 608.2b), and controlled by an opponent. The
// leave trigger returns the card the entry trigger exiled, if it is still
// in exile, under its owner's control; if the Stalker leaves first, the
// creature is exiled for good, as the Oracle rulings say.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0c31edd0-a471-4994-8209-b0a7370fc27b",
		Name:            "Slithery Stalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(slitheryStalkerExileLabel, effects.ExileFirstLegalCardTarget),
				effects.TargetCreature("target green or white creature an opponent controls",
					effects.Or(effects.OfColor("G"), effects.OfColor("W")), effects.OpponentControls())),
			effects.On(game.EventLTB, effects.Self, "Slithery Stalker — return the exiled card to the battlefield under its owner's control",
				effects.B41ReturnCardsExiledWithToTheBattlefield(slitheryStalkerExileLabel)),
		},
	})
}

// slitheryStalkerExileLabel is the entry trigger's key, which is what the
// leave trigger reads the linked exile by.
const slitheryStalkerExileLabel = "Slithery Stalker — exile target green or white creature an opponent controls"
