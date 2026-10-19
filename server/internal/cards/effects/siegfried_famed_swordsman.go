package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siegfried, Famed Swordsman — Legendary Creature — Human Warrior Rogue
// {3}{B}, 2/2 (deck request #3655):
//
//	"Menace
//	 When Siegfried enters, mill three cards. Then put X +1/+1 counters
//	 on Siegfried, where X is twice the number of creature cards in your
//	 graveyard."
//
// X is counted after the mill, so the milled creatures count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a7cac6c3-cbb1-41f4-905d-2eb581e8e2e0",
		Name:            "Siegfried, Famed Swordsman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Siegfried, Famed Swordsman — mill three cards, then +1/+1 counters equal to twice the creature cards in your graveyard",
				Do(
					// The counters wait for a prompt the mill asks (CR 608.2c, #4155). The
					// graveyard is counted again for them, after the mill.
					MillCards{N: 3},
					Instruction(func(ctx *Context) error {
						x := 2 * b11CreatureCardsInGraveyard(ctx.Game, ctx.Item.Controller)
						if x == 0 {
							return nil
						}
						return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
					}),
				)),
		},
	})
}
