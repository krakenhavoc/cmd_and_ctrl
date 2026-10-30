package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boomerang, Blade Flinger — Legendary Creature — Human Villain {1}{B},
// 1/3 (Marvel Super Heroes, #4455):
//
//	"Whenever Boomerang attacks, he deals 1 damage to each opponent and
//	 you gain 1 life."
//
// An attack trigger on the creature itself (ThisAttacked, the Cecily
// shape). The damage is damageToEachOpponent, the shared "pings the
// table" body, and the life gain is GainLife for the controller. The
// damage comes first, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bbd84035-3efd-4301-9778-913898efd57a",
		Name:         "Boomerang, Blade Flinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Boomerang, Blade Flinger — 1 damage to each opponent and gain 1 life",
				Do(Instruction(boomerangPingAndGain))),
		},
	})
}

// boomerangPingAndGain is the attack trigger's body: one damage to every
// opponent, then one life for the controller.
func boomerangPingAndGain(ctx *Context) error {
	if err := damageToEachOpponent(ctx.Game, ctx.Item, 1); err != nil {
		return err
	}
	return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
}
