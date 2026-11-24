package cardsbu

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Calamity Bearer — Creature — Giant Berserker {2}{R}{R}, 3/4 (Kaldheim):
//
//	"If a Giant source you control would deal damage to a permanent or player,
//	 it deals double that damage to that permanent or player instead."
//
// A CR 614 damage replacement (Obosh's shape). The source is read as the damage
// event snapshotted it (a Giant that has left deals its damage as it last was,
// CR 608.2h): its controller is the Bearer's and it was a Giant (or a changeling,
// which is every creature type) then. Each Bearer applies once to a given event,
// so two Bearers double twice (CR 616.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3f7c4ac-b788-4307-9bc3-d99ce310e07c",
		Name:         "Calamity Bearer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:  []game.EventKind{game.EventDealDamage},
			RowLabel: "If a Giant source you control would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.",
			Label:    "Calamity Bearer — double that damage",
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				ch, ok := effects.DamageSourceCharacteristics(ev, g)
				return ok && ch.Controller == src.Controller &&
					(ch.AllCreatureTypes || slices.Contains(ch.Subtypes, src.TextSubtype("Giant")))
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		}},
	})
}
