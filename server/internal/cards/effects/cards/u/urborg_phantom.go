package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Phantom — Creature — Spirit Minion {2}{B}, 3/1:
//
//	"This creature can't block.
//	 {U}: Prevent all combat damage that would be dealt to and dealt by
//	 this creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): Bloodghast's "can't block", and
// one to-and-by record (Mod.AndDealtBy) pinned to this creature.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8b5d2f4-58a8-408c-bb67-6b5f9ba5b678",
		Name:         "Urborg Phantom",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Activated: []effects.ActivatedAbility{effects.ToAndByThisRow(
			"{U}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.",
			effects.ManaCost("{U}"))},
	})
}
