package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impossible Inferno — Instant {4}{R} (Duskmourn: House of Horror,
// #4466):
//
//	"Impossible Inferno deals 6 damage to target creature.
//	 Delirium — If there are four or more card types among cards in your
//	 graveyard, exile the top card of your library. You may play it until
//	 the end of your next turn."
//
// Two sentences in order (CR 608.2c), the second waiting for any prompt
// the damage asks. The damage is Heated Argument's first sentence
// (socB26HeatedArgumentDamage). Delirium is read as the second sentence resolves
// (b16CardTypesInGraveyard). The spell itself is still on the stack, and a
// creature the damage kills dies only once the spell has finished
// resolving (CR 704.3), so neither counts. The exile is Snowslope Hunter's
// (b20ExileTopUntilEndOfNextTurn): "play", so a land can be played. A
// target gone by resolution leaves the spell with no legal target, so it
// does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa7c8092-94da-4147-aa90-0a951d2a8be8",
		Name:         "Impossible Inferno",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 6)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.SocB26HeatedArgumentDamage),
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.B16CardTypesInGraveyard(ctx.Game, ctx.Controller()) < 4 {
					return nil
				}
				return effects.B20ExileTopUntilEndOfNextTurn(ctx.Game, ctx.Item)
			}),
		),
	})
}
