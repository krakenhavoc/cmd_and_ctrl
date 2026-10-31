package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wakanda Forever! — {4}{G}{G} Sorcery (Marvel Super Heroes Commander,
// #4455):
//
//	"Reveal the top six cards of your library. You may put a permanent card
//	 from among them onto the battlefield with an indestructible counter on
//	 it. You may put a permanent card from among them into your hand. Put
//	 the rest into your graveyard."
//
// The six cards are revealed, then the caster picks (or declines) one
// permanent card to enter the battlefield with an indestructible counter
// (CR 122.1b, seeded onto the entry so it is there as the permanent enters).
// The rest of the sentence is that put's registered tail, told which cards
// remain in the library: a second optional pick of one permanent card among
// those, into the hand, and the cards still left go to the graveyard (the
// take's own rest). The card put onto the battlefield cannot also go to the
// hand, and a nonpermanent card can only go to the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8add484-8ed1-4a1c-aff9-dec65d60e4b1",
		Name:         "Wakanda Forever!",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 6,
				"Wakanda Forever! — revealed from the top of the library")
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    revealed,
				Max:      1,
				Optional: true,
				Counters: PutCounters{Kind: game.CounterIndestructible, N: 1},
				Label:    "Wakanda Forever! — put a permanent card onto the battlefield with an indestructible counter",
				ThenTail: wakandaForeverHand,
			}.Apply(ctx)
		},
	})
}

// wakandaForeverHand is the second sentence: you may put a permanent card
// from among the rest into your hand; the cards that remain go to the
// graveyard. An on-disk key.
var wakandaForeverHand = PutFromLibraryTail("wakanda-forever/hand-then-graveyard", func(ctx *Context, res PutFromLibraryResult) error {
	return TakeFromLibraryToHand{
		Player:   res.Player,
		Cards:    res.Rest,
		Match:    Permanent(),
		Max:      1,
		Optional: true,
		Label:    "Wakanda Forever! — put a permanent card from among them into your hand",
		Key:      takeRestGraveyard,
	}.Apply(ctx)
})
