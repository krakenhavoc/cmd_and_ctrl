package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Punishing Punch — Instant {2}{G} (Marvel Super Heroes, #4455):
//
//	"This spell costs {2} less to cast if there are two or more creature
//	 cards in your graveyard.
//	 Target creature you control deals damage equal to twice its power to
//	 target creature an opponent controls."
//
// Bite Down's two target clauses (slot 0 the damage source, slot 1 the
// recipient) with the amount doubled: the power is read as the spell
// resolves, so a pump in response scales it, and the damage carries the
// source's deathtouch and lifelink. A creature that left in response
// deals nothing; a recipient that left takes nothing (CR 608.2b). The
// reduction is a self cost modifier read when the cost is determined
// (CR 601.2f): the card being cast is not in the graveyard, so it never
// counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "454c4e8b-e417-4381-a931-6a11356ef64e",
		Name:         "Punishing Punch",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if there are two or more creature cards in your graveyard.",
				func(q game.CostQuery) bool { return CountCardsInYourGraveyard(Creature())(q) >= 2 }),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok {
				return nil
			}
			return DealDamage{Source: biter.ID, Target: victim.ID, Amount: 2 * c.CurrentPower()}.Apply(ctx)
		},
	})
}
