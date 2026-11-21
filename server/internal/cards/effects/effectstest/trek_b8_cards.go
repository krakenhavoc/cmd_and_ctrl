package effectstest

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trekB8Main walks the fresh table to seat 0's precombat main phase.
func TrekB8Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceTo(t, g, game.StepPrecombatMain)
	return g, g.Seats[0], g.Seats[1]
}

// trekB8Float floats a cost's mana into p's pool, each generic {N} as N
// colourless.
func TrekB8Float(t *testing.T, g *game.Game, p *game.Player, cost string) {
	t.Helper()
	mana := regexp.MustCompile(`\{(\d+)\}`).ReplaceAllStringFunc(cost, func(m string) string {
		n, _ := strconv.Atoi(strings.Trim(m, "{}"))
		return strings.Repeat("{C}", n)
	})
	FloatMana(t, g, p, mana)
}
