package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// You Meet in a Tavern — Sorcery {2}{G}{G} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Choose one —
//	 • Form a Party — Look at the top five cards of your library. You may
//	   reveal any number of creature cards from among them and put them into
//	   your hand. Put the rest on the bottom of your library in a random order.
//	 • Start a Brawl — Creatures you control get +2/+2 until end of turn."
//
// The look is Commune with Beavers' dig with no ceiling on the pick ("any
// number", Max 0) and Monumental Henge's rest, revealed and bottomed at
// random. The pump is a mass boost over the creatures you control as the
// spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c735077a-fae8-4c5f-a5bf-35969696460c",
		Name:         "You Meet in a Tavern",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Form a Party — Look at the top five cards of your library. You may reveal any number of creature cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player:   player,
						Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
						Match:    effects.Creature(),
						Max:      0,
						Optional: true,
						Label:    "You Meet in a Tavern — reveal any number of creature cards and put them into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(ctx)
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Start a Brawl — Creatures you control get +2/+2 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2, Label: "You Meet in a Tavern — Start a Brawl"}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
		),
	})
}
