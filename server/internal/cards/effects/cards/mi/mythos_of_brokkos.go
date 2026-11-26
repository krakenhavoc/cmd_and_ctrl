package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mythos of Brokkos — Sorcery {2}{G}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"If {U}{B} was spent to cast this spell, search your library for a card,
//	 put that card into your graveyard, then shuffle.
//	 Return up to two permanent cards from your graveyard to your hand."
//
// River's Grasp's mana-spent reading over the spell's own payment (a waived
// payment asks what was paid; a copy has no mana spent). The search is
// Buried Alive's, one card with no quality, shuffled; the return follows it
// as a registered tail, so the card just put into the graveyard is a
// candidate. The two cards are chosen as the clause resolves, not targeted
// (the text names no target), from the permanent cards the graveyard then
// holds; "up to two" lets the player take none, one or two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "14596108-a8ed-4e68-b787-d655b28bdab4",
		Name:           "Mythos of Brokkos",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{SelfMillTutor: 1},
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"U", "B"},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.ManaSpentOfColor("U") == 0 || ctx.ManaSpentOfColor("B") == 0 {
				return effects.S82OldB117ReturnUpToTwoPermanents(ctx.Game, ctx.Controller(), ctx.Source())
			}
			return effects.SearchLibrary{
				Player:   item.Controller,
				Dest:     game.ZoneGraveyard,
				Limit:    1,
				Shuffle:  true,
				Source:   ctx.Source(),
				Reason:   "Mythos of Brokkos — put a card into your graveyard",
				ThenTail: effects.S82OldB117ReturnPermanentsTail(item.Controller, ctx.Source()),
			}.Apply(ctx)
		},
	})
}
