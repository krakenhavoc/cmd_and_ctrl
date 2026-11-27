package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Latulla's Orders — Enchantment — Aura {1}{R} (Prophecy, S82 OLD b669, #5617):
//
//	"Flash
//	 Enchant creature
//	 Whenever enchanted creature deals combat damage to defending player, you
//	 may destroy target artifact that player controls."
//
// Destructive Urge's trigger condition (the enchanted creature dealt combat
// damage to a player; a blocker never deals combat damage to a player, so the
// player is the defending one), Blind Zealot's "target ... that player
// controls" chosen as the trigger goes on the stack, and an Optional gate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f3c66db-efc1-4571-a4db-371a3d532615",
		Name:            "Latulla's Orders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Latulla's Orders — you may destroy target artifact that player controls", func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheFirstLegalCardTarget(effects.NewContext(g, item))
			}),
				effects.ControlledByThatPlayer(effects.TargetPermanent("target artifact that player controls", effects.Artifact()))),
				"Latulla's Orders — destroy target artifact that player controls?"),
		},
	})
}
