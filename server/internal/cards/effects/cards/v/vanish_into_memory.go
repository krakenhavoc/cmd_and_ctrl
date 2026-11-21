package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vanish into Memory — Instant {2}{W}{U}:
//
//	"Exile target creature. You draw cards equal to that creature's
//	 power. At the beginning of your next upkeep, return that card to
//	 the battlefield under its owner's control. If you do, discard cards
//	 equal to that creature's toughness."
//
// The creature's power and toughness are read as the spell resolves,
// before it leaves (CR 608.2h), and ride the exile's continuation. The
// draw happens whether or not the exile took (a commander may have gone
// to the command zone). The return is a delayed trigger (CR 603.7) on
// the caster's next upkeep naming the exiled object, so a card that has
// left exile since is a new object and stays (CR 603.7c, 400.7), and a
// token, which ceased to exist, never returns. The discard is "if you
// do": it follows only a card that actually entered, and the caster
// chooses the cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3beb624-f589-4ed5-97a0-11ad61481250",
		Name:         "Vanish into Memory",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.VanishIntoMemoryResolve,
	})
}
