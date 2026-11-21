package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "83860749-1603-46e9-9065-e98001244193",
		Name:            "Treebeard, Gracious Host",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Treebeard, Gracious Host — ward {2}"),
			WhenThisEnters("Treebeard, Gracious Host — create two Food tokens",
				Do(CreateToken{Template: FoodToken(), N: 2})),
			Targeting(WheneverYouGainLife("Treebeard, Gracious Host — put that many +1/+1 counters on target Halfling or Treefolk",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 || item.Trigger == nil {
						return nil
					}
					return AddCounter{Target: ts[0].ID, Kind: game.CounterPlusOne, N: item.Trigger.Event.Amount}.Apply(ctx)
				}),
				TargetPermanent("target Halfling or Treefolk", Or(Subtype("Halfling"), Subtype("Treefolk")))),
		},
	})
}
