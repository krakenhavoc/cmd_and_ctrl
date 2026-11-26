package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archon of Absolution — Creature — Archon {3}{W}, 3/2 (Throne of Eldraine, S82
// old sets):
//
//	"Flying
//	 Protection from white
//	 Creatures can't attack you or planeswalkers you control unless their
//	 controller pays {1} for each of those creatures."
//
// Baird's attack tax (CR 508.1h), protecting its controller and their
// planeswalkers; flying and protection from white are printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0e3ac581-3d78-4879-9afb-3db69c5c47a5",
		Name:            "Archon of Absolution",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from white"},
		AttackTaxes: []game.AttackTax{
			effects.ProtectingPlaneswalkers(effects.AttackTax("{1}", "Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.")),
		},
	})
}
