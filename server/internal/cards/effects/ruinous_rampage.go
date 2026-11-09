package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5020628c-5e47-4a83-89c3-7bf5726e69ad",
		Name:         "Ruinous Rampage",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Ruinous Rampage deals 3 damage to each opponent."),
			ModeWithPurpose(Mode("Exile all artifacts with mana value 3 or less."),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepExile, Partial: true}}),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return DamageEachOpponent{N: 3}.Apply(ctx)
			case ctx.HasMode(1):
				return ExileAllMatching{Match: And(Artifact(), ManaValueLE(3))}.Apply(ctx)
			}
			return nil
		},
	})
}
