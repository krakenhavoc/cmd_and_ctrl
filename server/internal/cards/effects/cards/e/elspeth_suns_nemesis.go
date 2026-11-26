package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elspeth, Sun's Nemesis — Legendary Planeswalker — Elspeth {2}{W}{W}, loyalty 5:
//
//	"−1: Up to two target creatures you control each get +2/+1 until end of turn.
//	 −2: Create two 1/1 white Human Soldier creature tokens.
//	 −3: You gain 5 life.
//	 Escape—{4}{W}{W}, Exile four other cards from your graveyard."
func init() {
	effects.Register(effects.Spec{
		OracleID:         "faf296b0-59e5-4ec9-9912-e5e21c0ce9f5",
		Name:             "Elspeth, Sun's Nemesis",
		Completeness:     effects.CompletenessFull,
		StartingLoyalty:  5,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{4}{W}{W}", 4)},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "−1: Up to two target creatures you control each get +2/+1 until end of turn.",
				Cost:    effects.LoyaltyCost(-1),
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 1)),
				Targets: effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1, Label: "Elspeth, Sun's Nemesis — +2/+1 until end of turn"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:  "−2: Create two 1/1 white Human Soldier creature tokens.",
				Cost:   effects.LoyaltyCost(-2),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 2}),
			},
			{
				Label:  "−3: You gain 5 life.",
				Cost:   effects.LoyaltyCost(-3),
				Effect: effects.Do(effects.GainLife{Amount: 5}),
			},
		},
	})
}
