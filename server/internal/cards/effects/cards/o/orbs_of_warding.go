package cardso

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orbs of Warding — Artifact {5} (Magic Origins):
//
//	"You have hexproof.
//	 If a creature would deal damage to you, prevent 1 of that damage."
//
// Keen-Eared Sentry's player hexproof and Djeru's one-point prevention,
// read here off a creature source dealing damage to the controller. Each
// creature's damage is its own event, so each loses 1; damage of 1 is
// prevented entirely (CR 615.1a). The source is read as the event
// snapshotted it (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "c420d9c8-dc54-46b8-bd22-39c21019c3d7",
		Name:           "Orbs of Warding",
		Completeness:   effects.CompletenessFull,
		PlayerKeywords: []string{effects.KeywordHexproof},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget != src.Controller {
					return false
				}
				ch, ok := effects.DamageSourceCharacteristics(ev, g)
				return ok && slices.Contains(ch.Types, "Creature")
			},
			Replace:    effects.S82OldB317PreventOneOfThatDamage,
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
			Label:      "Orbs of Warding — prevent 1 of that damage",
		}},
	})
}
