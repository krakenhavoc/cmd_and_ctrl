package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Charmer — Creature — Worm Mercenary {2}{B}, 2/2 (Prophecy, S82 OLD b670,
// #5617):
//
//	"Whenever this creature deals combat damage to a creature, that creature's
//	 controller loses 2 life unless they pay {2}."
//
// The shared damaged-creature pay-or-else trigger: the creature's controller (as
// of the damage) is asked, and a decline makes that player lose 2 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "491606d8-e9c1-4f89-bbef-b7590d310097",
		Name:         "Death Charmer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB670WheneverThisDealsCombatDamageToACreatureUnlessItsControllerPays(
				"Death Charmer — that creature's controller loses 2 life unless they pay {2}", "{2}",
				"Death Charmer — pay {2}, or lose 2 life?", deathCharmerLoseLife),
		},
	})
}

// deathCharmerLoseLife is the decline: the payer loses 2 life. Keyed, an on-disk
// identity.
var deathCharmerLoseLife = effects.PayUnlessElse("pay-unless/s82-old-b670-death-charmer-lose-life",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.GainLife{Player: r.Chooser, Amount: -2}.Apply(ctx)
	})
