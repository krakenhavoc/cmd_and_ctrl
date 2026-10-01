package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Endangered Armodon — Creature {2}{G}{G}, 4/5:
//
//	"When you control a creature with toughness 2 or less, sacrifice
//	 this creature."
//
// ADR 0107 §1 (#1858). A CR 603.8 state trigger. Toughness is read as
// it is now (layer 7), so a creature shrunk to 2 by an effect counts.
// The Armodon is a creature too and would count itself at toughness 2 or
// less.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68879437-bbe2-4e99-b17b-ddec30bfd3d0",
		Name:         "Endangered Armodon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenState("Endangered Armodon — sacrifice it",
				func(g *game.Game, _ *game.Card, controller uuid.UUID) bool {
					return controlsAnyCard(g, controller, And(Creature(), ToughnessLE(2)))
				}, SacrificeThisIfStillOnBattlefield),
		},
	})
}
