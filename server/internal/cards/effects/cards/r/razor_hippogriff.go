package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Razor Hippogriff — Creature — Hippogriff {3}{W}{W}, 3/3 (Commander 2013, S82
// OLD b387):
//
//	"Flying
//	 When this creature enters, return target artifact card from your graveyard
//	 to your hand. You gain life equal to that card's mana value."
//
// Workshop Assistant's regrowth of an artifact card, with Sheoldred's
// Restoration's life read off the card as it sat in the graveyard (CR 608.2h;
// an X in its cost is 0). A target gone by resolution makes the ability do
// nothing (CR 608.2b), and no life is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d121108e-f0bc-469b-bf94-e5e5308014a2",
		Name:            "Razor Hippogriff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Razor Hippogriff — return target artifact card from your graveyard to your hand and gain life equal to its mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
						return nil
					}
					card, ok := g.LookupCardForEffect(item.Targets[0].ID)
					if !ok {
						return nil
					}
					if err := (effects.ReturnFromGraveyard{Target: card.InstanceID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: card.ManaValue()}.Apply(ctx)
				}),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn())),
		},
	})
}
