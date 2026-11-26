package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slayer's Cleaver — Artifact — Equipment {3} (Eldritch Moon, S82 OLD, #5541):
//
//	"Equipped creature gets +3/+1 and must be blocked by an Eldrazi if able.
//	 Equip {4}"
//
// The requirement is Ace's Baseball Bat's "must be blocked by a Dalek if
// able" with Eldrazi and with no attacking condition: it is a requirement on
// the equipped creature (CR 509.1c), and only a creature declared as an
// attacker is ever blocked, so it is live exactly when it matters.
//
// No simplification.
func init() {
	eldrazi := effects.MustBeBlockedBySubtypeWhere("Eldrazi", func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.AttachedToSource(target, g, source)
	})
	eldrazi.Label = "Equipped creature must be blocked by an Eldrazi if able."
	effects.Register(effects.Spec{
		OracleID:     "e51cbdce-017e-4b14-b3a7-2db50ddd8037",
		Name:         "Slayer's Cleaver",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 1), eldrazi},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
