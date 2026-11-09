package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Callous Sell-Sword // Burn Together — Creature — Human Soldier {1}{B},
// 2/2 // Sorcery — Adventure {R} (Wilds of Eldraine, #4471):
//
//	Callous Sell-Sword — "This creature enters with a +1/+1 counter on it
//	 for each creature that died under your control this turn."
//	Burn Together — "Target creature you control deals damage equal to
//	 its power to any other target. Then sacrifice it."
//
// The counters are Gorma's per-controller tally of creatures that died
// (tokens included, read off last-known information) as a CR 614.1c
// self-replacement. Burn Together is two target clauses read positionally
// (Bite Down's shape, the second any target and Distinct): the first
// creature deals damage if the recipient is still legal, and is sacrificed
// if it is still the legal pick, as each is re-checked on its own at
// resolution (CR 608.2b).
//
// No simplification.
const callousSellSwordOracleID = "0e7405e3-e9e6-4213-a73f-e0c01a7d6721"

func init() {
	Register(Spec{
		OracleID:     callousSellSwordOracleID,
		Name:         "Callous Sell-Sword",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return g.TurnTallyFor(src.Controller).CreaturesDied
			}, "Callous Sell-Sword enters with a +1/+1 counter on it for each creature that died under your control this turn"),
		},
	})
	Register(Spec{
		OracleID:     callousSellSwordOracleID + "#1",
		Name:         "Burn Together",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(TargetAny()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			biter, haveBiter := ctx.ClauseTarget(0)
			haveBiter = haveBiter && biter.Kind == game.TargetCard
			if !haveBiter {
				return nil
			}
			if victim, ok := ctx.ClauseTarget(1); ok {
				if c, found := ctx.Game.LookupCardForEffect(biter.ID); found {
					if err := (DealDamage{Source: biter.ID, Target: victim.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return SacrificePermanent{Target: biter.ID}.Apply(ctx)
		},
	})
}
