package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Druid of the Emerald Grove — {3}{G} Creature — Dwarf Druid 2/2
//
//	"When this creature enters, search your library for up to two basic land
//	 cards and reveal them, then roll a d20.
//	 9 or less | Put those cards into your hand, then shuffle.
//	 10—19 | Put one of those cards onto the battlefield tapped and the other
//	 into your hand, then shuffle.
//	 20+ | Put those cards onto the battlefield tapped, then shuffle."
//
// The search finds the cards first and leaves them in the library (CR 701.23),
// so the roll can't change which cards are chosen. The table then moves them:
// both to the hand; one tapped onto the battlefield (the controller's pick)
// and the other to the hand; or both tapped onto the battlefield. The library
// is shuffled afterwards in every case, including when nothing was found.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acf54a85-0e9e-43fb-99d9-c223c02f13c4",
		Name:         "Druid of the Emerald Grove",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Druid of the Emerald Grove — search your library for up to two basic land cards, reveal them, then roll a d20",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
						Player:   ctx.Controller(),
						Source:   ctx.Source(),
						Pred:     effects.IsBasicLand,
						Dest:     game.ZoneLibrary,
						Limit:    2,
						Reveal:   true,
						Shuffle:  false,
						ToTop:    true,
						Reason:   "Druid of the Emerald Grove — search your library for up to two basic land cards",
						ThenTail: ctx.ItemTail(druidOfTheEmeraldGroveFound),
					})
				}),
		},
	})
}

// druidOfTheEmeraldGroveFound is "then roll a d20", the found cards carried to
// the table. An on-disk key: never renamed, never reused.
var druidOfTheEmeraldGroveFound = effects.CardTail("tail/gap-b216-druid-emerald-grove-found", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.RollDice{Sides: 20, N: 1, Then: druidOfTheEmeraldGroveTable, TailCarry: append([]uuid.UUID(nil), o.IDs...)}.Apply(ctx)
})

// druidOfTheEmeraldGroveTable is the results table. An on-disk key.
var druidOfTheEmeraldGroveTable = effects.CardTail("dice/gap-b216-druid-emerald-grove-table", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	player := ctx.Controller()
	found := o.Tail.Carry
	if len(found) == 0 {
		return effects.ShuffleLibrary{Player: player}.Apply(ctx)
	}
	switch r := o.Rolls[0]; {
	case r >= 20:
		return effects.PutFromLibraryOntoBattlefield{
			Player: player, Cards: found, All: true, Tapped: true,
			Label:    "Druid of the Emerald Grove — put those cards onto the battlefield tapped",
			ThenTail: druidOfTheEmeraldGroveShuffle,
		}.Apply(ctx)
	case r >= 10:
		return effects.PutFromLibraryOntoBattlefield{
			Player: player, Cards: found, Max: 1, Tapped: true,
			Label:    "Druid of the Emerald Grove — put one of those cards onto the battlefield tapped and the other into your hand",
			ThenTail: druidOfTheEmeraldGroveRest,
		}.Apply(ctx)
	default:
		ctx.Game.TakeFromLibraryToHandForEffect(player, found)
		return effects.ShuffleLibrary{Player: player}.Apply(ctx)
	}
})

// druidOfTheEmeraldGroveRest puts the card not put onto the battlefield into
// the hand, then shuffles. An on-disk key.
var druidOfTheEmeraldGroveRest = effects.PutFromLibraryTail("druid-emerald-grove/rest-to-hand",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		ctx.Game.TakeFromLibraryToHandForEffect(res.Player, res.Rest)
		return effects.ShuffleLibrary{Player: res.Player}.Apply(ctx)
	})

// druidOfTheEmeraldGroveShuffle is "then shuffle". An on-disk key.
var druidOfTheEmeraldGroveShuffle = effects.PutFromLibraryTail("druid-emerald-grove/shuffle",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		return effects.ShuffleLibrary{Player: res.Player}.Apply(ctx)
	})
