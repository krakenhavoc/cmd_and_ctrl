package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gandalf, Shadow's Foe — Legendary Creature — Avatar Wizard {5}{U}{U},
// 3/4 (The Hobbit Eternal, HOC):
//
//	"Vigilance
//	 When Gandalf enters, exile up to three target lands you control, then
//	 return them to the battlefield tapped under their owner's control.
//	 Landfall — Whenever a land you control enters, draw a card and put a
//	 +1/+1 counter on Gandalf."
//
// The blink is Ghostly Flicker's: the lands leave together (one exile) and
// come back as one entry, so each sees the others enter (CR 603.6a), each as
// a new object (CR 400.7) tapped, and every one of them triggers the landfall
// ability (Gandalf is already on the battlefield). A target that left in
// response is skipped (CR 608.2b); a land that cannot return stays exiled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6760d7d1-098e-4131-8824-53e5427d047b",
		Name:            "Gandalf, Shadow's Foe",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Gandalf — exile up to three target lands you control, then return them tapped",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						var ids []uuid.UUID
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard && t.ID != uuid.Nil {
								ids = append(ids, t.ID)
							}
						}
						if len(ids) == 0 {
							return nil
						}
						return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(gandalfShadowsFoeReturnLands))
					}),
				TargetPermanent("up to three target lands you control", Land(), YouControl()).WithCount(0, 3)),
			Landfall("Gandalf — draw a card and put a +1/+1 counter on Gandalf",
				Do(DrawCards{N: 1}, Instruction(func(ctx *Context) error {
					return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}))),
		},
	})
}

// gandalfShadowsFoeReturnLands returns what reached exile tapped, under its
// owner's control, together.
var gandalfShadowsFoeReturnLands = CardTail("gandalf-shadows-foe/return-lands-tapped", func(ctx *Context, o game.TailOutcome) error {
	return ReturnFromExileTogether{Targets: o.IDs, Tapped: true}.Apply(ctx)
})
