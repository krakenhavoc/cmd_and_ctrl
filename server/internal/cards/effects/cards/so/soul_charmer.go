package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul Charmer — Creature — Human Rebel {2}{W}, 2/2 (Prophecy, S82 OLD b670,
// #5617):
//
//	"Whenever this creature deals combat damage to a creature, you gain 2 life
//	 unless that creature's controller pays {2}."
//
// The shared damaged-creature pay-or-else trigger: a decline gains the
// trigger's controller (not the payer) 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de56a8d9-faeb-4e73-bdc1-cc81a6b6d05c",
		Name:         "Soul Charmer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB670WheneverThisDealsCombatDamageToACreatureUnlessItsControllerPays(
				"Soul Charmer — you gain 2 life unless that creature's controller pays {2}", "{2}",
				"Soul Charmer — pay {2}, or its controller gains 2 life?", soulCharmerGainLife),
		},
	})
}

// soulCharmerGainLife is the decline: the Charmer's controller gains 2 life.
// Keyed, an on-disk identity.
var soulCharmerGainLife = effects.PayUnlessElse("pay-unless/s82-old-b670-soul-charmer-gain-life",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
	})
