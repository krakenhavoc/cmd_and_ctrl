package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootwater Shaman — Creature — Merfolk Shaman {2}{U}, 2/2 (Tempest, S82 OLD,
// #5628):
//
//	"You may cast Aura spells with enchant creature as though they had flash."
//
// Blur of Heroism's timing grant (Spec.CastTimings, derived from the battlefield
// on every query) narrowed to Aura spells whose own enchant clause is plain
// "enchant creature": an Aura that enchants a land, an artifact or a creature
// card in a graveyard still waits for its own timing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a449915-8bd7-4a9c-abb1-a95852f33695",
		Name:         "Rootwater Shaman",
		Completeness: effects.CompletenessFull,
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{Subtype: "Aura", EnchantsCreatureOnly: true},
				"You may cast Aura spells with enchant creature as though they had flash."),
		},
	})
}
