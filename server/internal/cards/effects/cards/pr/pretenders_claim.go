package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pretender's Claim — Enchantment — Aura {1}{B} (Mercadian Masques, S82 OLD batch 686, #5619):
//
//	"Enchant creature
//	Whenever enchanted creature becomes blocked, tap all lands defending player controls."
//
// EventBecomesBlocked is one event per blocked attacker however many creatures
// block it (CR 509.1h), and its Actor is the defending player, whose lands are
// tapped as the trigger resolves (Nature's Will's tap-every-land walk).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac285479-46c2-49f3-a7ff-31affe23c21f",
		Name:         "Pretender's Claim",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source.IsAttachedTo(ev.CardID)
				},
				"Pretender's Claim — tap all lands defending player controls",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
						return nil
					}
					return effects.B24TapAllLandsControlledBy(effects.NewContext(g, item), item.Trigger.Event.Actor)
				}),
		},
	})
}
