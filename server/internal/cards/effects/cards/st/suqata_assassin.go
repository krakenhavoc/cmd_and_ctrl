package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suq'Ata Assassin — Creature — Human Assassin {1}{B}{B}, 1/1 (Visions, S82
// older sets batch 791, #5631):
//
//	"Fear (This creature can't be blocked except by artifact creatures and/or
//	 black creatures.)
//	 Whenever this creature attacks and isn't blocked, defending player gets a
//	 poison counter."
//
// Fear is a printed keyword the block check reads (CR 702.36). The trigger is
// Bone Dancer's: it fires once the blockers are declared and this creature is
// unblocked, and the player it attacked, a planeswalker's controller included
// (CR 506.3d), gets one poison counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "db43b866-13e0-4336-a580-b64fe1ee5954",
		Name:            "Suq'Ata Assassin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Triggered: []game.TriggeredAbility{
			effects.WhenAttacksAndIsNotBlockedEffect("Suq'Ata Assassin — defending player gets a poison counter", effects.Old791DefenderGetsAPoisonCounter),
		},
	})
}
