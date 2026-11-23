package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Synapse Necromage — Creature — Fungus Wizard {2}{B}, 3/1 (The Lost
// Caverns of Ixalan, #4470):
//
//	"When this creature dies, create two 1/1 black Fungus creature tokens
//	 with "This token can't block.""
//
// A dies trigger (CR 700.4) that looks back in time (CR 603.10a). The
// Fungus is the catalog's token with its can't-block static
// (FungusCantBlockToken).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5eace41-0f53-4084-8bec-1e079c04287b",
		Name:         "Synapse Necromage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Synapse Necromage — create two 1/1 black Fungus tokens that can't block",
				effects.Do(effects.CreateToken{Template: effects.FungusCantBlockToken(), N: 2})),
		},
	})
}
