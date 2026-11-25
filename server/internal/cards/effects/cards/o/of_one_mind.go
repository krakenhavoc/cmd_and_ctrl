package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Of One Mind — Sorcery {2}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"This spell costs {2} less to cast if you control a Human creature and a
//	 non-Human creature.
//	 Draw two cards."
//
// The two creatures are two separate permanents: a creature is a Human or it
// is not, so one creature can never be both. Read through the layers as the
// spell is priced (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a23c734-cbfb-419a-9bf1-126cb07fe427",
		Name:         "Of One Mind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you control a Human creature and a non-Human creature.",
				effects.YouControlA(effects.And(effects.Creature(), effects.HasSubtype("Human"))),
				effects.YouControlA(effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Human"))))),
		},
		OnResolve: effects.ResolveSteps(effects.DrawCards{N: 2}),
	})
}
