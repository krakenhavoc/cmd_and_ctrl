package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stitched Drake — {1}{U}{U} Creature — Zombie Drake 3/4 (Foundations
// Jumpstart):
//
//	"As an additional cost to cast this spell, exile a creature card from
//	 your graveyard.
//	 Flying"
//
// Relentless Skaabs' additional cost: the caster names the card as
// exile_ids and it is exiled with the spell on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10fd868f-45b3-4e2f-89d9-5c0b8a92f9fe",
		Name:            "Stitched Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AdditionalCost: effects.ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool {
			return c.IsCreature()
		}),
	})
}
