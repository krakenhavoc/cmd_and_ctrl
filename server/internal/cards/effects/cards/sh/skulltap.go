package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skulltap — Sorcery {1}{B} (Scourge):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards."
//
// Village Rites at sorcery speed: the creature is sacrificed as the spell
// is cast (CR 601.2h), so it is gone even if the spell is countered, and
// with no creature to sacrifice the spell can't be cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "17c7878f-891f-40a6-80a7-47fcc7b8e29b",
		Name:           "Skulltap",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Draws: 2},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
