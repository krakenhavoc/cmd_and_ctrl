package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secrets of the Dead — Enchantment {2}{U} (Secret Lair Drop):
//
//	"Whenever you cast a spell from your graveyard, draw a card."
//
// Ominous Roost's cast condition (the spell's CastFromZone is the caster's own
// graveyard) with a draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9720c63-5aa7-4fb8-8d71-8f55d8b1337b",
		Name:         "Secrets of the Dead",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastASpellFromYourGraveyard,
				"Secrets of the Dead — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
