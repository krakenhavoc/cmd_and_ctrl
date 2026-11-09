package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thoughtweft Lieutenant — Creature — Kithkin Soldier {G}{W} (Lorwyn Eclipsed, #4458):
//
//	"Whenever this creature or another Kithkin you control enters, target
//	 creature you control gets +1/+1 and gains trample until end of turn."
func init() {
	Register(Spec{
		OracleID:     "50dd9e17-6dae-4ff6-8a38-50088c71283f",
		Name:         "Thoughtweft Lieutenant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return true
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Controller == source.Controller && c.IsCreature() && c.HasSubtype(source.TextSubtype("Kithkin"))
				}, "Thoughtweft Lieutenant — target creature you control gets +1/+1 and gains trample until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Thoughtweft Lieutenant — +1/+1 until end of turn"}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Thoughtweft Lieutenant — trample until end of turn"}.Apply(ctx)
					}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
