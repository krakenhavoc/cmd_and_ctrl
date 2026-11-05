package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Destruction of Praxis — Sorcery {3}{B}{B}:
//
//	"Strange new worlds — All creatures get -X/-X until end of turn,
//	 where X is the number of differently named lands you control. If X
//	 is 6 or more, create two 2/1 red Klingon Warrior creature tokens
//	 with haste."
//
// Strange new worlds is an ability word: X is counted once, as the spell
// resolves (socB23DifferentlyNamedLands, each land name once). The
// creatures it shrinks are fixed then too (CR 611.2c), so the two Klingon
// Warriors, made after the -X/-X, are not shrunk. With no land X is 0 and
// nothing shrinks; the tokens need X of 6 or more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "028d6812-55db-40df-9288-89c2e478c76c",
		Name:     "The Destruction of Praxis",
		// ADR 0126 §6: X is counted at resolution, so it is declared as the
		// partial creature wipe it usually is (Ichor Explosion's reading).
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := socB23DifferentlyNamedLands(ctx)
			if x > 0 {
				if err := (BoostUntilEOT{
					Match:     Creature(),
					Power:     -x,
					Toughness: -x,
					Label:     fmt.Sprintf("The Destruction of Praxis — -%d/-%d", x, x),
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if x < 6 {
				return nil
			}
			return CreateToken{Controller: item.Controller, Template: TokenCard("2/1 red Klingon Warrior with haste"), N: 2}.Apply(ctx)
		},
	})
}
