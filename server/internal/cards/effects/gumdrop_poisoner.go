package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const gumdropPoisonerOracleID = "4f5ebbb8-f49b-488f-bb85-888aaa918bdd"

func init() {
	Register(Spec{
		OracleID:        gumdropPoisonerOracleID,
		Name:            "Gumdrop Poisoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(
				"Gumdrop Poisoner — up to one target creature gets -X/-X until end of turn, where X is the life you gained this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					x := b15LifeGainedThisTurn(g, item.Controller)
					if !ok || x <= 0 {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: -x, Toughness: -x,
						Label: "Gumdrop Poisoner — shrink until end of turn"}.Apply(ctx)
				}),
				UpToOneTargetCreature("up to one target creature")),
		},
	})
	Register(Spec{
		OracleID:     gumdropPoisonerOracleID + "#1",
		Name:         "Tempt with Treats",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: FoodToken(), N: 1}.Apply(ctx)
		},
	})
}
