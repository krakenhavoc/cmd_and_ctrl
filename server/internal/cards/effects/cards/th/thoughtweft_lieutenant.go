package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thoughtweft Lieutenant — Creature — Kithkin Soldier {G}{W} (Lorwyn Eclipsed, #4458):
//
//	"Whenever this creature or another Kithkin you control enters, target
//	 creature you control gets +1/+1 and gains trample until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50dd9e17-6dae-4ff6-8a38-50088c71283f",
		Name:         "Thoughtweft Lieutenant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return true
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Controller == source.Controller && c.IsCreature() && c.HasSubtype(source.TextSubtype("Kithkin"))
				}, "Thoughtweft Lieutenant — target creature you control gets +1/+1 and gains trample until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Thoughtweft Lieutenant — +1/+1 until end of turn"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Thoughtweft Lieutenant — trample until end of turn"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
