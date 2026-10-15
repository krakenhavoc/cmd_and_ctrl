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
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (MillCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
						return err
					}
					x := 2 * b11CreatureCardsInGraveyard(g, item.Controller)
					if x == 0 {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
				}),
		},
	})
}
