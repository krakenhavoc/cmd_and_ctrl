package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ant-Man, Reformed Rogue — Legendary Creature — Human Rogue Hero
// {2}{G/U}, 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever you cast a green spell, Ant-Man gets +1/+0 and gains
//	 trample until end of turn.
//	 Whenever you cast a blue spell, Ant-Man gets -1/-0 until end of turn
//	 and can't be blocked this turn.
//	 Whenever Ant-Man deals combat damage to a player, draw a card."
//
// A two-colour spell fires both cast triggers. Each changes only the
// Ant-Man that was on the battlefield when it triggered (CR 400.7).
func init() {
	Register(Spec{
		OracleID:     "c51c429d-d8d7-40d6-9760-72c26acaed42",
		Name:         "Ant-Man, Reformed Rogue",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(OfColor("G"), "Ant-Man — gets +1/+0 and gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Ant-Man — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"}, Label: "Ant-Man — trample"}.Apply(ctx)
				}),
			WheneverYouCast(OfColor("U"), "Ant-Man — gets -1/-0 until end of turn and can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (BoostUntilEOT{Target: item.SourceCardID, Power: -1, Label: "Ant-Man — -1/-0"}).Apply(ctx); err != nil {
						return err
					}
					return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Ant-Man — can't be blocked"}.Apply(ctx)
				}),
			WheneverThisDealsCombatDamageToAPlayer("Ant-Man — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
