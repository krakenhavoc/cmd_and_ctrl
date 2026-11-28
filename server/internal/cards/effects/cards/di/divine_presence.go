package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Divine Presence — Enchantment {2}{W} (Invasion, S82 #5616):
//
//	"If a source would deal 4 or more damage to a permanent or player, that source
//	 deals 3 damage to that permanent or player instead."
//
// A CR 614 replacement ("instead", not prevention, CR 615.1a), so damage that
// can't be prevented is still cut to 3. It applies to every damage event of 4
// or more, anyone's, and the 3 it leaves is below its own threshold, so it does
// not apply again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff9a5529-83bd-46a8-b20b-235a5fd26d93",
		Name:         "Divine Presence",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: false, // CR 614.1a: "instead", not "prevent" (CR 615.1a)
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return src != nil && ev.Kind == game.RepEventDamage && ev.DamageTarget != uuid.Nil && ev.DamageAmount >= 4
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount = 3
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Divine Presence — damage of 4 or more becomes 3",
		}},
	})
}
