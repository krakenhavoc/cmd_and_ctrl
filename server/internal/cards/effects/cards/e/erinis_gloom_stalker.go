package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Erinis, Gloom Stalker — {2}{G} Legendary Creature — Halfling Ranger 3/3
//
//	"Deathtouch
//	 Whenever Erinis attacks, return target land card from your graveyard to
//	 the battlefield.
//	 Choose a Background"
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b); the land enters untapped under your control (it is
// not "put onto the battlefield tapped"). "Choose a Background" is the
// deck-construction rule (CR 702.124k) read from the oracle text.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b76a54d-5a9f-4b1b-a075-4b4c4de8ba0a",
		Name:            "Erinis, Gloom Stalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Purpose:         game.Purpose{Lands: 1, LandsUntapped: 1},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Erinis, Gloom Stalker — return target land card from your graveyard to the battlefield",
					effects.S82Mom16PutTargetOntoTheBattlefield),
				effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn())),
		},
	})
}
