package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wakandan Royal Guard — {4}{G} Creature — Human Soldier Hero 4/4 (Marvel
// Super Heroes, #4455):
//
//	"Vigilance
//	 When this creature enters, put a +1/+1 counter on target creature.
//	 If that creature is another Hero, put two +1/+1 counters on it
//	 instead."
//
// Shortcut to Mushrooms' shape (an enters trigger that targets a creature
// and counters it). The Hero test is read as the trigger resolves, from
// the target's current subtypes (CR 608.2b), and "another" means the
// target is not the Guard itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4ae15bd1-e0ea-4c1a-a311-dae421381ebc",
		Name:            "Wakandan Royal Guard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Wakandan Royal Guard — put a +1/+1 counter on target creature (two if it's another Hero)",
					wakandanRoyalGuardCounters),
				effects.TargetCreature("target creature")),
		},
	})
}

// wakandanRoyalGuardCounters puts one +1/+1 counter on the target, or two
// when it is another Hero.
func wakandanRoyalGuardCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		n := 1
		if c, ok := g.LookupCardForEffect(t.ID); ok && t.ID != item.SourceCardID && c.HasSubtype(ctx.TextSubtype("Hero")) {
			n = 2
		}
		return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	}
	return nil
}
