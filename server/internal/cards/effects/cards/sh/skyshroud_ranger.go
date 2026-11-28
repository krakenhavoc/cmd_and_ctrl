package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyshroud Ranger — Creature — Elf Ranger {G}, 1/1 (Tempest, S82 OLD b754,
// #5628):
//
//	"{T}: You may put a land card from your hand onto the battlefield.
//	 Activate only as a sorcery."
//
// Walking Atlas' put (a put, not a play, CR 305.4) at sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f47a230-7500-4afb-97cd-dbd444c307c8",
		Name:         "Skyshroud Ranger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: You may put a land card from your hand onto the battlefield. Activate only as a sorcery.",
			Purpose:      game.Purpose{Lands: 1, LandsUntapped: 1},
			Cost:         effects.TapCost(),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.MayPutALandFromHand("Skyshroud Ranger")),
		}},
	})
}
