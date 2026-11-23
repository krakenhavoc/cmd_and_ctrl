package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullsnap Nuisance — Creature — Insect Skeleton {U}{B}, 1/4:
//
//	"Flying
//	Eerie — Whenever an enchantment you control enters and whenever you
//	fully unlock a Room, surveil 1."
//
// Flying rides PrintedKeywords; the surveil is a real prompt (Surveil).
//
// Eerie is one ability with two conditions (Eerie, rooms.go): an
// enchantment entering under your control, or you fully unlocking a Room.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "99348e1d-e95e-4911-aae0-8fd1a7345a82",
		Name:            "Skullsnap Nuisance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Eerie("Skullsnap Nuisance — surveil 1 (eerie)", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
