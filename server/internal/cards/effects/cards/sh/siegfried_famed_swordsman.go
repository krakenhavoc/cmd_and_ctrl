package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "a7cac6c3-cbb1-41f4-905d-2eb581e8e2e0",
		Name:            "Siegfried, Famed Swordsman",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Siegfried, Famed Swordsman — mill three cards, then +1/+1 counters equal to twice the creature cards in your graveyard",
				effects.Do(
					// The counters wait for a prompt the mill asks (CR 608.2c, #4155). The
					// graveyard is counted again for them, after the mill.
					effects.MillCards{N: 3},
					effects.Instruction(func(ctx *effects.Context) error {
						x := 2 * effects.B11CreatureCardsInGraveyard(ctx.Game, ctx.Item.Controller)
						if x == 0 {
							return nil
						}
						return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
					}),
				)),
		},
	})
}
