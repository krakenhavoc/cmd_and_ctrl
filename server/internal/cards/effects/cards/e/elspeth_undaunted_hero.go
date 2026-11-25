package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth, Undaunted Hero — Legendary Planeswalker — Elspeth {2}{W}{W}{W},
// loyalty 5 (Theros Beyond Death, #5514):
//
//	"+2: Put a +1/+1 counter on each of up to two target creatures.
//	 −2: Search your library and/or graveyard for a card named Sunlit
//	 Hoplite and put it onto the battlefield. If you search your library
//	 this way, shuffle.
//	 −8: Until end of turn, creatures you control gain flying and get +X/+X,
//	 where X is your devotion to white."
//
// The +2 puts the counters through the counter pipeline on each target still
// legal as it resolves (CR 608.2b). The −2 is Finale of Devastation's search
// over both zones (the zones are asked first, so a graveyard-only search does
// not shuffle) for the one name. The −8 reads devotion once, as it resolves
// (CR 608.2h), and locks the set of creatures you control then (CR 611.2c):
// +X/+X and flying until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3f9a797b-85e7-474e-aefe-7e2ec08fc5b7",
		Name:            "Elspeth, Undaunted Hero",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Put a +1/+1 counter on each of up to two target creatures.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label: "−2: Search your library and/or graveyard for a card named Sunlit Hoplite and put it onto the battlefield. If you search your library this way, shuffle.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.Name == "Sunlit Hoplite" },
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "Elspeth, Undaunted Hero — a card named Sunlit Hoplite",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label: "−8: Until end of turn, creatures you control gain flying and get +X/+X, where X is your devotion to white.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.DevotionTo(g, item.Controller, ctx.TextColor("W"))
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{
						Match: yours, Power: x, Toughness: x,
						Label: "Elspeth, Undaunted Hero — +X/+X",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{
						Match: yours, Keywords: []string{"flying"},
						Label: "Elspeth, Undaunted Hero — flying",
					}.Apply(ctx)
				},
			},
		},
	})
}
