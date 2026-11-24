package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vectis Gloves — Artifact — Equipment {2} (Modern Horizons 2, #5501):
//
//	"Equipped creature gets +2/+0 and has artifact landwalk. (It can't be
//	 blocked as long as defending player controls an artifact land.)
//	 Equip {2}"
//
// Artifact landwalk is a token of the engine's landwalk table
// (game/landwalk.go): it switches on for a land that is also an artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41074ef0-8686-41b5-9b99-7103192da915",
		Name:         "Vectis Gloves",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("artifact landwalk"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
