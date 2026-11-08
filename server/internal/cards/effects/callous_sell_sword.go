package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Callous Sell-Sword // Burn Together — Creature — Human Soldier {1}{B},
// 2/2 // Sorcery — Adventure {R} (Wilds of Eldraine, #4471):
//
//	Callous Sell-Sword — "This creature enters with a +1/+1 counter on it
//	 for each creature that died under your control this turn."
//	Burn Together — "Target creature you control deals damage equal to
//	 its power to any other target. Then sacrifice it. (Then exile this
//	 card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The counters are a CR 614 entry replacement, Diregraf
// Colossus's counted shape, sized by your per-turn death tally
// (PlayerTurnTally.CreaturesDied, CR 700.4) as the creature enters, so a
// creature that died before the Sell-Sword was cast still counts.
//
// Burn Together is Special Move's Foot Toss as a whole spell: two clauses,
// the second "any other target" (Distinct, so the thrower is never its own
// target). The damage is the thrower's power as the spell resolves, and
// the thrower is sacrificed afterwards. When the other target has become
// illegal no damage is dealt but the still-legal thrower is sacrificed
// (CR 608.2b); when the thrower itself has become illegal nothing happens.
// The Adventure lifecycle is the engine's (CR 715.3d).
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
			}, "Callous Sell-Sword: enters with a +1/+1 counter for each creature that died under your control this turn"),
		},
	})
	Register(Spec{
		OracleID:     callousSellSwordOracleID + "#1",
		Name:         "Burn Together",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(specialMoveAnyOtherTarget()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			thrower, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if victim, ok := ctx.ClauseTarget(1); ok {
				ctx.Game.RecomputeLayersIfStaleLocked()
				if c, found := ctx.Game.LookupCardForEffect(thrower.ID); found && c.CurrentPower() > 0 {
					if err := (DealDamage{Source: thrower.ID, Target: victim.ID, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return SacrificePermanent{Target: thrower.ID}.Apply(ctx)
		},
	})
}
