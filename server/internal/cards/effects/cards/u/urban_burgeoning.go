package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urban Burgeoning — Enchantment — Aura {G} (Return to Ravnica, S82 OLD):
//
//	"Enchant land
//	 Enchanted land has "Untap this land during each other player's untap
//	 step.""
//
// An untap-step permission (Bender's Waterskin's, game/untap.go) declared on
// the Aura and aimed at its host: "this land" is the enchanted land, and
// "each other player" is each player but the land's controller, who may be
// an opponent of the Aura's controller. An Aura attached to nothing
// contributes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2772e3cc-10e0-420f-b183-71ebaf3580e1",
		Name:         "Urban Burgeoning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		UntapStep: []game.UntapStepPermission{{
			Label: "Urban Burgeoning — untap enchanted land during each other player's untap step",
			AppliesTo: func(g *game.Game, source *game.Card, activePlayer uuid.UUID) bool {
				host := g.AttachedHostOf(source)
				return host != nil && activePlayer != uuid.Nil && activePlayer != host.Controller
			},
			Untaps: func(_ *game.Game, source, target *game.Card) bool {
				return source.IsAttachedTo(target.InstanceID)
			},
		}},
	})
}
