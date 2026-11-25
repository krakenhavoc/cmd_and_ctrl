package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silumgar, the Drifting Death — Legendary Creature — Dragon {4}{U}{B}, 3/7
// (Commander 2017, #5533):
//
//	"Flying, hexproof
//	 Whenever a Dragon you control attacks, creatures defending player
//	 controls get -1/-1 until end of turn."
//
// One trigger per attacking Dragon (CR 508.3a), so two Dragons attacking
// give -2/-2. The creatures are those the defending player controls as the
// ability resolves (CR 611.2c); one that arrives later is unaffected. A
// Dragon attacking a planeswalker names that planeswalker's controller as
// the defending player. A defending player who has left the game gives no
// creatures to shrink.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b26c91d9-87de-45f7-a24c-bbe0cef3a35c",
		Name:            "Silumgar, the Drifting Death",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.DragonAttackTrigger("Silumgar — creatures defending player controls get -1/-1 until end of turn",
				func(ctx *effects.Context, defender uuid.UUID) error {
					if defender == uuid.Nil {
						return nil
					}
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.ControlledByPlayer(defender)), Power: -1, Toughness: -1,
						Label: "Silumgar — -1/-1",
					}.Apply(ctx)
				}),
		},
	})
}
