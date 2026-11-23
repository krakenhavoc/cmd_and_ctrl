package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamebreak — Sorcery {R}{R}{R}:
//
//	"Flamebreak deals 3 damage to each creature without flying and each
//	 player. Creatures dealt damage this way can't be regenerated this
//	 turn."
//
// One batch over the creatures without flying (read post-layer as it
// resolves) and every player still in the game, with "creatures dealt
// damage this way" read from the damage's continuation (ADR 0108 §2):
// a creature whose damage was all prevented can still be regenerated,
// and a player is never marked.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID: "9e608ace-2844-419b-9204-9054127557b2",
		Name:     "Flamebreak",
		// ADR 0126 §6: a creature with flying is spared.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatchingAndEachPlayerThen(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), 3,
				effects.CantBeRegeneratedIfDealtDamage(item))
		},
	})
}
