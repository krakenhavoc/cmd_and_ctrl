package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saheeli's Directive — Sorcery {X}{R}{R}{R} (Commander 2018, #5526):
//
//	"Improvise (Each artifact you tap after you're done activating mana
//	 abilities pays for {1}.)
//	 Reveal the top X cards of your library. You may put any number of
//	 artifact cards with mana value X or less from among them onto the
//	 battlefield. Then put all cards revealed this way that weren't put
//	 onto the battlefield into your graveyard."
//
// Genesis Wave narrowed to artifact cards, with improvise as the spell's
// tap cost (Arc Reactor's): the X cards are revealed, any number of the
// artifact cards with mana value X or less enter as one simultaneous entry,
// and the rest go to the graveyard as a plain move, not a mill. "Any
// number" includes none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1aee3a93-d064-4d10-9fe6-879a1f342c1f",
		Name:         "Saheeli's Directive",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Improvise(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), x,
				"Saheeli's Directive — revealed from the top of the library")
			return effects.PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    revealed,
				Match:    effects.And(effects.Artifact(), effects.ManaValueLE(x)),
				Optional: true,
				Label:    "Saheeli's Directive — put any number of artifact cards with mana value X or less onto the battlefield",
				ThenTail: effects.PutRestIntoGraveyardTail,
			}.Apply(ctx)
		},
	})
}
