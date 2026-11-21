package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treebeard, Gracious Host — Legendary Creature — Treefolk {2}{G}{W}, 0/5
// (Tales of Middle-earth Commander):
//
//	"Trample, ward {2}
//	 When Treebeard enters, create two Food tokens.
//	 Whenever you gain life, put that many +1/+1 counters on target
//	 Halfling or Treefolk."
//
// Light of Promise's "that many": the trigger reads the life actually
// gained off its event. Any Halfling or Treefolk permanent is a legal
// target, Treebeard himself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83860749-1603-46e9-9065-e98001244193",
		Name:            "Treebeard, Gracious Host",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Treebeard, Gracious Host — ward {2}"),
			effects.WhenThisEnters("Treebeard, Gracious Host — create two Food tokens",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 2})),
			effects.Targeting(effects.WheneverYouGainLife("Treebeard, Gracious Host — put that many +1/+1 counters on target Halfling or Treefolk",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 || item.Trigger == nil {
						return nil
					}
					return effects.AddCounter{Target: ts[0].ID, Kind: game.CounterPlusOne, N: item.Trigger.Event.Amount}.Apply(ctx)
				}),
				effects.TargetPermanent("target Halfling or Treefolk", effects.Or(effects.Subtype("Halfling"), effects.Subtype("Treefolk")))),
		},
	})
}
