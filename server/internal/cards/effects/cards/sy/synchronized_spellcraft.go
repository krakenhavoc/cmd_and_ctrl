package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Synchronized Spellcraft — Instant {4}{R} (Zendikar Rising, S82 OLD
// #5508):
//
//	"Synchronized Spellcraft deals 4 damage to target creature and X damage
//	 to that creature's controller, where X is the number of creatures in
//	 your party. (Your party consists of up to one each of Cleric, Rogue,
//	 Warrior, and Wizard.)"
//
// The creature's controller is read before the damage is dealt, and the
// party as the spell resolves (CR 700.8a); both are dealt by the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfdb13b9-7050-43a3-ac5f-d427942b1616",
		Name:         "Synchronized Spellcraft",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, found := ctx.Game.LookupCardForEffect(id)
			x := effects.FinB9PartySize(ctx.Game, ctx.Controller())
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 4}).Apply(ctx); err != nil {
				return err
			}
			if !found || x <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: x}.Apply(ctx)
		},
	})
}
