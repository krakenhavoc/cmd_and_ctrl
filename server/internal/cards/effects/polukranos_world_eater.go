package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polukranos, World Eater — 5/5 legendary Hydra for {2}{G}{G}:
//
//	"{X}{X}{G}: Monstrosity X. (If this creature isn't monstrous, put X
//	 +1/+1 counters on it and it becomes monstrous.)
//	 When Polukranos becomes monstrous, it deals X damage divided as
//	 you choose among any number of target creatures your opponents
//	 control. Each of those creatures deals damage equal to its power
//	 to Polukranos."
//
// X is the monstrosity's announced X (CR 701.37c), carried on the
// triggering event; a Doubling Season doubles the counters and not
// the damage.
//
// The division is #1563's, sized by the EVENT rather than by an
// announced X: the trigger's clause comes from TargetsFrom, which
// reads X off the trigger context and declares "up to X targets, X
// divided among them" — a fixed Total, so the CR 603.3d target walk
// settles the split exactly as it does for Inferno Titan's 3. "Any
// number" is capped at X because every chosen target must be assigned
// at least 1 (CR 601.2d). With X = 0 the clause is nil: nothing is
// targeted, nothing is dealt, nothing hits back.
//
// The fight-back reads each creature's power as the ability resolves
// and is dealt only by targets that are still legal (CR 608.2b); a
// target that left in response neither takes nor deals damage. The
// two sentences are sequential and state-based actions wait for the
// whole resolution, so a creature Polukranos has just dealt lethal
// damage still hits back.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecbf5560-853a-43a9-bf70-ac8d403b0ce8",
		Name:         "Polukranos, World Eater",
		Completeness: CompletenessFull,
		// X=0 makes Polukranos monstrous for nothing and spends the
		// ability for good (CR 701.37a) — not a move worth offering.
		XMatters:  true,
		Activated: []ActivatedAbility{MonstrosityX(ManaCost("{X}{X}{G}"))},
		Triggered: []game.TriggeredAbility{polukranosTrigger()},
	})
}

func polukranosTrigger() game.TriggeredAbility {
	t := WhenBecomesMonstrous("Polukranos — X damage divided among opponents' creatures", polukranosBites)
	t.TargetsFrom = polukranosTargets
	return t
}

// polukranosTargets is the X-sized clause. A pure read of the trigger
// context, so restore can re-derive it (no TargetsFromReadsBoard).
func polukranosTargets(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	x := tc.Amount()
	if x <= 0 {
		return nil
	}
	return TargetCreature("any number of target creatures your opponents control", OpponentControls()).
		WithCount(0, x).
		Dividing(Divide(x))
}

func polukranosBites(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := DealDividedDamage(ctx); err != nil {
		return err
	}
	self := ctx.Source()
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		if err := (DealDamage{Source: t.ID, Target: self, Amount: c.CurrentPower()}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
