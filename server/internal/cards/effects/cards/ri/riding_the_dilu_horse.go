package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Riding the Dilu Horse — Sorcery {2}{G}
// (Portal Three Kingdoms, #5622):
//
//	"Target creature gets +2/+2 and gains horsemanship. (It can't be blocked except by creatures with horsemanship. This effect lasts indefinitely.)"
//
// The effect has no stated duration, so it lasts indefinitely (CR 611.2a):
// a layer-6 and layer-7c effect pinned to the creature, which a copy or a
// later return to the battlefield does not carry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7162f90-8430-4c1d-9f9a-a912bf1e9f50",
		Name:         "Riding the Dilu Horse",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB710TargetGetsAndGainsIndefinitely(ctx, 2, 2, "horsemanship", "Riding the Dilu Horse")
		},
	})
}
