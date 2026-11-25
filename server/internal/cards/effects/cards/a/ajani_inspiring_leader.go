package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ajani, Inspiring Leader — Legendary Planeswalker — Ajani {4}{W}{W}, loyalty 5
// (Core Set 2020, #5518):
//
//	"+2: You gain 2 life. Put two +1/+1 counters on up to one target creature.
//	 −3: Exile target creature. Its controller gains 2 life.
//	 −10: Creatures you control gain flying and double strike until end of
//	 turn."
//
// The +2 gains the life whatever the target does, then counters the target if
// it is still legal (CR 608.2b). The −3 reads the creature's controller before
// the exile (CR 608.2h). The −10 locks the set of creatures you control as it
// resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15474e82-94b4-47e1-b91c-3deaf4dbffd6",
		Name:            "Ajani, Inspiring Leader",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: You gain 2 life. Put two +1/+1 counters on up to one target creature.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  effects.OldB155AjaniPlusTwo,
			},
			{
				Label:   "−3: Exile target creature. Its controller gains 2 life.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.OldB155AjaniExile,
			},
			{
				Label: "−10: Creatures you control gain flying and double strike until end of turn.",
				Cost:  effects.LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"flying", "double strike"},
						Label:    "Ajani, Inspiring Leader — flying and double strike",
					}.Apply(ctx)
				},
			},
		},
	})
}
