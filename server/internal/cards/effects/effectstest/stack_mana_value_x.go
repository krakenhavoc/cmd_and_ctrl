package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stack_mana_value_x_test.go — #788. CR 202.3e: while a spell is on
// the stack, the {X} in its mana cost counts as the value chosen for
// it. Every card here used to read the printed cost, where X is zero.
// They now read game.(*Game).ManaValueForEffect. The engine table for
// that read is game/mana_value_for_effect_test.go. These tests check
// each card that used the printed cost.

// seedLibraryOfCost replaces the active seat's library with `n`
// instants that all cost `manaCost`, so the cascade limit alone
// decides whether anything is a hit.
func SeedLibraryOfCost(g *game.Game, p *game.Player, n int, manaCost string) {
	p.Library.Cards = nil
	for i := 0; i < n; i++ {
		c := game.NewCard("Library Spell", p.ID)
		c.TypeLine = "Instant"
		c.ManaCost = manaCost
		p.Library.PushTop(c)
	}
}
