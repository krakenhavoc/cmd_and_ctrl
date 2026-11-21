package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Champion Lancer — Creature — Human Knight {4}{W}{W}, 3/3 (Starter
// 1999, #5620):
//
//	"Prevent all damage that would be dealt to this creature by
//	 creatures."
//
// Marble Priest's standing prevention (CR 615.1a) widened to all damage,
// combat or not, whose source is a creature. "Creatures" is the source's
// characteristics as the damage would be dealt, or as it last existed if
// it has left (ReplacementEvent.SourceLKI, the same read the engine's
// "by creatures" shields make), so a creature's ability that deals
// damage is prevented too, and damage from a spell is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2681bfb5-127b-4046-85ac-72be584999cb",
		Name:         "Champion Lancer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDamage && ev.DamageTarget == src.InstanceID &&
					ev.SourceLKI != nil && effects.LkiHasType(ev.SourceLKI, "Creature")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Champion Lancer — prevent all damage that would be dealt to it by creatures",
		}},
	})
}
