package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortal Obstinacy — Enchantment — Aura {W} (Journey into Nyx):
//
//	"Enchant creature you control
//	 Enchanted creature gets +1/+1.
//	 Whenever enchanted creature deals combat damage to a player, you may
//	 sacrifice this Aura. If you do, destroy target enchantment."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), before the
// sacrifice is decided; the destroy happens only if the Aura really left the
// battlefield (the keyed "if you do" tail, as Rotisserie Elemental's).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea29c489-bd04-4f36-8b59-bccaac153152",
		Name:         "Mortal Obstinacy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				},
					"Mortal Obstinacy — you may sacrifice this Aura; if you do, destroy target enchantment",
					effects.Do(effects.MayChoice{
						Question: "Mortal Obstinacy — sacrifice this Aura to destroy target enchantment?",
						YesLabel: "Sacrifice",
						NoLabel:  "Decline",
						Then:     effects.S82OldB367SacrificeAuraAnswer("confirm/mortal-obstinacy-sacrifice", effects.S82OldB367DestroyTargetTail("then/mortal-obstinacy-destroy")),
					})),
				func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec {
					return effects.TargetPermanent("target enchantment", effects.Enchantment())
				}),
		},
	})
}
