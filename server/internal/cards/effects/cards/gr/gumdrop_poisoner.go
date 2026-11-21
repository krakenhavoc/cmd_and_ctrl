package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gumdrop Poisoner // Tempt with Treats — Creature — Human Warlock {2}{B},
// 3/2 // Instant — Adventure {B} (Wilds of Eldraine, #4471):
//
//	Gumdrop Poisoner — "Lifelink
//	 When this creature enters, up to one target creature gets -X/-X until
//	 end of turn, where X is the amount of life you gained this turn."
//	Tempt with Treats — "Create a Food token. (Then exile this card. You may
//	 cast the creature later from exile.)"
//
// X is read as the trigger resolves, from the controller's life-gain tally
// for the turn (the tally counts the life this very Poisoner's lifelink
// gained earlier, and any gain since the trigger went on the stack). With
// no target chosen, or none still legal, nothing happens. Tempt with
// Treats is the Food an Adventure instant makes.
//
// No simplification.
const GumdropPoisonerOracleID = "4f5ebbb8-f49b-488f-bb85-888aaa918bdd"

func init() {
	effects.Register(effects.Spec{
		OracleID:        GumdropPoisonerOracleID,
		Name:            "Gumdrop Poisoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(
				"Gumdrop Poisoner — up to one target creature gets -X/-X until end of turn, where X is the life you gained this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					x := effects.B15LifeGainedThisTurn(g, item.Controller)
					if !ok || x <= 0 {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: -x,
						Label: "Gumdrop Poisoner — shrink until end of turn"}.Apply(ctx)
				}),
				effects.UpToOneTargetCreature("up to one target creature")),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     GumdropPoisonerOracleID + "#1",
		Name:         "Tempt with Treats",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.FoodToken(), N: 1}.Apply(ctx)
		},
	})
}
