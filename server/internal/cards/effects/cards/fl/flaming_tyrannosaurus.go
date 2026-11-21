package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flaming Tyrannosaurus — Creature — Dinosaur {5}{R}{R}, 5/5 (Doctor Who):
//
//	"Menace
//	 Paradox — Whenever you cast a spell from anywhere other than your hand, this
//	 creature deals 3 damage to any target. Then put a +1/+1 counter on this
//	 creature.
//	 When this creature dies, it deals damage equal to its power to each opponent."
//
// The Paradox trigger chooses its target as it goes on the stack and re-checks it as
// it resolves (CR 608.2b); the counter is placed afterwards whether or not the
// damage was dealt, as long as the Tyrannosaurus is still the same object on the
// battlefield. The death trigger deals damage equal to its power as it last was
// (CR 603.10a, so counters and anthems count) to each opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "89505f24-6c96-4062-8171-1851f7e7395b",
		Name:            "Flaming Tyrannosaurus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventCast, effects.GapB87YouCastFromOutsideYourHand,
					"Paradox — this creature deals 3 damage to any target, then put a +1/+1 counter on it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if ts := ctx.LegalTargets(); len(ts) > 0 {
							if err := (effects.DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: 3}).Apply(ctx); err != nil {
								return err
							}
						}
						if !effects.SourceIsStillThisPermanent(g, item) {
							return nil
						}
						return effects.PlusOneCountersOnThis(1)(g, item)
					}),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.WhenThisDies("When this creature dies, it deals damage equal to its power to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Power <= 0 {
						return nil
					}
					// One printed instruction: one instance of damage to every opponent (CR 615.8).
					return g.DamageInstanceForEffect(func() error {
						for _, opp := range ctx.Opponents() {
							if err := (effects.DealDamage{Source: item.SourceCardID, Target: opp, Amount: info.Power}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				}),
		},
	})
}
