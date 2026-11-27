package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gutter Grime — Enchantment {4}{G} (Innistrad, S82 OLD batch 450, #5572):
//
//	"Whenever a nontoken creature you control dies, put a slime counter on
//	 this enchantment, then create a green Ooze creature token with "This
//	 token's power and toughness are each equal to the number of slime
//	 counters on Gutter Grime.""
//
// Saproling Burst's linked token (see GutterGrimeOozeToken) behind a
// dies trigger on a nontoken creature the controller controlled as it
// died (last-known, CR 603.10a). Seam check: #6268.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b8adb9e-97a5-43d9-87d6-993f22edaf3b",
		Name:         "Gutter Grime",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !dead.IsToken() && effects.LeftUnderControlOf(ev, dead) == source.Controller
			}, "Gutter Grime — put a slime counter on this enchantment, then create an Ooze token sized by it", effects.GutterGrimeCreate),
		},
	})
}
