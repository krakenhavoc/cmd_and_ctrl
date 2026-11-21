package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruinous Rampage — Sorcery {1}{R}{R} (Edge of Eternities, #4461):
//
//	"Choose one —
//	 • Ruinous Rampage deals 3 damage to each opponent.
//	 • Exile all artifacts with mana value 3 or less."
//
// Two untargeted bullets. The first is one damage event from the spell to
// every opponent still in the game. The second exiles every artifact on
// the battlefield, whoever controls it, whose mana value is 3 or less;
// ManaValueLE reads {X} as 0 on the battlefield (the rulings), and a
// token, which has no mana cost, is mana value 0 and goes too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5020628c-5e47-4a83-89c3-7bf5726e69ad",
		Name:         "Ruinous Rampage",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Ruinous Rampage deals 3 damage to each opponent."),
			effects.ModeWithPurpose(effects.Mode("Exile all artifacts with mana value 3 or less."),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepExile, Partial: true}}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.DamageEachOpponent{N: 3}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.ExileAllMatching{Match: effects.And(effects.Artifact(), effects.ManaValueLE(3))}.Apply(ctx)
			}
			return nil
		},
	})
}
