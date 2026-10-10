package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Repulsor Blast — {3}{R} Sorcery:
//
//	"Teamwork 2 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 2 or more.)
//	 Repulsor Blast deals 5 damage to target creature. If this spell was
//	 cast using teamwork, it also deals 2 damage to that creature's
//	 controller."
//
// #1703's plainest teamwork card. "That creature's controller" is read
// as the spell resolves, before the 5 damage is dealt — the creature is
// still on the battlefield then, and a controller read after a lethal
// hit would find it in a graveyard. No simplification.
func init() {
	Register(Spec{
		OracleID:      "150f920c-c942-4feb-824f-79dd9b531687",
		Name:          "Repulsor Blast",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(2)},
		Targets:       TargetCreature("target creature"),
		Purpose:       ForTargets(DamageToTarget(0, 5)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			victim, found := ctx.Game.LookupCardForEffect(t.ID)
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			if !found || !ctx.UsedTeamwork() {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: 2}.Apply(ctx)
		},
	})
}
