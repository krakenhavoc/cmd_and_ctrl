package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gideon, Martial Paragon — Legendary Planeswalker — Gideon {4}{W},
// starting loyalty 5:
//
//	"+2: Untap all creatures you control. Those creatures get +1/+1
//	 until end of turn.
//	 0: Until end of turn, Gideon becomes a 5/5 Human Soldier creature
//	 with indestructible that's still a planeswalker. Prevent all damage
//	 that would be dealt to him this turn.
//	 −10: Creatures you control get +2/+2 until end of turn. Tap all
//	 creatures your opponents control."
//
// THE +2 untaps the set and pumps the same set: both read "creatures you
// control" as the ability resolves (CR 611.2c locks the pump to those
// creatures, so one that enters afterwards is not pumped). THE 0 is
// animateGideon (gideon_animate.go; #2046, ADR 0032 amendment of
// 2026-10-07). THE −10 pumps the creatures you control and taps every
// creature an opponent controls, the two clauses in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4cbce730-7b70-41bd-b463-bfec46afe3b0",
		Name:            "Gideon, Martial Paragon",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Untap all creatures you control. Those creatures get +1/+1 until end of turn.",
				Cost:  effects.LoyaltyCost(2),
				Effect: effects.Do(
					effects.UntapAllCreaturesYouControl{},
					effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl()),
						Power:     1,
						Toughness: 1,
						Label:     "Gideon, Martial Paragon — creatures you control get +1/+1",
					},
				),
			},
			{
				Label: "0: Until end of turn, Gideon becomes a 5/5 Human Soldier creature with indestructible that's still a planeswalker. Prevent all damage that would be dealt to him this turn.",
				Cost:  effects.LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.AnimateGideon(g, item, effects.GideonAnimation{
						Label:          "Gideon, Martial Paragon — a 5/5 Human Soldier creature with indestructible until end of turn",
						Subtypes:       []string{"Human", "Soldier"},
						Power:          5,
						Toughness:      5,
						Indestructible: true,
					})
				},
			},
			{
				Label: "−10: Creatures you control get +2/+2 until end of turn. Tap all creatures your opponents control.",
				Cost:  effects.LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.YouControl()),
						Power:     2,
						Toughness: 2,
						Label:     "Gideon, Martial Paragon — creatures you control get +2/+2",
					}).Apply(ctx); err != nil {
						return err
					}
					return effects.TapAllCreaturesYourOpponentsControl(g, item)
				},
			},
		},
	})
}
