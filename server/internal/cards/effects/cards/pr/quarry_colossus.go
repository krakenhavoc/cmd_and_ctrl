package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quarry Colossus — Creature — Giant {5}{W}{W}, 5/6 (Journey into Nyx, S82 OLD,
// #5558):
//
//	"When this creature enters, put target creature into its owner's library
//	 just beneath the top X cards of that library, where X is the number of
//	 Plains you control."
//
// Unexpectedly Absent's tuck with X read as the trigger resolves: "just
// beneath the top X cards" is X + 1 cards down, and a library of fewer cards
// takes the creature on the bottom (CR 401.4). It goes to its owner's library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad5e2208-df68-4119-b225-a7c133b70359",
		Name:         "Quarry Colossus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Quarry Colossus — put target creature into its owner's library just beneath the top X cards, where X is the number of Plains you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						x := effects.CountControlled(g, ctx.Controller(), func(p game.Card) bool { return p.HasSubtypeWord("Plains") })
						return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{t.ID}, game.TuckOptions{Depth: x + 1}, game.KeyedTail{})
					}
					return nil
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
