package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Draconic Roar — Instant {1}{R} (Jumpstart, #5510):
//
//	"As an additional cost to cast this spell, you may reveal a Dragon card
//	 from your hand.
//	 Draconic Roar deals 3 damage to target creature. If you revealed a
//	 Dragon card or controlled a Dragon as you cast this spell, Draconic
//	 Roar deals 3 damage to that creature's controller."
//
// The optional reveal is the engine's optional behold (a Dragon card in
// hand, revealed to the table, or a Dragon you control chosen): either way
// "revealed a Dragon card or controlled a Dragon" holds, and beholding a
// Dragon you control only happens when you control one, so the clause reads
// as printed. "Controlled a Dragon as you cast" is a cast-time value (CR
// 601.2i), so a Dragon that dies in response still counts. The creature's
// controller is read before the damage (CR 608.2h), and a target that has
// become illegal makes the whole spell fizzle (CR 608.2b).
//
// No simplification.
func init() {
	const dragonKey = "controlled_a_dragon"
	effects.Register(effects.Spec{
		OracleID:      "843af5d3-447e-40e4-a562-f5a7ed09ca3e",
		Name:          "Draconic Roar",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.OptionalBehold("a", "Dragon")},
		CastTimeValues: []game.CastTimeValue{{
			Key:   dragonKey,
			Value: effects.ControlsADragonAsCast,
		}},
		Targets: effects.TargetCreature("target creature"),
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := victim.Controller
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			controlled, _ := ctx.CastValue(dragonKey)
			if !ctx.WasBeheld() && controlled == 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 3}.Apply(ctx)
		},
	})
}
