package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Banewasp Affliction — Enchantment — Aura {1}{B} (Shards of Alara, S82 OLD
// batch 517, #5588):
//
//	"Enchant creature
//	 When enchanted creature dies, that creature's controller loses life equal
//	 to its toughness."
//
// The toughness is the creature's last-known toughness as it died (CR 608.2h,
// counters and effects included), and the life is lost by the player who
// controlled it as it died, which may be neither the Aura's controller nor the
// creature's owner.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1cb870a-8c1f-435c-b678-e456c4d2d627",
		Name:         "Banewasp Affliction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Banewasp Affliction — that creature's controller loses life equal to its toughness",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					info, ok := g.LastKnownPermanentForEffect(item.Trigger.Event.CardID)
					if !ok || info.Toughness <= 0 {
						return nil
					}
					return effects.GainLife{Player: info.Controller, Amount: -info.Toughness}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
