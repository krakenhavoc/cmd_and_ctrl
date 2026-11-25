package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Collision // Colossus — split card {1}{R/G} // {R}{G} (Ravnica Allegiance):
//
//	Collision (Instant): "Collision deals 6 damage to target creature with
//	 flying."
//	Colossus (Instant): "Target creature gets +4/+2 and gains trample until
//	 end of turn."
//
// Collision's target is legal only while the creature has flying (CR 608.2b),
// read through the layered view. Colossus is the plain pump-and-grant body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "197ce21b-3387-4e76-a49c-04b9c371c031",
		Name:         "Collision",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 6)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DealDamageToTheTarget(6)(ctx.Game, item)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "197ce21b-3387-4e76-a49c-04b9c371c031#1",
		Name:         "Colossus",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 2, "trample")),
		OnResolve:    effects.SocB27TargetGetsAndGains(4, 2, "trample", "Colossus"),
	})
}
