package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamespeaker's Will — Enchantment — Aura {R} (Journey into Nyx):
//
//	"Enchant creature you control
//	 Enchanted creature gets +1/+1.
//	 Whenever enchanted creature deals combat damage to a player, you may
//	 sacrifice this Aura. If you do, destroy target artifact."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), before the
// sacrifice is decided; the destroy happens only if the Aura really left the
// battlefield (the keyed "if you do" tail, as Rotisserie Elemental's).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3abc09a9-3baa-45ff-b048-53e42d4fa9f2",
		Name:         "Flamespeaker's Will",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
				},
					"Flamespeaker's Will — you may sacrifice this Aura; if you do, destroy target artifact",
					effects.Do(effects.MayChoice{
						Question: "Flamespeaker's Will — sacrifice this Aura to destroy target artifact?",
						YesLabel: "Sacrifice",
						NoLabel:  "Decline",
						Then:     effects.S82OldB367SacrificeAuraAnswer("confirm/flamespeaker-s-will-sacrifice", effects.S82OldB367DestroyTargetTail("then/flamespeaker-s-will-destroy")),
					})),
				func(game.TriggerContext, *game.Card, *game.Game) *game.TargetSpec {
					return effects.TargetPermanent("target artifact", effects.Artifact())
				}),
		},
	})
}
