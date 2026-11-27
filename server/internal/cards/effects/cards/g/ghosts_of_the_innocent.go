package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghosts of the Innocent — Creature — Spirit {5}{W}{W}, 4/5 (Ravnica: City of
// Guilds):
//
//	"If a source would deal damage to a permanent or player, it deals half that
//	 damage, rounded down, to that permanent or player instead."
//
// A damage REPLACEMENT (CR 614.1a), not prevention: it applies to every source
// and every recipient, the Ghosts' controller's included, and the source is
// still the one dealing the (smaller) damage, so lifelink, deathtouch and
// "deals damage" triggers read the halved amount. Gisela's prevention rounds
// the other way; this one rounds down (CR 107.1a), so one damage becomes none.
// Two such effects each apply (CR 616.1), and the second halves the first's
// result.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9243bd95-5467-490d-9bd2-4d5ce1ebc589",
		Name:         "Ghosts of the Innocent",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			// A replacement that rewrites the amount, not a prevention
			// (CR 615.1a): "can't be prevented" does not stop it.
			Prevention: false,
			RowLabel:   "If a source would deal damage to a permanent or player, it deals half that damage, rounded down, to that permanent or player instead.",
			Label:      "Ghosts of the Innocent — half that damage, rounded down",
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount /= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		}},
	})
}
