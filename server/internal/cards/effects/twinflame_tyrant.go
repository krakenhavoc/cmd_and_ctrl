package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twinflame Tyrant — Creature — Dragon {3}{R}{R}, 3/5 (EDHREC rank 795):
//
//	"Flying
//	 If a source you control would deal damage to an opponent or a
//	 permanent an opponent controls, it deals double that damage
//	 instead."
//
// Angrath's Marauders with a body and a narrower clause — only damage
// that lands on an opponent's side of the table. Combat damage is
// included (Solphim's noncombat-only gate is what this deliberately
// does not have), so an attacking Dragon hits for six.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "321b5cc6-8df6-4292-97a0-a6a3a22f3b55",
		Name:            "Twinflame Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
					return false
				}
				return damageSourceControlledBy(ev, g, src.Controller) &&
					damageHitsAnOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
		}},
	})
}
