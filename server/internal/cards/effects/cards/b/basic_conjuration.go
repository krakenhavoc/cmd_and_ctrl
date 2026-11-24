package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Basic Conjuration — {1}{G}{G} Sorcery — Lesson (Strixhaven: School of
// Mages, #5502):
//
//	"Look at the top six cards of your library. You may reveal a creature card
//	 from among them and put it into your hand. Put the rest on the bottom of
//	 your library in a random order. You gain 3 life."
//
// Monumental Henge's look-and-reveal (the registered take that reveals the
// taken card and bottoms the rest in a random order), over six cards and a
// creature filter. The life is gained after the take is set up, as the text
// orders it, and is gained whether or not a creature was taken.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "547ab16c-b8d1-4357-9219-7cabde80f5ba",
		Name:         "Basic Conjuration",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			if err := (effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 6),
				Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() },
				Max:      1,
				Optional: true,
				Label:    "Basic Conjuration — you may reveal a creature card from among them and put it into your hand",
				Key:      effects.TakeRevealRestBottomRandom,
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: player, Amount: 3}.Apply(ctx)
		},
	})
}
