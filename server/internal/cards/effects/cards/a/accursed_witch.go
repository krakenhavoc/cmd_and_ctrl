package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Accursed Witch // Infectious Curse — Creature — Human Shaman {3}{B}, 4/2 //
// Enchantment — Aura Curse (Shadows over Innistrad, S82 OLD):
//
//	Front: "Spells your opponents cast that target this creature cost {1} less
//	        to cast.
//	        When this creature dies, return it to the battlefield transformed
//	        under your control attached to target opponent."
//	Back:  "Enchant player
//	        Spells you cast that target enchanted player cost {1} less to cast.
//	        At the beginning of enchanted player's upkeep, that player loses 1
//	        life and you gain 1 life."
//
// The reduction is Boreal Elemental's target-reading modifier turned into a
// discount (CR 601.2c, read at 601.2f). The dies trigger is Vengeful
// Strangler's: it targets an opponent as it goes on the stack and returns the
// card back face up already attached to them
// (ReturnThisCardTransformedAttachedToTarget). The Curse's discount reads the
// player the Curse is attached to, and its upkeep trigger is Curse of Leeches'
// drain: the gain happens whether or not the loss went through.
//
// No simplification.
func init() {
	const oracle = "80b07882-e144-4815-8b6f-04b3ab343d97"
	cheaper := effects.CostsLess(1, "Spells your opponents cast that target this creature cost {1} less to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	cheaper.ReadsTargets = true
	curse := effects.CostsLess(1, "Spells you cast that target enchanted player cost {1} less to cast.",
		effects.YourSpell(), effects.OldB289ThatTargetEnchantedPlayer())
	curse.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:      oracle,
		Name:          "Accursed Witch",
		Completeness:  effects.CompletenessFull,
		CostModifiers: []game.CostModifier{cheaper},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Accursed Witch — return it to the battlefield transformed attached to target opponent",
				effects.ReturnThisCardTransformedAttachedToTarget),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
	effects.Register(effects.Spec{
		OracleID:      oracle + "#1",
		Name:          "Infectious Curse",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.EnchantPlayer(),
		Purpose:       game.Purpose{Aura: game.AuraHarm},
		CostModifiers: []game.CostModifier{curse},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			}, "Infectious Curse — they lose 1 life and you gain 1 life",
				// The gain waits for a prompt the loss asks (a CR 616 order; CR 608.2c, #4155).
				effects.EnchantedPlayerLosesOneYouGainOne()),
		},
	})
}
